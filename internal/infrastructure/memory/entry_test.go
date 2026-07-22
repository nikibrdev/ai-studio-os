package memory

import (
	"errors"
	"testing"
)

func TestNewEntry_RequiresFields(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		kind      string
		content   string
	}{
		{"missing project", "", "fact", "content"},
		{"missing kind", "proj-1", "", "content"},
		{"missing content", "proj-1", "fact", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewEntry(tt.projectID, tt.kind, tt.content, "human"); !errors.Is(err, ErrMissingField) {
				t.Errorf("NewEntry() error = %v, want ErrMissingField", err)
			}
		})
	}
}

func TestNewEntry_SetsFields(t *testing.T) {
	e, err := NewEntry("proj-1", "fact", "содержимое", "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}
	if e.ProjectID() != "proj-1" || e.Kind() != "fact" || e.Content() != "содержимое" || e.Source() != "human" {
		t.Errorf("NewEntry() = %+v", e)
	}
	if e.ID() == "" {
		t.Error("ID() is empty")
	}
	if e.RecordedAt().IsZero() {
		t.Error("RecordedAt() is zero")
	}
}

func TestNewEntry_GeneratesUniqueUUIDv4IDs(t *testing.T) {
	a, err := NewEntry("proj-1", "fact", "a", "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}
	b, err := NewEntry("proj-1", "fact", "b", "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}
	if a.ID() == b.ID() {
		t.Fatal("two entries got the same ID")
	}

	for _, id := range []string{a.ID(), b.ID()} {
		if len(id) != 36 || id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
			t.Errorf("ID() = %q, want UUID v4 shape (8-4-4-4-12)", id)
		}
		if id[14] != '4' {
			t.Errorf("ID() = %q, want version nibble 4 at position 14", id)
		}
	}
}
