package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz_RespondsOKWithoutDependencies(t *testing.T) {
	// Deps left zero-valued: /healthz must not touch any dependency.
	server := NewServer(Deps{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestUnknownRoute_Returns404(t *testing.T) {
	server := NewServer(Deps{})

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
