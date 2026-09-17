package main

import (
	"log/slog"
	"os"

	"github.com/vivek-344/diagon/services/auth/internal/app"
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

	a := app.New(cfg, logger)

	a.Run()
}
