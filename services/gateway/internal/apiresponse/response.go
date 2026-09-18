package apiresponse

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data any `json:"data"`
}

func Write(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(
		Response{
			Data: data,
		},
	)
}
