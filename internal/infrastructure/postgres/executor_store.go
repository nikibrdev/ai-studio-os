package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
)

// ExecutorStore persists Executor aggregates in PostgreSQL — implements
// application.ExecutorStore.
type ExecutorStore struct {
	pool *pgxpool.Pool
}

var _ application.ExecutorStore = (*ExecutorStore)(nil)

// NewExecutorStore creates an ExecutorStore backed by the given pool.
func NewExecutorStore(pool *pgxpool.Pool) *ExecutorStore {
	return &ExecutorStore{pool: pool}
}

// Get loads an Executor by id, or application.ErrNotFound if none exists.
func (s *ExecutorStore) Get(ctx context.Context, id string) (*executor.Executor, error) {
	const q = `SELECT id, backend, roles, registered_at, state FROM executors WHERE id = $1`

	var (
		gotID, backend, state string
		roles                 []string
		registeredAt          time.Time
	)
	err := s.pool.QueryRow(ctx, q, id).Scan(&gotID, &backend, &roles, &registeredAt, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get executor %s: %w", id, err)
	}

	return executor.Restore(gotID, backend, toRoles(roles), registeredAt, executor.State(state)), nil
}

// Save creates or updates an Executor (upsert on id).
func (s *ExecutorStore) Save(ctx context.Context, e *executor.Executor) error {
	const q = `
INSERT INTO executors (id, backend, roles, registered_at, state)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
	roles = EXCLUDED.roles,
	state = EXCLUDED.state`

	_, err := s.pool.Exec(ctx, q, e.ID(), e.Backend(), fromRoles(e.Roles()), e.RegisteredAt(), string(e.State()))
	if err != nil {
		return fmt.Errorf("postgres: save executor %s: %w", e.ID(), err)
	}
	return nil
}

// List returns every Executor, ordered by id.
func (s *ExecutorStore) List(ctx context.Context) ([]*executor.Executor, error) {
	const q = `SELECT id, backend, roles, registered_at, state FROM executors ORDER BY id`

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("postgres: list executors: %w", err)
	}
	defer rows.Close()

	var executors []*executor.Executor
	for rows.Next() {
		var (
			id, backend, state string
			roles              []string
			registeredAt       time.Time
		)
		if err := rows.Scan(&id, &backend, &roles, &registeredAt, &state); err != nil {
			return nil, fmt.Errorf("postgres: scan executor: %w", err)
		}
		executors = append(executors, executor.Restore(id, backend, toRoles(roles), registeredAt, executor.State(state)))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list executors: %w", err)
	}
	return executors, nil
}

func toRoles(raw []string) []shared.Role {
	out := make([]shared.Role, len(raw))
	for i, r := range raw {
		out[i] = shared.Role(r)
	}
	return out
}

func fromRoles(roles []shared.Role) []string {
	out := make([]string, len(roles))
	for i, r := range roles {
		out[i] = string(r)
	}
	return out
}
