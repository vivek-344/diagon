package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vivek-344/diagon/services/gateway/internal/client"
	"github.com/vivek-344/diagon/services/gateway/internal/config"
	"github.com/vivek-344/diagon/services/gateway/internal/handler"
	"github.com/vivek-344/diagon/services/gateway/internal/middleware"
	"github.com/vivek-344/diagon/services/gateway/internal/server"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)

	cfg, err := config.Load()
	if err != nil {
		logger.Error(
			"failed to load config",
			"error", err,
		)
		os.Exit(1)
	}

	logger = logger.With(
		"service", cfg.App.Name,
		"version", cfg.App.Version,
		"env", cfg.App.Env,
	)

	authClient, err := client.NewAuthClient(
		cfg.Auth.GRPCAddress,
	)
	if err != nil {
		logger.Error(
			"failed to create auth client",
			"error", err,
		)
		os.Exit(1)
	}
	defer authClient.Close()

	authMiddleware := middleware.NewAuthMiddleware(
		authClient,
	)

	authHandler := handler.NewAuthHandler(
		authClient,
	)

	httpServer := server.NewHTTPServer(
		cfg.App.HTTPPort,
		authHandler,
		authMiddleware,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		logger.Info(
			"gateway http server listening",
			"address", httpServer.Addr,
		)

		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"gateway http server stopped",
				"error", err,
			)
			os.Exit(1)
		}

	case <-ctx.Done():
		logger.Info(
			"shutting down gateway",
		)

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			cfg.App.ShutdownTimeout,
		)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error(
				"failed to shutdown gateway http server",
				"error", err,
			)
			os.Exit(1)
		}

		logger.Info(
			"gateway shutdown complete",
		)
	}
}
