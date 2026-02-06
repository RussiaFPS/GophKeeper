package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"gopthkeeper/internal/auth"
	"gopthkeeper/internal/models"
	storage "gopthkeeper/internal/repo"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestRegister tests the Register handler
func TestRegister(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			requestBody: models.User{
				Login:    "testuser",
				Password: "testpass",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var user models.User
				if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if user.ID == 0 {
					t.Error("Expected user ID to be set")
				}
				if user.Login != "testuser" {
					t.Errorf("Expected login 'testuser', got '%s'", user.Login)
				}
			},
		},
		{
			name: "duplicate user registration",
			requestBody: models.User{
				Login:    "duplicate",
				Password: "testpass",
			},
			expectedStatus: http.StatusConflict,
			checkResponse:  nil,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStore()
			jwtManager := auth.NewJWTManager("test-secret")
			api := New(store, jwtManager)

			if tt.name == "duplicate user registration" {
				hashedPass, _ := auth.HashPassword("testpass")
				store.CreateUser(context.Background(), models.User{Login: "duplicate", Password: hashedPass})
			}

			var body []byte
			var err error
			if user, ok := tt.requestBody.(models.User); ok {
				body, err = json.Marshal(user)
			} else {
				body = []byte(tt.requestBody.(string))
			}
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(body))
			resp := httptest.NewRecorder()

			api.Register(resp, req)

			if resp.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestLogin tests the Login handler
func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		setupStore     func(*storage.MemStore)
		requestBody    models.User
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			setupStore: func(store *storage.MemStore) {
				hashedPass, _ := auth.HashPassword("correctpass")
				store.CreateUser(context.Background(), models.User{Login: "testuser", Password: hashedPass})
			},
			requestBody: models.User{
				Login:    "testuser",
				Password: "correctpass",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]string
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if result["token"] == "" {
					t.Error("Expected token in response")
				}
			},
		},
		{
			name: "invalid password",
			setupStore: func(store *storage.MemStore) {
				hashedPass, _ := auth.HashPassword("correctpass")
				store.CreateUser(context.Background(), models.User{Login: "testuser", Password: hashedPass})
			},
			requestBody: models.User{
				Login:    "testuser",
				Password: "wrongpass",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
		{
			name:       "user not found",
			setupStore: func(store *storage.MemStore) {},
			requestBody: models.User{
				Login:    "nonexistent",
				Password: "anypass",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStore()
			jwtManager := auth.NewJWTManager("test-secret")
			api := New(store, jwtManager)

			if tt.setupStore != nil {
				tt.setupStore(store)
			}

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(body))
			resp := httptest.NewRecorder()

			api.Login(resp, req)

			if resp.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestCreateSecret tests the CreateSecret handler
func TestCreateSecret(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		requestBody    models.Secret
		expectedStatus int
		checkResponse  func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:   "successful secret creation",
			userID: 1,
			requestBody: models.Secret{
				Type:     models.LoginPasswordType,
				Data:     []byte("secret data"),
				Metadata: "test metadata",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var secret models.Secret
				if err := json.NewDecoder(resp.Body).Decode(&secret); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if secret.ID == 0 {
					t.Error("Expected secret ID to be set")
				}
				if secret.UserID != 1 {
					t.Errorf("Expected UserID 1, got %d", secret.UserID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStore()
			jwtManager := auth.NewJWTManager("test-secret")
			api := New(store, jwtManager)

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/secrets", bytes.NewBuffer(body))
			ctx := context.WithValue(req.Context(), auth.UserIDContextKey, tt.userID)
			req = req.WithContext(ctx)
			resp := httptest.NewRecorder()

			api.CreateSecret(resp, req)

			if resp.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestGetSecrets tests the GetSecrets handler
func TestGetSecrets(t *testing.T) {
	store := storage.NewMemStore()
	jwtManager := auth.NewJWTManager("test-secret")
	api := New(store, jwtManager)

	// Create test secrets
	secret1 := models.Secret{UserID: 1, Type: models.TextDataType, Data: []byte("data1")}
	secret2 := models.Secret{UserID: 1, Type: models.TextDataType, Data: []byte("data2")}
	secret3 := models.Secret{UserID: 2, Type: models.TextDataType, Data: []byte("data3")}

	store.CreateSecret(context.Background(), secret1)
	store.CreateSecret(context.Background(), secret2)
	store.CreateSecret(context.Background(), secret3)

	req := httptest.NewRequest(http.MethodGet, "/api/secrets", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, 1)
	req = req.WithContext(ctx)
	resp := httptest.NewRecorder()

	api.GetSecrets(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var secrets []models.Secret
	if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(secrets) != 2 {
		t.Errorf("Expected 2 secrets for user 1, got %d", len(secrets))
	}
}

// TestGetSecretByID tests the GetSecretByID handler
func TestGetSecretByID(t *testing.T) {
	tests := []struct {
		name           string
		setupStore     func(*storage.MemStore) int
		userID         int
		secretID       string
		expectedStatus int
	}{
		{
			name: "successful retrieval",
			setupStore: func(store *storage.MemStore) int {
				secret, _ := store.CreateSecret(context.Background(), models.Secret{
					UserID: 1,
					Type:   models.TextDataType,
					Data:   []byte("test data"),
				})
				return secret.ID
			},
			userID:         1,
			secretID:       "1",
			expectedStatus: http.StatusOK,
		},
		{
			name: "secret not found",
			setupStore: func(store *storage.MemStore) int {
				return 999
			},
			userID:         1,
			secretID:       "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "invalid secret ID",
			setupStore: func(store *storage.MemStore) int {
				return 0
			},
			userID:         1,
			secretID:       "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStore()
			jwtManager := auth.NewJWTManager("test-secret")
			api := New(store, jwtManager)

			if tt.setupStore != nil {
				tt.setupStore(store)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/secrets/"+tt.secretID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.secretID)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, auth.UserIDContextKey, tt.userID)
			req = req.WithContext(ctx)

			resp := httptest.NewRecorder()

			api.GetSecretByID(resp, req)

			if resp.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.Code)
			}
		})
	}
}

// TestUpdateSecret tests the UpdateSecret handler
func TestUpdateSecret(t *testing.T) {
	store := storage.NewMemStore()
	jwtManager := auth.NewJWTManager("test-secret")
	api := New(store, jwtManager)

	// Create initial secret
	secret, _ := store.CreateSecret(context.Background(), models.Secret{
		UserID:   1,
		Type:     models.TextDataType,
		Data:     []byte("original data"),
		Metadata: "original",
	})

	updatedSecret := models.Secret{
		ID:       secret.ID,
		Type:     models.TextDataType,
		Data:     []byte("updated data"),
		Metadata: "updated",
	}

	body, _ := json.Marshal(updatedSecret)
	req := httptest.NewRequest(http.MethodPut, "/api/secrets/1", bytes.NewBuffer(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, auth.UserIDContextKey, 1) // ✅ Используем конкретное значение
	req = req.WithContext(ctx)

	resp := httptest.NewRecorder()

	api.UpdateSecret(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var result models.Secret
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if string(result.Data) != "updated data" {
		t.Errorf("Expected updated data, got %s", string(result.Data))
	}
}

// TestDeleteSecret tests the DeleteSecret handler
func TestDeleteSecret(t *testing.T) {
	tests := []struct {
		name           string
		setupStore     func(*storage.MemStore) int
		userID         int
		secretID       string
		expectedStatus int
	}{
		{
			name: "successful deletion",
			setupStore: func(store *storage.MemStore) int {
				secret, _ := store.CreateSecret(context.Background(), models.Secret{
					UserID: 1,
					Type:   models.TextDataType,
					Data:   []byte("test data"),
				})
				return secret.ID
			},
			userID:         1,
			secretID:       "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "secret not found",
			setupStore: func(store *storage.MemStore) int {
				return 999
			},
			userID:         1,
			secretID:       "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStore()
			jwtManager := auth.NewJWTManager("test-secret")
			api := New(store, jwtManager)

			if tt.setupStore != nil {
				tt.setupStore(store)
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/secrets/"+tt.secretID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.secretID)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, auth.UserIDContextKey, tt.userID)
			req = req.WithContext(ctx)

			resp := httptest.NewRecorder()

			api.DeleteSecret(resp, req)

			if resp.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.Code)
			}
		})
	}
}
