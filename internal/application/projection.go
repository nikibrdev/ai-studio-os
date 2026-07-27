package application

import (
	"context"
	"encoding/json"
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

	// Title, Type, Scope and AcceptanceCriteria come from TaskCreated's
	// Envelope.WithData (EPIC-009, TASK-076) — the task detail page needs
	// them, and TaskProjection is the only read path for Task (ADR-014).
	//
	// Scope and AcceptanceCriteria are revisable: TaskRefined carries a better
	// version while the task is still in Backlog (EPIC-013, TASK-096 — what a
	// Project Manager prepares before a human accepts Definition of Ready).
	// Before that event existed these fields were captured once and never
	// revised; ignoring the revision would show the human the original task
	// and hide the agent's work, which is indistinguishable from "the agent
	// did nothing". Title and Type still have no revising event.
	Title              string
	Type               string
	Scope              string
	AcceptanceCriteria []string

	// Repository and PullRequestID identify the pull request under review,
	// captured from ReviewRequested's attached data (BUGFIX-009) — the
	// reference docs/architecture/events.md always required that event to
	// carry. Without it the platform opened a pull request and then forgot
	// it, making the acceptance decision impossible: CompleteTesting needs
	// it to merge, and nothing else remembered it.
	//
	// Empty until a review is requested with a pull request, and left alone
	// afterwards: a later ReviewRequested without a reference (a human
	// re-requesting review by hand) must not erase what is already known.
	Repository    string
	PullRequestID string

	// QAReportID identifies the published TestReport a QA agent produced for
	// this task, empty until one exists (TASK-100). Only the identifier: the
	// report itself stays in the artifact, and a reader fetches it by id —
	// duplicating the text here would put the same content in two places and in
	// the event journal.
	QAReportID string
}

// taskProjectionEvents are the event types TaskProjection subscribes to —
// exactly the ones docs/roadmap/EPIC-004-application-layer.md (TASK-045)
// names.
var taskProjectionEvents = []string{
	event.TaskCreated,
	event.TaskRefined,
	event.TaskPlanned,
	event.TaskStarted,
	event.ReviewRequested,
	event.ReviewCompleted,
	event.TestsFailed,
	event.TestsPassed,
	event.TaskCompleted,
	// Not a task event, but it carries the task it concerns (TASK-100): this is
	// how a QA report becomes reachable from the task whose acceptance decision
	// it informs.
	event.ArtifactPublished,
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

	// ArtifactPublished is the one subscribed event whose subject is not a task
	// — it is the artifact. Handled apart from everything below, which would
	// otherwise key a view by the artifact's identifier and invent a task that
	// does not exist.
	if e.Type() == event.ArtifactPublished {
		p.applyPublishedArtifact(e)
		return nil
	}

	key := viewKey(e.ProjectID(), e.SubjectID())
	v := p.views[key]
	v.ID = e.SubjectID()
	if e.ProjectID() != "" {
		v.ProjectID = e.ProjectID()
	}
	if to, ok := targetState(e); ok {
		v.State = to
	}
	if e.Type() == event.TaskCreated {
		applyCreatedData(&v, e)
	}
	if e.Type() == event.TaskRefined {
		applyRefinedData(&v, e)
	}
	if e.Type() == event.ReviewRequested {
		applyReviewRequestedData(&v, e)
	}
	v.UpdatedAt = e.OccurredAt()
	p.views[key] = v
	return nil
}

// dataCarrier is satisfied by any event that attaches event-type-specific
// data beyond the fixed platform.Event fields — Envelope via its
// WithData/Data methods, and equally an event reconstructed from the
// durable journal (internal/infrastructure/eventbus).
//
// Matched structurally, not by asserting the concrete Envelope type
// (BUGFIX-004): a journal row comes back as eventbus's own event type, so
// a concrete assertion silently dropped every WithData payload on replay —
// descriptive fields came back empty and ReviewCompleted lost its target
// state, defeating Rebuild's whole purpose. The eventbus package already
// reads this data the same structural way on its side of the boundary.
type dataCarrier interface {
	Data() map[string]string
}

// applyCreatedData populates v's immutable descriptive fields from
// TaskCreated's attached data — a no-op (fields stay zero) if e carries no
// data at all, which keeps Handle safe against any platform.Event
// implementation, not just this package's own.
func applyCreatedData(v *TaskView, e platform.Event) {
	carrier, ok := e.(dataCarrier)
	if !ok {
		return
	}
	data := carrier.Data()
	v.Title = data[dataKeyTitle]
	v.Type = data[dataKeyType]
	v.Scope = data[dataKeyScope]
	if raw := data[dataKeyAcceptanceCriteria]; raw != "" {
		var criteria []string
		if err := json.Unmarshal([]byte(raw), &criteria); err == nil {
			v.AcceptanceCriteria = criteria
		}
	}
}

// applyPublishedArtifact attaches a published artifact to the task it concerns
// (TASK-100). Called with p.mu held.
//
// Only TestReport is attached, and only when the event names a task: an artifact
// carries no task reference of its own, so anything without that data simply is
// not attributable — silently guessing which task a commit belongs to would be
// worse than leaving the field empty.
func (p *TaskProjection) applyPublishedArtifact(e platform.Event) {
	carrier, ok := e.(dataCarrier)
	if !ok {
		return
	}
	data := carrier.Data()

	taskID := data[dataKeyTaskID]
	if taskID == "" || data[dataKeyArtifactType] != artifactTypeTestReport {
		return
	}

	key := viewKey(e.ProjectID(), taskID)
	v, seen := p.views[key]
	if !seen {
		// A report for a task the projection has never seen would create a view
		// with nothing but this field. Skipped rather than half-invented: the
		// task's own events establish it, and replay delivers them first.
		return
	}
	v.QAReportID = e.SubjectID()
	v.UpdatedAt = e.OccurredAt()
	p.views[key] = v
}

// artifactTypeTestReport is the artifact type a QA run produces
// (docs/specifications/domain/artifact.md names it).
const artifactTypeTestReport = "TestReport"

// applyRefinedData applies a refinement of scope and/or acceptance criteria
// from TaskRefined's attached data (EPIC-013, TASK-096).
//
// Only fields actually present are touched: TaskRefined carries just what
// changed, so an absent key means "not refined", never "cleared". That reading
// is safe because the domain forbids clearing — SetScope("") is
// ErrMissingField — so an empty value is never a legitimate refinement. A
// refinement of scope alone must not wipe criteria a previous one recorded.
func applyRefinedData(v *TaskView, e platform.Event) {
	carrier, ok := e.(dataCarrier)
	if !ok {
		return
	}
	data := carrier.Data()

	if scope, present := data[dataKeyScope]; present && scope != "" {
		v.Scope = scope
	}
	if raw, present := data[dataKeyAcceptanceCriteria]; present && raw != "" {
		var criteria []string
		if err := json.Unmarshal([]byte(raw), &criteria); err == nil {
			v.AcceptanceCriteria = criteria
		}
	}
}

// applyReviewRequestedData records the pull request under review from
// ReviewRequested's attached data (BUGFIX-009).
//
// Only ever fills in, never clears: a review can legitimately be requested
// without a reference (a human doing it by hand), and treating that as "the
// pull request is gone" would lose a reference the platform already had and
// break the acceptance decision that depends on it.
func applyReviewRequestedData(v *TaskView, e platform.Event) {
	carrier, ok := e.(dataCarrier)
	if !ok {
		return
	}
	data := carrier.Data()
	if repo := data[dataKeyRepository]; repo != "" {
		v.Repository = repo
	}
	if prID := data[dataKeyPullRequestID]; prID != "" {
		v.PullRequestID = prID
	}
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
		if carrier, ok := e.(dataCarrier); ok {
			if to, ok := carrier.Data()[dataKeyTo]; ok {
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
