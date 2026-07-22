package memory

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestQdrantClient(t *testing.T, handler http.HandlerFunc) *QdrantClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewQdrantClient(server.URL)
}

func writeQdrantJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func TestEnsureCollection_AlreadyExists(t *testing.T) {
	calls := 0
	c := newTestQdrantClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet || r.URL.Path != "/collections/memory_entries" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeQdrantJSON(t, w, http.StatusOK, map[string]string{"status": "ok"})
	})

	if err := c.EnsureCollection(context.Background()); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected exactly one request (no create call), got %d", calls)
	}
}

func TestEnsureCollection_CreatesWhenMissing(t *testing.T) {
	var gotCreateBody map[string]any
	c := newTestQdrantClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeQdrantJSON(t, w, http.StatusNotFound, map[string]string{"status": "not found"})
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&gotCreateBody); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			writeQdrantJSON(t, w, http.StatusOK, map[string]string{"status": "ok"})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	if err := c.EnsureCollection(context.Background()); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}
	vectors, ok := gotCreateBody["vectors"].(map[string]any)
	if !ok {
		t.Fatalf("create body = %+v, want a vectors object", gotCreateBody)
	}
	if size, _ := vectors["size"].(float64); int(size) != embeddingDim {
		t.Errorf("vectors.size = %v, want %d", vectors["size"], embeddingDim)
	}
	if vectors["distance"] != "Cosine" {
		t.Errorf("vectors.distance = %v, want Cosine", vectors["distance"])
	}
}

func TestEnsureCollection_PropagatesUnexpectedError(t *testing.T) {
	c := newTestQdrantClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeQdrantJSON(t, w, http.StatusInternalServerError, map[string]string{"status": "error"})
	})

	err := c.EnsureCollection(context.Background())
	var apiErr *QdrantAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("EnsureCollection() error = %v, want wrapping *QdrantAPIError with status 500", err)
	}
}

func TestUpsert_SendsPointWithVectorAndPayload(t *testing.T) {
	var gotBody map[string]any
	c := newTestQdrantClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/collections/memory_entries/points" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeQdrantJSON(t, w, http.StatusOK, map[string]string{"status": "ok"})
	})

	err := c.Upsert(context.Background(), "id-1", []float32{0.1, 0.2}, map[string]any{"project_id": "proj-1"})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	points, ok := gotBody["points"].([]any)
	if !ok || len(points) != 1 {
		t.Fatalf("request body = %+v, want one point", gotBody)
	}
	point := points[0].(map[string]any)
	if point["id"] != "id-1" {
		t.Errorf("point id = %v, want id-1", point["id"])
	}
}

func TestUpsert_PropagatesError(t *testing.T) {
	c := newTestQdrantClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeQdrantJSON(t, w, http.StatusBadRequest, map[string]string{"status": "bad vector size"})
	})

	err := c.Upsert(context.Background(), "id-1", []float32{0.1}, nil)
	var apiErr *QdrantAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("Upsert() error = %v, want wrapping *QdrantAPIError with status 400", err)
	}
}

func TestSearch_FiltersByProjectIDAndParsesResults(t *testing.T) {
	var gotBody map[string]any
	c := newTestQdrantClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/collections/memory_entries/points/search" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeQdrantJSON(t, w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": "id-1", "score": 0.9, "payload": map[string]any{"content": "hello"}},
			},
		})
	})

	points, err := c.Search(context.Background(), "proj-1", []float32{0.1, 0.2}, 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(points) != 1 || points[0].ID != "id-1" || points[0].Payload["content"] != "hello" {
		t.Fatalf("Search() = %+v", points)
	}

	filter, ok := gotBody["filter"].(map[string]any)
	if !ok {
		t.Fatalf("request body = %+v, want a filter object", gotBody)
	}
	must, _ := filter["must"].([]any)
	if len(must) != 1 {
		t.Fatalf("filter.must = %+v, want one condition", must)
	}
	condition := must[0].(map[string]any)
	if condition["key"] != "project_id" {
		t.Errorf("filter condition key = %v, want project_id", condition["key"])
	}
	match := condition["match"].(map[string]any)
	if match["value"] != "proj-1" {
		t.Errorf("filter condition value = %v, want proj-1", match["value"])
	}
}

func TestSearch_PropagatesError(t *testing.T) {
	c := newTestQdrantClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeQdrantJSON(t, w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
	})

	_, err := c.Search(context.Background(), "proj-1", []float32{0.1}, 5)
	var apiErr *QdrantAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("Search() error = %v, want wrapping *QdrantAPIError with status 503", err)
	}
}
