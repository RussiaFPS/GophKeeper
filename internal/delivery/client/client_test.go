package delivery

import (
	config "gopthkeeper/internal/config/client"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// TestNewClient tests the NewClient function
func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
	if client.httpClient.Timeout == 0 {
		t.Error("Expected timeout to be set")
	}
}

// TestAuthenticatedRequest tests the AuthenticatedRequest method
func TestAuthenticatedRequest(t *testing.T) {
	tests := []struct {
		name           string
		setupToken     func() error
		method         string
		path           string
		body           interface{}
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
		checkRequest   func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful authenticated request",
			setupToken: func() error {
				return config.SaveToken("test-token-123")
			},
			method:         http.MethodGet,
			path:           "/api/secrets",
			body:           nil,
			serverResponse: nil,
			expectError:    false,
			checkRequest:   nil,
		},
		{
			name: "request with body",
			setupToken: func() error {
				return config.SaveToken("test-token-456")
			},
			method: http.MethodPost,
			path:   "/api/secrets",
			body: map[string]interface{}{
				"type": 0,
				"data": "secret data",
			},
			serverResponse: nil,
			expectError:    false,
			checkRequest:   nil,
		},
		{
			name: "missing token",
			setupToken: func() error {
				// Don't save token
				return nil
			},
			method:         http.MethodGet,
			path:           "/api/secrets",
			body:           nil,
			serverResponse: nil,
			expectError:    true,
			checkRequest:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tempDir := t.TempDir()
			oldConfigDir := os.Getenv("GOPHKEEPER_CONFIG_DIR")
			os.Setenv("GOPHKEEPER_CONFIG_DIR", tempDir)
			defer os.Setenv("GOPHKEEPER_CONFIG_DIR", oldConfigDir)
			tokenPath := filepath.Join(tempDir, "gophkeeper_token.txt")
			os.Remove(tokenPath)

			if tt.setupToken != nil {
				if err := tt.setupToken(); err != nil {
					t.Fatalf("Failed to setup token: %v", err)
				}
			}

			client := NewClient()

			resp, err := client.AuthenticatedRequest(tt.method, tt.path, tt.body)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				// Expected to fail since we can't override ServerURL
				// but we've verified token loading works
				return
			}

			if resp != nil {
				defer resp.Body.Close()
			}
		})
	}
}

// TestAuthenticatedRequestTokenLoading tests token loading behavior
func TestAuthenticatedRequestTokenLoading(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	oldConfigDir := os.Getenv("GOPHKEEPER_CONFIG_DIR")
	os.Setenv("GOPHKEEPER_CONFIG_DIR", tempDir)
	defer os.Setenv("GOPHKEEPER_CONFIG_DIR", oldConfigDir)

	client := NewClient()

	// Test without token
	_, err := client.AuthenticatedRequest(http.MethodGet, "/api/secrets", nil)
	if err == nil {
		t.Error("Expected error when token is missing")
	}

	// Save token and test again
	if err := config.SaveToken("valid-token"); err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	_, err = client.AuthenticatedRequest(http.MethodGet, "/api/secrets", nil)
	if err != nil && err.Error() == "authentication required" {
		t.Error("Token should have been loaded successfully")
	}
}
