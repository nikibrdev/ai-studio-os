package container

import (
	"strings"
	"testing"
)

func TestRenderSquidConf_AllowsOnlyGivenDomains(t *testing.T) {
	conf := renderSquidConf([]string{"github.com", "api.anthropic.com"})

	if !strings.Contains(conf, "dstdomain .github.com .api.anthropic.com") {
		t.Errorf("expected both domains in the allowed_domains ACL, got:\n%s", conf)
	}
	if !strings.Contains(conf, "http_access deny all") {
		t.Errorf("expected a trailing deny-all rule, got:\n%s", conf)
	}

	denyIdx := strings.Index(conf, "http_access deny all")
	allowIdx := strings.Index(conf, "http_access allow allowed_domains")
	if denyIdx < allowIdx {
		t.Errorf("deny all must come after the allow rules, got:\n%s", conf)
	}
}

func TestRenderSquidConf_NormalizesLeadingDot(t *testing.T) {
	conf := renderSquidConf([]string{".github.com", "github.com"})
	if !strings.Contains(conf, "dstdomain .github.com .github.com") {
		t.Errorf("expected normalized single leading dot per entry, got:\n%s", conf)
	}
}

func TestCloneAndRunScript_TokenOnlyViaEnvNotArgv(t *testing.T) {
	script := cloneAndRunScript("org/repo", "feature/x", []string{"echo", "hi"})

	if strings.Contains(script, "GIT_TOKEN=") && !strings.Contains(script, `echo "$GIT_TOKEN"`) {
		t.Errorf("script should only reference $GIT_TOKEN via the askpass helper, got:\n%s", script)
	}
	if !strings.Contains(script, "GIT_ASKPASS=/tmp/git-askpass.sh") {
		t.Errorf("expected GIT_ASKPASS to be configured, got:\n%s", script)
	}
	if !strings.Contains(script, "https://x-access-token@github.com/org/repo.git") {
		t.Errorf("expected clone URL without an embedded token, got:\n%s", script)
	}
	if !strings.Contains(script, "--branch 'feature/x'") {
		t.Errorf("expected branch to be passed and quoted, got:\n%s", script)
	}
	if !strings.Contains(script, "'echo' 'hi'") {
		t.Errorf("expected trailing command with each argument individually quoted, got:\n%s", script)
	}
}

func TestCloneAndRunScript_CapturesExitCodeAndStaysAlive(t *testing.T) {
	script := cloneAndRunScript("org/repo", "main", []string{"false"})

	if !strings.Contains(script, "echo $? > "+exitCodeFile) {
		t.Errorf("expected the exit code to be captured to %s, got:\n%s", exitCodeFile, script)
	}
	if !strings.Contains(script, "sleep 300") {
		t.Errorf("expected the container to idle after the command finishes, got:\n%s", script)
	}

	captureIdx := strings.Index(script, "echo $? > "+exitCodeFile)
	sleepIdx := strings.Index(script, "sleep 300")
	if captureIdx < 0 || sleepIdx < 0 || sleepIdx < captureIdx {
		t.Errorf("exit code must be captured before the idle sleep, got:\n%s", script)
	}
}

func TestShellQuote_EscapesSingleQuotes(t *testing.T) {
	got := shellQuote(`it's "quoted"`)
	want := `'it'\''s "quoted"'`
	if got != want {
		t.Errorf("shellQuote() = %s, want %s", got, want)
	}
}
