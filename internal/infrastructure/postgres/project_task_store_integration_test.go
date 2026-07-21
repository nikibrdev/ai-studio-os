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

	got, err := tasks.Get(ctx, "task-store-1")
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

	_, err := store.Get(context.Background(), "does-not-exist")
	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("Get() error = %v, want application.ErrNotFound", err)
	}
}
