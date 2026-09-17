package app

import (
	"log/slog"

	"github.com/vivek-344/diagon/services/auth/internal/config"
)

type App struct {
	cfg    *config.Config
	logger *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}
