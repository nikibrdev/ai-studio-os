package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrationsDir = "migrations"

// migrationLockKey is an arbitrary fixed key for the session-level
// PostgreSQL advisory lock Migrate holds for its duration: without it,
// two processes (or two test packages, as go test runs them concurrently)
// migrating a fresh database at the same time can both decide a migration
// is unapplied and both attempt CREATE TABLE, one of which fails with a
// duplicate pg_type catalog entry rather than a clean "already exists" —
// discovered by TASK-049's integration tests running the eventbus and
// postgres packages against the same database at once.
const migrationLockKey = 727694512

// dbConn is the subset of *pgxpool.Conn the migration steps need —
// narrowed so it is exercised through one dedicated connection (required
// for the session-level advisory lock below to mean anything).
type dbConn interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Migrate applies every embedded .sql migration not yet recorded in
// schema_migrations, in filename order (hence the numeric prefix
// convention, e.g. 0001_init.sql). Re-running Migrate against an
// already migrated database is a no-op. A PostgreSQL advisory lock held
// for the duration of Migrate serializes concurrent callers against the
// same database.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("postgres: acquire connection for migration lock: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("postgres: acquire migration lock: %w", err)
	}
	defer func() { _, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", migrationLockKey) }()

	if err := ensureMigrationsTable(ctx, conn); err != nil {
		return err
	}

	names, err := migrationNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		applied, err := isApplied(ctx, conn, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(ctx, conn, name); err != nil {
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

func ensureMigrationsTable(ctx context.Context, conn dbConn) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version    TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`
	if _, err := conn.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("postgres: ensure schema_migrations: %w", err)
	}
	return nil
}

func isApplied(ctx context.Context, conn dbConn, version string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`

	var exists bool
	if err := conn.QueryRow(ctx, q, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("postgres: check migration %s: %w", version, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, conn dbConn, name string) error {
	sqlBytes, err := migrationsFS.ReadFile(path.Join(migrationsDir, name))
	if err != nil {
		return fmt.Errorf("postgres: read migration %s: %w", name, err)
	}

	tx, err := conn.Begin(ctx)
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
