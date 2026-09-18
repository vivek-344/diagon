package server

import (
	"fmt"
	"net/http"

	"github.com/vivek-344/diagon/services/gateway/internal/apiresponse"
	"github.com/vivek-344/diagon/services/gateway/internal/handler"
	"github.com/vivek-344/diagon/services/gateway/internal/middleware"
)

func NewHTTPServer(
	port int,
	authHandler *handler.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
) *http.Server {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc(
		"POST /api/v1/auth/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/refresh",
		authHandler.Refresh,
	)

	// Protected routes
	mux.Handle(
		"GET /api/v1/auth/me",
		authMiddleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := middleware.UserID(r.Context())
				if !ok {
					apiresponse.WriteError(
						w,
						http.StatusInternalServerError,
						apiresponse.CodeInternalError,
						"An unexpected error occurred.",
						nil,
					)
					return
				}

				apiresponse.Write(
					w,
					http.StatusOK,
					map[string]string{
						"user_id": userID,
					},
				)
			}),
		),
	)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}
