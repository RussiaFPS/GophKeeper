// Package auth handles JWT-based authentication.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token generation and validation.
type JWTManager struct {
	jwtKey []byte
}

// NewJWTManager creates a new JWTManager with the given secret key.
func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{jwtKey: []byte(secret)}
}

// Claims contains the JWT claims.
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

// UserIDContextKey is the key for the user ID in the context.
const UserIDContextKey ContextKey = "userID"

// GenerateJWT creates a new JWT token for a given user ID.
func (j *JWTManager) GenerateJWT(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.jwtKey)

	return tokenString, err
}

// ValidateJWT validates a JWT token and returns the user ID from the claims if valid.
func (j *JWTManager) ValidateJWT(tokenString string) (int, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.jwtKey, nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

// AuthMiddleware is a middleware that validates the JWT token and sets the UserID in the context.
func (j *JWTManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		userID, err := j.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext retrieves the UserID from the request context.
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(int)
	return userID, ok
}
