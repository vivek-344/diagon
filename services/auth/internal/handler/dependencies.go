package handler

import (
	"context"

	"github.com/vivek-344/diagon/services/auth/internal/domain"
	"github.com/vivek-344/diagon/services/auth/internal/security"
)

type AuthService interface {
	Register(
		ctx context.Context,
		email string,
		password string,
	) (*domain.User, error)

	Login(
		ctx context.Context,
		email string,
		password string,
	) (*domain.User, *security.TokenPair, error)

	Refresh(
		refreshToken string,
	) (*security.TokenPair, error)

	ValidateAccessToken(
		accessToken string,
	) (string, error)

	GetUser(
		ctx context.Context,
		userID string,
	) (*domain.User, error)
}
