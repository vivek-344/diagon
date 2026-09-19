package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	authv1 "github.com/vivek-344/diagon/gen/auth/v1"
	"github.com/vivek-344/diagon/services/auth/internal/domain"
	"github.com/vivek-344/diagon/services/auth/internal/security"
	"github.com/vivek-344/diagon/services/auth/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	createdAt := time.Date(
		2026,
		9,
		19,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	authService := &mockAuthService{
		registerFn: func(
			_ context.Context,
			email string,
			password string,
		) (*domain.User, error) {
			if email != "user@example.com" {
				t.Fatalf(
					"expected email user@example.com, got %q",
					email,
				)
			}

			if password != "password123" {
				t.Fatalf(
					"expected password password123, got %q",
					password,
				)
			}

			return &domain.User{
				ID:        "user-123",
				Email:     email,
				CreatedAt: createdAt,
			}, nil
		},
	}

	handler := NewAuthHandler(authService)

	response, err := handler.Register(
		context.Background(),
		&authv1.RegisterRequest{
			Email:    "user@example.com",
			Password: "password123",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("expected response")
	}

	if response.GetUser() == nil {
		t.Fatal("expected user")
	}

	user := response.GetUser()

	if user.GetId() != "user-123" {
		t.Fatalf(
			"expected user-123, got %q",
			user.GetId(),
		)
	}

	if user.GetEmail() != "user@example.com" {
		t.Fatalf(
			"expected user@example.com, got %q",
			user.GetEmail(),
		)
	}

	if !user.GetCreatedAt().AsTime().Equal(createdAt) {
		t.Fatalf(
			"unexpected created_at: %v",
			user.GetCreatedAt().AsTime(),
		)
	}
}

func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	authService := &mockAuthService{
		registerFn: func(
			context.Context,
			string,
			string,
		) (*domain.User, error) {
			return nil, service.ErrInvalidEmail
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.Register(
		context.Background(),
		&authv1.RegisterRequest{
			Email:    "invalid",
			Password: "password123",
		},
	)

	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf(
			"expected InvalidArgument, got %s",
			status.Code(err),
		)
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	if st.Message() != service.ErrInvalidEmail.Error() {
		t.Fatalf(
			"unexpected error message: %q",
			st.Message(),
		)
	}
}

func TestAuthHandler_Register_InvalidPassword(t *testing.T) {
	authService := &mockAuthService{
		registerFn: func(
			context.Context,
			string,
			string,
		) (*domain.User, error) {
			return nil, service.ErrInvalidPassword
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.Register(
		context.Background(),
		&authv1.RegisterRequest{
			Email:    "user@example.com",
			Password: "short",
		},
	)

	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf(
			"expected InvalidArgument, got %s",
			status.Code(err),
		)
	}
}

func TestAuthHandler_Register_EmailExists(t *testing.T) {
	authService := &mockAuthService{
		registerFn: func(
			context.Context,
			string,
			string,
		) (*domain.User, error) {
			return nil, service.ErrEmailExists
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.Register(
		context.Background(),
		&authv1.RegisterRequest{
			Email:    "user@example.com",
			Password: "password123",
		},
	)

	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf(
			"expected AlreadyExists, got %s",
			status.Code(err),
		)
	}
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
	authService := &mockAuthService{
		registerFn: func(
			context.Context,
			string,
			string,
		) (*domain.User, error) {
			return nil, errors.New("database failure")
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.Register(
		context.Background(),
		&authv1.RegisterRequest{
			Email:    "user@example.com",
			Password: "password123",
		},
	)

	if status.Code(err) != codes.Internal {
		t.Fatalf(
			"expected Internal, got %s",
			status.Code(err),
		)
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	if st.Message() != "failed to register user" {
		t.Fatalf(
			"unexpected message: %q",
			st.Message(),
		)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	createdAt := time.Date(
		2026,
		9,
		19,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	expectedTokens := &security.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	authService := &mockAuthService{
		loginFn: func(
			_ context.Context,
			email string,
			password string,
		) (*domain.User, *security.TokenPair, error) {
			if email != "user@example.com" {
				t.Fatalf(
					"unexpected email: %q",
					email,
				)
			}

			if password != "password123" {
				t.Fatalf(
					"unexpected password: %q",
					password,
				)
			}

			return &domain.User{
					ID:        "user-123",
					Email:     email,
					CreatedAt: createdAt,
				},
				expectedTokens,
				nil
		},
	}

	handler := NewAuthHandler(authService)

	response, err := handler.Login(
		context.Background(),
		&authv1.LoginRequest{
			Email:    "user@example.com",
			Password: "password123",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("expected response")
	}

	user := response.GetUser()
	if user == nil {
		t.Fatal("expected user")
	}

	if user.GetId() != "user-123" {
		t.Fatalf(
			"expected user-123, got %q",
			user.GetId(),
		)
	}

	tokens := response.GetTokens()
	if tokens == nil {
		t.Fatal("expected tokens")
	}

	if tokens.GetAccessToken() != "access-token" {
		t.Fatalf(
			"unexpected access token: %q",
			tokens.GetAccessToken(),
		)
	}

	if tokens.GetRefreshToken() != "refresh-token" {
		t.Fatalf(
			"unexpected refresh token: %q",
			tokens.GetRefreshToken(),
		)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	authService := &mockAuthService{
		loginFn: func(
			context.Context,
			string,
			string,
		) (*domain.User, *security.TokenPair, error) {
			return nil, nil, service.ErrInvalidCredentials
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.Login(
		context.Background(),
		&authv1.LoginRequest{
			Email:    "user@example.com",
			Password: "wrong",
		},
	)

	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf(
			"expected Unauthenticated, got %s",
			status.Code(err),
		)
	}
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	authService := &mockAuthService{
		loginFn: func(
			context.Context,
			string,
			string,
		) (*domain.User, *security.TokenPair, error) {
			return nil, nil, errors.New("database failure")
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.Login(
		context.Background(),
		&authv1.LoginRequest{},
	)

	if status.Code(err) != codes.Internal {
		t.Fatalf(
			"expected Internal, got %s",
			status.Code(err),
		)
	}
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	authService := &mockAuthService{
		refreshFn: func(
			refreshToken string,
		) (*security.TokenPair, error) {
			if refreshToken != "refresh-token" {
				t.Fatalf(
					"unexpected refresh token: %q",
					refreshToken,
				)
			}

			return &security.TokenPair{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
			}, nil
		},
	}

	handler := NewAuthHandler(authService)

	response, err := handler.Refresh(
		context.Background(),
		&authv1.RefreshRequest{
			RefreshToken: "refresh-token",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("expected response")
	}

	tokens := response.GetTokens()

	if tokens.GetAccessToken() != "new-access-token" {
		t.Fatalf(
			"unexpected access token: %q",
			tokens.GetAccessToken(),
		)
	}

	if tokens.GetRefreshToken() != "new-refresh-token" {
		t.Fatalf(
			"unexpected refresh token: %q",
			tokens.GetRefreshToken(),
		)
	}
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	authService := &mockAuthService{
		refreshFn: func(
			string,
		) (*security.TokenPair, error) {
			return nil, service.ErrInvalidRefreshToken
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.Refresh(
		context.Background(),
		&authv1.RefreshRequest{
			RefreshToken: "invalid-token",
		},
	)

	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf(
			"expected Unauthenticated, got %s",
			status.Code(err),
		)
	}
}

func TestAuthHandler_Refresh_InternalError(t *testing.T) {
	authService := &mockAuthService{
		refreshFn: func(
			string,
		) (*security.TokenPair, error) {
			return nil, errors.New("redis/database failure")
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.Refresh(
		context.Background(),
		&authv1.RefreshRequest{
			RefreshToken: "refresh-token",
		},
	)

	if status.Code(err) != codes.Internal {
		t.Fatalf(
			"expected Internal, got %s",
			status.Code(err),
		)
	}
}

func TestAuthHandler_ValidateAccessToken_Success(t *testing.T) {
	authService := &mockAuthService{
		validateAccessTokenFn: func(
			accessToken string,
		) (string, error) {
			if accessToken != "access-token" {
				t.Fatalf(
					"unexpected access token: %q",
					accessToken,
				)
			}

			return "user-123", nil
		},
	}

	handler := NewAuthHandler(authService)

	response, err := handler.ValidateAccessToken(
		context.Background(),
		&authv1.ValidateAccessTokenRequest{
			AccessToken: "access-token",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.GetUserId() != "user-123" {
		t.Fatalf(
			"expected user-123, got %q",
			response.GetUserId(),
		)
	}
}

func TestAuthHandler_ValidateAccessToken_InvalidToken(t *testing.T) {
	authService := &mockAuthService{
		validateAccessTokenFn: func(
			string,
		) (string, error) {
			return "", service.ErrInvalidAccessToken
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.ValidateAccessToken(
		context.Background(),
		&authv1.ValidateAccessTokenRequest{
			AccessToken: "invalid-token",
		},
	)

	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf(
			"expected Unauthenticated, got %s",
			status.Code(err),
		)
	}
}

func TestAuthHandler_ValidateAccessToken_InternalError(t *testing.T) {
	authService := &mockAuthService{
		validateAccessTokenFn: func(
			string,
		) (string, error) {
			return "", errors.New("unexpected failure")
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.ValidateAccessToken(
		context.Background(),
		&authv1.ValidateAccessTokenRequest{
			AccessToken: "access-token",
		},
	)

	if status.Code(err) != codes.Internal {
		t.Fatalf(
			"expected Internal, got %s",
			status.Code(err),
		)
	}
}

func TestAuthHandler_GetUser_Success(t *testing.T) {
	createdAt := time.Date(
		2026,
		9,
		19,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	authService := &mockAuthService{
		getUserFn: func(
			_ context.Context,
			userID string,
		) (*domain.User, error) {
			if userID != "user-123" {
				t.Fatalf(
					"unexpected user ID: %q",
					userID,
				)
			}

			return &domain.User{
				ID:        "user-123",
				Email:     "user@example.com",
				CreatedAt: createdAt,
			}, nil
		},
	}

	handler := NewAuthHandler(authService)

	response, err := handler.GetUser(
		context.Background(),
		&authv1.GetUserRequest{
			UserId: "user-123",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user := response.GetUser()

	if user == nil {
		t.Fatal("expected user")
	}

	if user.GetId() != "user-123" {
		t.Fatalf(
			"expected user-123, got %q",
			user.GetId(),
		)
	}

	if user.GetEmail() != "user@example.com" {
		t.Fatalf(
			"unexpected email: %q",
			user.GetEmail(),
		)
	}

	if !user.GetCreatedAt().AsTime().Equal(createdAt) {
		t.Fatalf("unexpected created_at")
	}
}

func TestAuthHandler_GetUser_NotFound(t *testing.T) {
	authService := &mockAuthService{
		getUserFn: func(
			context.Context,
			string,
		) (*domain.User, error) {
			return nil, service.ErrUserNotFound
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.GetUser(
		context.Background(),
		&authv1.GetUserRequest{
			UserId: "missing-user",
		},
	)

	if status.Code(err) != codes.NotFound {
		t.Fatalf(
			"expected NotFound, got %s",
			status.Code(err),
		)
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	if st.Message() != "user not found" {
		t.Fatalf(
			"unexpected message: %q",
			st.Message(),
		)
	}
}

func TestAuthHandler_GetUser_InternalError(t *testing.T) {
	authService := &mockAuthService{
		getUserFn: func(
			context.Context,
			string,
		) (*domain.User, error) {
			return nil, errors.New("database failure")
		},
	}

	handler := NewAuthHandler(authService)

	_, err := handler.GetUser(
		context.Background(),
		&authv1.GetUserRequest{
			UserId: "user-123",
		},
	)

	if status.Code(err) != codes.Internal {
		t.Fatalf(
			"expected Internal, got %s",
			status.Code(err),
		)
	}
}
