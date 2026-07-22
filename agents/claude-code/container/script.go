package container

import "strings"

// exitCodeFile is where cloneAndRunScript records the clone-and-command
// exit code inside the container, and how Status reads it back.
const exitCodeFile = "/tmp/ai-studio-os-exit-code"

// workspaceDir is where cloneAndRunScript clones the task's branch, and
// the working directory Exec runs commands in (the container's own
// WORKDIR, set by docker/execution's Dockerfile, is its parent — cloning
// creates this subdirectory, it does not already exist in the image).
const workspaceDir = "/workspace/repo"

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
func cloneAndRunScript(repository, branch string, command []string) string {
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
		shellQuote("https://x-access-token@github.com/"+repository+".git") + " " + workspaceDir + " &&\n")
	b.WriteString("  cd " + workspaceDir + " &&\n")
	b.WriteString("  " + strings.Join(quoted, " ") + "\n")
	b.WriteString(")\n")
	b.WriteString("echo $? > " + exitCodeFile + "\n")
	b.WriteString("sleep 300\n")
	return b.String()
}

// shellQuote wraps s in single quotes, safe for embedding in a POSIX
// shell script regardless of its content.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
