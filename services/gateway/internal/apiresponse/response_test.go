package apiresponse

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestWrite(t *testing.T) {
	data := map[string]any{"message": "ok", "count": 2}
	recorder := httptest.NewRecorder()

	Write(recorder, 201, data)

	if recorder.Code != 201 {
		t.Fatalf("status code = %d, want %d", recorder.Code, 201)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := map[string]any{"message": "ok", "count": float64(2)}
	if !reflect.DeepEqual(response.Data, want) {
		t.Fatalf("response data = %#v, want %#v", response.Data, want)
	}
}
