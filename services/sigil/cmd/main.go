package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vivek-344/diagon/sigil/config"
	"github.com/vivek-344/diagon/sigil/internal/handler"
	"github.com/vivek-344/diagon/sigil/internal/repository"
	"github.com/vivek-344/diagon/sigil/internal/service"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Run the application
	if err := run(cfg); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database Connection
	dbPool, err := initDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	// Initialize Repositories, Services, and Handlers
	developerRepo := repository.NewDeveloperRepository(dbPool)
	developerSvc := service.NewDeveloperService(developerRepo)
	developerHandler := handler.NewDeveloperHandler(developerSvc)

	// HTTP Router
	router := setupRouter(developerHandler, dbPool)

	// HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return startServerWithGracefulShutdown(ctx, server)
}

func initDB(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	slog.Info("database connected successfully")
	return pool, nil
}

func setupRouter(developerHandler *handler.DeveloperHandler, dbPool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if err := dbPool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "unhealthy",
				"db":     "disconnected",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// API routes
	r.Route("/", func(r chi.Router) {
		r.Route("developers", func(r chi.Router) {
			r.Post("/", developerHandler.Create)
		})
	})

	return r
}

func startServerWithGracefulShutdown(ctx context.Context, server *http.Server) error {
	// Channel to receive shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Channel to receive server errors
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		slog.Info("server starting", "addr", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	// Block until signal or error
	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-shutdown:
		slog.Info("shutdown signal received", "signal", sig)

		shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			// Force close if graceful shutdown fails
			server.Close()
			return err
		}
	}

	slog.Info("server stopped gracefully")
	return nil
}
