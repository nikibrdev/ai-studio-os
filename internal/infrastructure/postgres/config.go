package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseURLEnv is the environment variable holding the PostgreSQL
// connection string (any DSN form accepted by pgx, e.g.
// "postgres://user:pass@host:5432/dbname").
const DatabaseURLEnv = "DATABASE_URL"

// ErrDatabaseURLNotSet is returned by NewPool when DatabaseURLEnv is empty.
var ErrDatabaseURLNotSet = errors.New("postgres: " + DatabaseURLEnv + " is not set")

// NewPool creates a connection pool from the DSN in DatabaseURLEnv and
// verifies connectivity with a ping before returning.
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv(DatabaseURLEnv)
	if dsn == "" {
		return nil, ErrDatabaseURLNotSet
	}
	return NewPoolFromDSN(ctx, dsn)
}

// NewPoolFromDSN creates a connection pool from an explicit DSN and
// verifies connectivity with a ping before returning. Callers own the
// returned pool and must Close it.
func NewPoolFromDSN(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return pool, nil
}
