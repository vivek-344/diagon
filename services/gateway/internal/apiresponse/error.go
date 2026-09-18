package apiresponse

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

const (
	CodeInvalidRequest      = "INVALID_REQUEST"
	CodeEmailAlreadyExists  = "EMAIL_ALREADY_EXISTS"
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"
	CodeInvalidRefreshToken = "INVALID_REFRESH_TOKEN"
	CodeInvalidAccessToken  = "INVALID_ACCESS_TOKEN"
	CodeInternalError       = "INTERNAL_ERROR"
)

func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	details any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(
		ErrorResponse{
			Error: APIError{
				Code:    code,
				Message: message,
				Details: details,
			},
		},
	)
}
