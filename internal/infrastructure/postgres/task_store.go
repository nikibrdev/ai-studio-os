package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/task"
)

// TaskStore persists Task aggregates in PostgreSQL — implements
// application.TaskStore. PostgreSQL is the source of truth for tasks
// (ADR-004).
type TaskStore struct {
	pool *pgxpool.Pool
}

var _ application.TaskStore = (*TaskStore)(nil)

// NewTaskStore creates a TaskStore backed by the given pool.
func NewTaskStore(pool *pgxpool.Pool) *TaskStore {
	return &TaskStore{pool: pool}
}

// Get loads a Task by id, or application.ErrNotFound if none exists.
func (s *TaskStore) Get(ctx context.Context, id string) (*task.Task, error) {
	const q = `
SELECT id, project_id, epic_id, title, task_type, scope, acceptance_criteria, created_at, state
FROM tasks WHERE id = $1`

	var (
		gotID, projectID, epicID, title, taskType, scope, state string
		acceptanceCriteria                                      []string
		createdAt                                               time.Time
	)
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&gotID, &projectID, &epicID, &title, &taskType, &scope, &acceptanceCriteria, &createdAt, &state,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get task %s: %w", id, err)
	}

	return task.Restore(
		gotID, projectID, epicID, title, taskType, scope, acceptanceCriteria, createdAt, shared.TaskState(state),
	), nil
}

// NextID atomically issues the next public TASK-NNN identifier for
// projectID (ADR-011) — implements application.TaskIDGenerator. A single
// INSERT ... ON CONFLICT DO UPDATE ... RETURNING statement means
// PostgreSQL's own row lock serializes concurrent callers; no
// application-level locking is needed.
func (s *TaskStore) NextID(ctx context.Context, projectID string) (string, error) {
	const q = `
INSERT INTO task_sequences (project_id, next_number)
VALUES ($1, 2)
ON CONFLICT (project_id) DO UPDATE SET next_number = task_sequences.next_number + 1
RETURNING next_number - 1`

	var n int
	if err := s.pool.QueryRow(ctx, q, projectID).Scan(&n); err != nil {
		return "", fmt.Errorf("postgres: next task id for project %s: %w", projectID, err)
	}
	return fmt.Sprintf("TASK-%03d", n), nil
}

// Save creates or updates a Task (upsert on id).
func (s *TaskStore) Save(ctx context.Context, t *task.Task) error {
	const q = `
INSERT INTO tasks (id, project_id, epic_id, title, task_type, scope, acceptance_criteria, created_at, state)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE SET
	epic_id             = EXCLUDED.epic_id,
	scope               = EXCLUDED.scope,
	acceptance_criteria = EXCLUDED.acceptance_criteria,
	state               = EXCLUDED.state`

	acceptanceCriteria := t.AcceptanceCriteria()
	if acceptanceCriteria == nil {
		acceptanceCriteria = []string{}
	}

	_, err := s.pool.Exec(ctx, q,
		t.ID(), t.ProjectID(), t.EpicID(), t.Title(), t.Type(), t.Scope(), acceptanceCriteria, t.CreatedAt(), string(t.State()),
	)
	if err != nil {
		return fmt.Errorf("postgres: save task %s: %w", t.ID(), err)
	}
	return nil
}
