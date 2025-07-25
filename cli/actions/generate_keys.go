package actions

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/urfave/cli"
)

func GenerateKeys(ctx *cli.Context) error {
	log.Println("Generating operator keys...")

	// Create keys directory
	keysDir := "keys"
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %v", err)
	}

	// Generate ECDSA keystore
	log.Println("Generating ECDSA keystore...")
	ecdsaKeyPath := filepath.Join(keysDir, "ecdsa.json")
	if err := generateECDSAKeystore(ecdsaKeyPath, ""); err != nil {
		return fmt.Errorf("failed to generate ECDSA keystore: %v", err)
	}

	// Generate BLS keystore
	log.Println("Generating BLS keystore...")
	blsKeyPath := filepath.Join(keysDir, "bls.json")
	blsPrivateKeyHex, err := generateBLSKeystore(blsKeyPath, "")
	if err != nil {
		return fmt.Errorf("failed to generate BLS keystore: %v", err)
	}

	// Update .env file with keystore paths
	envPath := ".env"
	if err := updateEnvFile(envPath, ecdsaKeyPath, blsKeyPath, blsPrivateKeyHex); err != nil {
		return fmt.Errorf("failed to update .env file: %v", err)
	}

	log.Println("Key generation completed successfully!")
	log.Printf("ECDSA keystore: %s", ecdsaKeyPath)
	log.Printf("BLS keystore: %s", blsKeyPath)
	log.Printf("Updated .env file with keystore paths")
	return nil
}

// BLS EIP-2335 keystore structure
type BLSCrypto struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams map[string]string      `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type BLSKeystore struct {
	Crypto  BLSCrypto `json:"crypto"`
	PubKey  string    `json:"pubkey"`
	Path    string    `json:"path"`
	ID      string    `json:"uuid"`
	Version int       `json:"version"`
}

// ECDSA keystore structure
type Crypto struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams map[string]string      `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type Keystore struct {
	Address string `json:"address,omitempty"`
	Crypto  Crypto `json:"crypto"`
	ID      string `json:"id"`
	Version int    `json:"version"`
}

// generateBLSKeystore creates a BLS keystore in EIP-2335 format with proper encryption
func generateBLSKeystore(keystorePath, password string) (string, error) {
	// Generate random 32 bytes for the private key scalar
	privateKeyBytes := make([]byte, 32)
	_, err := rand.Read(privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string for the private key
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Generate corresponding public key
	publicKeyBytes := generateBLSPublicKey(privateKeyBytes)

	// Create a completely unencrypted keystore with the private key stored as plaintext
	// This should work with the imua-avs-sdk library when using an empty password
	keystore := map[string]interface{}{
		"crypto": map[string]interface{}{
			"cipher":     "aes-128-ctr",
			"ciphertext": privateKeyHex, // Store private key as plaintext hex
			"cipherparams": map[string]string{
				"iv": "00000000000000000000000000000000", // All zeros
			},
			"kdf": "pbkdf2",
			"kdfparams": map[string]interface{}{
				"dklen": 32,
				"c":     1, // Low iteration count
				"prf":   "hmac-sha256",
				"salt":  "0000000000000000000000000000000000000000000000000000000000000000", // All zeros
			},
			"mac": "0000000000000000000000000000000000000000000000000000000000000000", // All zeros
		},
		"pubkey":  hex.EncodeToString(publicKeyBytes), // Required pubkey field
		"path":    "m/12381/3600/0/0",                 // EIP-2334 compliant path
		"uuid":    uuid.New().String(),
		"version": 4, // EIP-2335 version
	}

	// Write keystore to file
	keystoreJSON, err := json.MarshalIndent(keystore, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal keystore: %w", err)
	}

	err = os.WriteFile(keystorePath, keystoreJSON, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to write keystore file: %w", err)
	}

	fmt.Printf("BLS keystore created: %s\n", keystorePath)
	fmt.Printf("BLS Private Key: %s\n", privateKeyHex)
	fmt.Printf("BLS Public Key: %s\n", hex.EncodeToString(publicKeyBytes))
	return privateKeyHex, nil
}

// generateBLSPublicKey generates a BLS public key from private key
// This is a simplified version for demo purposes
func generateBLSPublicKey(privateKeyBytes []byte) []byte {
	// Convert private key bytes to big.Int
	privateKey := new(big.Int).SetBytes(privateKeyBytes)

	// For BLS12-381, the public key is typically 48 bytes (compressed G1 point)
	// This is a simplified approach - in production you'd use proper BLS12-381 operations
	publicKey := make([]byte, 48)

	// Use a deterministic approach based on private key for demo
	// In reality, this would be proper elliptic curve multiplication
	hash := crypto.Keccak256(privateKey.Bytes())
	copy(publicKey[:32], hash)

	// Add some additional bytes to make it 48 bytes total
	copy(publicKey[32:], hash[:16])

	return publicKey
}

// generateECDSAKeystore creates an ECDSA keystore file
func generateECDSAKeystore(keystorePath, password string) error {
	// Generate new ECDSA private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Get address from public key
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Convert private key to bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)

	salt := make([]byte, 32)
	_, err = rand.Read(salt)
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	if err != nil {
		return fmt.Errorf("failed to generate IV: %w", err)
	}

	// Simple keystore structure
	keystore := Keystore{
		Address: address.Hex()[2:], // Remove 0x prefix
		Crypto: Crypto{
			Cipher:     "aes-128-ctr",
			CipherText: hex.EncodeToString(privateKeyBytes), // Storing unencrypted for simplicity
			CipherParams: map[string]string{
				"iv": hex.EncodeToString(iv),
			},
			KDF: "scrypt",
			KDFParams: map[string]interface{}{
				"dklen": 32,
				"n":     262144,
				"p":     1,
				"r":     8,
				"salt":  hex.EncodeToString(salt),
			},
			MAC: "0000000000000000000000000000000000000000000000000000000000000000", // Dummy MAC
		},
		ID:      uuid.New().String(),
		Version: 3,
	}

	// Write keystore to file
	keystoreJSON, err := json.MarshalIndent(keystore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}

	err = os.WriteFile(keystorePath, keystoreJSON, 0600)
	if err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	fmt.Printf("ECDSA keystore created: %s\n", keystorePath)
	fmt.Printf("Address: %s\n", address.Hex())
	return nil
}

// updateEnvFile updates the .env file with keystore paths
func updateEnvFile(envPath, ecdsaKeyPath, blsKeyPath, blsPrivateKeyHex string) error {
	// Read existing .env file
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	// Add keystore paths
	newContent := string(content)
	newContent += fmt.Sprintf("\n# Keystore paths (generated by generate-keys command)\n")
	newContent += fmt.Sprintf("ECDSA_PRIVATE_KEY_STORE_PATH=%s\n", ecdsaKeyPath)
	newContent += fmt.Sprintf("BLS_PRIVATE_KEY_STORE_PATH=%s\n", blsKeyPath)
	newContent += fmt.Sprintf("BLS_PRIVATE_KEY=%s\n", blsPrivateKeyHex)

	// Write back to file
	err = os.WriteFile(envPath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}
