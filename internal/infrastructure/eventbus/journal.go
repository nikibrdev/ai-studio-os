package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/platform"
)

// journalEvent reconstructs a platform.Event (plus its dataCarrier
// payload) from a row of event_journal.
type journalEvent struct {
	id, typ, source, actor, projectID, subjectID string
	schemaVersion                                int
	occurredAt                                   time.Time
	data                                         map[string]string
}

func (e journalEvent) ID() string              { return e.id }
func (e journalEvent) Type() string            { return e.typ }
func (e journalEvent) SchemaVersion() int      { return e.schemaVersion }
func (e journalEvent) OccurredAt() time.Time   { return e.occurredAt }
func (e journalEvent) Source() string          { return e.source }
func (e journalEvent) Actor() string           { return e.actor }
func (e journalEvent) ProjectID() string       { return e.projectID }
func (e journalEvent) SubjectID() string       { return e.subjectID }
func (e journalEvent) Data() map[string]string { return e.data }

// journalEventColumns is the column list, in scan order, that every
// journal query selects for reconstructing a journalEvent.
const journalEventColumns = `id, type, schema_version, occurred_at, source, actor, project_id, subject_id, data`

// scanDest returns the Scan destinations for journalEventColumns, in the
// same order. Shared by every journal query so the column list and the
// scan order can only ever drift together.
func (e *journalEvent) scanDest(dataRaw *[]byte) []any {
	return []any{
		&e.id, &e.typ, &e.schemaVersion, &e.occurredAt, &e.source, &e.actor, &e.projectID, &e.subjectID, dataRaw,
	}
}

// decodeData parses the JSONB data column into the event's data map.
func (e *journalEvent) decodeData(dataRaw []byte) error {
	if err := json.Unmarshal(dataRaw, &e.data); err != nil {
		return fmt.Errorf("eventbus: unmarshal journal data for %s: %w", e.id, err)
	}
	return nil
}

// ReadJournal returns every event recorded in event_journal, ordered by
// seq (write order — see Journal.Since and migration 0007 on why not
// occurred_at), reconstructed as platform.Event values — the purpose
// ADR-002/event-model.md assigns to the journal: rebuilding read
// projections (e.g. application.TaskProjection.Rebuild) from durable
// history rather than from the live in-process bus.
func ReadJournal(ctx context.Context, pool *pgxpool.Pool) ([]platform.Event, error) {
	const q = `SELECT ` + journalEventColumns + ` FROM event_journal ORDER BY seq`

	rows, err := pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("eventbus: query journal: %w", err)
	}
	defer rows.Close()

	var events []platform.Event
	for rows.Next() {
		var (
			je      journalEvent
			dataRaw []byte
		)
		if err := rows.Scan(je.scanDest(&dataRaw)...); err != nil {
			return nil, fmt.Errorf("eventbus: scan journal row: %w", err)
		}
		if err := je.decodeData(dataRaw); err != nil {
			return nil, err
		}
		events = append(events, je)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("eventbus: iterate journal: %w", err)
	}
	return events, nil
}

// Journal is the production implementation of application.EventJournal:
// cursor-based reads of event_journal for a process that cannot subscribe
// to the in-process Bus (apps/orchestrator — ADR-002, EPIC-010).
type Journal struct {
	pool *pgxpool.Pool
}

// NewJournal creates a Journal reading through the given pool.
func NewJournal(pool *pgxpool.Pool) *Journal {
	return &Journal{pool: pool}
}

// Since implements application.EventJournal: every entry with seq strictly
// greater than afterSeq, in seq order.
//
// The cursor is event_journal.seq (a BIGSERIAL assigned at INSERT,
// migration 0007), never occurred_at: the latter is stamped by the domain
// before the row is written, so under concurrency a row can become
// readable with an occurred_at earlier than one already returned — a
// time-based cursor would step over it and never come back, silently.
func (j *Journal) Since(ctx context.Context, afterSeq int64) ([]application.JournalEntry, error) {
	const q = `SELECT seq, ` + journalEventColumns + ` FROM event_journal WHERE seq > $1 ORDER BY seq`

	rows, err := j.pool.Query(ctx, q, afterSeq)
	if err != nil {
		return nil, fmt.Errorf("eventbus: query journal since %d: %w", afterSeq, err)
	}
	defer rows.Close()

	var entries []application.JournalEntry
	for rows.Next() {
		var (
			seq     int64
			je      journalEvent
			dataRaw []byte
		)
		dest := append([]any{&seq}, je.scanDest(&dataRaw)...)
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("eventbus: scan journal row: %w", err)
		}
		if err := je.decodeData(dataRaw); err != nil {
			return nil, err
		}
		entries = append(entries, application.JournalEntry{Seq: seq, Event: je})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("eventbus: iterate journal since %d: %w", afterSeq, err)
	}
	return entries, nil
}
