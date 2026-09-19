package handler

import (
	"context"

	authv1 "github.com/vivek-344/diagon/gen/auth/v1"
)

type AuthClient interface {
	Register(
		context.Context,
		string,
		string,
	) (*authv1.RegisterResponse, error)

	Login(
		context.Context,
		string,
		string,
	) (*authv1.LoginResponse, error)

	Refresh(
		context.Context,
		string,
	) (*authv1.RefreshResponse, error)

	ValidateAccessToken(
		context.Context,
		string,
	) (*authv1.ValidateAccessTokenResponse, error)

	GetUser(
		context.Context,
		string,
	) (*authv1.GetUserResponse, error)
}
