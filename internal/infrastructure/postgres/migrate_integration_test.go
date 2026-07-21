//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"
)

// TestMigrate_AppliesAndIsIdempotent requires a reachable PostgreSQL —
// see docker-compose.yml. It is excluded from the default `go test ./...`
// (and therefore from `make verify`) and only runs with
// `go test -tags=integration ./...` against a database named by
// TEST_DATABASE_URL.
func TestMigrate_AppliesAndIsIdempotent(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; run docker compose up and set it to run this test")
	}

	ctx := context.Background()

	pool, err := NewPoolFromDSN(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("second Migrate (must be a no-op): %v", err)
	}

	var count int
	const q = `SELECT count(*) FROM schema_migrations`
	if err := pool.QueryRow(ctx, q).Scan(&count); err != nil {
		t.Fatalf("count applied migrations: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one applied migration, got 0")
	}
}
