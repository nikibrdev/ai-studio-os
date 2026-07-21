//go:build integration

package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"ai-studio-os/internal/infrastructure/postgres"
	"ai-studio-os/internal/platform"
)

func TestBus_Publish_WritesToEventJournal(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; run docker compose up and set it to run this test")
	}

	ctx := context.Background()
	pool, err := postgres.NewPoolFromDSN(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	bus := New(pool)
	id := fmt.Sprintf("evt-integration-%d", time.Now().UnixNano())
	e := testEventWithData{
		testEvent: testEvent{
			id: id, typ: "task.created", source: "test",
			occurredAt: time.Now(), schemaVersion: 1, projectID: "proj-1", subjectID: "task-1",
		},
		data: map[string]string{"to": "ready"},
	}

	delivered := false
	if _, err := bus.Subscribe("task.created", func(_ context.Context, _ platform.Event) error {
		delivered = true
		return nil
	}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	if err := bus.Publish(ctx, e); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if !delivered {
		t.Error("event was journaled but not delivered to the subscriber")
	}

	var (
		gotType, gotSource, gotProjectID, gotSubjectID string
		gotSchemaVersion                               int
		gotDataRaw                                     []byte
	)
	const q = `SELECT type, source, project_id, subject_id, schema_version, data FROM event_journal WHERE id = $1`
	err = pool.QueryRow(ctx, q, id).Scan(
		&gotType, &gotSource, &gotProjectID, &gotSubjectID, &gotSchemaVersion, &gotDataRaw,
	)
	if err != nil {
		t.Fatalf("query journal row: %v", err)
	}
	if gotType != "task.created" || gotSource != "test" || gotProjectID != "proj-1" || gotSubjectID != "task-1" {
		t.Errorf("journal row = type=%q source=%q project=%q subject=%q, want task.created/test/proj-1/task-1",
			gotType, gotSource, gotProjectID, gotSubjectID)
	}
	if gotSchemaVersion != 1 {
		t.Errorf("journal schema_version = %d, want 1", gotSchemaVersion)
	}

	var gotData map[string]string
	if err := json.Unmarshal(gotDataRaw, &gotData); err != nil {
		t.Fatalf("unmarshal journal data: %v", err)
	}
	if gotData["to"] != "ready" {
		t.Errorf("journal data = %v, want to=ready", gotData)
	}
}
