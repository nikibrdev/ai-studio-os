package application_test

import (
	"context"
	"reflect"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
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
	if err := completion.RequestReview(ctx, "proj-1", "task-1", ""); err != nil {
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

	if err := completion.RequestReview(ctx, "proj-1", "task-1", ""); err != nil {
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
