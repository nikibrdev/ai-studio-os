package main

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/platform"
)

// fakeBackend records what it was handed instead of starting a real Docker
// sandbox.
type fakeBackend struct {
	accepted  []platform.ExecutorTask
	acceptErr error
	// order appends a marker per observable step, shared with fakeRepos so a
	// test can assert the branch is created before the agent clones it.
	order *[]string
}

func (b *fakeBackend) Accept(_ context.Context, t platform.ExecutorTask) error {
	b.accepted = append(b.accepted, t)
	if b.order != nil {
		*b.order = append(*b.order, "Accept")
	}
	return b.acceptErr
}
func (b *fakeBackend) Artifacts(context.Context) ([]platform.Artifact, error) { return nil, nil }
func (b *fakeBackend) Status(context.Context) (platform.ExecutionStatus, error) {
	return platform.ExecutionStatus{}, nil
}
func (b *fakeBackend) Finish(context.Context) error { return nil }

// branchCall is one recorded CreateBranch invocation.
type branchCall struct{ repo, branch, base string }

// fakeRepos records branch creation and the order of calls relative to
// Accept. inmemory.RepositoryProvider's CreateBranch is a no-op that records
// nothing, and these assertions need the arguments and the ordering — so
// this test keeps its own fake rather than widening a shared fixture other
// packages depend on.
type fakeRepos struct {
	inmemory.RepositoryProvider
	branches        []branchCall
	createBranchErr error
	// order appends a marker per observable step, so a test can assert the
	// branch exists before the agent is told to clone it.
	order *[]string
}

func (r *fakeRepos) CreateBranch(_ context.Context, repo, branch, base string) error {
	if r.createBranchErr != nil {
		return r.createBranchErr
	}
	r.branches = append(r.branches, branchCall{repo: repo, branch: branch, base: base})
	if r.order != nil {
		*r.order = append(*r.order, "CreateBranch")
	}
	return nil
}

func (r *fakeRepos) created(repo, branch string) bool {
	for _, c := range r.branches {
		if c.repo == repo && c.branch == branch {
			return true
		}
	}
	return false
}

// dispatchFixture is a Dispatcher wired entirely onto in-memory fakes, plus
// handles on the pieces tests assert against.
type dispatchFixture struct {
	dispatcher *Dispatcher
	backend    *fakeBackend
	repos      *fakeRepos
	tasks      application.TaskStore
	views      *application.TaskProjection
	bus        *inmemory.EventBus
	// order records observable steps so tests can assert their sequence.
	order []string
}

// nowForTest is a fixed timestamp: these tests assert on identity and
// content, never on time, so a constant keeps events deterministic.
func nowForTest() time.Time {
	return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
}

func newDispatchFixture(t *testing.T, repositories []string, executors ...*executor.Executor) *dispatchFixture {
	t.Helper()
	ctx := context.Background()

	projects := inmemory.NewProjectStore()
	proj, _, err := project.New("proj-1", "AI Studio OS")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	for _, r := range repositories {
		if _, _, err := proj.ConnectRepository(r); err != nil {
			t.Fatalf("ConnectRepository: %v", err)
		}
	}
	if len(repositories) > 0 {
		if _, err := proj.Activate(); err != nil {
			t.Fatalf("Activate: %v", err)
		}
	}
	if err := projects.Save(ctx, proj); err != nil {
		t.Fatalf("Save project: %v", err)
	}

	executorStore := inmemory.NewExecutorStore()
	for _, e := range executors {
		if err := executorStore.Save(ctx, e); err != nil {
			t.Fatalf("Save executor: %v", err)
		}
	}

	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	views := application.NewTaskProjection()

	f := &dispatchFixture{tasks: tasks, views: views, bus: bus}
	backend := &fakeBackend{order: &f.order}
	repos := &fakeRepos{order: &f.order}

	d := &Dispatcher{
		Projects:  projects,
		Executors: executorStore,
		Work: &application.WorkService{
			Tasks: tasks, Executors: executorStore, Executions: inmemory.NewExecutionStore(),
			Events: bus, Rules: workflow.Machine{},
		},
		Views:       views,
		Repos:       repos,
		NewExecutor: func() (platform.Executor, error) { return backend, nil },
		Log:         log.New(io.Discard, "", 0),
	}
	f.dispatcher = d
	f.backend = backend
	f.repos = repos
	return f
}

func activeDeveloper(t *testing.T, id string) *executor.Executor {
	t.Helper()
	e, _, err := executor.New(id, "claude-code", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	return e
}

// seedPlannedTask puts a task into the fixture in the state a real
// TaskPlanned event would find it: Ready in storage, with its descriptive
// fields in the projection. It returns the TaskPlanned event to dispatch.
func seedPlannedTask(t *testing.T, f *dispatchFixture, title string) platform.Event {
	t.Helper()
	ctx := context.Background()

	planning := &application.TaskPlanningService{
		Projects: mustProjects(f), Tasks: f.tasks, Events: f.bus, Rules: workflow.Machine{},
	}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "TASK-001", ProjectID: "proj-1", Title: title, Type: "feature",
		Scope: "Что нужно сделать", AcceptanceCriteria: []string{"критерий один"},
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := planning.PlanTask(ctx, "proj-1", "TASK-001", "pm"); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}

	// Replay what the journal would have carried, through the same path the
	// orchestrator uses: everything before the last event is history.
	published := f.bus.Published()
	for _, e := range published[:len(published)-1] {
		if err := f.dispatcher.Observe(ctx, e); err != nil {
			t.Fatalf("Observe: %v", err)
		}
	}
	return published[len(published)-1]
}

func mustProjects(f *dispatchFixture) application.ProjectStore { return f.dispatcher.Projects }

func TestDispatch_TaskPlannedStartsWorkAndAcceptsTask(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/nikibrdev/ai-studio-os"}, activeDeveloper(t, "exec-dev"))
	planned := seedPlannedTask(t, f, "Заголовок задачи")

	if err := f.dispatcher.Handle(ctx, planned); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(f.backend.accepted) != 1 {
		t.Fatalf("Accept called %d times, want 1", len(f.backend.accepted))
	}
	got := f.backend.accepted[0]
	if got.TaskID != "TASK-001" || got.ProjectID != "proj-1" {
		t.Errorf("ExecutorTask identity = %s/%s, want proj-1/TASK-001", got.ProjectID, got.TaskID)
	}
	if got.Role != string(shared.RoleDeveloper) {
		t.Errorf("Role = %q, want %q", got.Role, shared.RoleDeveloper)
	}
	if got.Title != "Заголовок задачи" || got.Type != "feature" || got.Scope != "Что нужно сделать" {
		t.Errorf("planning content = {Title:%q Type:%q Scope:%q}, want the values from CreateTask", got.Title, got.Type, got.Scope)
	}
	if len(got.AcceptanceCriteria) != 1 || got.AcceptanceCriteria[0] != "критерий один" {
		t.Errorf("AcceptanceCriteria = %v, want [критерий один]", got.AcceptanceCriteria)
	}
	if got.Repository != "github.com/nikibrdev/ai-studio-os" {
		t.Errorf("Repository = %q, want the project's connected repository", got.Repository)
	}

	if !f.repos.created("github.com/nikibrdev/ai-studio-os", got.Branch) {
		t.Errorf("branch %q was never created", got.Branch)
	}
	if base := f.repos.branches[0].base; base != baseBranch {
		t.Errorf("branch cut from %q, want %q", base, baseBranch)
	}
	// The branch must exist before the agent is told to clone it.
	if len(f.order) != 2 || f.order[0] != "CreateBranch" || f.order[1] != "Accept" {
		t.Errorf("call order = %v, want [CreateBranch Accept]", f.order)
	}

	// The task must have moved through the Application Layer, not been
	// dispatched behind the state machine's back.
	task, err := f.tasks.Get(ctx, "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	if task.State() != shared.StateInProgress {
		t.Errorf("task state = %v, want %v", task.State(), shared.StateInProgress)
	}
}

func TestDispatch_IgnoresEventsOtherThanTaskPlanned(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, activeDeveloper(t, "exec-dev"))

	created := application.NewEvent(event.TaskCreated, "task", "human", "proj-1", "TASK-001", nowForTest()).
		WithData(map[string]string{"title": "Заголовок", "type": "feature"})

	if err := f.dispatcher.Handle(ctx, created); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(f.backend.accepted) != 0 {
		t.Errorf("Accept called %d times for TaskCreated, want 0", len(f.backend.accepted))
	}
	// The projection must still have been updated — observing is not
	// conditional on dispatching.
	if view, ok := f.views.Get("proj-1", "TASK-001"); !ok || view.Title != "Заголовок" {
		t.Errorf("projection view = (%+v, %v), want it updated from the observed event", view, ok)
	}
}

func TestDispatch_NoActiveDeveloperExecutorLeavesTaskInReady(t *testing.T) {
	ctx := context.Background()
	// Registered but never activated: not assignable.
	idle, _, err := executor.New("exec-idle", "claude-code", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, idle)
	planned := seedPlannedTask(t, f, "Заголовок")

	err = f.dispatcher.Handle(ctx, planned)
	if !errors.Is(err, ErrNoDeveloperExecutor) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrNoDeveloperExecutor)
	}
	if len(f.backend.accepted) != 0 {
		t.Errorf("Accept called despite no available executor")
	}

	task, err := f.tasks.Get(ctx, "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	if task.State() != shared.StateReady {
		t.Errorf("task state = %v, want it left at %v so a human can still start it", task.State(), shared.StateReady)
	}
}

// An executor holding only a non-Developer role must not be picked.
func TestDispatch_SkipsExecutorWithoutDeveloperRole(t *testing.T) {
	ctx := context.Background()
	qa, _, err := executor.New("exec-qa", "claude-code", []shared.Role{shared.RoleQA})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := qa.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, qa)
	planned := seedPlannedTask(t, f, "Заголовок")

	if err := f.dispatcher.Handle(ctx, planned); !errors.Is(err, ErrNoDeveloperExecutor) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrNoDeveloperExecutor)
	}
}

func TestDispatch_ProjectWithoutRepositoryReportsError(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, nil, activeDeveloper(t, "exec-dev"))

	// CreateTask requires an Active project, which requires a repository —
	// so drive the projection directly to reach the dispatch path.
	created := application.NewEvent(event.TaskCreated, "task", "human", "proj-1", "TASK-001", nowForTest()).
		WithData(map[string]string{"title": "Заголовок", "type": "feature"})
	if err := f.dispatcher.Observe(ctx, created); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	planned := application.NewEvent(event.TaskPlanned, "task", "pm", "proj-1", "TASK-001", nowForTest())

	if err := f.dispatcher.Handle(ctx, planned); !errors.Is(err, ErrNoRepository) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrNoRepository)
	}
	if len(f.backend.accepted) != 0 {
		t.Error("Accept called despite the project having no repository")
	}
}

func TestDispatch_UnknownTaskReportsError(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, activeDeveloper(t, "exec-dev"))

	// TaskPlanned for a task the projection never saw created.
	planned := application.NewEvent(event.TaskPlanned, "task", "pm", "proj-1", "TASK-404", nowForTest())

	err := f.dispatcher.Handle(ctx, planned)
	if err == nil {
		t.Fatal("Handle() error = nil, want an error for a task absent from the projection")
	}
	if len(f.backend.accepted) != 0 {
		t.Error("Accept called for a task with no known planning content")
	}
}

func TestDispatch_AcceptErrorPropagates(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, activeDeveloper(t, "exec-dev"))
	wantErr := errors.New("docker daemon unreachable")
	f.backend.acceptErr = wantErr
	planned := seedPlannedTask(t, f, "Заголовок")

	if err := f.dispatcher.Handle(ctx, planned); !errors.Is(err, wantErr) {
		t.Fatalf("Handle() error = %v, want wrapping %v", err, wantErr)
	}
}

func TestBranchName(t *testing.T) {
	tests := []struct {
		name, taskID, title, want string
	}{
		{
			name: "russian title yields no ascii slug",
			// Every title in this project is Russian; a transliteration
			// table would be invented data, so the id alone is used.
			taskID: "TASK-001", title: "Заголовок задачи", want: "feature/TASK-001",
		},
		{
			name:   "ascii title becomes a slug",
			taskID: "TASK-002", title: "Add Orchestrator dispatch", want: "feature/TASK-002-add-orchestrator-dispatch",
		},
		{
			name:   "punctuation collapses to single hyphens",
			taskID: "TASK-003", title: "Fix  API:  errors!!", want: "feature/TASK-003-fix-api-errors",
		},
		{
			name:   "empty title",
			taskID: "TASK-004", title: "", want: "feature/TASK-004",
		},
		{
			name:   "mixed script keeps the ascii part",
			taskID: "TASK-005", title: "Оркестратор dispatch", want: "feature/TASK-005-dispatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := branchName(tt.taskID, tt.title); got != tt.want {
				t.Errorf("branchName(%q, %q) = %q, want %q", tt.taskID, tt.title, got, tt.want)
			}
		})
	}
}

func TestBranchName_LongTitleIsTruncated(t *testing.T) {
	got := branchName("TASK-006", "This is an extremely long english title that would make an unwieldy git branch name")
	if len(got) > len("feature/TASK-006-")+maxSlugLength {
		t.Errorf("branchName() = %q (len %d), want the slug capped at %d", got, len(got), maxSlugLength)
	}
	if got == "feature/TASK-006" {
		t.Errorf("branchName() = %q, want a truncated slug rather than none", got)
	}
}

// TestDispatch_BranchFailureLeavesTaskInReady pins the ordering fixed after
// live-testing TASK-082: the branch is cut before the Task is moved, so a
// CreateBranch failure — misconfigured repository, bad token, GitHub down —
// leaves the task retryable in Ready instead of stranded In Progress with a
// running Execution and nowhere to commit.
func TestDispatch_BranchFailureLeavesTaskInReady(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, activeDeveloper(t, "exec-dev"))
	wantErr := errors.New("github: 404 Not Found")
	f.repos.createBranchErr = wantErr
	planned := seedPlannedTask(t, f, "Заголовок")

	if err := f.dispatcher.Handle(ctx, planned); !errors.Is(err, wantErr) {
		t.Fatalf("Handle() error = %v, want wrapping %v", err, wantErr)
	}

	task, err := f.tasks.Get(ctx, "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	if task.State() != shared.StateReady {
		t.Errorf("task state = %v, want %v — a branch failure must not strand the task",
			task.State(), shared.StateReady)
	}
	if len(f.backend.accepted) != 0 {
		t.Error("Accept called despite the branch not existing")
	}
}
