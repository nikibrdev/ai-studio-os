package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/artifact"
)

// ArtifactStore persists Artifact aggregates in PostgreSQL — implements
// application.ArtifactStore.
type ArtifactStore struct {
	pool *pgxpool.Pool
}

var _ application.ArtifactStore = (*ArtifactStore)(nil)

// NewArtifactStore creates an ArtifactStore backed by the given pool.
func NewArtifactStore(pool *pgxpool.Pool) *ArtifactStore {
	return &ArtifactStore{pool: pool}
}

// Get loads an Artifact by id, or application.ErrNotFound if none exists.
func (s *ArtifactStore) Get(ctx context.Context, id string) (*artifact.Artifact, error) {
	const q = `
SELECT id, project_id, type, origin, author, created_at, produced_by, payload, state
FROM artifacts WHERE id = $1`

	var (
		gotID, projectID, typ, origin, author, producedBy, state string
		payload                                                  []byte
		createdAt                                                time.Time
	)
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&gotID, &projectID, &typ, &origin, &author, &createdAt, &producedBy, &payload, &state,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get artifact %s: %w", id, err)
	}

	return artifact.Restore(
		gotID, projectID,
		artifact.Type(typ), artifact.Origin(origin), artifact.Author(author),
		createdAt, producedBy, payload, artifact.State(state),
	), nil
}

// Save creates or updates an Artifact (upsert on id).
func (s *ArtifactStore) Save(ctx context.Context, a *artifact.Artifact) error {
	const q = `
INSERT INTO artifacts (id, project_id, type, origin, author, created_at, produced_by, payload, state)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE SET
	author  = EXCLUDED.author,
	payload = EXCLUDED.payload,
	state   = EXCLUDED.state`

	_, err := s.pool.Exec(ctx, q,
		a.ID(), a.ProjectID(), string(a.ArtifactType()), string(a.Origin()), string(a.Author()),
		a.CreatedAt(), a.ProducedBy(), a.Payload(), string(a.State()),
	)
	if err != nil {
		return fmt.Errorf("postgres: save artifact %s: %w", a.ID(), err)
	}
	return nil
}
