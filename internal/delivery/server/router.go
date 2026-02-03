package delivery

import (
	"gopthkeeper/internal/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates a new router with the given API handlers and JWT manager.
func NewRouter(api *API, jwtManager *auth.JWTManager) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", api.Register)
		r.Post("/login", api.Login)
	})

	r.Route("/api/secrets", func(r chi.Router) {
		r.Use(jwtManager.AuthMiddleware)

		r.Post("/", api.CreateSecret)
		r.Get("/", api.GetSecrets)
		r.Get("/{id}", api.GetSecretByID)
		r.Put("/{id}", api.UpdateSecret)
		r.Delete("/{id}", api.DeleteSecret)
	})

	return r
}
