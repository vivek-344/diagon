package middleware

import (
	"net/http"
	"strings"

	"github.com/vivek-344/diagon/services/gateway/internal/apiresponse"
	"github.com/vivek-344/diagon/services/gateway/internal/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthMiddleware struct {
	authClient *client.AuthClient
}

func NewAuthMiddleware(authClient *client.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
	}
}

func (m *AuthMiddleware) RequireAuth(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			apiresponse.WriteError(
				w,
				http.StatusUnauthorized,
				apiresponse.CodeInvalidAccessToken,
				"Authentication is required.",
				nil,
			)
			return
		}

		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {
			apiresponse.WriteError(
				w,
				http.StatusUnauthorized,
				apiresponse.CodeInvalidAccessToken,
				"Invalid authorization header.",
				nil,
			)
			return
		}

		response, err := m.authClient.ValidateAccessToken(
			r.Context(),
			parts[1],
		)
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				apiresponse.WriteError(
					w,
					http.StatusUnauthorized,
					apiresponse.CodeInvalidAccessToken,
					"Invalid or expired access token.",
					nil,
				)
				return
			}

			apiresponse.WriteError(
				w,
				http.StatusInternalServerError,
				apiresponse.CodeInternalError,
				"An unexpected error occurred.",
				nil,
			)
			return
		}

		if response.GetUserId() == "" {
			apiresponse.WriteError(
				w,
				http.StatusUnauthorized,
				apiresponse.CodeInvalidAccessToken,
				"Invalid access token.",
				nil,
			)
			return
		}

		ctx := withUserID(r.Context(), response.GetUserId())

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}
