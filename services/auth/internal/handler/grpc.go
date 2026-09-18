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

	authService *service.AuthService
}

func NewAuthHandler(
	authService *service.AuthService,
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
