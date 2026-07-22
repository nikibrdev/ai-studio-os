//go:build integration

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/infrastructure/wiring"
)

// TestGoldenPath_HTTP proves apps/api carries a task through the whole
// golden path (docs/architecture/golden-path.md) via real HTTP requests
// to a real server backed by a real PostgreSQL — not httptest.Recorder
// calling ServeHTTP in-process (that is what tasks_test.go/work_test.go/
// etc. already do), and not internal/application's own fakes
// (e2e_test.go). Mirrors internal/application/e2e_test.go and
// internal/infrastructure/wiring's TestGoldenPath_Infrastructure,
// including both branches TASK-045 asked for (changes requested, tests
// failed).
//
// RepositoryProvider: no GitHub token is available in every environment
// this runs in (same reasoning as TestGoldenPath_Infrastructure) — this
// test wires the in-memory RepositoryProvider fake EPIC-004 used instead
// of sys.Repository, which is otherwise real end to end.
func TestGoldenPath_HTTP(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; run docker compose up and set it to run this test")
	}
	qdrantURL := os.Getenv("TEST_QDRANT_URL")

	ctx := context.Background()
	sys, err := wiring.New(ctx, dsn, qdrantURL)
	if err != nil {
		t.Fatalf("wiring.New: %v", err)
	}
	defer sys.Close()

	views := application.NewTaskProjection()
	if err := views.Subscribe(sys.Events); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	rules := workflow.Machine{}
	repos := inmemory.NewRepositoryProvider()

	deps := Deps{
		Projects: &application.ProjectService{Projects: sys.Projects, Events: sys.Events},
		Tasks: &application.TaskPlanningService{
			Projects: sys.Projects, Tasks: sys.Tasks, Events: sys.Events, Rules: rules, IDs: sys.Tasks,
		},
		Work: &application.WorkService{
			Tasks: sys.Tasks, Executors: sys.Executors, Executions: sys.Executions, Events: sys.Events, Rules: rules,
		},
		Results: &application.ResultService{
			Projects: sys.Projects, Tasks: sys.Tasks, Executions: sys.Executions, Artifacts: sys.Artifacts, Events: sys.Events,
		},
		Completion: &application.CompletionService{Tasks: sys.Tasks, Repositories: repos, Events: sys.Events, Rules: rules},
		Views:      views,
	}

	// httptest.NewServer starts a real TCP listener on loopback — requests
	// below go over real HTTP, not an in-process ServeHTTP call (unlike
	// doRequest in deps_test.go, used by this package's other tests).
	server := httptest.NewServer(NewServer(deps))
	defer server.Close()
	client := server.Client()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	projectID := "proj-http-golden-" + suffix
	executorID := "executor-http-golden-" + suffix

	// Создание и активация проекта — целиком через HTTP.
	httpDo(t, client, http.MethodPost, server.URL+"/projects", createProjectRequest{ID: projectID, Name: "HTTP Golden Path"}, http.StatusCreated, nil)
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/repositories", connectRepositoryRequest{Repository: "github.com/org/repo"}, http.StatusNoContent, nil)
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/activate", nil, http.StatusNoContent, nil)

	// Исполнитель — нет HTTP-маршрута для регистрации (вне scope этого
	// эпика, ADR-007 Decision Required), сохраняется напрямую в реальный
	// PostgreSQL через уже собранный sys.Executors.
	e, _, err := executor.New(executorID, "claude-code-instance", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("executor Activate: %v", err)
	}
	if err := sys.Executors.Save(ctx, e); err != nil {
		t.Fatalf("save executor: %v", err)
	}

	// Постановка задачи.
	var task taskResponse
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks",
		createTaskRequest{Title: "Golden path через HTTP на реальной инфраструктуре", Type: "feature", Scope: "Сквозной сценарий apps/api", AcceptanceCriteria: []string{"задача доходит до Done"}},
		http.StatusCreated, &task)
	taskID := task.ID
	requireHTTPTaskState(t, client, server.URL, projectID, taskID, "backlog")

	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/plan", actorRequest{Actor: "pm:executor-2"}, http.StatusNoContent, nil)
	requireHTTPTaskState(t, client, server.URL, projectID, taskID, "ready")

	// Запуск работы — порождает Execution.
	var run executionResponse
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/start",
		startTaskRequest{ExecutorID: executorID, Actor: "developer:executor-1"}, http.StatusCreated, &run)
	requireHTTPTaskState(t, client, server.URL, projectID, taskID, "in-progress")

	// Производство результата: черновик артефакта, доработка, публикация.
	artifactID := "art-http-golden-" + suffix
	var art artifactResponse
	httpDo(t, client, http.MethodPost, server.URL+"/artifacts",
		recordDraftArtifactRequest{
			ID: artifactID, ProjectID: projectID, ExecutionID: run.ExecutionID,
			Type: "PullRequest", Origin: "agent", Author: "developer", Actor: "developer:executor-1",
		}, http.StatusCreated, &art)
	httpDo(t, client, http.MethodPatch, server.URL+"/artifacts/"+artifactID,
		updateArtifactDraftRequest{Payload: []byte("diff --git a/x b/x")}, http.StatusNoContent, nil)
	httpDo(t, client, http.MethodPost, server.URL+"/artifacts/"+artifactID+"/publish",
		actorRequest{Actor: "developer:executor-1"}, http.StatusNoContent, nil)
	httpDo(t, client, http.MethodPost, server.URL+"/executions/"+run.ExecutionID+"/succeed",
		executionActionRequest{ProjectID: projectID, Actor: "developer:executor-1"}, http.StatusNoContent, nil)

	// Ревью, первый круг: правки запрошены.
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/request-review",
		actorRequest{Actor: "developer:executor-1"}, http.StatusNoContent, nil)
	requireHTTPTaskState(t, client, server.URL, projectID, taskID, "review")
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/complete-review",
		completeReviewRequest{Approved: false, Actor: "reviewer:executor-3"}, http.StatusNoContent, nil)
	requireHTTPTaskState(t, client, server.URL, projectID, taskID, "in-progress")

	// Второй круг: одобрено.
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/request-review",
		actorRequest{Actor: "developer:executor-1"}, http.StatusNoContent, nil)
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/complete-review",
		completeReviewRequest{Approved: true, Actor: "reviewer:executor-3"}, http.StatusNoContent, nil)
	requireHTTPTaskState(t, client, server.URL, projectID, taskID, "testing")

	// QA: первый прогон падает.
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/complete-testing",
		completeTestingRequest{Passed: false, Actor: "qa:executor-4"}, http.StatusNoContent, nil)
	requireHTTPTaskState(t, client, server.URL, projectID, taskID, "in-progress")

	// Возврат в Review -> Testing, повторный прогон — успешно, с merge.
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/request-review",
		actorRequest{Actor: "developer:executor-1"}, http.StatusNoContent, nil)
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/complete-review",
		completeReviewRequest{Approved: true, Actor: "reviewer:executor-3"}, http.StatusNoContent, nil)
	httpDo(t, client, http.MethodPost, server.URL+"/projects/"+projectID+"/tasks/"+taskID+"/complete-testing",
		completeTestingRequest{Passed: true, Repository: "github.com/org/repo", PullRequestID: "pr-1", Actor: "qa:executor-4"},
		http.StatusNoContent, nil)

	// Задача закрыта — конечная точка golden path.
	requireHTTPTaskState(t, client, server.URL, projectID, taskID, "done")
	if len(repos.MergeCalls) != 1 || repos.MergeCalls[0] != "github.com/org/repo/pr-1" {
		t.Errorf("MergeCalls = %v, want exactly one call for github.com/org/repo/pr-1", repos.MergeCalls)
	}
}

func requireHTTPTaskState(t *testing.T, client *http.Client, baseURL, projectID, taskID, want string) {
	t.Helper()
	var view taskViewResponse
	httpDo(t, client, http.MethodGet, baseURL+"/projects/"+projectID+"/tasks/"+taskID, nil, http.StatusOK, &view)
	if view.State != want {
		t.Fatalf("task %s state = %q, want %q", taskID, view.State, want)
	}
}

// httpDo sends a real HTTP request (JSON body if reqBody is non-nil) and
// decodes the response into out (if non-nil), failing the test if the
// status code does not match wantStatus.
func httpDo(t *testing.T, client *http.Client, method, url string, reqBody any, wantStatus int, out any) {
	t.Helper()

	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("build request %s %s: %v", method, url, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body for %s %s: %v", method, url, err)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("%s %s: status = %d, want %d, body = %s", method, url, resp.StatusCode, wantStatus, respBody)
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			t.Fatalf("decode response body %q for %s %s: %v", respBody, method, url, err)
		}
	}
}
