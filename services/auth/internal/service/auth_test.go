package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/vivek-344/diagon/services/auth/internal/domain"
	"github.com/vivek-344/diagon/services/auth/internal/repository"
	"github.com/vivek-344/diagon/services/auth/internal/security"
)

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	repo := &mockUserRepository{}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	_, err := svc.Register(
		context.Background(),
		"not-an-email",
		"password123",
	)

	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf(
			"expected ErrInvalidEmail, got %v",
			err,
		)
	}
}

func TestAuthService_Register_InvalidPassword(t *testing.T) {
	repo := &mockUserRepository{}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	_, err := svc.Register(
		context.Background(),
		"user@example.com",
		"short",
	)

	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf(
			"expected ErrInvalidPassword, got %v",
			err,
		)
	}
}

func TestAuthService_Register_Success(t *testing.T) {
	var (
		gotEmail        string
		gotPasswordHash string
	)

	repo := &mockUserRepository{
		createFn: func(
			_ context.Context,
			email string,
			passwordHash string,
		) (*domain.User, error) {
			gotEmail = email
			gotPasswordHash = passwordHash

			return &domain.User{
				ID:           "user-123",
				Email:        email,
				PasswordHash: passwordHash,
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	user, err := svc.Register(
		context.Background(),
		" USER@Example.COM ",
		"password123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotEmail != "user@example.com" {
		t.Fatalf(
			"expected normalized email, got %q",
			gotEmail,
		)
	}

	if gotPasswordHash == "" {
		t.Fatal("expected password hash")
	}

	if gotPasswordHash == "password123" {
		t.Fatal("password must not be stored in plaintext")
	}

	if user.ID != "user-123" {
		t.Fatalf(
			"expected user-123, got %s",
			user.ID,
		)
	}
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	repo := &mockUserRepository{
		createFn: func(
			_ context.Context,
			_ string,
			_ string,
		) (*domain.User, error) {
			return nil, repository.ErrUserAlreadyExists
		},
	}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	_, err := svc.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	if !errors.Is(err, ErrEmailExists) {
		t.Fatalf(
			"expected ErrEmailExists, got %v",
			err,
		)
	}
}

func TestAuthService_Register_RepositoryError(t *testing.T) {
	repo := &mockUserRepository{
		createFn: func(
			_ context.Context,
			_ string,
			_ string,
		) (*domain.User, error) {
			return nil, errors.New("database unavailable")
		},
	}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	_, err := svc.Register(
		context.Background(),
		"user@example.com",
		"password123",
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(
		err.Error(),
		"service: create user",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := &mockUserRepository{
		findByEmailFn: func(
			_ context.Context,
			_ string,
		) (*domain.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	_, _, err := svc.Login(
		context.Background(),
		"user@example.com",
		"password123",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	hash, err := security.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &mockUserRepository{
		findByEmailFn: func(
			_ context.Context,
			_ string,
		) (*domain.User, error) {
			return &domain.User{
				ID:           "user-123",
				Email:        "user@example.com",
				PasswordHash: hash,
			}, nil
		},
	}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	_, _, err = svc.Login(
		context.Background(),
		"user@example.com",
		"wrong-password",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestAuthService_Login_TokenGenerationError(t *testing.T) {
	hash, err := security.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &mockUserRepository{
		findByEmailFn: func(
			_ context.Context,
			_ string,
		) (*domain.User, error) {
			return &domain.User{
				ID:           "user-123",
				Email:        "user@example.com",
				PasswordHash: hash,
			}, nil
		},
	}

	tokenManager := &mockTokenManager{
		generateTokenPairFn: func(
			_ string,
		) (*security.TokenPair, error) {
			return nil, errors.New("signing failure")
		},
	}

	svc := NewAuthService(
		repo,
		tokenManager,
	)

	_, _, err = svc.Login(
		context.Background(),
		"user@example.com",
		"correct-password",
	)

	if err == nil {
		t.Fatal("expected token generation error")
	}

	if !strings.Contains(
		err.Error(),
		"service: generate tokens",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	hash, err := security.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           "user-123",
		Email:        "user@example.com",
		PasswordHash: hash,
	}

	repo := &mockUserRepository{
		findByEmailFn: func(
			_ context.Context,
			email string,
		) (*domain.User, error) {
			if email != "user@example.com" {
				t.Fatalf(
					"unexpected email: %s",
					email,
				)
			}

			return user, nil
		},
	}

	tokens := &security.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	tokenManager := &mockTokenManager{
		generateTokenPairFn: func(
			userID string,
		) (*security.TokenPair, error) {
			if userID != "user-123" {
				t.Fatalf(
					"unexpected user ID: %s",
					userID,
				)
			}

			return tokens, nil
		},
	}

	svc := NewAuthService(
		repo,
		tokenManager,
	)

	gotUser, gotTokens, err := svc.Login(
		context.Background(),
		" USER@Example.COM ",
		"correct-password",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotUser.ID != "user-123" {
		t.Fatalf(
			"expected user-123, got %s",
			gotUser.ID,
		)
	}

	if gotTokens != tokens {
		t.Fatal("expected returned token pair")
	}
}

func TestAuthService_Login_RepositoryError(t *testing.T) {
	repo := &mockUserRepository{
		findByEmailFn: func(
			_ context.Context,
			_ string,
		) (*domain.User, error) {
			return nil, errors.New("database unavailable")
		},
	}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	_, _, err := svc.Login(
		context.Background(),
		"user@example.com",
		"password123",
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(
		err.Error(),
		"service: find user",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	tokenManager := &mockTokenManager{
		validateRefreshTokenFn: func(
			_ string,
		) (string, error) {
			return "", errors.New("invalid token")
		},
	}

	svc := NewAuthService(
		&mockUserRepository{},
		tokenManager,
	)

	_, err := svc.Refresh("bad-token")

	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf(
			"expected ErrInvalidRefreshToken, got %v",
			err,
		)
	}
}

func TestAuthService_Refresh_Success(t *testing.T) {
	expectedTokens := &security.TokenPair{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
	}

	tokenManager := &mockTokenManager{
		validateRefreshTokenFn: func(
			token string,
		) (string, error) {
			if token != "old-refresh" {
				t.Fatalf(
					"unexpected token: %s",
					token,
				)
			}

			return "user-123", nil
		},

		generateTokenPairFn: func(
			userID string,
		) (*security.TokenPair, error) {
			if userID != "user-123" {
				t.Fatalf(
					"unexpected user ID: %s",
					userID,
				)
			}

			return expectedTokens, nil
		},
	}

	svc := NewAuthService(
		&mockUserRepository{},
		tokenManager,
	)

	tokens, err := svc.Refresh("old-refresh")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokens != expectedTokens {
		t.Fatal("expected new token pair")
	}
}

func TestAuthService_ValidateAccessToken_Success(t *testing.T) {
	tokenManager := &mockTokenManager{
		validateAccessTokenFn: func(
			token string,
		) (string, error) {
			if token != "valid-token" {
				t.Fatalf(
					"unexpected token: %s",
					token,
				)
			}

			return "user-123", nil
		},
	}

	svc := NewAuthService(
		&mockUserRepository{},
		tokenManager,
	)

	userID, err := svc.ValidateAccessToken("valid-token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if userID != "user-123" {
		t.Fatalf(
			"expected user-123, got %s",
			userID,
		)
	}
}

func TestAuthService_ValidateAccessToken_Invalid(t *testing.T) {
	tokenManager := &mockTokenManager{
		validateAccessTokenFn: func(
			_ string,
		) (string, error) {
			return "", errors.New("invalid token")
		},
	}

	svc := NewAuthService(
		&mockUserRepository{},
		tokenManager,
	)

	_, err := svc.ValidateAccessToken("bad-token")

	if !errors.Is(err, ErrInvalidAccessToken) {
		t.Fatalf(
			"expected ErrInvalidAccessToken, got %v",
			err,
		)
	}
}

func TestAuthService_GetUser_Success(t *testing.T) {
	user := &domain.User{
		ID:    "user-123",
		Email: "user@example.com",
	}

	repo := &mockUserRepository{
		findByIDFn: func(
			_ context.Context,
			id string,
		) (*domain.User, error) {
			if id != "user-123" {
				t.Fatalf(
					"unexpected user ID: %s",
					id,
				)
			}

			return user, nil
		},
	}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	got, err := svc.GetUser(
		context.Background(),
		"user-123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != user {
		t.Fatal("expected returned user")
	}
}

func TestAuthService_GetUser_NotFound(t *testing.T) {
	repo := &mockUserRepository{
		findByIDFn: func(
			_ context.Context,
			_ string,
		) (*domain.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}

	svc := NewAuthService(
		repo,
		&mockTokenManager{},
	)

	_, err := svc.GetUser(
		context.Background(),
		"user-123",
	)

	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf(
			"expected ErrUserNotFound, got %v",
			err,
		)
	}
}
