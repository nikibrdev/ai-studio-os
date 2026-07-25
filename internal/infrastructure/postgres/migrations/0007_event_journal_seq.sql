-- Adds a monotonic insertion sequence to event_journal so a separate
-- process can poll the journal by cursor without silently missing events
-- (EPIC-010/TASK-080: apps/orchestrator cannot use eventbus.Bus.Subscribe,
-- which delivers only within its own process — ADR-002).
--
-- Why not a cursor on occurred_at, as TASK-079 originally declared:
-- occurred_at is stamped by the domain (time.Now() inside a command)
-- BEFORE the row is written here, so it does not order rows by the moment
-- they became visible to a reader.
--
--   * Two concurrent requests: event A is stamped T1, event B is stamped
--     T2 > T1, but B commits first. A poller reads B, advances its cursor
--     to T2, and A's row — arriving afterwards with T1 < T2 — is never
--     read. Silently: the task simply stays in Ready forever, with no
--     error anywhere.
--   * `WHERE occurred_at > $1` also skips siblings sharing a timestamp
--     (WorkService.StartTask publishes three events in a row;
--     TIMESTAMPTZ resolves to microseconds).
--
-- seq is assigned at INSERT, so cursor order IS write order and the
-- domain's clock stops mattering for delivery. This is the standard
-- journal/outbox polling design.
--
-- Existing rows (journaled before this migration) get seq values from
-- BIGSERIAL automatically, in an arbitrary order among themselves — which
-- affects nothing: nothing read this journal by cursor before now.

ALTER TABLE event_journal ADD COLUMN seq BIGSERIAL;

CREATE INDEX event_journal_seq_idx ON event_journal (seq);
