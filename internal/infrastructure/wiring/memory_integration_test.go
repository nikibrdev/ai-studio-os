//go:build integration

package wiring

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-studio-os/internal/infrastructure/memory"
)

// TestNew_WiresMemoryWhenQdrantURLProvided proves System.Memory is a real,
// working platform.MemoryProvider when New is given a reachable Qdrant URL
// — not just that the field compiles (TASK-062). Requires TEST_DATABASE_URL
// (System still needs PostgreSQL for everything else) and TEST_QDRANT_URL.
func TestNew_WiresMemoryWhenQdrantURLProvided(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	qdrantURL := os.Getenv("TEST_QDRANT_URL")
	if dsn == "" || qdrantURL == "" {
		t.Skip("TEST_DATABASE_URL and TEST_QDRANT_URL must both be set; run docker compose up and set them to run this test")
	}

	ctx := context.Background()
	sys, err := New(ctx, dsn, qdrantURL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer sys.Close()

	if sys.Memory == nil {
		t.Fatal("System.Memory is nil, want a wired Provider when TEST_QDRANT_URL is set")
	}

	projectID := fmt.Sprintf("wiring-it-%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(memoryRootDir, projectID)) })

	entry, err := memory.NewEntry(projectID, "fact", "wiring собрал Memory Provider вживую", "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}
	if err := sys.Memory.Record(ctx, entry); err != nil {
		t.Fatalf("Record: %v", err)
	}

	found, err := sys.Memory.Search(ctx, projectID, "wiring Memory Provider", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(found) != 1 || found[0].ID() != entry.ID() {
		t.Fatalf("Search() = %+v, want exactly the recorded entry", found)
	}
}
