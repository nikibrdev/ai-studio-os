-- ai-studio-os: initial schema marker.
-- Aggregate tables are added by later migrations (TASK-047: projects,
-- tasks; TASK-048: executors, executions, artifacts; TASK-049:
-- event_journal). This migration exists so the runner (Migrate) has at
-- least one file to apply and record on a fresh database.
SELECT 1;
