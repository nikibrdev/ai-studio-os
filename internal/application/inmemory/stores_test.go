package inmemory_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/task"
)

// Compile-time checks: the fakes satisfy the ports they stand in for.
var (
	_ application.ProjectStore   = inmemory.NewProjectStore()
	_ application.TaskStore      = inmemory.NewTaskStore()
	_ application.ExecutorStore  = inmemory.NewExecutorStore()
	_ application.ExecutionStore = inmemory.NewExecutionStore()
	_ application.ArtifactStore  = inmemory.NewArtifactStore()
)

func TestProjectStore_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	store := inmemory.NewProjectStore()
	p, _, err := project.New("proj-1", "AI Studio OS")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Get(ctx, "proj-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID() != p.ID() || got.Name() != p.Name() {
		t.Errorf("Get() = %+v, want the saved project", got)
	}
}

func TestTaskStore_GetMissing_ReturnsErrNotFound(t *testing.T) {
	ctx := context.Background()
	store := inmemory.NewTaskStore()
	if _, err := store.Get(ctx, "proj-1", "nonexistent"); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestTaskStore_SaveOverwritesPreviousValue(t *testing.T) {
	ctx := context.Background()
	store := inmemory.NewTaskStore()
	tsk, _, err := task.New("task-1", "proj-1", "", "Название", "feature")
	if err != nil {
		t.Fatalf("task.New: %v", err)
	}
	if err := store.Save(ctx, tsk); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := tsk.SetScope("обновлённый scope"); err != nil {
		t.Fatalf("SetScope: %v", err)
	}
	if err := store.Save(ctx, tsk); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	got, err := store.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Scope() != "обновлённый scope" {
		t.Errorf("Scope() = %q, want the overwritten value", got.Scope())
	}
}

// TestTaskStore_SameIDDifferentProjectsDoNotCollide proves BUGFIX-003 at
// the fake level too: two different projects' tasks sharing the same
// public id must not overwrite one another.
func TestTaskStore_SameIDDifferentProjectsDoNotCollide(t *testing.T) {
	ctx := context.Background()
	store := inmemory.NewTaskStore()

	taskA, _, err := task.New("TASK-001", "proj-a", "", "Задача A", "feature")
	if err != nil {
		t.Fatalf("task.New A: %v", err)
	}
	if err := store.Save(ctx, taskA); err != nil {
		t.Fatalf("Save A: %v", err)
	}
	taskB, _, err := task.New("TASK-001", "proj-b", "", "Задача B", "bugfix")
	if err != nil {
		t.Fatalf("task.New B: %v", err)
	}
	if err := store.Save(ctx, taskB); err != nil {
		t.Fatalf("Save B: %v", err)
	}

	gotA, err := store.Get(ctx, "proj-a", "TASK-001")
	if err != nil {
		t.Fatalf("Get A: %v", err)
	}
	gotB, err := store.Get(ctx, "proj-b", "TASK-001")
	if err != nil {
		t.Fatalf("Get B: %v", err)
	}
	if gotA.Title() != "Задача A" || gotB.Title() != "Задача B" {
		t.Errorf("gotA = %+v, gotB = %+v, want each project's own task untouched", gotA, gotB)
	}
}
