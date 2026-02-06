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
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	user.Password = hashedPassword

	createdUser, err := a.store.CreateUser(ctx, user)
	if err != nil {
		var userExistsErr storage.ErrUserExists
		if errors.As(err, &userExistsErr) {
			RespondWithError(w, http.StatusConflict, err.Error())
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	RespondWithJSON(w, http.StatusCreated, createdUser)
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := a.store.GetUserByLogin(ctx, creds.Login)
	if err != nil {
		var userNotFoundErr storage.ErrUserNotFound
		if errors.As(err, &userNotFoundErr) {
			RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Server error")
		return
	}

	if !auth.CheckPasswordHash(creds.Password, user.Password) {
		RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := a.jwtManager.GenerateJWT(user.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (a *API) CreateSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "User ID not found in context")
		return
	}

	var secret models.Secret
	if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	secret.UserID = userID // Ensure secret is for the authenticated user

	createdSecret, err := a.store.CreateSecret(ctx, secret)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create secret")
		return
	}

	RespondWithJSON(w, http.StatusCreated, createdSecret)
}

func (a *API) GetSecrets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "User ID not found in context")
		return
	}

	secrets, err := a.store.GetSecrets(ctx, userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve secrets")
		return
	}

	RespondWithJSON(w, http.StatusOK, secrets)
}

func (a *API) GetSecretByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "User ID not found in context")
		return
	}

	secretIDStr := chi.URLParam(r, "id")
	if secretIDStr == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing secret ID")
		return
	}

	secretID, err := strconv.Atoi(secretIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid secret ID")
		return
	}

	secret, err := a.store.GetSecretByID(ctx, userID, secretID)
	if err != nil {
		var secretNotFoundErr storage.ErrSecretNotFound
		if errors.As(err, &secretNotFoundErr) {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve secret")
		return
	}

	RespondWithJSON(w, http.StatusOK, secret)
}

func (a *API) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "User ID not found in context")
		return
	}

	secretIDStr := chi.URLParam(r, "id")
	if secretIDStr == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing secret ID")
		return
	}

	secretID, err := strconv.Atoi(secretIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid secret ID")
		return
	}

	var secret models.Secret
	if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	secret.ID = secretID
	secret.UserID = userID

	updatedSecret, err := a.store.UpdateSecret(ctx, secret)
	if err != nil {
		var secretNotFoundErr storage.ErrSecretNotFound
		if errors.As(err, &secretNotFoundErr) {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to update secret")
		return
	}

	RespondWithJSON(w, http.StatusOK, updatedSecret)
}

func (a *API) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "User ID not found in context")
		return
	}

	secretIDStr := chi.URLParam(r, "id")
	if secretIDStr == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing secret ID")
		return
	}

	secretID, err := strconv.Atoi(secretIDStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid secret ID")
		return
	}

	err = a.store.DeleteSecret(ctx, userID, secretID)
	if err != nil {
		var secretNotFoundErr storage.ErrSecretNotFound
		if errors.As(err, &secretNotFoundErr) {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete secret")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion
}
