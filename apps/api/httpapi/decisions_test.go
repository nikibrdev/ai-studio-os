package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// activeProject creates and activates a project under a chosen id, for tests
// that need more than one (createActiveProject in deps_test.go fixes the id).
func activeProject(t *testing.T, server http.Handler, id string) {
	t.Helper()
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: id, Name: id})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+id+"/repositories",
		jsonBody(t, connectRepositoryRequest{Repository: "github.com/org/repo"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+id+"/activate", nil), nil)
}

func TestListDecisions_EmptyReturnsEmptyArrayNotNull(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/decisions", nil), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	// Nothing awaiting a decision is a normal state; a client must not have to
	// distinguish [] from null.
	if got := rec.Body.String(); got != "{\"decisions\":[]}\n" {
		t.Errorf("body = %q, want an empty array rather than null", got)
	}
}

// A freshly created task sits in Backlog, which is the Definition of Ready
// checkpoint.
func TestListDecisions_NewTaskAwaitsDefinitionOfReady(t *testing.T) {
	server := NewServer(testDeps())
	activeProject(t, server, "proj-1")
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-1/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача", Type: "feature"})), nil)

	var got listDecisionsResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/decisions", nil), &got)

	if len(got.Decisions) != 1 {
		t.Fatalf("decisions = %+v, want exactly one", got.Decisions)
	}
	d := got.Decisions[0]
	if d.Decision != "definition-of-ready" {
		t.Errorf("decision = %q, want definition-of-ready", d.Decision)
	}
	if d.Task.ID != "TASK-001" || d.Task.ProjectID != "proj-1" || d.Task.State != "backlog" {
		t.Errorf("task = %+v, want TASK-001 in proj-1 in backlog", d.Task)
	}
	// The annotation on the task itself must agree with the list.
	if d.Task.AwaitingDecision != "definition-of-ready" {
		t.Errorf("task.awaitingDecision = %q, want it to match the list entry", d.Task.AwaitingDecision)
	}
}

// Accepting Definition of Ready moves the task to Ready, which is not a
// checkpoint — so it must leave the list.
func TestListDecisions_AcceptingDefinitionOfReadyRemovesTaskFromList(t *testing.T) {
	server := NewServer(testDeps())
	activeProject(t, server, "proj-1")
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-1/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача", Type: "feature"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-1/tasks/TASK-001/plan", nil), nil)

	var got listDecisionsResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/decisions", nil), &got)

	if len(got.Decisions) != 0 {
		t.Errorf("decisions = %+v, want empty once the task left Backlog", got.Decisions)
	}

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/proj-1/tasks/TASK-001", nil), &view)
	if view.State != "ready" {
		t.Fatalf("state = %q, want ready", view.State)
	}
	if view.AwaitingDecision != "" {
		t.Errorf("awaitingDecision = %q, want empty for a task in Ready", view.AwaitingDecision)
	}
}

// The list spans projects — the question is "what awaits a decision at all".
func TestListDecisions_SpansProjects(t *testing.T) {
	server := NewServer(testDeps())
	activeProject(t, server, "proj-a")
	activeProject(t, server, "proj-b")
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-a/tasks",
		jsonBody(t, createTaskRequest{Title: "A", Type: "feature"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-b/tasks",
		jsonBody(t, createTaskRequest{Title: "B", Type: "feature"})), nil)

	var got listDecisionsResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/decisions", nil), &got)

	if len(got.Decisions) != 2 {
		t.Fatalf("decisions = %+v, want one per project", got.Decisions)
	}
	// Deterministic order: project first, so a UI does not reshuffle.
	if got.Decisions[0].Task.ProjectID != "proj-a" || got.Decisions[1].Task.ProjectID != "proj-b" {
		t.Errorf("order = %s, %s; want proj-a before proj-b",
			got.Decisions[0].Task.ProjectID, got.Decisions[1].Task.ProjectID)
	}
}
