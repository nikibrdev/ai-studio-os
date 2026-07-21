// Package claudecode implements platform.Executor (ADR-005) for Claude
// Code: the first real Executor adapter, running the AI Developer agent
// inside an isolated Docker sandbox (container/, ADR-006).
package claudecode

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"ai-studio-os/agents/claude-code/container"
	"ai-studio-os/internal/platform"
)

// ErrNotAccepted is returned by Artifacts, Status and Finish when called
// before Accept.
var ErrNotAccepted = errors.New("claudecode: Accept must be called before Artifacts/Status/Finish")

// sandbox is the subset of *container.Manager this adapter needs —
// narrowed so tests can inject a fake without a real Docker daemon.
type sandbox interface {
	Start(ctx context.Context, p container.StartParams) (*container.Handle, error)
	Status(ctx context.Context, h *container.Handle) (container.Status, error)
	Exec(ctx context.Context, h *container.Handle, cmd []string) (string, error)
	Stop(ctx context.Context, h *container.Handle) error
}

// Executor implements platform.Executor (ADR-005) by running Claude Code
// inside an isolated Docker container (ADR-006) via container.Manager
// (TASK-054). One Executor value serves exactly one Execution's
// Accept -> Finish lifecycle; a fresh value is constructed per Execution.
type Executor struct {
	sandbox        sandbox
	gitToken       string
	providerAPIKey string

	executionID string
	handle      *container.Handle
}

var _ platform.Executor = (*Executor)(nil)

// New creates an Executor that runs the given execution image
// (docker/execution, TASK-053) and authenticates git and the AI provider
// with the given short-lived credentials (ADR-006). Passing an empty
// providerAPIKey is valid — Accept will still start the sandbox, useful
// for exercising the container lifecycle without a real AI-provider call
// (see TASK-056's Open Question on credential availability).
func New(image, gitToken, providerAPIKey string) (*Executor, error) {
	id, err := randomID()
	if err != nil {
		return nil, fmt.Errorf("claudecode: generate execution id: %w", err)
	}
	return &Executor{
		sandbox:        container.NewManager(image),
		gitToken:       gitToken,
		providerAPIKey: providerAPIKey,
		executionID:    id,
	}, nil
}

// Accept implements platform.Executor: it starts the sandbox, clones the
// task's branch and launches Claude Code non-interactively against a
// prompt built from the task's planning content.
func (e *Executor) Accept(ctx context.Context, task platform.ExecutorTask) error {
	h, err := e.sandbox.Start(ctx, container.StartParams{
		ExecutionID:    e.executionID,
		Repository:     task.Repository,
		Branch:         task.Branch,
		GitToken:       e.gitToken,
		ProviderAPIKey: e.providerAPIKey,
		Allowlist:      []string{"api.anthropic.com"},
		Command:        claudeCommand(task),
	})
	if err != nil {
		return fmt.Errorf("claudecode: accept task %s: %w", task.TaskID, err)
	}
	e.handle = h
	return nil
}

// Artifacts implements platform.Executor: it reports the commits Claude
// Code produced on the task branch. Turning that into a Pull Request is
// the calling application service's job (ResultService/CompletionService,
// EPIC-004, via platform.RepositoryProvider) — this adapter only reports
// what happened inside its own sandbox.
func (e *Executor) Artifacts(ctx context.Context) ([]platform.Artifact, error) {
	if e.handle == nil {
		return nil, ErrNotAccepted
	}
	out, err := e.sandbox.Exec(ctx, e.handle, []string{"git", "log", "--format=%H%n%s", "-n", "20"})
	if err != nil {
		return nil, fmt.Errorf("claudecode: list artifacts: %w", err)
	}
	return parseCommitArtifacts(out), nil
}

// Status implements platform.Executor.
func (e *Executor) Status(ctx context.Context) (platform.ExecutionStatus, error) {
	if e.handle == nil {
		return platform.ExecutionStatus{}, ErrNotAccepted
	}
	status, err := e.sandbox.Status(ctx, e.handle)
	if err != nil {
		return platform.ExecutionStatus{}, fmt.Errorf("claudecode: status: %w", err)
	}
	if status.Running {
		return platform.ExecutionStatus{State: "running"}, nil
	}
	if status.ExitCode == 0 {
		return platform.ExecutionStatus{State: "succeeded"}, nil
	}
	return platform.ExecutionStatus{State: "failed", Message: fmt.Sprintf("exit code %d", status.ExitCode)}, nil
}

// Finish implements platform.Executor: it tears down the sandbox — the
// ephemeral working copy dies with the container (ADR-006).
func (e *Executor) Finish(ctx context.Context) error {
	if e.handle == nil {
		return ErrNotAccepted
	}
	return e.sandbox.Stop(ctx, e.handle)
}

func randomID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
