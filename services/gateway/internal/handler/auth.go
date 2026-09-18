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
