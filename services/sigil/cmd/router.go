package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vivek-344/diagon/sigil/internal/handler"
)

func setupRouter(
	authMiddleware func(http.Handler) http.Handler,
	authHandler *handler.AuthHandler,
	developerHandler *handler.DeveloperHandler,
	dbPool *pgxpool.Pool,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := dbPool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "unhealthy",
				"db":     "disconnected",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	// API routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", developerHandler.Create)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/profile", authHandler.GetProfile)
		})
	})
	r.Route("/developers", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", developerHandler.GetAll)
		r.Get("/{id}", developerHandler.GetByID)
		r.Put("/{id}", developerHandler.Update)
		r.Delete("/{id}", developerHandler.Delete)
		r.Put("/{id}/password", developerHandler.UpdatePassword)
		r.Post("/{id}/suspend", developerHandler.Suspend)
	})

	return r
}
