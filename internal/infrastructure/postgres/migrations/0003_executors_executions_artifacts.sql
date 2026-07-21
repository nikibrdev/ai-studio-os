-- Executor, Execution and Artifact aggregates
-- (application.ExecutorStore/ExecutionStore/ArtifactStore).

CREATE TABLE executors (
	id            TEXT PRIMARY KEY,
	backend       TEXT NOT NULL,
	roles         TEXT[] NOT NULL DEFAULT '{}',
	registered_at TIMESTAMPTZ NOT NULL,
	state         TEXT NOT NULL
);

CREATE TABLE executions (
	id           TEXT PRIMARY KEY,
	task_id      TEXT NOT NULL REFERENCES tasks (id),
	executor_id  TEXT NOT NULL REFERENCES executors (id),
	created_at   TIMESTAMPTZ NOT NULL,
	artifact_ids TEXT[] NOT NULL DEFAULT '{}',
	state        TEXT NOT NULL
);

CREATE INDEX executions_task_id_idx ON executions (task_id);
CREATE INDEX executions_executor_id_idx ON executions (executor_id);

-- produced_by references an Execution informally (ADR-016: Artifact is an
-- independent Aggregate Root, not owned by Execution) — no foreign key, and
-- '' is a legitimate value meaning "not produced by an Execution".
CREATE TABLE artifacts (
	id          TEXT PRIMARY KEY,
	project_id  TEXT NOT NULL REFERENCES projects (id),
	type        TEXT NOT NULL,
	origin      TEXT NOT NULL,
	author      TEXT NOT NULL,
	created_at  TIMESTAMPTZ NOT NULL,
	produced_by TEXT NOT NULL DEFAULT '',
	payload     BYTEA,
	state       TEXT NOT NULL
);

CREATE INDEX artifacts_project_id_idx ON artifacts (project_id);
