package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/project"
)

// ProjectStore persists Project aggregates in PostgreSQL — implements
// application.ProjectStore.
type ProjectStore struct {
	pool *pgxpool.Pool
}

var _ application.ProjectStore = (*ProjectStore)(nil)

// NewProjectStore creates a ProjectStore backed by the given pool.
func NewProjectStore(pool *pgxpool.Pool) *ProjectStore {
	return &ProjectStore{pool: pool}
}

// Get loads a Project by id, or application.ErrNotFound if none exists.
func (s *ProjectStore) Get(ctx context.Context, id string) (*project.Project, error) {
	const q = `SELECT id, name, repositories, created_at, state FROM projects WHERE id = $1`

	var (
		gotID, name, state string
		repositories       []string
		createdAt          time.Time
	)
	err := s.pool.QueryRow(ctx, q, id).Scan(&gotID, &name, &repositories, &createdAt, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get project %s: %w", id, err)
	}

	return project.Restore(gotID, name, repositories, createdAt, project.State(state)), nil
}

// List returns every Project, ordered by id for a deterministic result
// (EPIC-009, TASK-072 — apps/dashboard has no other way to enumerate
// projects).
func (s *ProjectStore) List(ctx context.Context) ([]*project.Project, error) {
	const q = `SELECT id, name, repositories, created_at, state FROM projects ORDER BY id`

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("postgres: list projects: %w", err)
	}
	defer rows.Close()

	var projects []*project.Project
	for rows.Next() {
		var (
			id, name, state string
			repositories    []string
			createdAt       time.Time
		)
		if err := rows.Scan(&id, &name, &repositories, &createdAt, &state); err != nil {
			return nil, fmt.Errorf("postgres: scan project: %w", err)
		}
		projects = append(projects, project.Restore(id, name, repositories, createdAt, project.State(state)))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list projects: %w", err)
	}
	return projects, nil
}

// Save creates or updates a Project (upsert on id).
func (s *ProjectStore) Save(ctx context.Context, p *project.Project) error {
	const q = `
INSERT INTO projects (id, name, repositories, created_at, state)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
	name         = EXCLUDED.name,
	repositories = EXCLUDED.repositories,
	state        = EXCLUDED.state`

	repositories := p.Repositories()
	if repositories == nil {
		repositories = []string{}
	}

	_, err := s.pool.Exec(ctx, q, p.ID(), p.Name(), repositories, p.CreatedAt(), string(p.State()))
	if err != nil {
		return fmt.Errorf("postgres: save project %s: %w", p.ID(), err)
	}
	return nil
}
