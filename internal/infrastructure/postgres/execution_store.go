package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/execution"
)

// ExecutionStore persists Execution aggregates in PostgreSQL — implements
// application.ExecutionStore.
type ExecutionStore struct {
	pool *pgxpool.Pool
}

var _ application.ExecutionStore = (*ExecutionStore)(nil)

// NewExecutionStore creates an ExecutionStore backed by the given pool.
func NewExecutionStore(pool *pgxpool.Pool) *ExecutionStore {
	return &ExecutionStore{pool: pool}
}

// Get loads an Execution by id, or application.ErrNotFound if none exists.
func (s *ExecutionStore) Get(ctx context.Context, id string) (*execution.Execution, error) {
	const q = `SELECT id, task_id, executor_id, created_at, artifact_ids, state FROM executions WHERE id = $1`

	var (
		gotID, taskID, executorID, state string
		artifactIDs                      []string
		createdAt                        time.Time
	)
	err := s.pool.QueryRow(ctx, q, id).Scan(&gotID, &taskID, &executorID, &createdAt, &artifactIDs, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get execution %s: %w", id, err)
	}

	return execution.Restore(gotID, taskID, executorID, createdAt, artifactIDs, execution.State(state)), nil
}

// Save creates or updates an Execution (upsert on id).
func (s *ExecutionStore) Save(ctx context.Context, e *execution.Execution) error {
	const q = `
INSERT INTO executions (id, task_id, executor_id, created_at, artifact_ids, state)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
	artifact_ids = EXCLUDED.artifact_ids,
	state        = EXCLUDED.state`

	artifactIDs := e.ArtifactIDs()
	if artifactIDs == nil {
		artifactIDs = []string{}
	}

	_, err := s.pool.Exec(ctx, q, e.ID(), e.TaskID(), e.ExecutorID(), e.CreatedAt(), artifactIDs, string(e.State()))
	if err != nil {
		return fmt.Errorf("postgres: save execution %s: %w", e.ID(), err)
	}
	return nil
}
