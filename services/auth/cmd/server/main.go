package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vivek-344/diagon/services/auth/internal/app"
	"github.com/vivek-344/diagon/services/auth/internal/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := app.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger = logger.With(
		"service", cfg.App.Name,
		"version", cfg.App.Version,
		"env", cfg.App.Env,
	)

	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	a := app.New(cfg, logger, db)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := a.Run(ctx); err != nil {
		logger.Error(
			"auth service stopped",
			"error", err,
		)
		os.Exit(1)
	}
}
