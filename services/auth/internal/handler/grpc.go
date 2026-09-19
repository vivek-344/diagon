package handler

import (
	"context"
	"errors"

	authv1 "github.com/vivek-344/diagon/gen/auth/v1"
	"github.com/vivek-344/diagon/services/auth/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer

	authService AuthService
}

func NewAuthHandler(
	authService AuthService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	user, err := h.authService.Register(
		ctx,
		req.GetEmail(),
		req.GetPassword(),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			return nil, status.Error(
				codes.InvalidArgument,
				err.Error(),
			)

		case errors.Is(err, service.ErrInvalidPassword):
			return nil, status.Error(
				codes.InvalidArgument,
				err.Error(),
			)

		case errors.Is(err, service.ErrEmailExists):
			return nil, status.Error(
				codes.AlreadyExists,
				err.Error(),
			)

		default:
			return nil, status.Error(
				codes.Internal,
				"failed to register user",
			)
		}
	}

	return &authv1.RegisterResponse{
		User: &authv1.User{
			Id:        user.ID,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
		},
	}, nil
}

func (h *AuthHandler) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	user, tokens, err := h.authService.Login(
		ctx,
		req.GetEmail(),
		req.GetPassword(),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			return nil, status.Error(
				codes.Unauthenticated,
				err.Error(),
			)

		default:
			return nil, status.Error(
				codes.Internal,
				"failed to authenticate user",
			)
		}
	}

	return &authv1.LoginResponse{
		User: &authv1.User{
			Id:        user.ID,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
		},
		Tokens: &authv1.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
	}, nil
}

func (h *AuthHandler) Refresh(
	ctx context.Context,
	req *authv1.RefreshRequest,
) (*authv1.RefreshResponse, error) {
	tokens, err := h.authService.Refresh(
		req.GetRefreshToken(),
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			return nil, status.Error(
				codes.Unauthenticated,
				err.Error(),
			)
		}

		return nil, status.Error(
			codes.Internal,
			"failed to refresh token",
		)
	}

	return &authv1.RefreshResponse{
		Tokens: &authv1.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		},
	}, nil
}

func (h *AuthHandler) ValidateAccessToken(
	ctx context.Context,
	req *authv1.ValidateAccessTokenRequest,
) (*authv1.ValidateAccessTokenResponse, error) {
	userID, err := h.authService.ValidateAccessToken(
		req.GetAccessToken(),
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAccessToken) {
			return nil, status.Error(
				codes.Unauthenticated,
				err.Error(),
			)
		}

		return nil, status.Error(
			codes.Internal,
			"failed to validate access token",
		)
	}

	return &authv1.ValidateAccessTokenResponse{
		UserId: userID,
	}, nil
}

func (h *AuthHandler) GetUser(
	ctx context.Context,
	req *authv1.GetUserRequest,
) (*authv1.GetUserResponse, error) {
	user, err := h.authService.GetUser(
		ctx,
		req.GetUserId(),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			return nil, status.Error(
				codes.NotFound,
				"user not found",
			)

		default:
			return nil, status.Error(
				codes.Internal,
				"failed to get user",
			)
		}
	}

	return &authv1.GetUserResponse{
		User: &authv1.User{
			Id:        user.ID,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
		},
	}, nil
}
