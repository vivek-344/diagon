package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/vivek-344/diagon/sigil/internal/domain"
	"github.com/vivek-344/diagon/sigil/utils"
)

// AuthMiddleware validates JWT tokens and adds claims to context
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Check Bearer format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := utils.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				if err == utils.ErrExpiredToken {
					http.Error(w, `{"error": "token has expired"}`, http.StatusUnauthorized)
					return
				}
				http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), domain.DeveloperIDKey, claims.DeveloperID)
			ctx = context.WithValue(ctx, domain.EmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetDeveloperIDFromContext extracts developer ID from context
func GetDeveloperIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(domain.DeveloperIDKey).(uuid.UUID)
	return id, ok
}

// GetEmailFromContext extracts email from context
func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(domain.EmailKey).(string)
	return email, ok
}
