//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/task"
)

func TestExecutorStore_SaveThenGet(t *testing.T) {
	pool := testPool(t)
	store := NewExecutorStore(pool)
	ctx := context.Background()

	e, _, err := executor.New("exec-store-1", "claude-code-instance-1", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	if err := store.Save(ctx, e); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get(ctx, "exec-store-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID() != e.ID() || got.Backend() != e.Backend() || got.State() != e.State() {
		t.Errorf("Get() = %+v, want fields matching %+v", got, e)
	}
	if !got.HasRole(shared.RoleDeveloper) {
		t.Errorf("Get().HasRole(Developer) = false, want true")
	}
}

func TestExecutorStore_Get_NotFound(t *testing.T) {
	store := NewExecutorStore(testPool(t))

	_, err := store.Get(context.Background(), "does-not-exist")
	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("Get() error = %v, want application.ErrNotFound", err)
	}
}

func TestExecutionStore_SaveThenGet(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	proj, _, err := project.New("proj-for-execution-1", "Delta")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if err := NewProjectStore(pool).Save(ctx, proj); err != nil {
		t.Fatalf("Save project: %v", err)
	}

	tk, _, err := task.New("task-for-execution-1", proj.ID(), "", "Заголовок", "feature")
	if err != nil {
		t.Fatalf("task.New: %v", err)
	}
	if err := NewTaskStore(pool).Save(ctx, tk); err != nil {
		t.Fatalf("Save task: %v", err)
	}

	ex, _, err := executor.New("executor-for-execution-1", "claude-code-instance-1", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if err := NewExecutorStore(pool).Save(ctx, ex); err != nil {
		t.Fatalf("Save executor: %v", err)
	}

	e, _, err := execution.New("execution-store-1", tk.ID(), ex.ID())
	if err != nil {
		t.Fatalf("execution.New: %v", err)
	}
	if _, err := e.Accept(); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	if err := e.RecordArtifact("art-1"); err != nil {
		t.Fatalf("RecordArtifact: %v", err)
	}

	store := NewExecutionStore(pool)
	if err := store.Save(ctx, e); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get(ctx, "execution-store-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID() != e.ID() || got.TaskID() != e.TaskID() || got.ExecutorID() != e.ExecutorID() || got.State() != e.State() {
		t.Errorf("Get() = %+v, want fields matching %+v", got, e)
	}
	if len(got.ArtifactIDs()) != 1 || got.ArtifactIDs()[0] != "art-1" {
		t.Errorf("Get().ArtifactIDs() = %v, want [art-1]", got.ArtifactIDs())
	}
}

func TestExecutionStore_Get_NotFound(t *testing.T) {
	store := NewExecutionStore(testPool(t))

	_, err := store.Get(context.Background(), "does-not-exist")
	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("Get() error = %v, want application.ErrNotFound", err)
	}
}

func TestArtifactStore_SaveThenGet(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	proj, _, err := project.New("proj-for-artifact-1", "Epsilon")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if err := NewProjectStore(pool).Save(ctx, proj); err != nil {
		t.Fatalf("Save project: %v", err)
	}

	a, _, err := artifact.New("artifact-store-1", proj.ID(), artifact.Type("PullRequest"), artifact.OriginProduced, artifact.Author("nikita"), "")
	if err != nil {
		t.Fatalf("artifact.New: %v", err)
	}
	if err := a.UpdateDraft([]byte("payload"), ""); err != nil {
		t.Fatalf("UpdateDraft: %v", err)
	}
	if _, err := a.Publish(); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	store := NewArtifactStore(pool)
	if err := store.Save(ctx, a); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get(ctx, "artifact-store-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID() != a.ID() || got.ProjectID() != a.ProjectID() || got.State() != a.State() {
		t.Errorf("Get() = %+v, want fields matching %+v", got, a)
	}
	if string(got.Payload()) != "payload" {
		t.Errorf("Get().Payload() = %q, want payload", got.Payload())
	}
}

func TestArtifactStore_Get_NotFound(t *testing.T) {
	store := NewArtifactStore(testPool(t))

	_, err := store.Get(context.Background(), "does-not-exist")
	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("Get() error = %v, want application.ErrNotFound", err)
	}
}
