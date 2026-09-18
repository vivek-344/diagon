package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vivek-344/diagon/services/gateway/internal/apiresponse"
	"github.com/vivek-344/diagon/services/gateway/internal/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxRequestBodySize = 1 << 20

type AuthHandler struct {
	authClient *client.AuthClient
}

func NewAuthHandler(
	authClient *client.AuthClient,
) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	User   loginUserResponse `json:"user"`
	Tokens tokenResponse     `json:"tokens"`
}

type loginUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodySize,
	)

	decoder := json.NewDecoder(r.Body)

	var req registerRequest

	if err := decoder.Decode(&req); err != nil {
		apiresponse.WriteError(
			w,
			http.StatusBadRequest,
			apiresponse.CodeInvalidRequest,
			"The request body is invalid.",
			nil,
		)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		apiresponse.WriteError(
			w,
			http.StatusBadRequest,
			apiresponse.CodeInvalidRequest,
			"The request body is invalid.",
			nil,
		)
		return
	}

	response, err := h.authClient.Register(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			apiresponse.WriteError(
				w,
				http.StatusBadRequest,
				apiresponse.CodeInvalidRequest,
				"The request contains invalid fields.",
				nil,
			)

		case codes.AlreadyExists:
			apiresponse.WriteError(
				w,
				http.StatusConflict,
				apiresponse.CodeEmailAlreadyExists,
				"An account with this email already exists.",
				nil,
			)

		default:
			apiresponse.WriteError(
				w,
				http.StatusInternalServerError,
				apiresponse.CodeInternalError,
				"An unexpected error occurred.",
				nil,
			)
		}

		return
	}

	user := response.GetUser()
	if user == nil {
		apiresponse.WriteError(
			w,
			http.StatusInternalServerError,
			apiresponse.CodeInternalError,
			"An unexpected error occurred.",
			nil,
		)
		return
	}

	result := registerResponse{
		ID:    user.GetId(),
		Email: user.GetEmail(),
		CreatedAt: user.GetCreatedAt().AsTime().Format(
			"2006-01-02T15:04:05Z07:00",
		),
	}

	apiresponse.Write(
		w,
		http.StatusCreated,
		result,
	)
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodySize,
	)

	decoder := json.NewDecoder(r.Body)

	var req loginRequest

	if err := decoder.Decode(&req); err != nil {
		apiresponse.WriteError(
			w,
			http.StatusBadRequest,
			apiresponse.CodeInvalidRequest,
			"The request body is invalid.",
			nil,
		)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		apiresponse.WriteError(
			w,
			http.StatusBadRequest,
			apiresponse.CodeInvalidRequest,
			"The request body is invalid.",
			nil,
		)
		return
	}

	response, err := h.authClient.Login(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		switch status.Code(err) {
		case codes.Unauthenticated:
			apiresponse.WriteError(
				w,
				http.StatusUnauthorized,
				apiresponse.CodeInvalidCredentials,
				"Invalid email or password.",
				nil,
			)

		default:
			apiresponse.WriteError(
				w,
				http.StatusInternalServerError,
				apiresponse.CodeInternalError,
				"An unexpected error occurred.",
				nil,
			)
		}

		return
	}

	user := response.GetUser()
	tokens := response.GetTokens()

	if user == nil || tokens == nil {
		apiresponse.WriteError(
			w,
			http.StatusInternalServerError,
			apiresponse.CodeInternalError,
			"An unexpected error occurred.",
			nil,
		)
		return
	}

	result := loginResponse{
		User: loginUserResponse{
			ID:    user.GetId(),
			Email: user.GetEmail(),
			CreatedAt: user.GetCreatedAt().AsTime().Format(
				"2006-01-02T15:04:05Z07:00",
			),
		},
		Tokens: tokenResponse{
			AccessToken:  tokens.GetAccessToken(),
			RefreshToken: tokens.GetRefreshToken(),
		},
	}

	apiresponse.Write(
		w,
		http.StatusOK,
		result,
	)
}

func (h *AuthHandler) Refresh(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodySize,
	)

	decoder := json.NewDecoder(r.Body)

	var req refreshRequest

	if err := decoder.Decode(&req); err != nil {
		apiresponse.WriteError(
			w,
			http.StatusBadRequest,
			apiresponse.CodeInvalidRequest,
			"The request body is invalid.",
			nil,
		)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		apiresponse.WriteError(
			w,
			http.StatusBadRequest,
			apiresponse.CodeInvalidRequest,
			"The request body is invalid.",
			nil,
		)
		return
	}

	if req.RefreshToken == "" {
		apiresponse.WriteError(
			w,
			http.StatusBadRequest,
			apiresponse.CodeInvalidRequest,
			"The request contains invalid fields.",
			nil,
		)
		return
	}

	response, err := h.authClient.Refresh(
		r.Context(),
		req.RefreshToken,
	)
	if err != nil {
		switch status.Code(err) {
		case codes.Unauthenticated:
			apiresponse.WriteError(
				w,
				http.StatusUnauthorized,
				apiresponse.CodeInvalidRefreshToken,
				"The refresh token is invalid or expired.",
				nil,
			)

		default:
			apiresponse.WriteError(
				w,
				http.StatusInternalServerError,
				apiresponse.CodeInternalError,
				"An unexpected error occurred.",
				nil,
			)
		}

		return
	}

	tokens := response.GetTokens()

	if tokens == nil {
		apiresponse.WriteError(
			w,
			http.StatusInternalServerError,
			apiresponse.CodeInternalError,
			"An unexpected error occurred.",
			nil,
		)
		return
	}

	result := refreshResponse{
		AccessToken:  tokens.GetAccessToken(),
		RefreshToken: tokens.GetRefreshToken(),
	}

	apiresponse.Write(
		w,
		http.StatusOK,
		result,
	)
}
