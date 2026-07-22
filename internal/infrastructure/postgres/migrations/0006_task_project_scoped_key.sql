-- Fixes a real bug found live-testing EPIC-008/TASK-069: the public
-- TASK-NNN identifier (ADR-011) is unique only within a Project — TASK-065's
-- generator issues it per project_id — but `tasks.id` was the sole PRIMARY
-- KEY here (0002_projects_tasks.sql), so two different projects' first task
-- both named TASK-001 collided: the second project's Save silently
-- overwrote epic_id/scope/state on the first project's row via
-- `ON CONFLICT (id) DO UPDATE`.
--
-- Fix: tasks' identity becomes the pair (project_id, id), matching what
-- ADR-011 actually decided ("уникальная в рамках Project"). executions.task_id
-- becomes an informal reference (no FK) — the same treatment already given
-- to artifacts.produced_by in 0003 for the same reason (ADR-016: a
-- cross-aggregate reference by id does not need referential integrity
-- enforced by the referenced aggregate's now-composite key).

ALTER TABLE executions DROP CONSTRAINT executions_task_id_fkey;

ALTER TABLE tasks DROP CONSTRAINT tasks_pkey;
ALTER TABLE tasks ADD PRIMARY KEY (project_id, id);
