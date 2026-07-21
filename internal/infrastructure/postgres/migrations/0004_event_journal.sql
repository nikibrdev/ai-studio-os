-- Event journal (ADR-002: the bus's journal is a subscriber-like
-- responsibility that persists every published event in PostgreSQL,
-- append-only, for audit and future projection replay).

CREATE TABLE event_journal (
	id             TEXT PRIMARY KEY,
	type           TEXT NOT NULL,
	schema_version INT NOT NULL,
	occurred_at    TIMESTAMPTZ NOT NULL,
	source         TEXT NOT NULL,
	actor          TEXT NOT NULL DEFAULT '',
	project_id     TEXT NOT NULL DEFAULT '',
	subject_id     TEXT NOT NULL DEFAULT '',
	data           JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX event_journal_type_idx ON event_journal (type);
CREATE INDEX event_journal_project_id_idx ON event_journal (project_id);
