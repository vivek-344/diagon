package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authv1 "github.com/vivek-344/diagon/gen/auth/v1"
	"github.com/vivek-344/diagon/services/gateway/internal/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockAuthClient struct {
	AuthClient
	registerFn func(context.Context, string, string) (*authv1.RegisterResponse, error)
	loginFn    func(context.Context, string, string) (*authv1.LoginResponse, error)
	refreshFn  func(context.Context, string) (*authv1.RefreshResponse, error)
	getUserFn  func(context.Context, string) (*authv1.GetUserResponse, error)
}

func (m *mockAuthClient) Register(
	ctx context.Context,
	email string,
	password string,
) (*authv1.RegisterResponse, error) {
	return m.registerFn(ctx, email, password)
}

func (m *mockAuthClient) Login(
	ctx context.Context,
	email string,
	password string,
) (*authv1.LoginResponse, error) {
	return m.loginFn(ctx, email, password)
}

func (m *mockAuthClient) Refresh(
	ctx context.Context,
	refreshToken string,
) (*authv1.RefreshResponse, error) {
	return m.refreshFn(ctx, refreshToken)
}

func (m *mockAuthClient) GetUser(
	ctx context.Context,
	userID string,
) (*authv1.GetUserResponse, error) {
	return m.getUserFn(ctx, userID)
}

func TestAuthHandler_Success(t *testing.T) {
	authClient := &mockAuthClient{
		registerFn: func(
			_ context.Context,
			email string,
			password string,
		) (*authv1.RegisterResponse, error) {
			return &authv1.RegisterResponse{
				User: &authv1.User{
					Id:    "user-123",
					Email: email,
					CreatedAt: timestamppb.New(
						time.Date(
							2026,
							1,
							1,
							0,
							0,
							0,
							0,
							time.UTC,
						),
					),
				},
			}, nil
		},
		loginFn: func(
			_ context.Context,
			email string,
			password string,
		) (*authv1.LoginResponse, error) {
			if email != "user@example.com" || password != "password123" {
				t.Fatalf(
					"unexpected login credentials: %s, %s",
					email,
					password,
				)
			}

			return &authv1.LoginResponse{
				User: &authv1.User{
					Id:    "user-123",
					Email: email,
					CreatedAt: timestamppb.New(
						time.Date(
							2026,
							1,
							1,
							0,
							0,
							0,
							0,
							time.UTC,
						),
					),
				},
				Tokens: &authv1.TokenPair{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
				},
			}, nil
		},
		refreshFn: func(
			_ context.Context,
			refreshToken string,
		) (*authv1.RefreshResponse, error) {
			if refreshToken != "refresh-token" {
				t.Fatalf(
					"unexpected refresh token: %s",
					refreshToken,
				)
			}

			return &authv1.RefreshResponse{
				Tokens: &authv1.TokenPair{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
				},
			}, nil
		},
		getUserFn: func(
			_ context.Context,
			userID string,
		) (*authv1.GetUserResponse, error) {
			if userID != "user-123" {
				t.Fatalf(
					"unexpected user ID: %s",
					userID,
				)
			}

			return &authv1.GetUserResponse{
				User: &authv1.User{
					Id:    "user-123",
					Email: "user@example.com",
					CreatedAt: timestamppb.New(
						time.Date(
							2026,
							1,
							1,
							0,
							0,
							0,
							0,
							time.UTC,
						),
					),
				},
			}, nil
		},
	}

	handler := NewAuthHandler(authClient)

	// Register
	body := strings.NewReader(`{
		"email": "user@example.com",
		"password": "password123"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/register",
		body,
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d",
			rec.Code,
		)
	}

	// Login
	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{
			"email": "user@example.com",
			"password": "password123"
		}`),
	)

	loginReq.Header.Set(
		"Content-Type",
		"application/json",
	)

	loginRec := httptest.NewRecorder()

	handler.Login(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf(
			"expected login 200, got %d",
			loginRec.Code,
		)
	}

	// Refresh
	refreshReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh",
		strings.NewReader(`{"refresh_token":"refresh-token"}`),
	)

	refreshReq.Header.Set(
		"Content-Type",
		"application/json",
	)

	refreshRec := httptest.NewRecorder()

	handler.Refresh(refreshRec, refreshReq)

	if refreshRec.Code != http.StatusOK {
		t.Fatalf(
			"expected refresh 200, got %d",
			refreshRec.Code,
		)
	}

	// Get current user
	meReq := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/auth/me",
		nil,
	)

	meReq = meReq.WithContext(
		middleware.WithUserID(
			meReq.Context(),
			"user-123",
		),
	)

	meRec := httptest.NewRecorder()

	handler.GetCurrentUser(meRec, meReq)

	if meRec.Code != http.StatusOK {
		t.Fatalf(
			"expected me 200, got %d",
			meRec.Code,
		)
	}
}

func TestAuthHandler_Register_Errors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		registerFn func(context.Context, string, string) (*authv1.RegisterResponse, error)
		wantStatus int
	}{
		{name: "invalid JSON", body: `{`, wantStatus: http.StatusBadRequest},
		{
			name: "multiple JSON objects", body: `{"email":"user@example.com","password":"password123"}{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Auth InvalidArgument", body: `{"email":"user@example.com","password":"password123"}`,
			registerFn: func(context.Context, string, string) (*authv1.RegisterResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid request")
			}, wantStatus: http.StatusBadRequest,
		},
		{
			name: "Auth AlreadyExists", body: `{"email":"user@example.com","password":"password123"}`,
			registerFn: func(context.Context, string, string) (*authv1.RegisterResponse, error) {
				return nil, status.Error(codes.AlreadyExists, "user already exists")
			}, wantStatus: http.StatusConflict,
		},
		{
			name: "Auth unexpected error", body: `{"email":"user@example.com","password":"password123"}`,
			registerFn: func(context.Context, string, string) (*authv1.RegisterResponse, error) {
				return nil, errors.New("unexpected error")
			}, wantStatus: http.StatusInternalServerError,
		},
		{
			name: "missing user response", body: `{"email":"user@example.com","password":"password123"}`,
			registerFn: func(context.Context, string, string) (*authv1.RegisterResponse, error) {
				return &authv1.RegisterResponse{}, nil
			}, wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAuthClient{registerFn: tt.registerFn}
			handler := NewAuthHandler(client)
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestAuthHandler_Login_Errors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		loginFn  func(context.Context, string, string) (*authv1.LoginResponse, error)
		wantCode int
	}{
		{
			name:     "invalid JSON",
			body:     `{`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: `{"email":"user@example.com","password":"password123"}`,
			loginFn: func(context.Context, string, string) (*authv1.LoginResponse, error) {
				return nil, status.Error(codes.Unauthenticated, "invalid credentials")
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "unexpected error",
			body: `{"email":"user@example.com","password":"password123"}`,
			loginFn: func(context.Context, string, string) (*authv1.LoginResponse, error) {
				return nil, errors.New("unexpected error")
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAuthClient{loginFn: tt.loginFn}
			handler := NewAuthHandler(client)
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestAuthHandler_Refresh_Errors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		refreshFn func(context.Context, string) (*authv1.RefreshResponse, error)
		wantCode  int
	}{
		{name: "invalid JSON", body: `{`, wantCode: http.StatusBadRequest},
		{
			name: "invalid token",
			body: `{"refresh_token":"invalid-token"}`,
			refreshFn: func(context.Context, string) (*authv1.RefreshResponse, error) {
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "unexpected error",
			body: `{"refresh_token":"refresh-token"}`,
			refreshFn: func(context.Context, string) (*authv1.RefreshResponse, error) {
				return nil, errors.New("unexpected error")
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAuthClient{refreshFn: tt.refreshFn}
			handler := NewAuthHandler(client)
			req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Refresh(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestAuthHandler_GetCurrentUser_Errors(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		getUserFn func(context.Context, string) (*authv1.GetUserResponse, error)
		wantCode  int
	}{
		{
			name:     "missing user ID",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:   "user not found",
			userID: "user-123",
			getUserFn: func(
				context.Context,
				string,
			) (*authv1.GetUserResponse, error) {
				return nil, status.Error(
					codes.NotFound,
					"user not found",
				)
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:   "unexpected error",
			userID: "user-123",
			getUserFn: func(
				context.Context,
				string,
			) (*authv1.GetUserResponse, error) {
				return nil, errors.New("unexpected error")
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:   "missing user response",
			userID: "user-123",
			getUserFn: func(
				context.Context,
				string,
			) (*authv1.GetUserResponse, error) {
				return &authv1.GetUserResponse{}, nil
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAuthClient{
				getUserFn: tt.getUserFn,
			}

			handler := NewAuthHandler(client)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/auth/me",
				nil,
			)

			if tt.userID != "" {
				req = req.WithContext(
					middleware.WithUserID(
						req.Context(),
						tt.userID,
					),
				)
			}

			rec := httptest.NewRecorder()

			handler.GetCurrentUser(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf(
					"expected %d, got %d",
					tt.wantCode,
					rec.Code,
				)
			}
		})
	}
}
