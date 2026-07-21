//go:build integration

package wiring

import (
	"context"
	"os"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/infrastructure/eventbus"
	"ai-studio-os/internal/platform"
)

// TestGoldenPath_Infrastructure is TASK-045's TestGoldenPath_Application
// run on real infrastructure instead of in-memory fakes: the same four
// Application Layer services, the same workflow.Machine, not a single
// line of internal/application or internal/domain changed. It proves the
// result EPIC-005 exists for — "платформа работает end-to-end на реальных
// хранилищах и интеграциях" (ROADMAP.md v0.5).
//
// RepositoryProvider is the one exception: no GitHub token is available
// in every environment this runs in (TASK-050's Open Question), so this
// test uses the same in-memory RepositoryProvider fake EPIC-004 used —
// everything else (five PostgreSQL Store adapters, the production
// EventBus with its PostgreSQL journal) is real.
func TestGoldenPath_Infrastructure(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; run docker compose up and set it to run this test")
	}

	ctx := context.Background()
	sys, err := New(ctx, dsn)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer sys.Close()

	repos := inmemory.NewRepositoryProvider()
	rules := workflow.Machine{}

	proj := application.NewTaskProjection()
	if err := proj.Subscribe(sys.Events); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: sys.Projects, Tasks: sys.Tasks, Events: sys.Events, Rules: rules}
	work := &application.WorkService{Tasks: sys.Tasks, Executors: sys.Executors, Executions: sys.Executions, Events: sys.Events, Rules: rules}
	result := &application.ResultService{Projects: sys.Projects, Tasks: sys.Tasks, Executions: sys.Executions, Artifacts: sys.Artifacts, Events: sys.Events}
	completion := &application.CompletionService{Tasks: sys.Tasks, Repositories: repos, Events: sys.Events, Rules: rules}

	newActiveProject(ctx, t, sys.Projects)
	saveActiveExecutor(ctx, t, sys.Executors)

	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "Golden path на реальной инфраструктуре", Type: "feature",
		Scope: "Сквозной сценарий Infrastructure Layer", AcceptanceCriteria: []string{"задача доходит до Done"},
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	requireState(t, proj, "task-1", shared.StateBacklog)

	if err := planning.PlanTask(ctx, "task-1", "pm:executor-2"); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
	requireState(t, proj, "task-1", shared.StateReady)

	run, err := work.StartTask(ctx, application.StartTaskParams{TaskID: "task-1", ExecutorID: "executor-1", Actor: "developer:executor-1"})
	if err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	requireState(t, proj, "task-1", shared.StateInProgress)

	art, err := result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: run.ID(),
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
		Payload: []byte("diff --git a/x b/x"), Actor: "developer:executor-1",
	})
	if err != nil {
		t.Fatalf("RecordDraftArtifact: %v", err)
	}
	if err := result.PublishArtifact(ctx, art.ID(), "developer:executor-1"); err != nil {
		t.Fatalf("PublishArtifact: %v", err)
	}
	if err := result.SucceedExecution(ctx, run.ID(), "developer:executor-1"); err != nil {
		t.Fatalf("SucceedExecution: %v", err)
	}

	if err := completion.RequestReview(ctx, "task-1", "developer:executor-1"); err != nil {
		t.Fatalf("RequestReview: %v", err)
	}
	requireState(t, proj, "task-1", shared.StateReview)

	// Changes requested — back to the developer, matching TASK-045's branch coverage.
	if err := completion.CompleteReview(ctx, "task-1", false, "reviewer:executor-3"); err != nil {
		t.Fatalf("CompleteReview(changes requested): %v", err)
	}
	requireState(t, proj, "task-1", shared.StateInProgress)

	if err := completion.RequestReview(ctx, "task-1", "developer:executor-1"); err != nil {
		t.Fatalf("second RequestReview: %v", err)
	}
	if err := completion.CompleteReview(ctx, "task-1", true, "reviewer:executor-3"); err != nil {
		t.Fatalf("CompleteReview(approved): %v", err)
	}
	requireState(t, proj, "task-1", shared.StateTesting)

	// Tests fail once — back to the developer, matching TASK-045's branch coverage.
	if err := completion.CompleteTesting(ctx, application.CompleteTestingParams{TaskID: "task-1", Passed: false, Actor: "qa:executor-4"}); err != nil {
		t.Fatalf("CompleteTesting(failed): %v", err)
	}
	requireState(t, proj, "task-1", shared.StateInProgress)

	if err := completion.RequestReview(ctx, "task-1", "developer:executor-1"); err != nil {
		t.Fatalf("third RequestReview: %v", err)
	}
	if err := completion.CompleteReview(ctx, "task-1", true, "reviewer:executor-3"); err != nil {
		t.Fatalf("third CompleteReview: %v", err)
	}
	if err := completion.CompleteTesting(ctx, application.CompleteTestingParams{
		TaskID: "task-1", Passed: true, Repository: "github.com/nikibrdev/ai-studio-os", PullRequestID: "pr-1", Actor: "qa:executor-4",
	}); err != nil {
		t.Fatalf("CompleteTesting(passed): %v", err)
	}

	requireState(t, proj, "task-1", shared.StateDone)
	if state, err := repos.PullRequestState(ctx, "github.com/nikibrdev/ai-studio-os", "pr-1"); err != nil {
		t.Fatalf("PullRequestState: %v", err)
	} else if state != platform.PullRequestMerged {
		t.Errorf("PullRequestState() = %v, want %v", state, platform.PullRequestMerged)
	}

	// Rebuild from the real PostgreSQL journal (not the live bus) must
	// reach the same state — proves the journal, not just in-process
	// memory, carries everything the projection needs (ADR-002).
	events, err := eventbus.ReadJournal(ctx, sys.Pool)
	if err != nil {
		t.Fatalf("ReadJournal: %v", err)
	}
	rebuilt := application.NewTaskProjection()
	if err := rebuilt.Rebuild(ctx, events); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	requireState(t, rebuilt, "task-1", shared.StateDone)
}

func newActiveProject(ctx context.Context, t *testing.T, store application.ProjectStore) {
	t.Helper()
	p, _, err := project.New("proj-1", "AI Studio OS")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if _, _, err := p.ConnectRepository("github.com/org/repo"); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}
	if _, err := p.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("Save project: %v", err)
	}
}

func saveActiveExecutor(ctx context.Context, t *testing.T, store application.ExecutorStore) {
	t.Helper()
	e, _, err := executor.New("executor-1", "claude-code-instance", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := store.Save(ctx, e); err != nil {
		t.Fatalf("Save executor: %v", err)
	}
}

func requireState(t *testing.T, proj *application.TaskProjection, taskID string, want shared.TaskState) {
	t.Helper()
	view, ok := proj.Get(taskID)
	if !ok {
		t.Fatalf("projection has no view for %q", taskID)
	}
	if view.State != want {
		t.Fatalf("projection State() for %q = %v, want %v", taskID, view.State, want)
	}
}
