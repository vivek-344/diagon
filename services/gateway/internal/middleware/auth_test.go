package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authv1 "github.com/vivek-344/diagon/gen/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockAuthClient struct {
	validateFn func(
		context.Context,
		string,
	) (*authv1.ValidateAccessTokenResponse, error)
}

func (m *mockAuthClient) ValidateAccessToken(
	ctx context.Context,
	token string,
) (*authv1.ValidateAccessTokenResponse, error) {
	return m.validateFn(ctx, token)
}

func TestAuthMiddleware_MissingAuthorization(t *testing.T) {
	authClient := &mockAuthClient{}

	middleware := NewAuthMiddleware(authClient)

	nextCalled := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		nextCalled = true
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	middleware.RequireAuth(next).ServeHTTP(
		rec,
		req,
	)

	if nextCalled {
		t.Fatal("next handler should not be called")
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			rec.Code,
		)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	authClient := &mockAuthClient{
		validateFn: func(
			_ context.Context,
			token string,
		) (*authv1.ValidateAccessTokenResponse, error) {
			if token != "valid-token" {
				t.Fatalf(
					"unexpected token: %s",
					token,
				)
			}

			return &authv1.ValidateAccessTokenResponse{
				UserId: "user-123",
			}, nil
		},
	}

	authMiddleware := NewAuthMiddleware(authClient)

	nextCalled := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		nextCalled = true

		userID, ok := UserID(r.Context())
		if !ok {
			t.Fatal("expected user ID in context")
		}

		if userID != "user-123" {
			t.Fatalf(
				"expected user-123, got %s",
				userID,
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer valid-token",
	)

	rec := httptest.NewRecorder()

	authMiddleware.RequireAuth(next).ServeHTTP(
		rec,
		req,
	)

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			rec.Code,
		)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	authClient := &mockAuthClient{
		validateFn: func(
			_ context.Context,
			_ string,
		) (*authv1.ValidateAccessTokenResponse, error) {
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid token",
			)
		},
	}

	authMiddleware := NewAuthMiddleware(authClient)

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		t.Fatal("next handler should not be called")
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer invalid-token",
	)

	rec := httptest.NewRecorder()

	authMiddleware.RequireAuth(next).ServeHTTP(
		rec,
		req,
	)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			rec.Code,
		)
	}
}

func TestAuthMiddleware_MalformedAuthorization(t *testing.T) {
	for _, authorization := range []string{
		"Basic abc",
		"Bearer",
		"Bearer one two",
		"abc",
	} {
		t.Run(authorization, func(t *testing.T) {
			authClient := &mockAuthClient{}
			next := http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				t.Fatal("next handler should not be called")
			})

			req := httptest.NewRequest(
				http.MethodGet,
				"/",
				nil,
			)
			req.Header.Set("Authorization", authorization)

			rec := httptest.NewRecorder()
			NewAuthMiddleware(authClient).RequireAuth(next).ServeHTTP(
				rec,
				req,
			)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf(
					"expected 401, got %d",
					rec.Code,
				)
			}
		})
	}
}
