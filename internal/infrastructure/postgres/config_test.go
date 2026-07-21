package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestNewPool_MissingDatabaseURL(t *testing.T) {
	t.Setenv(DatabaseURLEnv, "")

	_, err := NewPool(context.Background())
	if !errors.Is(err, ErrDatabaseURLNotSet) {
		t.Fatalf("NewPool() error = %v, want ErrDatabaseURLNotSet", err)
	}
}

func TestNewPoolFromDSN_InvalidDSN(t *testing.T) {
	_, err := NewPoolFromDSN(context.Background(), "not a valid dsn")
	if err == nil {
		t.Fatal("NewPoolFromDSN() with an invalid DSN: expected error, got nil")
	}
}
