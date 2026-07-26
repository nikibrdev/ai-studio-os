package application_test

import (
	"context"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
)

// checkpointEvent drives a projection to a given task state directly, so a
// test can cover every state without walking the whole golden path for each.
type checkpointEvent struct {
	typ, projectID, subjectID string
	data                      map[string]string
}

func (e checkpointEvent) ID() string              { return "evt-" + e.projectID + "-" + e.subjectID + "-" + e.typ }
func (e checkpointEvent) Type() string            { return e.typ }
func (e checkpointEvent) SchemaVersion() int      { return 1 }
func (e checkpointEvent) OccurredAt() time.Time   { return time.Now() }
func (e checkpointEvent) Source() string          { return "task" }
func (e checkpointEvent) Actor() string           { return "" }
func (e checkpointEvent) ProjectID() string       { return e.projectID }
func (e checkpointEvent) SubjectID() string       { return e.subjectID }
func (e checkpointEvent) Data() map[string]string { return e.data }

// projectionAt builds a projection holding one task in the given state.
func projectionAt(t *testing.T, projectID, taskID string, state shared.TaskState) *application.TaskProjection {
	t.Helper()
	ctx := context.Background()
	proj := application.NewTaskProjection()

	feed := func(typ string, data map[string]string) {
		e := checkpointEvent{typ: typ, projectID: projectID, subjectID: taskID, data: data}
		if err := proj.Handle(ctx, e); err != nil {
			t.Fatalf("Handle(%s): %v", typ, err)
		}
	}

	feed(event.TaskCreated, map[string]string{"title": "Задача", "type": "feature"})
	switch state {
	case shared.StateBacklog:
		// TaskCreated already leaves it in Backlog.
	case shared.StateReady:
		feed(event.TaskPlanned, nil)
	case shared.StateInProgress:
		feed(event.TaskPlanned, nil)
		feed(event.TaskStarted, nil)
	case shared.StateReview:
		feed(event.TaskPlanned, nil)
		feed(event.TaskStarted, nil)
		feed(event.ReviewRequested, nil)
	case shared.StateTesting:
		feed(event.TaskPlanned, nil)
		feed(event.TaskStarted, nil)
		feed(event.ReviewRequested, nil)
		feed(event.ReviewCompleted, map[string]string{"to": string(shared.StateTesting)})
	case shared.StateDone:
		feed(event.TaskPlanned, nil)
		feed(event.TaskStarted, nil)
		feed(event.ReviewRequested, nil)
		feed(event.ReviewCompleted, map[string]string{"to": string(shared.StateTesting)})
		feed(event.TaskCompleted, nil)
	default:
		t.Fatalf("projectionAt does not know how to reach %v", state)
	}

	if got, _ := proj.Get(projectID, taskID); got.State != state {
		t.Fatalf("projection reached %v, want %v — the fixture itself is wrong", got.State, state)
	}
	return proj
}

// Only the two states ADR-007 names are checkpoints. Every other state must
// stay out of the list, including Done — a finished task owes nobody anything.
func TestListAwaitingDecision_OnlyCheckpointStatesAppear(t *testing.T) {
	cases := []struct {
		state shared.TaskState
		want  application.DecisionKind
	}{
		{shared.StateBacklog, application.DecisionDefinitionOfReady},
		{shared.StateTesting, application.DecisionAcceptance},
		{shared.StateReady, ""},
		{shared.StateInProgress, ""},
		{shared.StateReview, ""},
		{shared.StateDone, ""},
	}

	for _, tc := range cases {
		t.Run(string(tc.state), func(t *testing.T) {
			proj := projectionAt(t, "proj-1", "TASK-001", tc.state)
			got := proj.ListAwaitingDecision()

			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("ListAwaitingDecision() = %+v, want empty for state %v", got, tc.state)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("ListAwaitingDecision() = %+v, want exactly one entry", got)
			}
			if got[0].Decision != tc.want {
				t.Errorf("Decision = %q, want %q", got[0].Decision, tc.want)
			}
			if got[0].Task.ID != "TASK-001" || got[0].Task.Title != "Задача" {
				t.Errorf("Task = %+v, want the task's identity and title carried through", got[0].Task)
			}
		})
	}
}

func TestListAwaitingDecision_EmptyIsNotAnError(t *testing.T) {
	proj := application.NewTaskProjection()
	if got := proj.ListAwaitingDecision(); len(got) != 0 {
		t.Errorf("ListAwaitingDecision() = %+v, want empty", got)
	}
}

// The list spans projects, and its order is deterministic — project first,
// then task id — so a UI does not reorder between reloads.
func TestListAwaitingDecision_SpansProjectsInDeterministicOrder(t *testing.T) {
	ctx := context.Background()
	proj := application.NewTaskProjection()

	// Fed out of order on purpose.
	for _, e := range []checkpointEvent{
		{typ: event.TaskCreated, projectID: "proj-b", subjectID: "TASK-002"},
		{typ: event.TaskCreated, projectID: "proj-a", subjectID: "TASK-002"},
		{typ: event.TaskCreated, projectID: "proj-b", subjectID: "TASK-001"},
		{typ: event.TaskCreated, projectID: "proj-a", subjectID: "TASK-001"},
	} {
		if err := proj.Handle(ctx, e); err != nil {
			t.Fatalf("Handle: %v", err)
		}
	}

	got := proj.ListAwaitingDecision()
	if len(got) != 4 {
		t.Fatalf("ListAwaitingDecision() = %d entries, want 4", len(got))
	}

	want := []string{"proj-a/TASK-001", "proj-a/TASK-002", "proj-b/TASK-001", "proj-b/TASK-002"}
	for i, w := range want {
		if key := got[i].Task.ProjectID + "/" + got[i].Task.ID; key != w {
			t.Errorf("entry %d = %s, want %s", i, key, w)
		}
	}
}

// DecisionFor is what a delivery layer uses to annotate a single task. It must
// agree with the list, or the same task would be shown as awaiting a decision
// in one place and not in another.
func TestDecisionFor_AgreesWithTheList(t *testing.T) {
	for _, state := range []shared.TaskState{
		shared.StateBacklog, shared.StateReady, shared.StateInProgress,
		shared.StateReview, shared.StateTesting, shared.StateDone,
	} {
		proj := projectionAt(t, "proj-1", "TASK-001", state)
		inList := len(proj.ListAwaitingDecision()) == 1
		annotated := application.DecisionFor(state) != ""

		if inList != annotated {
			t.Errorf("state %v: in list = %t, but DecisionFor says %t", state, inList, annotated)
		}
		if inList {
			if got := application.DecisionFor(state); got != proj.ListAwaitingDecision()[0].Decision {
				t.Errorf("state %v: DecisionFor = %q, list says %q", state, got, proj.ListAwaitingDecision()[0].Decision)
			}
		}
	}
}
