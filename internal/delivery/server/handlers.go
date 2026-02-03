package delivery

import (
	"encoding/json"
	"errors"
	"gopthkeeper/internal/auth"
	"gopthkeeper/internal/models"
	storage "gopthkeeper/internal/repo"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// API holds the dependencies for the API handlers.
type API struct {
	store      storage.Store
	jwtManager *auth.JWTManager
}

// New creates a new API structure.
func New(store storage.Store, jwtManager *auth.JWTManager) *API {
	return &API{store: store, jwtManager: jwtManager}
}

func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	createdUser, err := a.store.CreateUser(ctx, user)
	if err != nil {
		var userExistsErr storage.ErrUserExists
		if errors.As(err, &userExistsErr) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := a.store.GetUserByLogin(ctx, creds.Login)
	if err != nil {
		var userNotFoundErr storage.ErrUserNotFound
		if errors.As(err, &userNotFoundErr) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError) // Generic error for other storage issues
		return
	}

	if !auth.CheckPasswordHash(creds.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := a.jwtManager.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (a *API) CreateSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	var secret models.Secret
	if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	secret.UserID = userID // Ensure secret is for the authenticated user

	createdSecret, err := a.store.CreateSecret(ctx, secret)
	if err != nil {
		http.Error(w, "Failed to create secret", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSecret)
}

func (a *API) GetSecrets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	secrets, err := a.store.GetSecrets(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve secrets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secrets)
}

func (a *API) GetSecretByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	secretIDStr := chi.URLParam(r, "id")
	if secretIDStr == "" {
		http.Error(w, "Missing secret ID", http.StatusBadRequest)
		return
	}

	secretID, err := strconv.Atoi(secretIDStr)
	if err != nil {
		http.Error(w, "Invalid secret ID", http.StatusBadRequest)
		return
	}

	secret, err := a.store.GetSecretByID(ctx, userID, secretID)
	if err != nil {
		var secretNotFoundErr storage.ErrSecretNotFound
		if errors.As(err, &secretNotFoundErr) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve secret", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

func (a *API) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	secretIDStr := chi.URLParam(r, "id")
	if secretIDStr == "" {
		http.Error(w, "Missing secret ID", http.StatusBadRequest)
		return
	}

	secretID, err := strconv.Atoi(secretIDStr)
	if err != nil {
		http.Error(w, "Invalid secret ID", http.StatusBadRequest)
		return
	}

	var secret models.Secret
	if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	secret.ID = secretID
	secret.UserID = userID

	updatedSecret, err := a.store.UpdateSecret(ctx, secret)
	if err != nil {
		var secretNotFoundErr storage.ErrSecretNotFound
		if errors.As(err, &secretNotFoundErr) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update secret", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSecret)
}

func (a *API) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	secretIDStr := chi.URLParam(r, "id")
	if secretIDStr == "" {
		http.Error(w, "Missing secret ID", http.StatusBadRequest)
		return
	}

	secretID, err := strconv.Atoi(secretIDStr)
	if err != nil {
		http.Error(w, "Invalid secret ID", http.StatusBadRequest)
		return
	}

	err = a.store.DeleteSecret(ctx, userID, secretID)
	if err != nil {
		var secretNotFoundErr storage.ErrSecretNotFound
		if errors.As(err, &secretNotFoundErr) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete secret", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion
}
