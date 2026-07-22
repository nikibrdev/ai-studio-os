package memory

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFileStore_WriteThenRead(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	entry, err := NewEntry("proj-1", "decision", "Решили использовать pgx/v5\nвторая строка", "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}
	if err := store.Write(ctx, entry); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := store.Read(ctx, "proj-1", entry.ID())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.ID() != entry.ID() || got.ProjectID() != entry.ProjectID() || got.Kind() != entry.Kind() ||
		got.Content() != entry.Content() || got.Source() != entry.Source() {
		t.Errorf("Read() = %+v, want fields matching %+v", got, entry)
	}
	// RFC3339 (used on disk for readability) has no sub-second precision -
	// compare at whole-second granularity, not exact equality.
	if !got.RecordedAt().Truncate(time.Second).Equal(entry.RecordedAt().UTC().Truncate(time.Second)) {
		t.Errorf("RecordedAt() = %v, want %v", got.RecordedAt(), entry.RecordedAt())
	}
}

func TestFileStore_Read_NotFound(t *testing.T) {
	store := NewFileStore(t.TempDir())
	_, err := store.Read(context.Background(), "proj-1", "does-not-exist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Read() error = %v, want ErrNotFound", err)
	}
}

func TestFileStore_Write_Overwrites(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	entry, err := NewEntry("proj-1", "fact", "первая версия", "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}
	if err := store.Write(ctx, entry); err != nil {
		t.Fatalf("first Write: %v", err)
	}

	updated := restoreEntry(entry.ID(), entry.ProjectID(), entry.Kind(), entry.Source(), "вторая версия", entry.RecordedAt())
	if err := store.Write(ctx, updated); err != nil {
		t.Fatalf("second Write: %v", err)
	}

	got, err := store.Read(ctx, "proj-1", entry.ID())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Content() != "вторая версия" {
		t.Errorf("Read().Content() = %q, want вторая версия", got.Content())
	}
}

func TestFileStore_List_IsolatesByProject(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	e1, _ := NewEntry("proj-1", "fact", "a", "human")
	e2, _ := NewEntry("proj-1", "fact", "b", "human")
	e3, _ := NewEntry("proj-2", "fact", "c", "human")
	for _, e := range []Entry{e1, e2, e3} {
		if err := store.Write(ctx, e); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	got, err := store.List(ctx, "proj-1")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List(proj-1) = %d entries, want 2: %+v", len(got), got)
	}
	for _, e := range got {
		if e.ProjectID() != "proj-1" {
			t.Errorf("List(proj-1) returned entry from project %q", e.ProjectID())
		}
	}
}

func TestFileStore_List_UnknownProjectReturnsEmptyNotError(t *testing.T) {
	store := NewFileStore(t.TempDir())
	got, err := store.List(context.Background(), "never-written-to")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %v, want empty", got)
	}
}

func TestFileStore_RejectsPathTraversal(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	if _, err := store.Read(ctx, "../escape", "id"); !errors.Is(err, ErrInvalidID) {
		t.Errorf("Read() with path-traversal projectID: error = %v, want ErrInvalidID", err)
	}
	if _, err := store.Read(ctx, "proj-1", "../escape"); !errors.Is(err, ErrInvalidID) {
		t.Errorf("Read() with path-traversal id: error = %v, want ErrInvalidID", err)
	}
	if _, err := store.List(ctx, ".."); !errors.Is(err, ErrInvalidID) {
		t.Errorf("List() with path-traversal projectID: error = %v, want ErrInvalidID", err)
	}
}
