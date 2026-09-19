package handler

import (
	"context"

	"github.com/vivek-344/diagon/services/auth/internal/domain"
	"github.com/vivek-344/diagon/services/auth/internal/security"
)

type mockAuthService struct {
	registerFn func(
		context.Context,
		string,
		string,
	) (*domain.User, error)

	loginFn func(
		context.Context,
		string,
		string,
	) (*domain.User, *security.TokenPair, error)

	refreshFn func(
		string,
	) (*security.TokenPair, error)

	validateAccessTokenFn func(
		string,
	) (string, error)

	getUserFn func(
		context.Context,
		string,
	) (*domain.User, error)
}

func (m *mockAuthService) Register(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, error) {
	return m.registerFn(ctx, email, password)
}

func (m *mockAuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, *security.TokenPair, error) {
	return m.loginFn(ctx, email, password)
}

func (m *mockAuthService) Refresh(
	refreshToken string,
) (*security.TokenPair, error) {
	return m.refreshFn(refreshToken)
}

func (m *mockAuthService) ValidateAccessToken(
	accessToken string,
) (string, error) {
	return m.validateAccessTokenFn(accessToken)
}

func (m *mockAuthService) GetUser(
	ctx context.Context,
	userID string,
) (*domain.User, error) {
	return m.getUserFn(ctx, userID)
}
