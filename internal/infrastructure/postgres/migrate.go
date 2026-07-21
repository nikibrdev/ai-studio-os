package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrationsDir = "migrations"

// Migrate applies every embedded .sql migration not yet recorded in
// schema_migrations, in filename order (hence the numeric prefix
// convention, e.g. 0001_init.sql). Re-running Migrate against an
// already migrated database is a no-op.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := ensureMigrationsTable(ctx, pool); err != nil {
		return err
	}

	names, err := migrationNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		applied, err := isApplied(ctx, pool, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(ctx, pool, name); err != nil {
			return err
		}
	}
	return nil
}

func migrationNames() ([]string, error) {
	entries, err := fs.ReadDir(migrationsFS, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("postgres: read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version    TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`
	if _, err := pool.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("postgres: ensure schema_migrations: %w", err)
	}
	return nil
}

func isApplied(ctx context.Context, pool *pgxpool.Pool, version string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`

	var exists bool
	if err := pool.QueryRow(ctx, q, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("postgres: check migration %s: %w", version, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, name string) error {
	sqlBytes, err := migrationsFS.ReadFile(path.Join(migrationsDir, name))
	if err != nil {
		return fmt.Errorf("postgres: read migration %s: %w", name, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: begin migration %s: %w", name, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("postgres: apply migration %s: %w", name, err)
	}
	const record = `INSERT INTO schema_migrations (version) VALUES ($1)`
	if _, err := tx.Exec(ctx, record, name); err != nil {
		return fmt.Errorf("postgres: record migration %s: %w", name, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: commit migration %s: %w", name, err)
	}
	return nil
}
