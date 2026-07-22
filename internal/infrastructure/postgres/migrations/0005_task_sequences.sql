-- Per-project sequential counter for the public TASK-NNN identifier
-- (ADR-011): one row per project, atomically incremented by
-- INSERT ... ON CONFLICT DO UPDATE ... RETURNING (TaskStore.NextID) so
-- concurrent callers (e.g. apps/api, EPIC-008) never collide on the same
-- number. Not a native PostgreSQL SEQUENCE object: those cannot be created
-- dynamically per project without DDL on every new project.

CREATE TABLE task_sequences (
	project_id  TEXT PRIMARY KEY,
	next_number INTEGER NOT NULL DEFAULT 1
);
