package application_test

import (
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/platform"
)

// Compile-time check: Envelope satisfies platform.Event.
var _ platform.Event = application.Envelope{}

func TestNewEvent_FieldsRoundTrip(t *testing.T) {
	occurredAt := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	e := application.NewEvent(event.TaskCreated, "task", "developer:executor-1", "proj-1", "task-1", occurredAt)

	if e.Type() != event.TaskCreated {
		t.Errorf("Type() = %q, want %q", e.Type(), event.TaskCreated)
	}
	if e.Source() != "task" {
		t.Errorf("Source() = %q, want %q", e.Source(), "task")
	}
	if e.Actor() != "developer:executor-1" {
		t.Errorf("Actor() = %q", e.Actor())
	}
	if e.ProjectID() != "proj-1" || e.SubjectID() != "task-1" {
		t.Errorf("ProjectID/SubjectID = %q/%q", e.ProjectID(), e.SubjectID())
	}
	if !e.OccurredAt().Equal(occurredAt) {
		t.Errorf("OccurredAt() = %v, want %v", e.OccurredAt(), occurredAt)
	}
	if e.SchemaVersion() != 1 {
		t.Errorf("SchemaVersion() = %d, want 1", e.SchemaVersion())
	}
	if e.ID() == "" {
		t.Error("ID() is empty, want a generated identifier")
	}
}

func TestNewEvent_ActorMayBeEmpty(t *testing.T) {
	e := application.NewEvent(event.TaskArchived, "task", "", "proj-1", "task-1", time.Now())
	if e.Actor() != "" {
		t.Errorf("Actor() = %q, want empty for a system-initiated fact", e.Actor())
	}
}

func TestNewEvent_GeneratesUniqueIDs(t *testing.T) {
	a := application.NewEvent(event.TaskCreated, "task", "", "proj-1", "task-1", time.Now())
	b := application.NewEvent(event.TaskCreated, "task", "", "proj-1", "task-1", time.Now())
	if a.ID() == b.ID() {
		t.Errorf("two events got the same ID: %q", a.ID())
	}
}
