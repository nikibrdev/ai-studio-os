//go:build integration

package eventbus

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/infrastructure/postgres"
)

// journalTestSetup connects, migrates and returns a Bus (to write through)
// and a Journal (to read through), plus the cursor position covering
// everything already in the journal.
//
// The test journal is shared and written to concurrently by other
// packages' integration tests (Go runs packages in parallel), so these
// tests never assert on the whole window after the baseline — only on
// their own uniquely-suffixed events and their order relative to each
// other, which is what the cursor actually promises. Same approach as
// postgres.TestProjectStore_List_ReturnsCreatedProjectsInOrder.
func journalTestSetup(t *testing.T) (*Bus, *Journal, int64) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; run docker compose up and set it to run this test")
	}

	ctx := context.Background()
	pool, err := postgres.NewPoolFromDSN(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	journal := NewJournal(pool)
	existing, err := journal.Since(ctx, 0)
	if err != nil {
		t.Fatalf("Since(0) to establish baseline: %v", err)
	}
	var baseline int64
	if len(existing) > 0 {
		baseline = existing[len(existing)-1].Seq
	}

	return New(pool), journal, baseline
}

// mine returns only the entries this test wrote, identified by its unique
// suffix — filtering out rows other packages' tests wrote concurrently.
func mine(entries []application.JournalEntry, suffix string) []application.JournalEntry {
	var out []application.JournalEntry
	for _, e := range entries {
		if strings.HasSuffix(e.Event.ID(), suffix) {
			out = append(out, e)
		}
	}
	return out
}

func ids(entries []application.JournalEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Event.ID()
	}
	return out
}

func TestJournal_Since_ReturnsOnlyEntriesAfterCursorInSeqOrder(t *testing.T) {
	ctx := context.Background()
	bus, journal, baseline := journalTestSetup(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	first := newTestEvent("evt-cursor-1-"+suffix, "task.created")
	second := newTestEvent("evt-cursor-2-"+suffix, "task.started")
	third := newTestEvent("evt-cursor-3-"+suffix, "task.completed")

	for _, e := range []testEvent{first, second, third} {
		if err := bus.Publish(ctx, e); err != nil {
			t.Fatalf("Publish %s: %v", e.ID(), err)
		}
	}

	all, err := journal.Since(ctx, baseline)
	if err != nil {
		t.Fatalf("Since(baseline): %v", err)
	}
	entries := mine(all, suffix)
	if got := ids(entries); len(got) != 3 || got[0] != first.ID() || got[1] != second.ID() || got[2] != third.ID() {
		t.Fatalf("Since(baseline) own entries = %v, want the three published events in write order", got)
	}

	// seq must strictly increase in write order — the whole premise of the cursor.
	if !(entries[0].Seq < entries[1].Seq && entries[1].Seq < entries[2].Seq) {
		t.Errorf("seq not strictly increasing in write order: %d, %d, %d", entries[0].Seq, entries[1].Seq, entries[2].Seq)
	}

	// Advancing the cursor past the first entry must drop exactly it.
	rest, err := journal.Since(ctx, entries[0].Seq)
	if err != nil {
		t.Fatalf("Since(first.Seq): %v", err)
	}
	if got := ids(mine(rest, suffix)); len(got) != 2 || got[0] != second.ID() || got[1] != third.ID() {
		t.Errorf("Since(first.Seq) own entries = %v, want only the second and third events", got)
	}

	// Cursor at the last own entry must yield none of this test's events —
	// an idle poller's steady state (other tests' rows may still appear).
	tail, err := journal.Since(ctx, entries[2].Seq)
	if err != nil {
		t.Fatalf("Since(third.Seq): %v", err)
	}
	if got := ids(mine(tail, suffix)); len(got) != 0 {
		t.Errorf("Since(third.Seq) own entries = %v, want none", got)
	}
}

// TestJournal_Since_DoesNotSkipRowWrittenWithEarlierOccurredAt is the
// regression test for the defect that made TASK-080 abandon a timestamp
// cursor: occurred_at is stamped by the domain BEFORE the row is written,
// so a row can become readable carrying a time earlier than one already
// returned. A `WHERE occurred_at > cursor` query would step over it and
// never come back — silently, with the task left sitting in Ready.
//
// Here the event stamped an hour in the past is written second. A seq
// cursor must still return it; a time cursor could not.
func TestJournal_Since_DoesNotSkipRowWrittenWithEarlierOccurredAt(t *testing.T) {
	ctx := context.Background()
	bus, journal, baseline := journalTestSetup(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	late := testEvent{
		id: "evt-late-stamp-" + suffix, typ: "task.created", source: "test",
		occurredAt: time.Now(), schemaVersion: 1,
	}
	early := testEvent{
		id: "evt-early-stamp-" + suffix, typ: "task.started", source: "test",
		occurredAt: time.Now().Add(-time.Hour), schemaVersion: 1,
	}

	if err := bus.Publish(ctx, late); err != nil {
		t.Fatalf("Publish late-stamped: %v", err)
	}
	if err := bus.Publish(ctx, early); err != nil {
		t.Fatalf("Publish early-stamped: %v", err)
	}

	all, err := journal.Since(ctx, baseline)
	if err != nil {
		t.Fatalf("Since(baseline): %v", err)
	}
	entries := mine(all, suffix)
	if got := ids(entries); len(got) != 2 || got[0] != late.ID() || got[1] != early.ID() {
		t.Fatalf("Since(baseline) own entries = %v, want [%s %s] — write order, not timestamp order", got, late.ID(), early.ID())
	}
	if !entries[1].Event.OccurredAt().Before(entries[0].Event.OccurredAt()) {
		t.Fatalf("test setup no longer reproduces the condition: entry 2 must carry the earlier occurred_at")
	}

	// Advancing past the late-stamped row must still yield the
	// earlier-stamped one written after it: this is the exact step a
	// timestamp cursor got wrong.
	rest, err := journal.Since(ctx, entries[0].Seq)
	if err != nil {
		t.Fatalf("Since(late.Seq): %v", err)
	}
	if got := ids(mine(rest, suffix)); len(got) != 1 || got[0] != early.ID() {
		t.Fatalf("Since(late.Seq) own entries = %v, want [%s] — the earlier-stamped row must not be skipped", got, early.ID())
	}
}

func TestJournal_Since_PreservesEventFieldsAndData(t *testing.T) {
	ctx := context.Background()
	bus, journal, baseline := journalTestSetup(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	e := testEventWithData{
		testEvent: testEvent{
			id: "evt-fields-" + suffix, typ: "task.review-completed", source: "git", actor: "reviewer:executor-3",
			projectID: "proj-1", subjectID: "TASK-001", occurredAt: time.Now(), schemaVersion: 1,
		},
		data: map[string]string{"to": "testing"},
	}
	if err := bus.Publish(ctx, e); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	all, err := journal.Since(ctx, baseline)
	if err != nil {
		t.Fatalf("Since(baseline): %v", err)
	}
	entries := mine(all, suffix)
	if len(entries) != 1 {
		t.Fatalf("Since(baseline) own entries = %v, want exactly 1", ids(entries))
	}

	got := entries[0].Event
	if got.Type() != e.Type() || got.Source() != e.Source() || got.Actor() != e.Actor() {
		t.Errorf("reconstructed event = type=%q source=%q actor=%q, want %q/%q/%q",
			got.Type(), got.Source(), got.Actor(), e.Type(), e.Source(), e.Actor())
	}
	if got.ProjectID() != e.ProjectID() || got.SubjectID() != e.SubjectID() || got.SchemaVersion() != e.SchemaVersion() {
		t.Errorf("reconstructed event = project=%q subject=%q schema=%d, want %q/%q/%d",
			got.ProjectID(), got.SubjectID(), got.SchemaVersion(), e.ProjectID(), e.SubjectID(), e.SchemaVersion())
	}

	dc, ok := got.(interface{ Data() map[string]string })
	if !ok {
		t.Fatalf("reconstructed event does not implement Data()")
	}
	if dc.Data()["to"] != "testing" {
		t.Errorf("Data() = %v, want to=testing", dc.Data())
	}
}

// TestJournal_ImplementsPort pins the port conformance at compile time in
// this package too — wiring.System's field assignment is the primary
// check, but this keeps the guarantee local to the implementation.
func TestJournal_ImplementsPort(t *testing.T) {
	var _ application.EventJournal = (*Journal)(nil)
}
