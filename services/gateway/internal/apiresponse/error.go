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
	CodeInvalidRequest     = "INVALID_REQUEST"
	CodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
	CodeInternalError      = "INTERNAL_ERROR"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
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
