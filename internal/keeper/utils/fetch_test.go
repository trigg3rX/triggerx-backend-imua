package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchDataFromUrl_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello world"))
		if err != nil {
			t.Errorf("error writing response: %v", err)
		}
	}))
	defer server.Close()

	data, err := FetchDataFromUrl(server.URL, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data != "hello world" {
		t.Errorf("expected 'hello world', got %q", data)
	}
}

func TestFetchDataFromUrl_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("not found"))
		if err != nil {
			t.Errorf("error writing response: %v", err)
		}
	}))
	defer server.Close()

	data, err := FetchDataFromUrl(server.URL, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(data, "not found") {
		t.Errorf("expected 'not found' in response, got %q", data)
	}
}

func TestFetchDataFromUrl_InvalidURL(t *testing.T) {
	_, err := FetchDataFromUrl(":badurl:", nil)
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestFetchDataFromUrl_EmptyURL(t *testing.T) {
	_, err := FetchDataFromUrl("", nil)
	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}
}
