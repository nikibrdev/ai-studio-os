-- Project and Task aggregates (application.ProjectStore / TaskStore,
-- ADR-004: PostgreSQL is the source of truth for tasks).

CREATE TABLE projects (
	id           TEXT PRIMARY KEY,
	name         TEXT NOT NULL,
	repositories TEXT[] NOT NULL DEFAULT '{}',
	created_at   TIMESTAMPTZ NOT NULL,
	state        TEXT NOT NULL
);

CREATE TABLE tasks (
	id                  TEXT PRIMARY KEY,
	project_id          TEXT NOT NULL REFERENCES projects (id),
	epic_id             TEXT NOT NULL DEFAULT '',
	title               TEXT NOT NULL,
	task_type           TEXT NOT NULL,
	scope               TEXT NOT NULL DEFAULT '',
	acceptance_criteria TEXT[] NOT NULL DEFAULT '{}',
	created_at          TIMESTAMPTZ NOT NULL,
	state               TEXT NOT NULL
);

CREATE INDEX tasks_project_id_idx ON tasks (project_id);
