package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vivek-344/diagon/services/auth/internal/domain"
	"github.com/vivek-344/diagon/services/auth/internal/repository"
	"github.com/vivek-344/diagon/services/auth/internal/security"
)

type AuthService struct {
	users  *repository.UserRepository
	tokens *security.TokenManager
}

func NewAuthService(
	users *repository.UserRepository,
	tokens *security.TokenManager,
) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || !strings.Contains(email, "@") {
		return nil, ErrInvalidEmail
	}

	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("service: hash password: %w", err)
	}

	user, err := s.users.Create(
		ctx,
		email,
		passwordHash,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrEmailExists
		}

		return nil, fmt.Errorf("service: create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*domain.User, *security.TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || !strings.Contains(email, "@") {
		return nil, nil, ErrInvalidCredentials
	}

	if password == "" {
		return nil, nil, ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(
		ctx,
		email,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}

		return nil, nil, fmt.Errorf(
			"service: find user: %w",
			err,
		)
	}

	ok, err := security.VerifyPassword(
		password,
		user.PasswordHash,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"service: verify password: %w",
			err,
		)
	}

	if !ok {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.tokens.GenerateTokenPair(
		user.ID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"service: generate tokens: %w",
			err,
		)
	}

	return user, tokens, nil
}
