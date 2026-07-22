//go:build integration

package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestProvider_RecordSearchIsolationAndReindex requires a real Qdrant
// instance (docker-compose.yml) and a temporary FileStore directory.
// Excluded from `make verify`'s default `go test ./...`; opt-in via
// TEST_QDRANT_URL, the same pattern as the standalone QdrantClient
// integration test and internal/infrastructure's TEST_DATABASE_URL.
func TestProvider_RecordSearchIsolationAndReindex(t *testing.T) {
	url := os.Getenv("TEST_QDRANT_URL")
	if url == "" {
		t.Skip("TEST_QDRANT_URL not set; run docker compose up and set it to run this test")
	}

	ctx := context.Background()
	qdrant := NewQdrantClient(url)
	if err := qdrant.EnsureCollection(ctx); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}

	files := NewFileStore(t.TempDir())
	provider := NewProvider(files, qdrant)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	projectA := "provider-it-a-" + suffix
	projectB := "provider-it-b-" + suffix

	entryA, err := NewEntry(projectA, "fact", "мы решили использовать pgx для PostgreSQL", "human")
	if err != nil {
		t.Fatalf("NewEntry A: %v", err)
	}
	entryB, err := NewEntry(projectB, "fact", "совершенно другой проект и другая тема", "human")
	if err != nil {
		t.Fatalf("NewEntry B: %v", err)
	}

	if err := provider.Record(ctx, entryA); err != nil {
		t.Fatalf("Record A: %v", err)
	}
	if err := provider.Record(ctx, entryB); err != nil {
		t.Fatalf("Record B: %v", err)
	}

	// Record -> Search finds the just-recorded entry (relevance is a
	// token-match check, honestly expected of the naive embedding).
	found, err := provider.Search(ctx, projectA, "pgx PostgreSQL", 5)
	if err != nil {
		t.Fatalf("Search projectA: %v", err)
	}
	if len(found) != 1 || found[0].ID() != entryA.ID() {
		t.Fatalf("Search(%s) = %+v, want exactly entryA", projectA, found)
	}

	// Project isolation: searching projectA never returns projectB's entry.
	for _, e := range found {
		if e.ProjectID() != projectA {
			t.Errorf("Search(%s) returned entry of project %s", projectA, e.ProjectID())
		}
	}

	// Simulate the index diverging from the files (e.g. a Qdrant outage
	// or data loss) by deleting projectA's points directly, then confirm
	// Reindex rebuilds the index from the still-durable file.
	if err := deleteProjectPoints(ctx, url, projectA); err != nil {
		t.Fatalf("deleteProjectPoints: %v", err)
	}
	lost, err := provider.Search(ctx, projectA, "pgx PostgreSQL", 5)
	if err != nil {
		t.Fatalf("Search after simulated index loss: %v", err)
	}
	if len(lost) != 0 {
		t.Fatalf("Search after simulated index loss = %+v, want empty (index cleared)", lost)
	}

	if err := provider.Reindex(ctx, projectA); err != nil {
		t.Fatalf("Reindex: %v", err)
	}
	recovered, err := provider.Search(ctx, projectA, "pgx PostgreSQL", 5)
	if err != nil {
		t.Fatalf("Search after Reindex: %v", err)
	}
	if len(recovered) != 1 || recovered[0].ID() != entryA.ID() {
		t.Fatalf("Search after Reindex = %+v, want entryA recovered", recovered)
	}
}

// deleteProjectPoints removes every point tagged with projectID directly
// via Qdrant's REST API — a test-only way to simulate the index losing a
// project's data while its files remain intact, without adding a Delete
// method to QdrantClient that production code has no use for.
func deleteProjectPoints(ctx context.Context, baseURL, projectID string) error {
	body := map[string]any{
		"filter": map[string]any{
			"must": []map[string]any{
				{"key": "project_id", "match": map[string]any{"value": projectID}},
			},
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/collections/"+collectionName+"/points/delete", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("delete points: unexpected status %d", resp.StatusCode)
	}
	return nil
}
