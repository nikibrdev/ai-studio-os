package container

import "strings"

// cloneAndRunScript builds the shell script executed inside the
// execution container: set up a GIT_ASKPASS reading the token from the
// GIT_TOKEN environment variable (never placed in argv — unlike
// embedding it in the clone URL or a `-c http.extraHeader` flag, both of
// which would appear verbatim in the container's own process list),
// clone repository/branch, then exec command.
func cloneAndRunScript(repository, branch string, command []string) string {
	var b strings.Builder
	b.WriteString("set -e\n")
	b.WriteString("cat > /tmp/git-askpass.sh <<'ASKPASS'\n#!/bin/sh\necho \"$GIT_TOKEN\"\nASKPASS\n")
	b.WriteString("chmod +x /tmp/git-askpass.sh\n")
	b.WriteString("export GIT_ASKPASS=/tmp/git-askpass.sh GIT_TERMINAL_PROMPT=0\n")
	b.WriteString("git clone --branch " + shellQuote(branch) + " " +
		shellQuote("https://x-access-token@github.com/"+repository+".git") + " /workspace/repo\n")
	b.WriteString("cd /workspace/repo\n")

	quoted := make([]string, len(command))
	for i, arg := range command {
		quoted[i] = shellQuote(arg)
	}
	b.WriteString(strings.Join(quoted, " ") + "\n")
	return b.String()
}

// shellQuote wraps s in single quotes, safe for embedding in a POSIX
// shell script regardless of its content.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
