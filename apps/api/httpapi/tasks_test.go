package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTask_GeneratesSequentialID(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	var first, second taskResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks",
		jsonBody(t, createTaskRequest{Title: "Первая задача", Type: "feature"})), &first)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks",
		jsonBody(t, createTaskRequest{Title: "Вторая задача", Type: "feature"})), &second)
	if rec.Code != http.StatusCreated {
		t.Fatalf("second create status = %d, body = %s", rec.Code, rec.Body.String())
	}

	if first.ID != "TASK-001" || second.ID != "TASK-002" {
		t.Errorf("ids = %q, %q, want TASK-001, TASK-002", first.ID, second.ID)
	}
	if first.State != "backlog" {
		t.Errorf("State = %q, want backlog", first.State)
	}
}

func TestCreateTask_ProjectNotActiveReturns409(t *testing.T) {
	server := NewServer(testDeps())
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-1", Name: "AI Studio OS"})), nil)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-1/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача", Type: "feature"})), nil)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestCreateTask_ProjectNotFoundReturns404(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/missing/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача", Type: "feature"})), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetTask_UnknownReturns404(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/proj-1/tasks/missing", nil), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetTask_ReflectsCreatedState(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	var created taskResponse
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача", Type: "feature"})), &created)

	var view taskViewResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks/"+created.ID, nil), &view)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if view.ID != created.ID || view.ProjectID != projectID || view.State != "backlog" {
		t.Errorf("view = %+v", view)
	}
}

// TestGetTask_ReturnsDescriptiveFields proves TASK-076's addition:
// TaskProjection is the only read path for Task (ADR-014), so title/type/
// scope/acceptanceCriteria must reach the client through this endpoint,
// not a separate one backed directly by TaskStore.
func TestGetTask_ReturnsDescriptiveFields(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	var created taskResponse
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks",
		jsonBody(t, createTaskRequest{
			Title: "Заголовок", Type: "feature", Scope: "Описание",
			AcceptanceCriteria: []string{"критерий раз", "критерий два"},
		})), &created)

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks/"+created.ID, nil), &view)

	if view.Title != "Заголовок" || view.Type != "feature" || view.Scope != "Описание" {
		t.Errorf("view = %+v, want Title/Type/Scope from creation", view)
	}
	if len(view.AcceptanceCriteria) != 2 || view.AcceptanceCriteria[0] != "критерий раз" || view.AcceptanceCriteria[1] != "критерий два" {
		t.Errorf("AcceptanceCriteria = %v, want the two supplied criteria", view.AcceptanceCriteria)
	}
}

func TestPlanTask_TransitionsToReady(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	var created taskResponse
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача", Type: "feature"})), &created)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+created.ID+"/plan",
		jsonBody(t, actorRequest{Actor: "pm:executor-1"})), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("plan status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks/"+created.ID, nil), &view)
	if view.State != "ready" {
		t.Errorf("State = %q, want ready", view.State)
	}
}

func TestPlanTask_UnknownTaskReturns404(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-1/tasks/missing/plan", nil), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestCreateTask_SameIDDifferentProjectsDoNotCollide proves BUGFIX-003 at
// the HTTP layer: two different projects each creating their first task
// both legitimately get TASK-001, and each project's own task keeps its
// own data — no cross-project overwrite.
func TestCreateTask_SameIDDifferentProjectsDoNotCollide(t *testing.T) {
	server := NewServer(testDeps())

	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-a", Name: "A"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-a/repositories",
		jsonBody(t, connectRepositoryRequest{Repository: "github.com/org/repo"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-a/activate", nil), nil)

	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-b", Name: "B"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-b/repositories",
		jsonBody(t, connectRepositoryRequest{Repository: "github.com/org/repo"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-b/activate", nil), nil)

	var taskA, taskB taskResponse
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-a/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача A", Type: "feature"})), &taskA)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-b/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача B", Type: "bugfix"})), &taskB)

	if taskA.ID != "TASK-001" || taskB.ID != "TASK-001" {
		t.Fatalf("taskA.ID = %q, taskB.ID = %q, want both TASK-001 (independent per-project sequences)", taskA.ID, taskB.ID)
	}

	var viewA, viewB taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/proj-a/tasks/TASK-001", nil), &viewA)
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/proj-b/tasks/TASK-001", nil), &viewB)
	if viewA.ProjectID != "proj-a" || viewB.ProjectID != "proj-b" {
		t.Errorf("viewA.ProjectID = %q, viewB.ProjectID = %q, want proj-a and proj-b respectively", viewA.ProjectID, viewB.ProjectID)
	}
}

func TestListTasks_IsolatesProjects(t *testing.T) {
	server := NewServer(testDeps())

	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-a", Name: "A"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-a/repositories",
		jsonBody(t, connectRepositoryRequest{Repository: "github.com/org/repo"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-a/activate", nil), nil)

	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-b", Name: "B"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-b/repositories",
		jsonBody(t, connectRepositoryRequest{Repository: "github.com/org/repo"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-b/activate", nil), nil)

	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-a/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача A1", Type: "feature"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-a/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача A2", Type: "feature"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-b/tasks",
		jsonBody(t, createTaskRequest{Title: "Задача B1", Type: "feature"})), nil)

	var viewsA []taskViewResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/proj-a/tasks", nil), &viewsA)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(viewsA) != 2 {
		t.Fatalf("proj-a tasks = %+v, want exactly 2 (not proj-b's)", viewsA)
	}
	for _, v := range viewsA {
		if v.ProjectID != "proj-a" {
			t.Errorf("proj-a list contains a task from project %q", v.ProjectID)
		}
	}

	var viewsB []taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/proj-b/tasks", nil), &viewsB)
	if len(viewsB) != 1 || viewsB[0].ProjectID != "proj-b" {
		t.Fatalf("proj-b tasks = %+v, want exactly 1 task belonging to proj-b", viewsB)
	}
}

func TestListTasks_EmptyReturnsEmptyArrayNotNull(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks", nil), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "[]\n" {
		t.Errorf("body = %q, want an empty JSON array, not null", rec.Body.String())
	}
}
