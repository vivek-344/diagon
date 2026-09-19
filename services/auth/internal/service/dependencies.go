package service

import (
	"context"

	"github.com/vivek-344/diagon/services/auth/internal/domain"
	"github.com/vivek-344/diagon/services/auth/internal/security"
)

type UserRepository interface {
	Create(
		ctx context.Context,
		email string,
		passwordHash string,
	) (*domain.User, error)

	FindByEmail(
		ctx context.Context,
		email string,
	) (*domain.User, error)

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.User, error)
}

type TokenManager interface {
	GenerateTokenPair(userID string) (*security.TokenPair, error)

	ValidateAccessToken(accessToken string) (string, error)

	ValidateRefreshToken(refreshToken string) (string, error)
}
