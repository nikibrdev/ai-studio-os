//go:build integration

package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestTaskStore_NextID_SequentialFromOne requires a real PostgreSQL
// (docker compose up); see testPool in project_task_store_integration_test.go.
func TestTaskStore_NextID_SequentialFromOne(t *testing.T) {
	pool := testPool(t)
	store := NewTaskStore(pool)
	ctx := context.Background()

	projectID := fmt.Sprintf("proj-seq-%d", time.Now().UnixNano())

	for want := 1; want <= 3; want++ {
		id, err := store.NextID(ctx, projectID)
		if err != nil {
			t.Fatalf("NextID: %v", err)
		}
		wantID := fmt.Sprintf("TASK-%03d", want)
		if id != wantID {
			t.Fatalf("NextID() = %q, want %q", id, wantID)
		}
	}
}

// TestTaskStore_NextID_ConcurrentCallsAreUniqueAndGapless requires a real
// PostgreSQL: N concurrent callers for the same project must receive N
// distinct numbers covering exactly [1, N] — the property ADR-011 asked
// the single write path to guarantee (no collisions between humans and
// agents creating tasks in parallel through apps/api, EPIC-008).
func TestTaskStore_NextID_ConcurrentCallsAreUniqueAndGapless(t *testing.T) {
	pool := testPool(t)
	store := NewTaskStore(pool)
	ctx := context.Background()

	projectID := fmt.Sprintf("proj-seq-concurrent-%d", time.Now().UnixNano())
	const n = 50

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = make([]string, 0, n)
		errs    = make([]error, 0)
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := store.NextID(ctx, projectID)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			results = append(results, id)
		}()
	}
	wg.Wait()

	if len(errs) != 0 {
		t.Fatalf("NextID returned %d errors, first: %v", len(errs), errs[0])
	}
	if len(results) != n {
		t.Fatalf("got %d results, want %d", len(results), n)
	}

	seen := make(map[string]bool, n)
	for _, id := range results {
		if seen[id] {
			t.Fatalf("duplicate id %q among concurrent NextID calls: %v", id, results)
		}
		seen[id] = true
	}
	for want := 1; want <= n; want++ {
		wantID := fmt.Sprintf("TASK-%03d", want)
		if !seen[wantID] {
			t.Fatalf("missing %q among concurrent NextID calls: %v", wantID, results)
		}
	}
}
