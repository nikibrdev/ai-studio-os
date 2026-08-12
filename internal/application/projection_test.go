package application_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/platform"
)

func TestTaskProjection_IncrementalUpdatesTrackState(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}
	newActiveProject(t, projects)

	proj := application.NewTaskProjection()
	if err := proj.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	view, ok := proj.Get("proj-1", "task-1")
	if !ok || view.State != shared.StateBacklog || view.ProjectID != "proj-1" {
		t.Fatalf("Get() after CreateTask = (%+v, %v), want Backlog/proj-1", view, ok)
	}

	if err := planning.PlanTask(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
	view, _ = proj.Get("proj-1", "task-1")
	if view.State != shared.StateReady {
		t.Errorf("State() after PlanTask = %v, want %v", view.State, shared.StateReady)
	}
}

func TestTaskProjection_ReviewCompletedDisambiguatesOutcome(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	executors := inmemory.NewExecutorStore()
	executions := inmemory.NewExecutionStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}
	newActiveProject(t, projects)

	proj := application.NewTaskProjection()
	if err := proj.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	work := &application.WorkService{Tasks: tasks, Executors: executors, Executions: executions, Events: bus, Rules: rules}
	completion := &application.CompletionService{Tasks: tasks, Repositories: inmemory.NewRepositoryProvider(), Events: bus, Rules: rules}

	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := planning.PlanTask(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
	saveExecutor(t, executors, true, shared.RoleDeveloper)
	if _, err := work.StartTask(ctx, application.StartTaskParams{ProjectID: "proj-1", TaskID: "task-1", ExecutorID: "executor-1"}); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := completion.RequestReview(ctx, application.RequestReviewParams{ProjectID: "proj-1", TaskID: "task-1", Actor: ""}); err != nil {
		t.Fatalf("RequestReview: %v", err)
	}

	// Changes requested: the projection must land back in In Progress,
	// not guess Testing just because a ReviewCompleted event fired.
	if err := completion.CompleteReview(ctx, "proj-1", "task-1", false, ""); err != nil {
		t.Fatalf("CompleteReview(false): %v", err)
	}
	view, _ := proj.Get("proj-1", "task-1")
	if view.State != shared.StateInProgress {
		t.Fatalf("State() after changes-requested = %v, want %v", view.State, shared.StateInProgress)
	}

	if err := completion.RequestReview(ctx, application.RequestReviewParams{ProjectID: "proj-1", TaskID: "task-1", Actor: ""}); err != nil {
		t.Fatalf("second RequestReview: %v", err)
	}
	if err := completion.CompleteReview(ctx, "proj-1", "task-1", true, ""); err != nil {
		t.Fatalf("CompleteReview(true): %v", err)
	}
	view, _ = proj.Get("proj-1", "task-1")
	if view.State != shared.StateTesting {
		t.Fatalf("State() after approval = %v, want %v", view.State, shared.StateTesting)
	}
}

func TestTaskProjection_RebuildFromJournalMatchesIncremental(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}
	newActiveProject(t, projects)

	live := application.NewTaskProjection()
	if err := live.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := planning.PlanTask(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}

	rebuilt := application.NewTaskProjection()
	if err := rebuilt.Rebuild(ctx, bus.Published()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	liveView, liveOK := live.Get("proj-1", "task-1")
	rebuiltView, rebuiltOK := rebuilt.Get("proj-1", "task-1")
	if liveOK != rebuiltOK || !reflect.DeepEqual(liveView, rebuiltView) {
		t.Errorf("rebuilt = (%+v, %v), want it to match live = (%+v, %v)", rebuiltView, rebuiltOK, liveView, liveOK)
	}
}

func TestTaskProjection_GetUnknownTask(t *testing.T) {
	proj := application.NewTaskProjection()
	if _, ok := proj.Get("proj-1", "missing"); ok {
		t.Error("Get() for unseen task ok = true, want false")
	}
}

// TestTaskProjection_ListByProject_IsolatesProjects proves the same
// property BUGFIX-003 established for TaskStore also holds for the
// projection's list operation: a project's task list never includes
// another project's tasks, even when both use the fixture's default
// "task-1" (mirrors task_planning_test.go's newActiveProject, which is
// hardcoded to "proj-1" — a second project is seeded directly here).
func TestTaskProjection_ListByProject_IsolatesProjects(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}
	newActiveProject(t, projects) // proj-1

	projB, _, err := project.New("proj-2", "B")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if _, _, err := projB.ConnectRepository("github.com/org/repo"); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}
	if _, err := projB.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := projects.Save(ctx, projB); err != nil {
		t.Fatalf("Save projB: %v", err)
	}

	proj := application.NewTaskProjection()
	if err := proj.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-1", Title: "Task in proj-1", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask proj-1: %v", err)
	}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-2", Title: "Task in proj-2", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask proj-2: %v", err)
	}

	viewsA := proj.ListByProject("proj-1")
	viewsB := proj.ListByProject("proj-2")
	if len(viewsA) != 1 || viewsA[0].ProjectID != "proj-1" {
		t.Fatalf("ListByProject(proj-1) = %+v, want exactly one view for proj-1", viewsA)
	}
	if len(viewsB) != 1 || viewsB[0].ProjectID != "proj-2" {
		t.Fatalf("ListByProject(proj-2) = %+v, want exactly one view for proj-2", viewsB)
	}
}

func TestTaskProjection_ListByProject_EmptyIsNotError(t *testing.T) {
	proj := application.NewTaskProjection()
	if views := proj.ListByProject("proj-1"); len(views) != 0 {
		t.Errorf("ListByProject() = %v, want empty", views)
	}
}

// TestTaskProjection_CapturesDescriptiveFieldsFromCreation proves TASK-076's
// TaskView additions: title/type/scope/acceptanceCriteria are set once
// from TaskCreated and survive later transitions untouched.
func TestTaskProjection_CapturesDescriptiveFieldsFromCreation(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	newActiveProject(t, projects)

	proj := application.NewTaskProjection()
	if err := proj.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: workflow.Machine{}}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "Заголовок", Type: "feature",
		Scope: "Описание", AcceptanceCriteria: []string{"критерий"},
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	view, ok := proj.Get("proj-1", "task-1")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if view.Title != "Заголовок" || view.Type != "feature" || view.Scope != "Описание" {
		t.Errorf("view = %+v, want Title/Type/Scope from CreateTask", view)
	}
	if len(view.AcceptanceCriteria) != 1 || view.AcceptanceCriteria[0] != "критерий" {
		t.Errorf("AcceptanceCriteria = %v, want [критерий]", view.AcceptanceCriteria)
	}

	if err := planning.PlanTask(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
	view, ok = proj.Get("proj-1", "task-1")
	if !ok {
		t.Fatal("Get() after PlanTask ok = false, want true")
	}
	if view.Title != "Заголовок" {
		t.Errorf("Title after PlanTask = %q, want unchanged \"Заголовок\"", view.Title)
	}
	if view.State != shared.StateReady {
		t.Errorf("State() = %v, want %v", view.State, shared.StateReady)
	}
}

// journalLikeEvent mimics an event reconstructed from the durable journal
// (internal/infrastructure/eventbus): it carries WithData's payload through
// a Data() method but is NOT an application.Envelope.
//
// BUGFIX-004: applyCreatedData and targetState used to assert the concrete
// Envelope type, so replaying a real journal silently dropped every
// payload — descriptive fields came back empty and ReviewCompleted lost its
// target state. This fake is deliberately a foreign type: it is the only
// thing that distinguishes the bug from the fix, since bus.Published()
// returns genuine Envelopes on which the old assertion worked.
type journalLikeEvent struct {
	typ, projectID, subjectID string
	data                      map[string]string
}

func (e journalLikeEvent) ID() string              { return "evt-" + e.typ }
func (e journalLikeEvent) Type() string            { return e.typ }
func (e journalLikeEvent) SchemaVersion() int      { return 1 }
func (e journalLikeEvent) OccurredAt() time.Time   { return time.Now() }
func (e journalLikeEvent) Source() string          { return "task" }
func (e journalLikeEvent) Actor() string           { return "" }
func (e journalLikeEvent) ProjectID() string       { return e.projectID }
func (e journalLikeEvent) SubjectID() string       { return e.subjectID }
func (e journalLikeEvent) Data() map[string]string { return e.data }

// TestTaskProjection_RebuildIgnoresEventsAboutOtherSubjects is the regression
// test for BUGFIX-013. Subscribe registers Handle per event type, so the live
// path never delivers anything else — but apps/api replays the whole durable
// journal through Rebuild at startup (BUGFIX-010), and that journal holds
// events about projects, executions, artifacts and executors too. Each of them
// keyed a view by its own subject and invented a task with an empty state,
// which then appeared in GET /projects/{id}/tasks.
//
// The journal below is deliberately "dirty" in exactly the way a real one is:
// on the live database every project showed two such phantoms.
func TestTaskProjection_RebuildIgnoresEventsAboutOtherSubjects(t *testing.T) {
	ctx := context.Background()
	proj := application.NewTaskProjection()

	journal := []platform.Event{
		journalLikeEvent{typ: event.ProjectCreated, projectID: "proj-1", subjectID: "proj-1"},
		journalLikeEvent{typ: event.RepositoryConnected, projectID: "proj-1", subjectID: "proj-1"},
		journalLikeEvent{typ: event.ProjectActivated, projectID: "proj-1", subjectID: "proj-1"},
		journalLikeEvent{
			typ: event.TaskCreated, projectID: "proj-1", subjectID: "TASK-001",
			data: map[string]string{"title": "Настоящая задача", "type": "feature"},
		},
		journalLikeEvent{typ: event.ExecutionQueued, projectID: "proj-1", subjectID: "exec-abc"},
		journalLikeEvent{typ: event.ExecutionStarted, projectID: "proj-1", subjectID: "exec-abc"},
		journalLikeEvent{typ: event.ArtifactCreated, projectID: "proj-1", subjectID: "art-xyz"},
		journalLikeEvent{typ: event.ExecutorRegistered, projectID: "proj-1", subjectID: "executor-1"},
		journalLikeEvent{typ: event.MergeCompleted, projectID: "proj-1", subjectID: "TASK-001"},
	}

	if err := proj.Rebuild(ctx, journal); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	views := proj.ListByProject("proj-1")
	if len(views) != 1 {
		ids := make([]string, len(views))
		for i, v := range views {
			ids[i] = v.ID
		}
		t.Fatalf("ListByProject() returned %d views %v, want exactly 1 (the only task in the journal)", len(views), ids)
	}
	if views[0].ID != "TASK-001" {
		t.Errorf("view.ID = %q, want %q", views[0].ID, "TASK-001")
	}
	if views[0].Title != "Настоящая задача" {
		t.Errorf("view.Title = %q, want the real task's title", views[0].Title)
	}
}

func TestTaskProjection_RebuildFromNonEnvelopeEventsKeepsAttachedData(t *testing.T) {
	ctx := context.Background()
	proj := application.NewTaskProjection()

	journal := []platform.Event{
		journalLikeEvent{
			typ: event.TaskCreated, projectID: "proj-1", subjectID: "task-1",
			data: map[string]string{
				"title": "Заголовок", "type": "feature", "scope": "Описание",
				"acceptanceCriteria": `["критерий"]`,
			},
		},
		journalLikeEvent{typ: event.TaskPlanned, projectID: "proj-1", subjectID: "task-1"},
		journalLikeEvent{typ: event.TaskStarted, projectID: "proj-1", subjectID: "task-1"},
		journalLikeEvent{typ: event.ReviewRequested, projectID: "proj-1", subjectID: "task-1"},
		journalLikeEvent{
			typ: event.ReviewCompleted, projectID: "proj-1", subjectID: "task-1",
			data: map[string]string{"to": string(shared.StateTesting)},
		},
	}

	if err := proj.Rebuild(ctx, journal); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	view, ok := proj.Get("proj-1", "task-1")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}

	// The descriptive fields the orchestrator needs to build a
	// platform.ExecutorTask (EPIC-010, TASK-082).
	if view.Title != "Заголовок" || view.Type != "feature" || view.Scope != "Описание" {
		t.Errorf("view = %+v, want Title/Type/Scope recovered from journal data", view)
	}
	if len(view.AcceptanceCriteria) != 1 || view.AcceptanceCriteria[0] != "критерий" {
		t.Errorf("AcceptanceCriteria = %v, want [критерий]", view.AcceptanceCriteria)
	}

	// ReviewCompleted is ambiguous without its attached target state.
	if view.State != shared.StateTesting {
		t.Errorf("State = %v, want %v — ReviewCompleted's target must survive replay", view.State, shared.StateTesting)
	}
}
