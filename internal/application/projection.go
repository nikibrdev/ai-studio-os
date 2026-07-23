package application

import (
	"context"
	"sort"
	"sync"
	"time"

	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// TaskView is the read model TaskProjection builds: enough to answer
// "what state is this task in, and since when" without touching TaskStore
// (ADR-014: projections are built only from events, never by reading a
// sibling module's storage).
type TaskView struct {
	ID        string
	ProjectID string
	State     shared.TaskState
	UpdatedAt time.Time
}

// taskProjectionEvents are the event types TaskProjection subscribes to —
// exactly the ones docs/roadmap/EPIC-004-application-layer.md (TASK-045)
// names.
var taskProjectionEvents = []string{
	event.TaskCreated,
	event.TaskPlanned,
	event.TaskStarted,
	event.ReviewRequested,
	event.ReviewCompleted,
	event.TestsFailed,
	event.TestsPassed,
	event.TaskCompleted,
}

// TaskProjection is a read-only view of Task state, built exclusively
// from the events published on the golden path. It is not the source of
// truth (TaskStore is) and is fully rebuildable from the event journal at
// any time — Rebuild proves that by replaying every event this test's
// EventBus fake recorded.
//
// views is keyed by (ProjectID, SubjectID), not SubjectID alone: TASK-NNN
// is unique only within a Project (ADR-011) — a bare-id key would let two
// different projects' tasks collide in this map the same way they used to
// collide in TaskStore before BUGFIX-003.
type TaskProjection struct {
	mu    sync.Mutex
	views map[string]TaskView
}

// NewTaskProjection creates an empty projection.
func NewTaskProjection() *TaskProjection {
	return &TaskProjection{views: make(map[string]TaskView)}
}

func viewKey(projectID, id string) string { return projectID + "\x00" + id }

// Subscribe registers Handle for every event type this projection reacts
// to.
func (p *TaskProjection) Subscribe(bus platform.EventBus) error {
	for _, t := range taskProjectionEvents {
		if _, err := bus.Subscribe(t, p.Handle); err != nil {
			return err
		}
	}
	return nil
}

// Handle applies one event to the projection. Exported separately from
// Subscribe so the exact same logic can replay a recorded event journal
// (e.g. an EventBus fake's Published()) into a fresh projection, proving
// rebuildability, without going through a live bus.
func (p *TaskProjection) Handle(_ context.Context, e platform.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := viewKey(e.ProjectID(), e.SubjectID())
	v := p.views[key]
	v.ID = e.SubjectID()
	if e.ProjectID() != "" {
		v.ProjectID = e.ProjectID()
	}
	if to, ok := targetState(e); ok {
		v.State = to
	}
	v.UpdatedAt = e.OccurredAt()
	p.views[key] = v
	return nil
}

// targetState derives the Task state an event moves the projection to.
// ReviewCompleted alone is ambiguous (Testing or back to In Progress) —
// its target is carried explicitly via Envelope.WithData (spec: CompleteReview).
// TestsPassed does not move the state on its own: per ADR-008, Done is
// reached only together with TaskCompleted, after the merge.
func targetState(e platform.Event) (shared.TaskState, bool) {
	switch e.Type() {
	case event.TaskCreated:
		return shared.StateBacklog, true
	case event.TaskPlanned:
		return shared.StateReady, true
	case event.TaskStarted:
		return shared.StateInProgress, true
	case event.ReviewRequested:
		return shared.StateReview, true
	case event.ReviewCompleted:
		if env, ok := e.(Envelope); ok {
			if to, ok := env.Data()["to"]; ok {
				return shared.TaskState(to), true
			}
		}
		return "", false
	case event.TestsFailed:
		return shared.StateInProgress, true
	case event.TaskCompleted:
		return shared.StateDone, true
	default:
		return "", false
	}
}

// Get returns the current view of a task, or false if the projection has
// not seen any event for it yet.
func (p *TaskProjection) Get(projectID, id string) (TaskView, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.views[viewKey(projectID, id)]
	return v, ok
}

// ListByProject returns every view currently known for projectID, ordered
// by ID for a deterministic result (EPIC-009, TASK-072 — apps/dashboard
// has no other way to show a project's task list). TaskView already
// carries ProjectID, so a linear scan filtering on it is enough — no
// restructuring of the (ProjectID, ID)-keyed map (BUGFIX-003) into a
// nested one was needed for this.
func (p *TaskProjection) ListByProject(projectID string) []TaskView {
	p.mu.Lock()
	defer p.mu.Unlock()

	var out []TaskView
	for _, v := range p.views {
		if v.ProjectID == projectID {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Rebuild resets the projection and replays the given event journal
// (typically an EventBus fake's Published() slice, in publication order)
// through Handle — proving the projection can be reconstructed from
// scratch and is not itself a source of truth.
func (p *TaskProjection) Rebuild(ctx context.Context, journal []platform.Event) error {
	p.mu.Lock()
	p.views = make(map[string]TaskView)
	p.mu.Unlock()

	for _, e := range journal {
		if err := p.Handle(ctx, e); err != nil {
			return err
		}
	}
	return nil
}
