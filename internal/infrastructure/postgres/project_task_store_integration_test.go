//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/task"
)

// testPool connects to TEST_DATABASE_URL, migrates it and returns a pool
// ready for the aggregate-store integration tests in this package. It
// skips the calling test if TEST_DATABASE_URL is not set — see
// migrate_integration_test.go and README.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; run docker compose up and set it to run this test")
	}

	ctx := context.Background()
	pool, err := NewPoolFromDSN(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestProjectStore_SaveThenGet(t *testing.T) {
	pool := testPool(t)
	store := NewProjectStore(pool)
	ctx := context.Background()

	p, _, err := project.New("proj-store-1", "Alpha")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if _, _, err := p.ConnectRepository("github.com/org/repo"); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}
	if _, err := p.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get(ctx, "proj-store-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID() != p.ID() || got.Name() != p.Name() || got.State() != p.State() {
		t.Errorf("Get() = %+v, want fields matching %+v", got, p)
	}
	if len(got.Repositories()) != 1 || got.Repositories()[0] != "github.com/org/repo" {
		t.Errorf("Get().Repositories() = %v, want [github.com/org/repo]", got.Repositories())
	}
}

func TestProjectStore_Get_NotFound(t *testing.T) {
	store := NewProjectStore(testPool(t))

	_, err := store.Get(context.Background(), "does-not-exist")
	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("Get() error = %v, want application.ErrNotFound", err)
	}
}

func TestProjectStore_Save_UpsertsExistingRow(t *testing.T) {
	pool := testPool(t)
	store := NewProjectStore(pool)
	ctx := context.Background()

	p, _, err := project.New("proj-store-2", "Beta")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	if _, _, err := p.ConnectRepository("github.com/org/repo2"); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}
	if _, err := p.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	got, err := store.Get(ctx, "proj-store-2")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.State() != project.StateActive {
		t.Errorf("Get().State() = %q, want %q (upsert did not apply)", got.State(), project.StateActive)
	}
}

func TestTaskStore_SaveThenGet(t *testing.T) {
	pool := testPool(t)
	projects := NewProjectStore(pool)
	tasks := NewTaskStore(pool)
	ctx := context.Background()

	proj, _, err := project.New("proj-for-task-1", "Gamma")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if err := projects.Save(ctx, proj); err != nil {
		t.Fatalf("Save project: %v", err)
	}

	tk, _, err := task.New("task-store-1", proj.ID(), "", "Заголовок", "feature")
	if err != nil {
		t.Fatalf("task.New: %v", err)
	}
	if err := tk.SetScope("сделать что-то полезное"); err != nil {
		t.Fatalf("SetScope: %v", err)
	}
	if err := tk.SetAcceptanceCriteria([]string{"критерий раз", "критерий два"}); err != nil {
		t.Fatalf("SetAcceptanceCriteria: %v", err)
	}

	if err := tasks.Save(ctx, tk); err != nil {
		t.Fatalf("Save task: %v", err)
	}

	got, err := tasks.Get(ctx, proj.ID(), "task-store-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID() != tk.ID() || got.ProjectID() != tk.ProjectID() || got.Title() != tk.Title() {
		t.Errorf("Get() = %+v, want fields matching %+v", got, tk)
	}
	if got.State() != shared.StateBacklog {
		t.Errorf("Get().State() = %q, want %q", got.State(), shared.StateBacklog)
	}
	if len(got.AcceptanceCriteria()) != 2 {
		t.Errorf("Get().AcceptanceCriteria() = %v, want 2 entries", got.AcceptanceCriteria())
	}
}

func TestTaskStore_Get_NotFound(t *testing.T) {
	store := NewTaskStore(testPool(t))

	_, err := store.Get(context.Background(), "does-not-exist-project", "does-not-exist")
	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("Get() error = %v, want application.ErrNotFound", err)
	}
}

// TestTaskStore_SameIDDifferentProjectsDoNotCollide proves BUGFIX-003:
// TASK-NNN is unique only within a Project (ADR-011) — two different
// projects each saving their own task under the same public id must not
// overwrite one another (they used to, via `ON CONFLICT (id) DO UPDATE`
// alone, discovered live-testing EPIC-008/TASK-069).
func TestTaskStore_SameIDDifferentProjectsDoNotCollide(t *testing.T) {
	pool := testPool(t)
	projects := NewProjectStore(pool)
	tasks := NewTaskStore(pool)
	ctx := context.Background()

	projA, _, err := project.New("proj-collide-a", "A")
	if err != nil {
		t.Fatalf("project.New A: %v", err)
	}
	if err := projects.Save(ctx, projA); err != nil {
		t.Fatalf("Save project A: %v", err)
	}
	projB, _, err := project.New("proj-collide-b", "B")
	if err != nil {
		t.Fatalf("project.New B: %v", err)
	}
	if err := projects.Save(ctx, projB); err != nil {
		t.Fatalf("Save project B: %v", err)
	}

	const sharedID = "TASK-001"
	taskA, _, err := task.New(sharedID, projA.ID(), "", "Задача A", "feature")
	if err != nil {
		t.Fatalf("task.New A: %v", err)
	}
	if err := tasks.Save(ctx, taskA); err != nil {
		t.Fatalf("Save task A: %v", err)
	}
	taskB, _, err := task.New(sharedID, projB.ID(), "", "Задача B", "bugfix")
	if err != nil {
		t.Fatalf("task.New B: %v", err)
	}
	if err := tasks.Save(ctx, taskB); err != nil {
		t.Fatalf("Save task B: %v", err)
	}

	gotA, err := tasks.Get(ctx, projA.ID(), sharedID)
	if err != nil {
		t.Fatalf("Get A: %v", err)
	}
	gotB, err := tasks.Get(ctx, projB.ID(), sharedID)
	if err != nil {
		t.Fatalf("Get B: %v", err)
	}
	if gotA.Title() != "Задача A" || gotA.Type() != "feature" {
		t.Errorf("project A's task = %+v, want its own title/type untouched by project B", gotA)
	}
	if gotB.Title() != "Задача B" || gotB.Type() != "bugfix" {
		t.Errorf("project B's task = %+v, want its own title/type untouched by project A", gotB)
	}
}
