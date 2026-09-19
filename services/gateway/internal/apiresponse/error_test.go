package apiresponse

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(
		rec,
		http.StatusBadRequest,
		CodeInvalidRequest,
		"invalid request",
		nil,
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %s", got)
	}

	var response ErrorResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Error.Code != CodeInvalidRequest {
		t.Fatalf(
			"expected code %q, got %q",
			CodeInvalidRequest,
			response.Error.Code,
		)
	}
}
