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
	// ErrNoDeveloperExecutor is returned when no Active Executor holds the
	// Developer role at dispatch time (e.g. it was disabled after startup).
	ErrNoDeveloperExecutor = errors.New("orchestrator: no active developer executor available")

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
	Views      *application.TaskProjection
	Repos      platform.RepositoryProvider

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
	NewExecutor func() (platform.Executor, error)

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
	if e.Type() != event.TaskPlanned {
		return nil
	}
	return d.dispatchTaskPlanned(ctx, e)
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

	executorID, err := d.pickDeveloperExecutor(ctx)
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

	backend, err := d.NewExecutor()
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

// pickDeveloperExecutor returns the id of an Active Executor holding the
// Developer role. List is ordered by id, so the choice is deterministic
// rather than dependent on map iteration order. Filtering happens here in
// memory: at one-executor-per-role scale (EPIC-010's accepted limitation)
// that is cheaper than a parameterised query.
func (d *Dispatcher) pickDeveloperExecutor(ctx context.Context) (string, error) {
	all, err := d.Executors.List(ctx)
	if err != nil {
		return "", fmt.Errorf("orchestrator: list executors: %w", err)
	}
	for _, e := range all {
		if e.AvailableForAssignment() && e.HasRole(shared.RoleDeveloper) {
			return e.ID(), nil
		}
	}
	return "", ErrNoDeveloperExecutor
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
