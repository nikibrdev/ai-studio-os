package wiring

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/infrastructure/eventbus"
	"ai-studio-os/internal/infrastructure/github"
	"ai-studio-os/internal/infrastructure/postgres"
	"ai-studio-os/internal/platform"
)

// System is every real adapter the Application Layer needs, wired up and
// ready to use. Repository is nil when GITHUB_TOKEN is not set — the
// GitHub adapter is independent of PostgreSQL and its absence should not
// prevent using the rest of System (see TASK-050's Open Question: no
// GitHub token is available in every environment this runs in).
type System struct {
	Pool *pgxpool.Pool

	Projects   *postgres.ProjectStore
	Tasks      *postgres.TaskStore
	Executors  *postgres.ExecutorStore
	Executions *postgres.ExecutionStore
	Artifacts  *postgres.ArtifactStore

	Events platform.EventBus

	Repository platform.RepositoryProvider
}

// New connects to PostgreSQL at dsn, applies pending migrations, and
// assembles System. Callers own the returned System and must call Close.
func New(ctx context.Context, dsn string) (*System, error) {
	pool, err := postgres.NewPoolFromDSN(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("wiring: connect: %w", err)
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("wiring: migrate: %w", err)
	}

	sys := &System{
		Pool:       pool,
		Projects:   postgres.NewProjectStore(pool),
		Tasks:      postgres.NewTaskStore(pool),
		Executors:  postgres.NewExecutorStore(pool),
		Executions: postgres.NewExecutionStore(pool),
		Artifacts:  postgres.NewArtifactStore(pool),
		Events:     eventbus.New(pool),
	}

	if repo, err := github.New(); err == nil {
		sys.Repository = repo
	}

	return sys, nil
}

// Close releases the underlying connection pool.
func (s *System) Close() {
	s.Pool.Close()
}
