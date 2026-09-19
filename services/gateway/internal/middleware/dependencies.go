package middleware

import (
	"context"

	authv1 "github.com/vivek-344/diagon/gen/auth/v1"
)

type AuthClient interface {
	ValidateAccessToken(
		ctx context.Context,
		accessToken string,
	) (*authv1.ValidateAccessTokenResponse, error)
}
