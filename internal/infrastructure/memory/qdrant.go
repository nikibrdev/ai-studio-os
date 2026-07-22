package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// collectionName is the single Qdrant collection all projects' entries
// share — project_id is a payload field used as a search filter, not a
// separate collection per project (ADR-018; same reasoning as project_id
// being a column, not a table, in the PostgreSQL adapters, EPIC-005).
const collectionName = "memory_entries"

// QdrantClient talks to Qdrant's REST API directly over net/http — no
// client library. Qdrant stores and searches vectors; it does not
// generate them (see embed).
type QdrantClient struct {
	baseURL string
	client  *http.Client
}

// NewQdrantClient creates a QdrantClient for the given base URL (e.g.
// "http://localhost:6333").
func NewQdrantClient(baseURL string) *QdrantClient {
	return &QdrantClient{baseURL: baseURL, client: &http.Client{Timeout: 10 * time.Second}}
}

// QdrantAPIError is returned when Qdrant responds with a non-2xx status;
// it carries enough context to diagnose the failure.
type QdrantAPIError struct {
	Method, Path string
	StatusCode   int
	Body         string
}

func (e *QdrantAPIError) Error() string {
	return fmt.Sprintf("qdrant: %s %s: unexpected status %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// Point is one search result: identifier, similarity score and payload.
type Point struct {
	ID      string
	Score   float32
	Payload map[string]any
}

// EnsureCollection creates the collection if it does not already exist.
// Safe to call on every startup (idempotent).
func (c *QdrantClient) EnsureCollection(ctx context.Context) error {
	err := c.do(ctx, http.MethodGet, "/collections/"+collectionName, nil, nil)
	if err == nil {
		return nil
	}

	var apiErr *QdrantAPIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		return fmt.Errorf("qdrant: check collection: %w", err)
	}

	body := map[string]any{
		"vectors": map[string]any{"size": embeddingDim, "distance": "Cosine"},
	}
	if err := c.do(ctx, http.MethodPut, "/collections/"+collectionName, body, nil); err != nil {
		return fmt.Errorf("qdrant: create collection: %w", err)
	}
	return nil
}

// Upsert indexes one point. id must be a UUID or unsigned integer per
// Qdrant's point ID constraints — entry IDs are already UUID v4
// (ADR-018).
func (c *QdrantClient) Upsert(ctx context.Context, id string, vector []float32, payload map[string]any) error {
	body := map[string]any{
		"points": []map[string]any{
			{"id": id, "vector": vector, "payload": payload},
		},
	}
	if err := c.do(ctx, http.MethodPut, "/collections/"+collectionName+"/points", body, nil); err != nil {
		return fmt.Errorf("qdrant: upsert point %s: %w", id, err)
	}
	return nil
}

// Search returns up to limit points closest to vector, restricted to
// projectID via a payload filter — project isolation is enforced at the
// query, not just by convention (memory.md: knowledge of different
// projects is never mixed).
func (c *QdrantClient) Search(ctx context.Context, projectID string, vector []float32, limit int) ([]Point, error) {
	body := map[string]any{
		"vector": vector,
		"limit":  limit,
		"filter": map[string]any{
			"must": []map[string]any{
				{"key": "project_id", "match": map[string]any{"value": projectID}},
			},
		},
		"with_payload": true,
	}

	var resp struct {
		Result []struct {
			ID      any            `json:"id"`
			Score   float32        `json:"score"`
			Payload map[string]any `json:"payload"`
		} `json:"result"`
	}
	if err := c.do(ctx, http.MethodPost, "/collections/"+collectionName+"/points/search", body, &resp); err != nil {
		return nil, fmt.Errorf("qdrant: search: %w", err)
	}

	points := make([]Point, len(resp.Result))
	for i, r := range resp.Result {
		points[i] = Point{ID: fmt.Sprintf("%v", r.ID), Score: r.Score, Payload: r.Payload}
	}
	return points, nil
}

func (c *QdrantClient) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("qdrant: marshal request body for %s %s: %w", method, path, err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("qdrant: build request %s %s: %w", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("qdrant: read response body for %s %s: %w", method, path, err)
	}

	if resp.StatusCode >= 300 {
		return &QdrantAPIError{Method: method, Path: path, StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("qdrant: decode response for %s %s: %w", method, path, err)
		}
	}
	return nil
}
