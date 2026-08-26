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

	logger.Info(
		"starting auth service",
		"service", cfg.App.Name,
		"env", cfg.App.Env,
		"port", cfg.App.GRPCPort,
	)
}
