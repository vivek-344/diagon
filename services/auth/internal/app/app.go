package app

import (
	"database/sql"
	"log/slog"

	"github.com/vivek-344/diagon/services/auth/internal/config"
	"github.com/vivek-344/diagon/services/auth/internal/handler"
	"github.com/vivek-344/diagon/services/auth/internal/repository"
	"github.com/vivek-344/diagon/services/auth/internal/security"
	"github.com/vivek-344/diagon/services/auth/internal/service"
)

type App struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sql.DB

	userRepository *repository.UserRepository
	authService    *service.AuthService
	authHandler    *handler.AuthHandler
}

func New(
	cfg *config.Config,
	logger *slog.Logger,
	db *sql.DB,
) *App {
	userRepository := repository.NewUserRepository(db)

	tokenManager := security.NewTokenManager(
		cfg.JWT.SigningKey,
		cfg.JWT.AccessDuration,
		cfg.JWT.RefreshDuration,
	)

	authService := service.NewAuthService(
		userRepository,
		tokenManager,
	)

	authHandler := handler.NewAuthHandler(
		authService,
	)

	return &App{
		cfg:    cfg,
		logger: logger,
		db:     db,

		userRepository: userRepository,
		authService:    authService,
		authHandler:    authHandler,
	}
}
