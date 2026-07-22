package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON_EncodesBodyAndStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusCreated, map[string]string{"id": "proj-1"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if !strings.Contains(rec.Body.String(), `"id":"proj-1"`) {
		t.Errorf("body = %q, want it to contain the encoded id", rec.Body.String())
	}
}

func TestWriteJSON_NilValueWritesNoBody(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusNoContent, nil)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}
}

func TestDecodeJSON_DecodesRequestBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"AI Studio OS"}`))

	var body struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(req, &body); err != nil {
		t.Fatalf("decodeJSON: %v", err)
	}
	if body.Name != "AI Studio OS" {
		t.Errorf("Name = %q, want %q", body.Name, "AI Studio OS")
	}
}

func TestDecodeJSON_InvalidBodyReturnsError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not json`))

	var body map[string]string
	if err := decodeJSON(req, &body); err == nil {
		t.Error("decodeJSON() error = nil, want a decode error for invalid JSON")
	}
}
