package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// Errors reported when the state a dispatch needs is missing. All are
// returned rather than fatal: Poller logs them, advances its cursor, and
// the task simply stays in Ready — where a human can still start it
// through apps/api. Retries are deliberately outside EPIC-010's scope.
var (
	// ErrNoExecutorForRole is returned when this orchestrator's own registry
	// entry for the role does not exist (e.g. bootstrap never ran).
	ErrNoExecutorForRole = errors.New("orchestrator: no registry entry for this role")

	// ErrExecutorNotUsable is returned when the entry exists but cannot take
	// work — not Active, or no longer holding the role.
	ErrExecutorNotUsable = errors.New("orchestrator: registry entry cannot take work")

	// ErrNoRepository is returned when the task's Project has no repository
	// connected — there is nowhere to create a branch.
	ErrNoRepository = errors.New("orchestrator: project has no connected repository")

	// ErrTaskNotInProjection is returned when the projection holds no view
	// for the task the event names, so its planning content is unknown.
	ErrTaskNotInProjection = errors.New("orchestrator: task is not in the projection")
)

// baseBranch is what task branches are cut from (docs/development/git-workflow.md).
const baseBranch = "main"

// Dispatcher turns task lifecycle events into Executor work. It holds no
// domain rules: every state change goes through internal/application, which
// asks the task/workflow modules whether a transition is legal
// (module-boundaries.md).
type Dispatcher struct {
	Projects   application.ProjectStore
	Executors  application.ExecutorStore
	Work       *application.WorkService
	Results    *application.ResultService
	Completion *application.CompletionService

	// Planning applies what a Project Manager agent proposed (EPIC-013).
	// RefineTask is the only planning command the orchestrator calls: PlanTask
	// would pass a human checkpoint, which no automation may do.
	Planning *application.TaskPlanningService
	Views    *application.TaskProjection
	Repos    platform.RepositoryProvider

	// StatusPollInterval and ExecutionTimeout override the defaults for
	// watching an execution (monitor.go). Zero means use the default; tests
	// set them small so they do not wait minutes.
	StatusPollInterval time.Duration
	ExecutionTimeout   time.Duration

	// NewExecutor builds the Executor for one Execution's Accept -> Finish
	// lifecycle. A factory rather than a single value because
	// agents/claude-code explicitly serves exactly one Execution per value
	// — and because it lets tests substitute a fake for the real Docker
	// sandbox.
	//
	// Takes the role so the backend that runs corresponds to the registry
	// entry that was selected (TASK-086): the entry's identifier is derived
	// from the same role, which is what keeps the two from diverging. One
	// adapter serves every role, differing only by prompt (ADR-007).
	NewExecutor func(role shared.Role) (platform.Executor, error)

	// NewReviewer builds the Executor for a review. Separate from NewExecutor
	// because reviewing needs one capability beyond the Executor contract —
	// reporting a decision — and because that keeps this file free of any
	// concrete adapter: ReviewExecutor is satisfied here through primitives,
	// and adapting a real adapter to it happens in main.go, the only place
	// permitted to name one (module-boundaries.md).
	NewReviewer func() (ReviewExecutor, error)

	// NewPlanner builds the Executor for a planning run (Project Manager).
	// Separate for the same reason as NewReviewer: preparing a Definition of
	// Ready needs a capability beyond the Executor contract, and keeping it
	// behind a factory keeps this file free of any concrete adapter.
	NewPlanner func() (PlanExecutor, error)

	// NewReporter builds the Executor for a QA run. Separate for the same reason
	// as the others: checking a task needs a capability beyond the Executor
	// contract, and the factory keeps this file free of any concrete adapter.
	NewReporter func() (ReportExecutor, error)

	Log *log.Logger
}

// Observe applies an event to the read models this process maintains,
// without any side effects. Wired as Poller.Seed so replayed history makes
// the projection current without re-dispatching finished work.
func (d *Dispatcher) Observe(ctx context.Context, e platform.Event) error {
	return d.Views.Handle(ctx, e)
}

// Handle applies an event and then acts on it. Wired as Poller.Handle.
//
// The projection is updated before dispatching, never after: dispatch reads
// the task's planning content back out of it, so the event that carries that
// content must already be applied.
func (d *Dispatcher) Handle(ctx context.Context, e platform.Event) error {
	if err := d.Observe(ctx, e); err != nil {
		return err
	}
	switch {
	case taskCreated(e):
		return d.dispatchTaskCreated(ctx, e)
	case e.Type() == event.TaskPlanned:
		return d.dispatchTaskPlanned(ctx, e)
	case reviewRequested(e):
		return d.dispatchReviewRequested(ctx, e)
	case enteredTesting(e):
		return d.dispatchEnteredTesting(ctx, e)
	default:
		return nil
	}
}

// dispatchTaskPlanned takes a task from Ready to a running Execution:
// pick an Executor, move the Task through the Application Layer, cut the
// branch, and hand the work to the backend.
//
// Watching that Execution to completion — collecting Artifacts, opening the
// Pull Request, concluding the Execution and requesting review — is
// TASK-083. Until then Accept leaves the sandbox running: this task
// deliberately stops at the point EPIC-010's decomposition specified.
func (d *Dispatcher) dispatchTaskPlanned(ctx context.Context, e platform.Event) error {
	projectID, taskID := e.ProjectID(), e.SubjectID()

	view, ok := d.Views.Get(projectID, taskID)
	if !ok {
		return fmt.Errorf("%w: %s/%s", ErrTaskNotInProjection, projectID, taskID)
	}

	executorID, err := d.executorForRole(ctx, shared.RoleDeveloper)
	if err != nil {
		return err
	}

	repo, err := d.repositoryOf(ctx, projectID)
	if err != nil {
		return err
	}

	// The branch is cut before the Task is moved, so the most likely failure
	// here — misconfigured repository, missing or invalid token, GitHub
	// unreachable — leaves the Task untouched in Ready, where a human can
	// retry it through apps/api. Live-testing TASK-082 showed the other
	// order is worse: CreateBranch failed after StartTask had already
	// published TaskStarted/ExecutionQueued/ExecutionStarted, stranding the
	// task In Progress with a running Execution and no branch to work in,
	// and nothing retries (Poller advances its cursor past a failure).
	// An orphan branch, the cost of this order, harms nothing.
	branch := branchName(taskID, view.Title)
	if err := d.Repos.CreateBranch(ctx, repo, branch, baseBranch); err != nil {
		return fmt.Errorf("orchestrator: create branch %s in %s: %w", branch, repo, err)
	}

	run, err := d.Work.StartTask(ctx, application.StartTaskParams{
		ProjectID: projectID, TaskID: taskID, ExecutorID: executorID, Actor: actor,
	})
	if err != nil {
		return fmt.Errorf("orchestrator: start task %s/%s: %w", projectID, taskID, err)
	}

	backend, err := d.NewExecutor(shared.RoleDeveloper)
	if err != nil {
		return fmt.Errorf("orchestrator: build executor for %s/%s: %w", projectID, taskID, err)
	}

	task := platform.ExecutorTask{
		TaskID:             taskID,
		ProjectID:          projectID,
		Role:               string(shared.RoleDeveloper),
		Title:              view.Title,
		Type:               view.Type,
		Scope:              view.Scope,
		AcceptanceCriteria: view.AcceptanceCriteria,
		Repository:         repo,
		Branch:             branch,
	}
	if err := backend.Accept(ctx, task); err != nil {
		return fmt.Errorf("orchestrator: accept %s/%s: %w", projectID, taskID, err)
	}

	d.Log.Printf("dispatched %s/%s to %s: execution %s, branch %s", projectID, taskID, executorID, run.ID(), branch)

	// Accept only starts the work; watching it through to a Pull Request and
	// Review is monitorExecution, which also guarantees the sandbox is torn
	// down (TASK-083).
	return d.monitorExecution(ctx, backend, executionContext{
		executionID: run.ID(),
		projectID:   projectID,
		taskID:      taskID,
		repository:  repo,
		branch:      branch,
		view:        view,
	})
}

// executorForRole returns the id of this orchestrator's own registry entry
// for the role, verifying it is usable.
//
// It looks the entry up by exact identifier rather than scanning for "the
// first Active executor holding this role" (TASK-086). That earlier form was
// not a choice at all but an accident of id ordering: the live run of
// TASK-085 selected exec-store-1 — a record left behind by integration
// tests — while actually running this orchestrator's own backend, so the
// events named one executor and a different one did the work. Looking up by
// derived identifier makes foreign records unable to influence the choice at
// all, rather than merely unlikely to.
//
// A present-but-unusable entry is an error, never a silent fallback to some
// other record: falling back is exactly how the two could diverge again.
func (d *Dispatcher) executorForRole(ctx context.Context, role shared.Role) (string, error) {
	want := executorIDForRole(role)

	all, err := d.Executors.List(ctx)
	if err != nil {
		return "", fmt.Errorf("orchestrator: list executors: %w", err)
	}
	for _, e := range all {
		if e.ID() != want {
			continue
		}
		if !e.AvailableForAssignment() {
			return "", fmt.Errorf("%w: %s is %s, not active", ErrExecutorNotUsable, want, e.State())
		}
		if !e.HasRole(role) {
			return "", fmt.Errorf("%w: %s does not hold the %s role", ErrExecutorNotUsable, want, role)
		}
		return e.ID(), nil
	}
	return "", fmt.Errorf("%w: expected registry entry %s", ErrNoExecutorForRole, want)
}

// repositoryOf returns the repository a task's work belongs in: the
// Project's first connected repository. One repository per project is
// enough for v1.0 — choosing among several would need a rule nothing in the
// domain expresses yet (Project.repositories is an unordered set with no
// designated primary), and inventing one here would put a domain decision
// in the orchestrator.
func (d *Dispatcher) repositoryOf(ctx context.Context, projectID string) (string, error) {
	proj, err := d.Projects.Get(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("orchestrator: get project %s: %w", projectID, err)
	}
	repos := proj.Repositories()
	if len(repos) == 0 {
		return "", fmt.Errorf("%w: %s", ErrNoRepository, projectID)
	}
	return repos[0], nil
}

// branchName builds the task branch name per docs/development/git-workflow.md
// (feature/<task-id>-<short-name>).
//
// Titles in this project are Russian, and a naive ASCII slug of one is
// empty. Rather than inventing a transliteration table, an empty slug
// simply yields feature/<taskID> — still unique, since TASK-NNN is unique
// within a project (ADR-011).
func branchName(taskID, title string) string {
	slug := asciiSlug(title)
	if slug == "" {
		return "feature/" + taskID
	}
	return "feature/" + taskID + "-" + slug
}

// maxSlugLength keeps branch names short enough to stay readable in git
// output; the task id already carries the identity.
const maxSlugLength = 40

// asciiSlug lowercases title and keeps only ASCII letters and digits,
// collapsing every other run of characters into a single hyphen.
func asciiSlug(title string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(title) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastHyphen = false
		case !lastHyphen && b.Len() > 0:
			b.WriteByte('-')
			lastHyphen = true
		}
		if b.Len() >= maxSlugLength {
			break
		}
	}
	return strings.Trim(b.String(), "-")
}
