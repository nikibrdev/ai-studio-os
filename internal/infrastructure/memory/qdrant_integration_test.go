//go:build integration

package memory

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestQdrantClient_EnsureCollectionUpsertSearch requires a real Qdrant
// instance (docker-compose.yml). Excluded from the default `go test
// ./...` (and therefore from `make verify`) and only runs with
// `go test -tags=integration ./...` when TEST_QDRANT_URL is set — the
// same opt-in pattern as internal/infrastructure's TEST_DATABASE_URL and
// agents/claude-code/container's TEST_DOCKER.
func TestQdrantClient_EnsureCollectionUpsertSearch(t *testing.T) {
	url := os.Getenv("TEST_QDRANT_URL")
	if url == "" {
		t.Skip("TEST_QDRANT_URL not set; run docker compose up and set it to run this test")
	}

	ctx := context.Background()
	c := NewQdrantClient(url)

	if err := c.EnsureCollection(ctx); err != nil {
		t.Fatalf("EnsureCollection: %v", err)
	}
	if err := c.EnsureCollection(ctx); err != nil {
		t.Fatalf("second EnsureCollection (must be a no-op): %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	projectA := "qdrant-it-a-" + suffix
	projectB := "qdrant-it-b-" + suffix

	entryA, err := NewEntry(projectA, "fact", "мы решили использовать pgx для PostgreSQL", "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}
	entryB, err := NewEntry(projectB, "fact", "совершенно другой проект и другая тема", "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}

	if err := c.Upsert(ctx, entryA.ID(), embed(entryA.Content()), map[string]any{
		"project_id": entryA.ProjectID(), "content": entryA.Content(),
	}); err != nil {
		t.Fatalf("Upsert entryA: %v", err)
	}
	if err := c.Upsert(ctx, entryB.ID(), embed(entryB.Content()), map[string]any{
		"project_id": entryB.ProjectID(), "content": entryB.Content(),
	}); err != nil {
		t.Fatalf("Upsert entryB: %v", err)
	}

	points, err := c.Search(ctx, projectA, embed("pgx PostgreSQL"), 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("Search(%s) = %d points, want 1 (project isolation): %+v", projectA, len(points), points)
	}
	if points[0].ID != entryA.ID() {
		t.Errorf("Search(%s) returned id %s, want %s", projectA, points[0].ID, entryA.ID())
	}
	if points[0].Payload["content"] != entryA.Content() {
		t.Errorf("Search result payload content = %v, want %q", points[0].Payload["content"], entryA.Content())
	}
}
