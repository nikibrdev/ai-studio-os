package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

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

// ReadJournal returns every event recorded in event_journal, ordered by
// occurred_at, reconstructed as platform.Event values — the purpose
// ADR-002/event-model.md assigns to the journal: rebuilding read
// projections (e.g. application.TaskProjection.Rebuild) from durable
// history rather than from the live in-process bus.
func ReadJournal(ctx context.Context, pool *pgxpool.Pool) ([]platform.Event, error) {
	const q = `
SELECT id, type, schema_version, occurred_at, source, actor, project_id, subject_id, data
FROM event_journal ORDER BY occurred_at`

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
		err := rows.Scan(
			&je.id, &je.typ, &je.schemaVersion, &je.occurredAt, &je.source, &je.actor, &je.projectID, &je.subjectID, &dataRaw,
		)
		if err != nil {
			return nil, fmt.Errorf("eventbus: scan journal row: %w", err)
		}
		if err := json.Unmarshal(dataRaw, &je.data); err != nil {
			return nil, fmt.Errorf("eventbus: unmarshal journal data for %s: %w", je.id, err)
		}
		events = append(events, je)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("eventbus: iterate journal: %w", err)
	}
	return events, nil
}
