// Package postgres provides the PostgreSQL connection, migration runner
// and (from TASK-047/048) the Store adapters for the five aggregates —
// PostgreSQL is the source of truth for tasks (ADR-004) and, in this
// layer, for every other aggregate the Application Layer persists.
//
// The driver is pgx/v5 accessed through pgxpool.Pool, chosen by ADR-017.
// Schema migrations are plain .sql files embedded at build time and applied
// by Migrate in filename order, tracked in the schema_migrations table —
// no external migration library is introduced (ADR-017).
package postgres
