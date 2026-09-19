package service

import (
	"context"

	"github.com/vivek-344/diagon/services/auth/internal/domain"
	"github.com/vivek-344/diagon/services/auth/internal/security"
)

type mockUserRepository struct {
	createFn      func(context.Context, string, string) (*domain.User, error)
	findByEmailFn func(context.Context, string) (*domain.User, error)
	findByIDFn    func(context.Context, string) (*domain.User, error)
}

func (m *mockUserRepository) Create(
	ctx context.Context,
	email string,
	passwordHash string,
) (*domain.User, error) {
	if m.createFn == nil {
		panic("mockUserRepository.Create called without createFn")
	}

	return m.createFn(ctx, email, passwordHash)
}

func (m *mockUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	return m.findByEmailFn(ctx, email)
}

func (m *mockUserRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	return m.findByIDFn(ctx, id)
}

type mockTokenManager struct {
	generateTokenPairFn    func(string) (*security.TokenPair, error)
	validateAccessTokenFn  func(string) (string, error)
	validateRefreshTokenFn func(string) (string, error)
}

func (m *mockTokenManager) GenerateTokenPair(
	userID string,
) (*security.TokenPair, error) {
	return m.generateTokenPairFn(userID)
}

func (m *mockTokenManager) ValidateAccessToken(
	token string,
) (string, error) {
	return m.validateAccessTokenFn(token)
}

func (m *mockTokenManager) ValidateRefreshToken(
	token string,
) (string, error) {
	return m.validateRefreshTokenFn(token)
}
