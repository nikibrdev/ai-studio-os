package memory

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeFileStore struct {
	writeCalls []Entry
	writeErr   error

	listResult []Entry
	listErr    error
	listCalls  []string
}

func (f *fakeFileStore) Write(_ context.Context, entry Entry) error {
	f.writeCalls = append(f.writeCalls, entry)
	return f.writeErr
}

func (f *fakeFileStore) List(_ context.Context, projectID string) ([]Entry, error) {
	f.listCalls = append(f.listCalls, projectID)
	return f.listResult, f.listErr
}

type upsertCall struct {
	id      string
	vector  []float32
	payload map[string]any
}

type fakeVectorIndex struct {
	upsertCalls []upsertCall
	upsertErr   error

	searchResult []Point
	searchErr    error
	searchQuery  []float32
	searchLimit  int
}

func (f *fakeVectorIndex) Upsert(_ context.Context, id string, vector []float32, payload map[string]any) error {
	f.upsertCalls = append(f.upsertCalls, upsertCall{id: id, vector: vector, payload: payload})
	return f.upsertErr
}

func (f *fakeVectorIndex) Search(_ context.Context, _ string, vector []float32, limit int) ([]Point, error) {
	f.searchQuery = vector
	f.searchLimit = limit
	return f.searchResult, f.searchErr
}

func newTestEntry(t *testing.T, projectID, content string) Entry {
	t.Helper()
	e, err := NewEntry(projectID, "fact", content, "human")
	if err != nil {
		t.Fatalf("NewEntry: %v", err)
	}
	return e
}

func TestRecord_WritesFileThenUpsertsWithEmbeddingAndPayload(t *testing.T) {
	files := &fakeFileStore{}
	index := &fakeVectorIndex{}
	p := &Provider{files: files, index: index}

	entry := newTestEntry(t, "proj-1", "мы решили использовать pgx")

	if err := p.Record(context.Background(), entry); err != nil {
		t.Fatalf("Record: %v", err)
	}

	if len(files.writeCalls) != 1 || files.writeCalls[0].ID() != entry.ID() {
		t.Fatalf("files.Write calls = %+v, want one call for entry %s", files.writeCalls, entry.ID())
	}

	if len(index.upsertCalls) != 1 {
		t.Fatalf("index.Upsert calls = %+v, want exactly one", index.upsertCalls)
	}
	call := index.upsertCalls[0]
	if call.id != entry.ID() {
		t.Errorf("Upsert id = %q, want %q", call.id, entry.ID())
	}
	wantVector := embed(entry.Content())
	if len(call.vector) != len(wantVector) {
		t.Fatalf("Upsert vector length = %d, want %d", len(call.vector), len(wantVector))
	}
	for i := range wantVector {
		if call.vector[i] != wantVector[i] {
			t.Fatalf("Upsert vector = %v, want %v", call.vector, wantVector)
		}
	}
	if call.payload["project_id"] != "proj-1" {
		t.Errorf("payload[project_id] = %v, want proj-1", call.payload["project_id"])
	}
	if call.payload["kind"] != "fact" {
		t.Errorf("payload[kind] = %v, want fact", call.payload["kind"])
	}
	if call.payload["content"] != entry.Content() {
		t.Errorf("payload[content] = %v, want %q", call.payload["content"], entry.Content())
	}
	if call.payload["source"] != "human" {
		t.Errorf("payload[source] = %v, want human", call.payload["source"])
	}
	if _, err := time.Parse(time.RFC3339, call.payload["recorded_at"].(string)); err != nil {
		t.Errorf("payload[recorded_at] = %v, not RFC3339: %v", call.payload["recorded_at"], err)
	}
}

func TestRecord_FileWriteErrorPropagates_IndexNotCalled(t *testing.T) {
	wantErr := errors.New("disk full")
	files := &fakeFileStore{writeErr: wantErr}
	index := &fakeVectorIndex{}
	p := &Provider{files: files, index: index}

	err := p.Record(context.Background(), newTestEntry(t, "proj-1", "content"))
	if !errors.Is(err, wantErr) {
		t.Fatalf("Record() error = %v, want wrapping %v", err, wantErr)
	}
	if len(index.upsertCalls) != 0 {
		t.Errorf("index.Upsert called %d times, want 0 (file write failed first)", len(index.upsertCalls))
	}
}

func TestRecord_IndexErrorPropagates(t *testing.T) {
	wantErr := errors.New("qdrant unavailable")
	files := &fakeFileStore{}
	index := &fakeVectorIndex{upsertErr: wantErr}
	p := &Provider{files: files, index: index}

	err := p.Record(context.Background(), newTestEntry(t, "proj-1", "content"))
	if !errors.Is(err, wantErr) {
		t.Fatalf("Record() error = %v, want wrapping %v", err, wantErr)
	}
	if len(files.writeCalls) != 1 {
		t.Errorf("files.Write called %d times, want 1 (file already durably written)", len(files.writeCalls))
	}
}

func TestSearch_EmbedsQueryAndReconstructsEntriesFromPayload(t *testing.T) {
	recordedAt := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	index := &fakeVectorIndex{
		searchResult: []Point{
			{
				ID: "id-1", Score: 0.9,
				Payload: map[string]any{
					"project_id":  "proj-1",
					"kind":        "fact",
					"content":     "мы решили использовать pgx",
					"source":      "human",
					"recorded_at": recordedAt.Format(time.RFC3339),
				},
			},
		},
	}
	p := &Provider{files: &fakeFileStore{}, index: index}

	entries, err := p.Search(context.Background(), "proj-1", "pgx", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	wantQuery := embed("pgx")
	if len(index.searchQuery) != len(wantQuery) {
		t.Fatalf("Search query vector length = %d, want %d", len(index.searchQuery), len(wantQuery))
	}
	if index.searchLimit != 5 {
		t.Errorf("Search limit = %d, want 5", index.searchLimit)
	}

	if len(entries) != 1 {
		t.Fatalf("Search() = %d entries, want 1", len(entries))
	}
	e := entries[0]
	if e.ID() != "id-1" || e.ProjectID() != "proj-1" || e.Kind() != "fact" ||
		e.Content() != "мы решили использовать pgx" || e.Source() != "human" {
		t.Errorf("Search() entry = %+v, fields do not match payload", e)
	}
	if !e.RecordedAt().Equal(recordedAt) {
		t.Errorf("RecordedAt() = %v, want %v", e.RecordedAt(), recordedAt)
	}
}

func TestSearch_IndexErrorPropagates(t *testing.T) {
	wantErr := errors.New("qdrant unavailable")
	p := &Provider{files: &fakeFileStore{}, index: &fakeVectorIndex{searchErr: wantErr}}

	_, err := p.Search(context.Background(), "proj-1", "query", 5)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Search() error = %v, want wrapping %v", err, wantErr)
	}
}

func TestSearch_InvalidPayloadReturnsError(t *testing.T) {
	index := &fakeVectorIndex{
		searchResult: []Point{{ID: "id-1", Payload: map[string]any{"recorded_at": "not a time"}}},
	}
	p := &Provider{files: &fakeFileStore{}, index: index}

	_, err := p.Search(context.Background(), "proj-1", "query", 5)
	if err == nil {
		t.Fatal("Search() error = nil, want error for invalid recorded_at")
	}
}

func TestReindex_UpsertsEveryFileForProject(t *testing.T) {
	entryA := newTestEntry(t, "proj-1", "первая запись")
	entryB := newTestEntry(t, "proj-1", "вторая запись")
	files := &fakeFileStore{listResult: []Entry{entryA, entryB}}
	index := &fakeVectorIndex{}
	p := &Provider{files: files, index: index}

	if err := p.Reindex(context.Background(), "proj-1"); err != nil {
		t.Fatalf("Reindex: %v", err)
	}

	if len(files.listCalls) != 1 || files.listCalls[0] != "proj-1" {
		t.Fatalf("files.List calls = %v, want one call for proj-1", files.listCalls)
	}
	if len(index.upsertCalls) != 2 {
		t.Fatalf("index.Upsert calls = %d, want 2", len(index.upsertCalls))
	}
	if index.upsertCalls[0].id != entryA.ID() || index.upsertCalls[1].id != entryB.ID() {
		t.Errorf("Upsert order/ids = %+v, want %s then %s", index.upsertCalls, entryA.ID(), entryB.ID())
	}
}

func TestReindex_ListErrorPropagates(t *testing.T) {
	wantErr := errors.New("disk error")
	p := &Provider{files: &fakeFileStore{listErr: wantErr}, index: &fakeVectorIndex{}}

	err := p.Reindex(context.Background(), "proj-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Reindex() error = %v, want wrapping %v", err, wantErr)
	}
}

func TestReindex_UpsertErrorPropagates(t *testing.T) {
	wantErr := errors.New("qdrant unavailable")
	files := &fakeFileStore{listResult: []Entry{newTestEntry(t, "proj-1", "content")}}
	p := &Provider{files: files, index: &fakeVectorIndex{upsertErr: wantErr}}

	err := p.Reindex(context.Background(), "proj-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Reindex() error = %v, want wrapping %v", err, wantErr)
	}
}
