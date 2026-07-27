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
	"strings"

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
	// Built before the sandbox starts: an unsupported role must fail loudly
	// without spending a container (TASK-087).
	command, err := claudeCommand(task)
	if err != nil {
		return err
	}

	h, err := e.sandbox.Start(ctx, container.StartParams{
		ExecutionID:    e.executionID,
		Repository:     task.Repository,
		Branch:         task.Branch,
		GitToken:       e.gitToken,
		ProviderAPIKey: e.providerAPIKey,
		Allowlist:      []string{"api.anthropic.com"},
		Command:        command,
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

	// Only commits made after the clone count as produced. Asking git for
	// plain `git log` returned the branch's inherited history — on any
	// repository with commits, someone else's work was reported as produced
	// by this execution (BUGFIX-007, found by the first live end-to-end run).
	baseOut, err := e.sandbox.Exec(ctx, e.handle, []string{"cat", container.BaseSHAFile})
	if err != nil {
		return nil, fmt.Errorf("claudecode: read base commit: %w", err)
	}
	base := strings.TrimSpace(baseOut)
	if base == "" {
		return nil, fmt.Errorf("claudecode: base commit is empty — the clone did not complete")
	}

	// maxReportedCommits is a safety bound on output size, not the selection
	// itself: the range already limits results to this execution's work.
	out, err := e.sandbox.Exec(ctx, e.handle,
		[]string{"git", "log", "--format=%H%n%s", "-n", maxReportedCommits, base + "..HEAD"})
	if err != nil {
		return nil, fmt.Errorf("claudecode: list artifacts: %w", err)
	}
	return parseCommitArtifacts(out), nil
}

// maxReportedCommits caps how many produced commits are reported at once —
// a guard against a runaway agent, not a correctness mechanism.
const maxReportedCommits = "100"

// Verdict returns the decision a reviewing agent wrote (TASK-087).
//
// Not part of platform.Executor: ADR-005 fixes exactly four capabilities,
// and a verdict is neither an artifact nor a status. It is a method on this
// concrete adapter, reached by the caller that dispatched a reviewing role —
// the same shape as the platform's other "data from the agent, decision by
// the platform" seam (container.BaseSHAFile).
//
// Returns ErrNoVerdict when the agent wrote nothing and
// ErrUnrecognizedVerdict when it wrote something unreadable; both leave the
// decision to a human, and neither is turned into a default verdict.
func (e *Executor) Verdict(ctx context.Context) (Verdict, error) {
	if e.handle == nil {
		return Verdict{}, ErrNotAccepted
	}

	out, err := e.sandbox.Exec(ctx, e.handle, []string{"cat", container.VerdictFile})
	if err != nil {
		// A missing file surfaces as a failing `cat`, which is the ordinary
		// case of "the agent never wrote one" — not an infrastructure fault.
		return Verdict{}, fmt.Errorf("%w: %v", ErrNoVerdict, err)
	}
	return parseVerdict(out)
}

// Proposal reports the Definition of Ready a Project Manager agent prepared
// (EPIC-013). Like Verdict, this is a method of this adapter rather than a
// fifth capability of the Executor contract (ADR-005 fixes four): a proposal is
// neither an artifact nor a status, and the platform — not the agent — acts on
// it.
func (e *Executor) Proposal(ctx context.Context) (Proposal, error) {
	if e.handle == nil {
		return Proposal{}, ErrNotAccepted
	}

	out, err := e.sandbox.Exec(ctx, e.handle, []string{"cat", container.ProposalFile})
	if err != nil {
		return Proposal{}, fmt.Errorf("%w: %v", ErrNoProposal, err)
	}
	return parseProposal(out)
}

// Report reports what a QA agent found (EPIC-013). Returned as text: it is read
// by the human making the acceptance decision, not parsed by the platform.
func (e *Executor) Report(ctx context.Context) (string, error) {
	if e.handle == nil {
		return "", ErrNotAccepted
	}

	out, err := e.sandbox.Exec(ctx, e.handle, []string{"cat", container.ReportFile})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNoReport, err)
	}
	return parseReport(out)
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
