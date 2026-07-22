package memory

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ErrNotFound is returned by Read when no entry exists for the given
// (projectID, id).
var ErrNotFound = errors.New("memory: not found")

// ErrInvalidID is returned when projectID or id would escape the
// project's directory (e.g. contains a path separator or "..").
var ErrInvalidID = errors.New("memory: invalid identifier")

// FileStore is the durable, human-readable store of knowledge entries —
// memory/<projectID>/<id>.md (2026-07-22-memory-file-format.md). It is
// the source of truth Qdrant is indexed from and rebuilt from (ADR-018).
type FileStore struct {
	rootDir string
}

// NewFileStore creates a FileStore rooted at rootDir (the memory/
// directory at the repository root, in production).
func NewFileStore(rootDir string) *FileStore {
	return &FileStore{rootDir: rootDir}
}

// Write persists entry as memory/<projectID>/<id>.md, creating the
// project's directory if needed. Writing an entry with an id that
// already exists overwrites it.
func (s *FileStore) Write(_ context.Context, entry Entry) error {
	path, err := s.entryPath(entry.ProjectID(), entry.ID())
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("memory: create project directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(serialize(entry)), 0o644); err != nil {
		return fmt.Errorf("memory: write entry %s: %w", entry.ID(), err)
	}
	return nil
}

// Read loads the entry at (projectID, id). Returns ErrNotFound if it does
// not exist.
func (s *FileStore) Read(_ context.Context, projectID, id string) (Entry, error) {
	path, err := s.entryPath(projectID, id)
	if err != nil {
		return Entry{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Entry{}, ErrNotFound
	}
	if err != nil {
		return Entry{}, fmt.Errorf("memory: read entry %s: %w", id, err)
	}
	return parse(projectID, id, data)
}

// List returns every entry recorded for projectID, ordered by ID for a
// deterministic result. Returns an empty slice (not an error) if the
// project has no directory yet.
func (s *FileStore) List(_ context.Context, projectID string) ([]Entry, error) {
	if strings.ContainsAny(projectID, `/\`) || projectID == ".." || projectID == "" {
		return nil, ErrInvalidID
	}

	dir := filepath.Join(s.rootDir, projectID)
	files, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("memory: list project %s: %w", projectID, err)
	}

	var entries []Entry
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(f.Name(), ".md")
		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("memory: read entry %s: %w", id, err)
		}
		entry, err := parse(projectID, id, data)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID() < entries[j].ID() })
	return entries, nil
}

func (s *FileStore) entryPath(projectID, id string) (string, error) {
	if strings.ContainsAny(projectID, `/\`) || projectID == ".." || projectID == "" {
		return "", ErrInvalidID
	}
	if strings.ContainsAny(id, `/\`) || id == ".." || id == "" {
		return "", ErrInvalidID
	}
	return filepath.Join(s.rootDir, projectID, id+".md"), nil
}

// serialize renders entry as frontmatter + Markdown body
// (2026-07-22-memory-file-format.md). ProjectID is not repeated in the
// frontmatter — it is already expressed by the directory.
func serialize(entry Entry) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", entry.ID())
	fmt.Fprintf(&b, "kind: %s\n", entry.Kind())
	fmt.Fprintf(&b, "source: %s\n", entry.Source())
	fmt.Fprintf(&b, "recorded_at: %s\n", entry.RecordedAt().UTC().Format(time.RFC3339))
	b.WriteString("---\n\n")
	b.WriteString(entry.Content())
	return b.String()
}

// parse reads frontmatter + body back into an Entry. projectID and id are
// supplied by the caller (from the file's own path) rather than trusted
// solely from frontmatter content.
func parse(projectID, id string, data []byte) (Entry, error) {
	text := string(data)
	const delim = "---\n"

	if !strings.HasPrefix(text, delim) {
		return Entry{}, fmt.Errorf("memory: entry %s: missing frontmatter", id)
	}
	rest := text[len(delim):]

	end := strings.Index(rest, delim)
	if end < 0 {
		return Entry{}, fmt.Errorf("memory: entry %s: unterminated frontmatter", id)
	}
	frontmatter := rest[:end]
	content := strings.TrimPrefix(rest[end+len(delim):], "\n")

	fields := map[string]string{}
	for _, line := range strings.Split(frontmatter, "\n") {
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			return Entry{}, fmt.Errorf("memory: entry %s: malformed frontmatter line %q", id, line)
		}
		fields[key] = value
	}

	recordedAt, err := time.Parse(time.RFC3339, fields["recorded_at"])
	if err != nil {
		return Entry{}, fmt.Errorf("memory: entry %s: invalid recorded_at: %w", id, err)
	}

	return restoreEntry(id, projectID, fields["kind"], fields["source"], content, recordedAt), nil
}
