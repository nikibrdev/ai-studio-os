package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
)

// sequentialTaskIDGenerator is a deterministic application.TaskIDGenerator
// fake for tests — the real generator (postgres.TaskStore.NextID,
// TASK-065) requires a database; these tests exercise the HTTP layer
// (routing, (de)serialization, error mapping) against real use-case
// services backed by internal/application/inmemory, the same fakes
// internal/application's own tests use. Counts per projectID, not
// globally — TASK-NNN is unique only within a Project (ADR-011,
// BUGFIX-003), and two different projects must each get their own
// TASK-001.
type sequentialTaskIDGenerator struct {
	mu sync.Mutex
	n  map[string]int
}

func (g *sequentialTaskIDGenerator) NextID(_ context.Context, projectID string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.n == nil {
		g.n = make(map[string]int)
	}
	g.n[projectID]++
	return fmt.Sprintf("TASK-%03d", g.n[projectID]), nil
}

// testDeps wires a fresh Deps against in-memory fakes. Deps holds
// concrete *application.*Service types, not interfaces, so exercising the
// HTTP layer end-to-end through real services is the natural seam here —
// the same choice EPIC-004's own tests made for these services.
func testDeps() Deps {
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	executors := inmemory.NewExecutorStore()
	executions := inmemory.NewExecutionStore()
	artifacts := inmemory.NewArtifactStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}

	views := application.NewTaskProjection()
	if err := views.Subscribe(bus); err != nil {
		panic(err)
	}

	return Deps{
		Projects: &application.ProjectService{Projects: projects, Events: bus},
		Tasks: &application.TaskPlanningService{
			Projects: projects, Tasks: tasks, Events: bus, Rules: rules, IDs: &sequentialTaskIDGenerator{},
		},
		Work: &application.WorkService{
			Tasks: tasks, Executors: executors, Executions: executions, Events: bus, Rules: rules,
		},
		Results: &application.ResultService{
			Projects: projects, Tasks: tasks, Executions: executions, Artifacts: artifacts, Events: bus,
		},
		Completion: &application.CompletionService{
			Tasks: tasks, Repositories: inmemory.NewRepositoryProvider(), Events: bus, Rules: rules,
		},
		Views: views,
	}
}

// jsonBody marshals v for use as an httptest.NewRequest body.
func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	return bytes.NewReader(b)
}

// doRequest sends req through server and decodes a JSON response body
// into out (if out is non-nil).
func doRequest(t *testing.T, server http.Handler, req *http.Request, out any) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if out != nil && rec.Body.Len() > 0 {
		if err := json.NewDecoder(rec.Body).Decode(out); err != nil {
			t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
		}
	}
	return rec
}

// createActiveProject drives the full create -> connect-repository ->
// activate sequence (docs/api/projects.md) through real HTTP requests and
// returns the project id.
func createActiveProject(t *testing.T, server http.Handler) string {
	t.Helper()
	const id = "proj-1"

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: id, Name: "AI Studio OS"})), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create project status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+id+"/repositories",
		jsonBody(t, connectRepositoryRequest{Repository: "github.com/org/repo"})), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("connect repository status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+id+"/activate", nil), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("activate status = %d, body = %s", rec.Code, rec.Body.String())
	}

	return id
}

// seedActiveDeveloperExecutor registers an Active Executor with the
// Developer role directly in deps.Work's store — there is no HTTP route
// to register an Executor in this version of the API (out of scope,
// ADR-007 Decision Required), so tests exercising StartTask seed one the
// same way internal/application's own work_test.go does.
func seedActiveDeveloperExecutor(t *testing.T, deps Deps, id string) {
	t.Helper()
	e, _, err := executor.New(id, "claude-code-instance", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := deps.Work.Executors.Save(context.Background(), e); err != nil {
		t.Fatalf("save executor: %v", err)
	}
}

// createReadyTask creates and plans a task within projectID
// (/projects/{projectId}/tasks — BUGFIX-003), returning its ID.
func createReadyTask(t *testing.T, server http.Handler, projectID string) string {
	t.Helper()
	var created taskResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача", Type: "feature"})), &created)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create task status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+created.ID+"/plan", nil), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("plan task status = %d, body = %s", rec.Code, rec.Body.String())
	}
	return created.ID
}

// startExecution seeds an Active Developer Executor and starts the given
// (Ready) task through real HTTP requests, returning the execution id.
func startExecution(t *testing.T, server http.Handler, deps Deps, projectID, taskID string) string {
	t.Helper()
	const executorID = "executor-1"
	seedActiveDeveloperExecutor(t, deps, executorID)

	var run executionResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/start",
		jsonBody(t, startTaskRequest{ExecutorID: executorID})), &run)
	if rec.Code != http.StatusCreated {
		t.Fatalf("start task status = %d, body = %s", rec.Code, rec.Body.String())
	}
	return run.ExecutionID
}
