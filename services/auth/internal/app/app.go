package app

import (
	"database/sql"
	"log/slog"

	"github.com/vivek-344/diagon/services/auth/internal/config"
	"github.com/vivek-344/diagon/services/auth/internal/repository"
)

type App struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sql.DB

	userRepository *repository.UserRepository
}

func New(
	cfg *config.Config,
	logger *slog.Logger,
	db *sql.DB,
) *App {
	return &App{
		cfg:            cfg,
		logger:         logger,
		db:             db,
		userRepository: repository.NewUserRepository(db),
	}
}
