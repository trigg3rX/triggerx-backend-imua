package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/types"
)

// Test setup helper
func setupTestLeaderboardHandler() (*Handler, *MockKeeperRepository, *MockUserRepository) {
	mockKeeperRepo := new(MockKeeperRepository)
	mockUserRepo := new(MockUserRepository)

	handler := &Handler{
		keeperRepository: mockKeeperRepo,
		userRepository:   mockUserRepo,
		logger:           &MockLogger{},
	}

	return handler, mockKeeperRepo, mockUserRepo
}

func TestGetKeeperLeaderboard(t *testing.T) {
	handler, mockKeeperRepo, _ := setupTestLeaderboardHandler()

	tests := []struct {
		name          string
		host          string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Default Domain (All Keepers)",
			host: "localhost:9002",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperLeaderboard").Return([]types.KeeperLeaderboardEntry{
					{KeeperID: 1, KeeperAddress: "0x123", KeeperName: "Keeper 1", NoExecutedTasks: 5, KeeperPoints: 100},
					{KeeperID: 2, KeeperAddress: "0x456", KeeperName: "Keeper 2", NoExecutedTasks: 3, KeeperPoints: 50},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Success - App Domain (on_imua = false)",
			host: "app.triggerx.network",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperLeaderboardByOnImua", false).Return([]types.KeeperLeaderboardEntry{
					{KeeperID: 1, KeeperAddress: "0x123", KeeperName: "Keeper 1", NoExecutedTasks: 5, KeeperPoints: 100},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Success - Imua Domain (on_imua = true)",
			host: "imua.triggerx.network",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperLeaderboardByOnImua", true).Return([]types.KeeperLeaderboardEntry{
					{KeeperID: 2, KeeperAddress: "0x456", KeeperName: "Keeper 2", NoExecutedTasks: 3, KeeperPoints: 50},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Error - Database Error (Default Domain)",
			host: "localhost:9002",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperLeaderboard").Return([]types.KeeperLeaderboardEntry{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "LEADERBOARD_FETCH_ERROR",
		},
		{
			name: "Error - Database Error (App Domain)",
			host: "app.triggerx.network",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperLeaderboardByOnImua", false).Return([]types.KeeperLeaderboardEntry{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "LEADERBOARD_FETCH_ERROR",
		},
		{
			name: "Error - Database Error (Imua Domain)",
			host: "imua.triggerx.network",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperLeaderboardByOnImua", true).Return([]types.KeeperLeaderboardEntry{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "LEADERBOARD_FETCH_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockKeeperRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request with specific host
			req := httptest.NewRequest("GET", "/leaderboard/keepers", nil)
			req.Host = tt.host
			c.Request = req

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetKeeperLeaderboard(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["code"])
			} else {
				var response []types.KeeperLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response)
			}
		})
	}
}

func TestGetUserLeaderboard(t *testing.T) {
	handler, _, mockUserRepo := setupTestLeaderboardHandler()

	tests := []struct {
		name          string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Get User Leaderboard",
			setupMocks: func() {
				mockUserRepo.On("GetUserLeaderboard").Return([]types.UserLeaderboardEntry{
					{UserID: 1, UserAddress: "0x123", TotalJobs: 5, TotalTasks: 10, UserPoints: 100},
					{UserID: 2, UserAddress: "0x456", TotalJobs: 3, TotalTasks: 6, UserPoints: 50},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Error - Database Error",
			setupMocks: func() {
				mockUserRepo.On("GetUserLeaderboard").Return([]types.UserLeaderboardEntry{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockUserRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetUserLeaderboard(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response []types.UserLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)
			}
		})
	}
}

func TestGetKeeperByIdentifier(t *testing.T) {
	handler, mockKeeperRepo, _ := setupTestLeaderboardHandler()

	tests := []struct {
		name          string
		keeperAddress string
		keeperName    string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:          "Success - Get Keeper by Address",
			keeperAddress: "0x123",
			keeperName:    "",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperLeaderboardByIdentifierInDB", "0x123", "").Return(types.KeeperLeaderboardEntry{
					KeeperID: 1, KeeperAddress: "0x123", KeeperName: "Keeper 1", NoExecutedTasks: 5, KeeperPoints: 100,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Missing Identifier",
			keeperAddress: "",
			keeperName:    "",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Either keeper_address or keeper_name must be provided",
		},
		{
			name:          "Error - Keeper Not Found",
			keeperAddress: "0x123",
			keeperName:    "",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperLeaderboardByIdentifierInDB", "0x123", "").Return(types.KeeperLeaderboardEntry{}, assert.AnError)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "Keeper not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockKeeperRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Request = httptest.NewRequest("GET", "/", nil)
			q := c.Request.URL.Query()
			if tt.keeperAddress != "" {
				q.Add("keeper_address", tt.keeperAddress)
			}
			if tt.keeperName != "" {
				q.Add("keeper_name", tt.keeperName)
			}
			c.Request.URL.RawQuery = q.Encode()

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetKeeperByIdentifier(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response types.KeeperLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "0x123", response.KeeperAddress)
			}
		})
	}
}

func TestGetUserLeaderboardByAddress(t *testing.T) {
	handler, _, mockUserRepo := setupTestLeaderboardHandler()

	tests := []struct {
		name          string
		userAddress   string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:        "Success - Get User by Address",
			userAddress: "0x123",
			setupMocks: func() {
				mockUserRepo.On("GetUserLeaderboardByAddress", "0x123").Return(types.UserLeaderboardEntry{
					UserID: 1, UserAddress: "0x123", TotalJobs: 5, TotalTasks: 10, UserPoints: 100,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Missing Address",
			userAddress:   "",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "user_address must be provided",
		},
		{
			name:        "Error - User Not Found",
			userAddress: "0x123",
			setupMocks: func() {
				mockUserRepo.On("GetUserLeaderboardByAddress", "0x123").Return(types.UserLeaderboardEntry{}, assert.AnError)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockUserRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Request = httptest.NewRequest("GET", "/", nil)
			q := c.Request.URL.Query()
			if tt.userAddress != "" {
				q.Add("user_address", tt.userAddress)
			}
			c.Request.URL.RawQuery = q.Encode()

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetUserLeaderboardByAddress(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response types.UserLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "0x123", response.UserAddress)
			}
		})
	}
}
