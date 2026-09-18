package server

import (
	"fmt"
	"net/http"

	"github.com/vivek-344/diagon/services/gateway/internal/handler"
)

func NewHTTPServer(
	port int,
	authHandler *handler.AuthHandler,
) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"POST /api/v1/auth/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}
