package app

import (
	"database/sql"
	"log/slog"

	"github.com/vivek-344/diagon/services/auth/internal/config"
)

type App struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sql.DB
}

func New(
	cfg *config.Config,
	logger *slog.Logger,
	db *sql.DB,
) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
		db:     db,
	}
}
