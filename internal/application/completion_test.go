package application_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
)

// completionFixture gets a Task all the way to Review, ready for the
// tests in this file to drive Review/Testing/Done.
type completionFixture struct {
	completion *application.CompletionService
	tasks      application.TaskStore
	bus        *inmemory.EventBus
	repos      *inmemory.RepositoryProvider
	views      *application.TaskProjection
}

// newCompletionFixtureAtInProgress stops one step short of Review, so a test
// can call RequestReview itself — needed to exercise the pull request reference
// it carries (BUGFIX-009).
func newCompletionFixtureAtInProgress(t *testing.T) completionFixture {
	t.Helper()
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	executors := inmemory.NewExecutorStore()
	executions := inmemory.NewExecutionStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}

	views := application.NewTaskProjection()
	if err := views.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	newActiveProject(t, projects)
	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	work := &application.WorkService{Tasks: tasks, Executors: executors, Executions: executions, Events: bus, Rules: rules}
	repos := inmemory.NewRepositoryProvider()
	completion := &application.CompletionService{
		Tasks: tasks, Repositories: repos, Events: bus, Rules: rules, Views: views,
	}

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

	return completionFixture{completion: completion, tasks: tasks, bus: bus, repos: repos, views: views}
}

func newCompletionFixture(t *testing.T) completionFixture {
	t.Helper()
	f := newCompletionFixtureAtInProgress(t)
	if err := f.completion.RequestReview(context.Background(), application.RequestReviewParams{
		ProjectID: "proj-1", TaskID: "task-1",
	}); err != nil {
		t.Fatalf("RequestReview: %v", err)
	}
	return f
}

func TestRequestReview_TransitionsToReview(t *testing.T) {
	f := newCompletionFixture(t) // fixture itself calls RequestReview
	tsk, err := f.tasks.Get(context.Background(), "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateReview {
		t.Errorf("State() = %v, want %v", tsk.State(), shared.StateReview)
	}
	last := f.bus.Published()[len(f.bus.Published())-1]
	if last.Type() != event.ReviewRequested {
		t.Errorf("last event Type() = %q, want %q", last.Type(), event.ReviewRequested)
	}
}

func TestCompleteReview_ApprovedGoesToTesting(t *testing.T) {
	ctx := context.Background()
	f := newCompletionFixture(t)
	if err := f.completion.CompleteReview(ctx, "proj-1", "task-1", true, "reviewer:executor-3"); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}
	tsk, err := f.tasks.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateTesting {
		t.Errorf("State() = %v, want %v", tsk.State(), shared.StateTesting)
	}
	last := f.bus.Published()[len(f.bus.Published())-1]
	if last.Type() != event.ReviewCompleted || last.Source() != "git" {
		t.Errorf("last event = type=%q source=%q, want ReviewCompleted/git", last.Type(), last.Source())
	}
}

func TestCompleteReview_ChangesRequestedReturnsToInProgress(t *testing.T) {
	ctx := context.Background()
	f := newCompletionFixture(t)
	if err := f.completion.CompleteReview(ctx, "proj-1", "task-1", false, ""); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}
	tsk, err := f.tasks.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateInProgress {
		t.Errorf("State() = %v, want %v", tsk.State(), shared.StateInProgress)
	}
}

func TestCompleteTesting_Failed_ReturnsToInProgress(t *testing.T) {
	ctx := context.Background()
	f := newCompletionFixture(t)
	if err := f.completion.CompleteReview(ctx, "proj-1", "task-1", true, ""); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}

	if err := f.completion.CompleteTesting(ctx, application.CompleteTestingParams{ProjectID: "proj-1", TaskID: "task-1", Passed: false}); err != nil {
		t.Fatalf("CompleteTesting: %v", err)
	}
	tsk, err := f.tasks.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateInProgress {
		t.Errorf("State() = %v, want %v", tsk.State(), shared.StateInProgress)
	}
	if len(f.repos.MergeCalls) != 0 {
		t.Errorf("MergeCalls = %v, want none on test failure", f.repos.MergeCalls)
	}
	last := f.bus.Published()[len(f.bus.Published())-1]
	if last.Type() != event.TestsFailed || last.Source() != "execution" {
		t.Errorf("last event = type=%q source=%q, want TestsFailed/execution", last.Type(), last.Source())
	}
}

// TestCompleteTesting_Passed_EventOrderMatchesADR008 is the ADR-008
// decision expressed as a test, not just a doc comment: TestsPassed,
// then MergeCompleted, then TaskCompleted — in exactly that order.
func TestCompleteTesting_Passed_EventOrderMatchesADR008(t *testing.T) {
	ctx := context.Background()
	f := newCompletionFixture(t)
	if err := f.completion.CompleteReview(ctx, "proj-1", "task-1", true, ""); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}
	before := len(f.bus.Published())

	if err := f.completion.CompleteTesting(ctx, application.CompleteTestingParams{
		ProjectID: "proj-1", TaskID: "task-1", Passed: true, Repository: "org/repo", PullRequestID: "pr-1", Actor: "qa:executor-4",
	}); err != nil {
		t.Fatalf("CompleteTesting: %v", err)
	}

	tsk, err := f.tasks.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateDone {
		t.Errorf("State() = %v, want %v", tsk.State(), shared.StateDone)
	}

	newEvents := f.bus.Published()[before:]
	wantOrder := []struct{ typ, source string }{
		{event.TestsPassed, "execution"},
		{event.MergeCompleted, "git"},
		{event.TaskCompleted, "task"},
	}
	if len(newEvents) != len(wantOrder) {
		t.Fatalf("published %d events, want %d (%v)", len(newEvents), len(wantOrder), wantOrder)
	}
	for i, want := range wantOrder {
		if newEvents[i].Type() != want.typ || newEvents[i].Source() != want.source {
			t.Errorf("event[%d] = type=%q source=%q, want type=%q source=%q", i, newEvents[i].Type(), newEvents[i].Source(), want.typ, want.source)
		}
	}
	if len(f.repos.MergeCalls) != 1 || f.repos.MergeCalls[0] != "org/repo/pr-1" {
		t.Errorf("MergeCalls = %v, want exactly one call for org/repo/pr-1", f.repos.MergeCalls)
	}
}

// TestCompleteTesting_MergeFailure_BlocksDone is the other half of ADR-008
// as code: if the merge itself fails, Done is unreachable — the merge is
// a guard, not a side effect that happens to usually succeed.
func TestCompleteTesting_MergeFailure_BlocksDone(t *testing.T) {
	ctx := context.Background()
	f := newCompletionFixture(t)
	if err := f.completion.CompleteReview(ctx, "proj-1", "task-1", true, ""); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}
	f.repos.MergeErr = errors.New("merge conflict")

	if err := f.completion.CompleteTesting(ctx, application.CompleteTestingParams{
		ProjectID: "proj-1", TaskID: "task-1", Passed: true, Repository: "org/repo", PullRequestID: "pr-1",
	}); err == nil {
		t.Fatal("CompleteTesting() error = nil, want the merge failure propagated")
	}

	tsk, err := f.tasks.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateTesting {
		t.Errorf("State() = %v, want still %v (Done is unreachable without a successful merge)", tsk.State(), shared.StateTesting)
	}
	for _, e := range f.bus.Published() {
		if e.Type() == event.TaskCompleted {
			t.Error("TaskCompleted was published despite the merge failing")
		}
	}
}

func TestRequestReview_TaskNotFound(t *testing.T) {
	completion := &application.CompletionService{
		Tasks:        inmemory.NewTaskStore(),
		Repositories: inmemory.NewRepositoryProvider(),
		Events:       inmemory.NewEventBus(),
		Rules:        workflow.Machine{},
	}
	params := application.RequestReviewParams{ProjectID: "proj-1", TaskID: "missing"}
	if err := completion.RequestReview(context.Background(), params); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("RequestReview() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestCompleteTesting_TaskNotFound(t *testing.T) {
	completion := &application.CompletionService{
		Tasks:        inmemory.NewTaskStore(),
		Repositories: inmemory.NewRepositoryProvider(),
		Events:       inmemory.NewEventBus(),
		Rules:        workflow.Machine{},
	}
	if err := completion.CompleteTesting(context.Background(), application.CompleteTestingParams{ProjectID: "proj-1", TaskID: "missing", Passed: true}); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("CompleteTesting() error = %v, want %v", err, application.ErrNotFound)
	}
}

// TestRequestReview_CarriesPullRequestReference is the regression test for
// BUGFIX-009. docs/architecture/events.md always required ReviewRequested to
// carry a reference to the pull request ("Данные: идентификатор задачи, ссылка
// на PR, автор изменений") and nothing did — so the platform opened a pull
// request and then forgot it, leaving the acceptance decision impossible.
func TestRequestReview_CarriesPullRequestReference(t *testing.T) {
	ctx := context.Background()
	f := newCompletionFixtureAtInProgress(t)

	err := f.completion.RequestReview(ctx, application.RequestReviewParams{
		ProjectID: "proj-1", TaskID: "task-1",
		Repository: "github.com/org/repo", PullRequestID: "42",
	})
	if err != nil {
		t.Fatalf("RequestReview: %v", err)
	}

	view, ok := f.views.Get("proj-1", "task-1")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if view.Repository != "github.com/org/repo" || view.PullRequestID != "42" {
		t.Errorf("view = {Repository:%q PullRequestID:%q}, want the reference RequestReview was given",
			view.Repository, view.PullRequestID)
	}
}

// The whole point of carrying the reference: the acceptance decision must work
// without the caller knowing it. A UI has no way to learn a pull request id.
func TestCompleteTesting_MergesWithoutCallerSupplyingPullRequest(t *testing.T) {
	ctx := context.Background()
	f := newCompletionFixtureAtInProgress(t)

	if err := f.completion.RequestReview(ctx, application.RequestReviewParams{
		ProjectID: "proj-1", TaskID: "task-1",
		Repository: "github.com/org/repo", PullRequestID: "42",
	}); err != nil {
		t.Fatalf("RequestReview: %v", err)
	}
	if err := f.completion.CompleteReview(ctx, "proj-1", "task-1", true, ""); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}

	// No Repository, no PullRequestID — exactly what a UI can supply.
	if err := f.completion.CompleteTesting(ctx, application.CompleteTestingParams{
		ProjectID: "proj-1", TaskID: "task-1", Passed: true,
	}); err != nil {
		t.Fatalf("CompleteTesting: %v", err)
	}

	if got := f.repos.MergeCalls; len(got) != 1 || got[0] != "github.com/org/repo/42" {
		t.Errorf("MergeCalls = %v, want one merge of github.com/org/repo/42", got)
	}
	tsk, err := f.tasks.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateDone {
		t.Errorf("State() = %v, want %v", tsk.State(), shared.StateDone)
	}
}

// Without a known pull request the positive decision must fail before
// announcing anything: publishing TestsPassed and then stopping would claim a
// passing run for a task that stays in Testing.
func TestCompleteTesting_UnknownPullRequestFailsBeforePublishing(t *testing.T) {
	ctx := context.Background()
	// Review requested without a reference — a human driving the task by hand.
	f := newCompletionFixture(t)
	if err := f.completion.CompleteReview(ctx, "proj-1", "task-1", true, ""); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}
	before := len(f.bus.Published())

	err := f.completion.CompleteTesting(ctx, application.CompleteTestingParams{
		ProjectID: "proj-1", TaskID: "task-1", Passed: true,
	})
	if !errors.Is(err, application.ErrPullRequestUnknown) {
		t.Fatalf("CompleteTesting() error = %v, want %v", err, application.ErrPullRequestUnknown)
	}
	if got := len(f.bus.Published()) - before; got != 0 {
		t.Errorf("published %d events, want none — nothing may be announced before the reference resolves", got)
	}
	if len(f.repos.MergeCalls) != 0 {
		t.Errorf("MergeCalls = %v, want none", f.repos.MergeCalls)
	}
}

// A later review request without a reference must not erase one already known:
// re-requesting review by hand would otherwise lose the pull request.
func TestRequestReview_WithoutReferenceKeepsKnownPullRequest(t *testing.T) {
	ctx := context.Background()
	f := newCompletionFixtureAtInProgress(t)

	if err := f.completion.RequestReview(ctx, application.RequestReviewParams{
		ProjectID: "proj-1", TaskID: "task-1",
		Repository: "github.com/org/repo", PullRequestID: "42",
	}); err != nil {
		t.Fatalf("RequestReview: %v", err)
	}
	if err := f.completion.CompleteReview(ctx, "proj-1", "task-1", false, ""); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}
	if err := f.completion.RequestReview(ctx, application.RequestReviewParams{
		ProjectID: "proj-1", TaskID: "task-1",
	}); err != nil {
		t.Fatalf("second RequestReview: %v", err)
	}

	view, _ := f.views.Get("proj-1", "task-1")
	if view.PullRequestID != "42" {
		t.Errorf("PullRequestID = %q, want it kept at 42", view.PullRequestID)
	}
}
