package application_test

import (
	"context"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/platform"
)

// TestGoldenPath_Application drives the whole golden path
// (docs/architecture/golden-path.md) through the four Application Layer
// services on in-memory adapters, exercising both branches TASK-045 asks
// for (changes requested, tests failed) before the task finally reaches
// Done — and reads every intermediate and final state through
// TaskProjection, never through TaskStore directly (ADR-014).
func TestGoldenPath_Application(t *testing.T) {
	ctx := context.Background()

	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	executors := inmemory.NewExecutorStore()
	executions := inmemory.NewExecutionStore()
	artifacts := inmemory.NewArtifactStore()
	repos := inmemory.NewRepositoryProvider()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}

	proj := application.NewTaskProjection()
	if err := proj.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	work := &application.WorkService{Tasks: tasks, Executors: executors, Executions: executions, Events: bus, Rules: rules}
	result := &application.ResultService{Projects: projects, Tasks: tasks, Executions: executions, Artifacts: artifacts, Events: bus}
	completion := &application.CompletionService{Tasks: tasks, Repositories: repos, Events: bus, Rules: rules}

	newActiveProject(t, projects)
	saveExecutor(t, executors, true, shared.RoleDeveloper)

	// Пользователь создаёт задачу.
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "Golden path целиком", Type: "feature",
		Scope: "Сквозной сценарий Application Layer", AcceptanceCriteria: []string{"задача доходит до Done"},
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	requireState(t, proj, "task-1", shared.StateBacklog)

	// PM доводит до готовности.
	if err := planning.PlanTask(ctx, "task-1", "pm:executor-2"); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
	requireState(t, proj, "task-1", shared.StateReady)

	// Developer берёт задачу — порождается Execution.
	run, err := work.StartTask(ctx, application.StartTaskParams{TaskID: "task-1", ExecutorID: "executor-1", Actor: "developer:executor-1"})
	if err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	requireState(t, proj, "task-1", shared.StateInProgress)

	// Пишет код → производит Artifact (Pull Request) → публикует его.
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

	// Открывает Pull Request → Review.
	if err := completion.RequestReview(ctx, "task-1", "developer:executor-1"); err != nil {
		t.Fatalf("RequestReview: %v", err)
	}
	requireState(t, proj, "task-1", shared.StateReview)

	// Первый круг ревью: правки запрошены — задача возвращается разработчику.
	if err := completion.CompleteReview(ctx, "task-1", false, "reviewer:executor-3"); err != nil {
		t.Fatalf("CompleteReview(changes requested): %v", err)
	}
	requireState(t, proj, "task-1", shared.StateInProgress)

	// Разработчик поправил, снова отправляет на ревью — на этот раз одобрено.
	if err := completion.RequestReview(ctx, "task-1", "developer:executor-1"); err != nil {
		t.Fatalf("second RequestReview: %v", err)
	}
	if err := completion.CompleteReview(ctx, "task-1", true, "reviewer:executor-3"); err != nil {
		t.Fatalf("CompleteReview(approved): %v", err)
	}
	requireState(t, proj, "task-1", shared.StateTesting)

	// QA: первый прогон падает.
	if err := completion.CompleteTesting(ctx, application.CompleteTestingParams{TaskID: "task-1", Passed: false, Actor: "qa:executor-4"}); err != nil {
		t.Fatalf("CompleteTesting(failed): %v", err)
	}
	requireState(t, proj, "task-1", shared.StateInProgress)

	// Возврат в Review -> Testing и повторный прогон — на этот раз успешно, с merge.
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

	// Задача закрыта — конечная точка golden path.
	requireState(t, proj, "task-1", shared.StateDone)
	if state, err := repos.PullRequestState(ctx, "github.com/nikibrdev/ai-studio-os", "pr-1"); err != nil {
		t.Fatalf("PullRequestState: %v", err)
	} else if state != platform.PullRequestMerged {
		t.Errorf("PullRequestState() = %v, want %v", state, platform.PullRequestMerged)
	}

	// Пересборка проекции с нуля из журнала должна дать тот же результат.
	rebuilt := application.NewTaskProjection()
	if err := rebuilt.Rebuild(ctx, bus.Published()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	requireState(t, rebuilt, "task-1", shared.StateDone)
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
