package container

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidRepository is returned when a repository identifier cannot be
// reduced to GitHub's "owner/name" for the clone URL (BUGFIX-006).
// Reported before the container starts: previously a malformed identifier
// produced a nonsense clone URL and surfaced only as "exit code 128" from
// inside git, saying nothing about the cause.
var ErrInvalidRepository = errors.New("container: repository identifier must be owner/name, optionally prefixed with a host")

// hostPrefixes are the leading forms stripped from the platform's stored
// repository identifier before it becomes part of the clone URL.
var hostPrefixes = []string{"https://", "http://", "github.com/", "www.github.com/"}

// exitCodeFile is where cloneAndRunScript records the clone-and-command
// exit code inside the container, and how Status reads it back.
const exitCodeFile = "/tmp/ai-studio-os-exit-code"

// workspaceDir is where cloneAndRunScript clones the task's branch, and
// the working directory Exec runs commands in (the container's own
// WORKDIR, set by docker/execution's Dockerfile, is its parent — cloning
// creates this subdirectory, it does not already exist in the image).
const workspaceDir = "/workspace/repo"

// BaseSHAFile holds the commit the working copy was at immediately after
// cloning — the boundary between history the execution inherited and
// whatever it produces (BUGFIX-007). Exported because the adapter reads it
// back to ask git for that range only; without it, `git log` reported the
// branch's inherited commits as produced artifacts.
//
// Written inside the clone's && chain, so a failed clone leaves no file and
// the adapter can tell "nothing was produced" from "the clone never ran".
const BaseSHAFile = "/tmp/ai-studio-os-base-sha"

// VerdictFile is where a role that produces a decision rather than code
// writes it (Reviewer — TASK-087), and where the adapter reads it back.
//
// Deliberately in /tmp rather than the working copy: inside the repository
// it would show up in `git status` as an untracked file and could end up
// committed, putting a verdict about the branch into the branch itself.
// Exec runs with --workdir on the clone, but an absolute path reads fine.
//
// A verdict is not an Artifact and is not reported through Artifacts: the
// Executor contract's four capabilities (ADR-005) stay as they are.
const VerdictFile = "/tmp/ai-studio-os-verdict"

// cloneAndRunScript builds the shell script executed inside the
// execution container: set up a GIT_ASKPASS reading the token from the
// GIT_TOKEN environment variable (never placed in argv — unlike
// embedding it in the clone URL or a `-c http.extraHeader` flag, both of
// which would appear verbatim in the container's own process list),
// clone repository/branch, run command, then keep the container alive
// (idle) instead of exiting immediately.
//
// The idle period exists so Artifacts/Exec can still inspect the working
// copy after the command finishes: `docker exec` only works on a running
// container, and Docker considers a container's main process exiting the
// end of the container — without this, a fast-failing command (e.g.
// Claude Code erroring out immediately on a missing API key, discovered
// running TASK-056's live demo) leaves no window at all between "Status
// turns terminal" and "docker exec no longer works". The exit code is
// captured to exitCodeFile before the idle sleep so Status does not have
// to rely on Docker's own (now unreliable, since the container stays
// running either way) State.Running.
// repository arrives in the platform's canonical stored form,
// "<host>/<owner>/<name>" (platform.RepositoryProvider) — the clone URL
// already carries the host, so it is stripped here (BUGFIX-006). Returns an
// error rather than building a URL that cannot work.
func cloneAndRunScript(repository, branch string, command []string) (string, error) {
	repoPath, err := normalizeRepository(repository)
	if err != nil {
		return "", err
	}

	quoted := make([]string, len(command))
	for i, arg := range command {
		quoted[i] = shellQuote(arg)
	}

	var b strings.Builder
	b.WriteString("cat > /tmp/git-askpass.sh <<'ASKPASS'\n#!/bin/sh\necho \"$GIT_TOKEN\"\nASKPASS\n")
	b.WriteString("chmod +x /tmp/git-askpass.sh\n")
	b.WriteString("export GIT_ASKPASS=/tmp/git-askpass.sh GIT_TERMINAL_PROMPT=0\n")
	b.WriteString("(\n")
	b.WriteString("  git clone --branch " + shellQuote(branch) + " " +
		shellQuote("https://x-access-token@github.com/"+repoPath+".git") + " " + workspaceDir + " &&\n")
	b.WriteString("  cd " + workspaceDir + " &&\n")
	b.WriteString("  git rev-parse HEAD > " + BaseSHAFile + " &&\n")
	b.WriteString("  " + strings.Join(quoted, " ") + "\n")
	b.WriteString(")\n")
	b.WriteString("echo $? > " + exitCodeFile + "\n")
	b.WriteString("sleep 300\n")
	return b.String(), nil
}

// normalizeRepository reduces a repository identifier to "owner/name".
//
// The same translation internal/infrastructure/github performs for API
// paths (BUGFIX-005) is needed here for the clone URL: the platform stores
// the host, and both the URL template and the API path supply their own.
// Duplicating a few lines rather than sharing them is deliberate —
// agents/ may not import internal/ (module-boundaries.md), and a shared
// helper in pkg/ for two call sites would be a public contract invented
// ahead of need.
func normalizeRepository(repository string) (string, error) {
	s := strings.TrimSpace(repository)
	for changed := true; changed; {
		changed = false
		for _, prefix := range hostPrefixes {
			if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
				s, changed = s[len(prefix):], true
			}
		}
	}
	s = strings.TrimSuffix(strings.Trim(s, "/"), ".git")

	owner, name, found := strings.Cut(s, "/")
	if !found || owner == "" || name == "" || strings.Contains(name, "/") {
		return "", fmt.Errorf("%w: %q", ErrInvalidRepository, repository)
	}
	return owner + "/" + name, nil
}

// shellQuote wraps s in single quotes, safe for embedding in a POSIX
// shell script regardless of its content.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
