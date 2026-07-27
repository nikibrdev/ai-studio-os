package main

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/execution"
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

	// statuses is returned one per Status call; the last value repeats once
	// exhausted. Defaults to a single "succeeded" so a test that only cares
	// about dispatch does not have to describe an execution's whole life.
	statuses  []platform.ExecutionStatus
	statusErr error
	statusN   int

	artifacts    []platform.Artifact
	artifactsErr error

	finished  int
	finishErr error
}

func (b *fakeBackend) Accept(_ context.Context, t platform.ExecutorTask) error {
	b.accepted = append(b.accepted, t)
	if b.order != nil {
		*b.order = append(*b.order, "Accept")
	}
	return b.acceptErr
}

func (b *fakeBackend) Artifacts(context.Context) ([]platform.Artifact, error) {
	if b.artifactsErr != nil {
		return nil, b.artifactsErr
	}
	return b.artifacts, nil
}

func (b *fakeBackend) Status(context.Context) (platform.ExecutionStatus, error) {
	if b.statusErr != nil {
		return platform.ExecutionStatus{}, b.statusErr
	}
	if len(b.statuses) == 0 {
		return platform.ExecutionStatus{State: statusSucceeded}, nil
	}
	i := b.statusN
	if i >= len(b.statuses) {
		i = len(b.statuses) - 1
	}
	b.statusN++
	return b.statuses[i], nil
}

func (b *fakeBackend) Finish(context.Context) error {
	b.finished++
	if b.order != nil {
		*b.order = append(*b.order, "Finish")
	}
	return b.finishErr
}

// branchCall is one recorded CreateBranch invocation.
type branchCall struct{ repo, branch, base string }

// prCall is one recorded OpenPullRequest invocation.
type prCall struct{ repo, branch, title, body string }

// fakeRepos records branch and pull-request creation, and the order of calls
// relative to Accept. inmemory.RepositoryProvider records none of that, and
// these assertions need the arguments and the ordering — so this test keeps
// its own fake rather than widening a shared fixture other packages depend on.
type fakeRepos struct {
	inmemory.RepositoryProvider
	branches        []branchCall
	createBranchErr error
	// order appends a marker per observable step, so a test can assert the
	// branch exists before the agent is told to clone it.
	order *[]string

	pullRequests     []prCall
	openPRErr        error
	requestReviewErr error
}

func (r *fakeRepos) OpenPullRequest(_ context.Context, repo, branch, title, body string) (string, error) {
	if r.openPRErr != nil {
		return "", r.openPRErr
	}
	r.pullRequests = append(r.pullRequests, prCall{repo: repo, branch: branch, title: title, body: body})
	return "1", nil
}

func (r *fakeRepos) RequestReview(_ context.Context, _, _ string) error {
	return r.requestReviewErr
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
	executions application.ExecutionStore
	artifacts  application.ArtifactStore
	// order records observable steps so tests can assert their sequence.
	order []string
}

// lastExecution returns the Execution the dispatch created — found through
// the event bus, since its identifier is generated inside WorkService.
func (f *dispatchFixture) lastExecution(t *testing.T) *execution.Execution {
	t.Helper()
	var id string
	for _, e := range f.bus.Published() {
		if e.Type() == event.ExecutionQueued {
			id = e.SubjectID()
		}
	}
	if id == "" {
		t.Fatal("no ExecutionQueued event was published — no execution was created")
	}
	run, err := f.executions.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get execution %s: %v", id, err)
	}
	return run
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

	executions := inmemory.NewExecutionStore()
	artifacts := inmemory.NewArtifactStore()
	f.executions = executions
	f.artifacts = artifacts

	d := &Dispatcher{
		Projects:  projects,
		Executors: executorStore,
		Work: &application.WorkService{
			Tasks: tasks, Executors: executorStore, Executions: executions,
			Events: bus, Rules: workflow.Machine{},
		},
		Results: &application.ResultService{
			Projects: projects, Tasks: tasks, Executions: executions,
			Artifacts: artifacts, Events: bus,
		},
		Completion: &application.CompletionService{
			// Views wired exactly as apps/orchestrator/main.go wires it, so
			// CompleteTesting can resolve the pull request the platform recorded
			// (BUGFIX-009). Leaving it nil made the fixture weaker than
			// production: CompleteTesting could not succeed even when wrongly
			// called, so tests asserting "nothing merged" passed for that reason
			// rather than because the checkpoint was respected.
			Tasks: tasks, Repositories: repos, Events: bus,
			Rules: workflow.Machine{}, Views: views,
		},
		Views:       views,
		Repos:       repos,
		NewExecutor: func(shared.Role) (platform.Executor, error) { return backend, nil },
		Log:         log.New(io.Discard, "", 0),
		// Small enough that watching an execution costs no real time; the
		// defaults are minutes.
		StatusPollInterval: time.Millisecond,
		ExecutionTimeout:   2 * time.Second,
	}
	f.dispatcher = d
	f.backend = backend
	f.repos = repos
	return f
}

// ownDeveloper is the registry entry this orchestrator owns for the Developer
// role — the one dispatch looks up by derived identifier (TASK-086). Most
// tests want this; activeDeveloper with an arbitrary id builds a foreign
// record, used to prove foreign records cannot influence the choice.
func ownDeveloper(t *testing.T) *executor.Executor {
	t.Helper()
	return activeDeveloper(t, developerID)
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
	f := newDispatchFixture(t, []string{"github.com/nikibrdev/ai-studio-os"}, ownDeveloper(t))
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
	// The branch must exist before the agent is told to clone it, and the
	// sandbox must always be torn down (TASK-083).
	if len(f.order) != 3 || f.order[0] != "CreateBranch" || f.order[1] != "Accept" || f.order[2] != "Finish" {
		t.Errorf("call order = %v, want [CreateBranch Accept Finish]", f.order)
	}

	// The task must have moved through the Application Layer, not been
	// dispatched behind the state machine's back. With monitoring (TASK-083)
	// the default fake reports success, so the full path ends in Review.
	task, err := f.tasks.Get(ctx, "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	if task.State() != shared.StateReview {
		t.Errorf("task state = %v, want %v", task.State(), shared.StateReview)
	}
}

// Events no role reacts to are observed but not dispatched. TaskCreated left
// this list in EPIC-013 (it now starts a Project Manager run — see plan_test.go);
// TaskCompleted is dispatched by nobody, and never should be — the task is done.
func TestDispatch_ObservesButDoesNotDispatchUnrelatedEvents(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))

	completed := application.NewEvent(event.TaskCompleted, "task", "human", "proj-1", "TASK-001", nowForTest())

	if err := f.dispatcher.Handle(ctx, completed); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(f.backend.accepted) != 0 {
		t.Errorf("Accept called %d times for TaskCompleted, want 0", len(f.backend.accepted))
	}
	// The projection must still have been updated — observing is not
	// conditional on dispatching.
	if view, ok := f.views.Get("proj-1", "TASK-001"); !ok || view.State != shared.StateDone {
		t.Errorf("projection view = (%+v, %v), want it updated from the observed event", view, ok)
	}
}

func TestDispatch_NoActiveDeveloperExecutorLeavesTaskInReady(t *testing.T) {
	ctx := context.Background()
	// This orchestrator's own entry exists but was never activated.
	idle, _, err := executor.New(developerID, "claude-code", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, idle)
	planned := seedPlannedTask(t, f, "Заголовок")

	err = f.dispatcher.Handle(ctx, planned)
	if !errors.Is(err, ErrExecutorNotUsable) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrExecutorNotUsable)
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

// TestDispatch_ForeignActiveDeveloperRecordIsNotUsed is the regression test
// for TASK-086, reproducing exactly what the live run of TASK-085 hit: the
// registry also holds a foreign Active record with the Developer role and a
// lower id (exec-store-1, left behind by integration tests). The old
// "first Active with this role" scan picked it, so the events named that
// executor while this orchestrator's own backend did the work.
func TestDispatch_ForeignActiveDeveloperRecordIsNotUsed(t *testing.T) {
	ctx := context.Background()
	// "exec-store-1" sorts before "executor-claude-code-developer", so an
	// id-ordered scan would reach it first — the actual live-run failure.
	foreign := activeDeveloper(t, "exec-store-1")
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, foreign, ownDeveloper(t))
	planned := seedPlannedTask(t, f, "Заголовок")

	if err := f.dispatcher.Handle(ctx, planned); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// The executor named in the events must be the one whose backend ran.
	var started string
	for _, e := range f.bus.Published() {
		if e.Type() == event.ExecutionStarted {
			started = e.Actor()
		}
	}
	run := f.lastExecution(t)
	if run.ExecutorID() != developerID {
		t.Errorf("execution executor = %q, want this orchestrator's own %q (not the foreign record)", run.ExecutorID(), developerID)
	}
	if started == "" {
		t.Error("no ExecutionStarted event was published")
	}
}

// A foreign record holding a different role is likewise irrelevant — but the
// point is that its role never even gets examined: the identifier decides.
func TestDispatch_MissingOwnEntryIsReportedEvenWhenOthersExist(t *testing.T) {
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

	err = f.dispatcher.Handle(ctx, planned)
	if !errors.Is(err, ErrNoExecutorForRole) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrNoExecutorForRole)
	}
	// The error must name the entry that was expected, so an operator can act.
	if !strings.Contains(err.Error(), developerID) {
		t.Errorf("error = %q, want it to name the expected entry %q", err, developerID)
	}
}

func TestDispatch_ProjectWithoutRepositoryReportsError(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, nil, ownDeveloper(t))

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
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))

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
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
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
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
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
