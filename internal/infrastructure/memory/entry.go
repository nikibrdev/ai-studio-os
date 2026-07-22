package memory

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"ai-studio-os/internal/platform"
)

// ErrMissingField is returned by NewEntry when a required field is empty.
var ErrMissingField = errors.New("memory: required field is missing")

// Entry is the concrete platform.MemoryEntry this package persists and
// indexes.
type Entry struct {
	id         string
	projectID  string
	kind       string
	content    string
	source     string
	recordedAt time.Time
}

var _ platform.MemoryEntry = Entry{}

// NewEntry creates an Entry recorded at the current time, with a freshly
// generated identifier (a UUID v4 — usable both as a filename and, per
// ADR-018, directly as a Qdrant point ID).
func NewEntry(projectID, kind, content, source string) (Entry, error) {
	if projectID == "" || kind == "" || content == "" {
		return Entry{}, ErrMissingField
	}

	id, err := newID()
	if err != nil {
		return Entry{}, fmt.Errorf("memory: generate id: %w", err)
	}
	return Entry{
		id:         id,
		projectID:  projectID,
		kind:       kind,
		content:    content,
		source:     source,
		recordedAt: time.Now(),
	}, nil
}

// restoreEntry reconstructs an Entry from previously persisted fields,
// without generating a new ID or timestamp — used when reading an entry
// back (from a file or from a Qdrant search result payload).
func restoreEntry(id, projectID, kind, source, content string, recordedAt time.Time) Entry {
	return Entry{
		id:         id,
		projectID:  projectID,
		kind:       kind,
		source:     source,
		content:    content,
		recordedAt: recordedAt,
	}
}

// ID implements platform.MemoryEntry.
func (e Entry) ID() string { return e.id }

// ProjectID implements platform.MemoryEntry.
func (e Entry) ProjectID() string { return e.projectID }

// Kind implements platform.MemoryEntry.
func (e Entry) Kind() string { return e.kind }

// Content implements platform.MemoryEntry.
func (e Entry) Content() string { return e.content }

// Source implements platform.MemoryEntry.
func (e Entry) Source() string { return e.source }

// RecordedAt implements platform.MemoryEntry.
func (e Entry) RecordedAt() time.Time { return e.recordedAt }

// newID generates a UUID v4 string from crypto/rand — no external UUID
// library (the same reasoning as internal/application/id.go: a few lines
// of stdlib code do not justify a new dependency).
func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
