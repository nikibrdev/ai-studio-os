//go:build integration

package container

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestManager_Start_ClonesAndEnforcesAllowlist requires a real Docker
// daemon and the image built by TASK-053 (docker/execution). It is
// excluded from the default `go test ./...` (and therefore from
// `make verify`) and only runs with `go test -tags=integration ./...`
// when TEST_DOCKER is set — the same opt-in pattern as
// internal/infrastructure's TEST_DATABASE_URL, since this test creates
// real Docker networks and containers as a side effect.
func TestManager_Start_ClonesAndEnforcesAllowlist(t *testing.T) {
	if os.Getenv("TEST_DOCKER") == "" {
		t.Skip("TEST_DOCKER not set; build docker/execution and set it to run this test")
	}

	m := NewManager("ai-studio-os-execution")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	h, err := m.Start(ctx, StartParams{
		ExecutionID: "it-" + time.Now().Format("150405"),
		Repository:  "nikibrdev/ai-studio-os",
		Branch:      "main",
		Command: []string{
			"sh", "-c",
			"git log -1 --oneline > /tmp/out.txt; " +
				"curl -s -o /dev/null -w 'api:%{http_code}\n' https://api.github.com >> /tmp/out.txt; " +
				"(curl -s -m 5 -o /dev/null -w 'blocked:%{http_code}\n' https://example.com >> /tmp/out.txt || echo 'blocked:refused' >> /tmp/out.txt); " +
				"sleep 30",
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		if err := m.Stop(context.Background(), h); err != nil {
			t.Errorf("Stop (cleanup): %v", err)
		}
	}()

	deadline := time.Now().Add(60 * time.Second)
	var out string
	for time.Now().Before(deadline) {
		out, err = m.Exec(ctx, h, []string{"cat", "/tmp/out.txt"})
		if err == nil && strings.Contains(out, "blocked:") {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	t.Logf("execution output:\n%s", out)

	if !strings.Contains(out, "api:200") {
		t.Errorf("expected the allowlisted host (api.github.com) to be reachable, got:\n%s", out)
	}
	if strings.Contains(out, "blocked:200") {
		t.Errorf("expected the non-allowlisted host (example.com) to be blocked, got:\n%s", out)
	}
}

// rootRefusal is what the Claude Code CLI prints instead of running when it is
// asked to bypass permissions as root. Asserted verbatim: this test exists
// because that refusal, and nothing else, silently stopped every role before
// any network call (BUGFIX-011).
const rootRefusal = "cannot be used with root/sudo privileges"

// TestManager_Start_RunsUnprivileged pins the fix for BUGFIX-011 to an
// observable property of a real container rather than to a line in the
// Dockerfile: the execution container must not run as root, and the CLI must
// get past its own root check.
//
// A unit test cannot cover this — the user is a property of the image, and the
// refusal comes from the CLI inside it. Same opt-in as the test above.
func TestManager_Start_RunsUnprivileged(t *testing.T) {
	if os.Getenv("TEST_DOCKER") == "" {
		t.Skip("TEST_DOCKER not set; build docker/execution and set it to run this test")
	}

	m := NewManager("ai-studio-os-execution")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	h, err := m.Start(ctx, StartParams{
		ExecutionID: "it-priv-" + time.Now().Format("150405"),
		Repository:  "nikibrdev/ai-studio-os",
		Branch:      "main",
		Allowlist:   []string{"api.anthropic.com"},
		Command: []string{
			"sh", "-c",
			"whoami > /tmp/priv.txt; " +
				// No key is configured, so the CLI is expected to stop at
				// authentication. That is the success condition here: reaching
				// authentication means the root check was passed.
				"claude --print --permission-mode bypassPermissions 'say hi' >> /tmp/priv.txt 2>&1; " +
				"echo done >> /tmp/priv.txt; sleep 30",
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		if err := m.Stop(context.Background(), h); err != nil {
			t.Errorf("Stop (cleanup): %v", err)
		}
	}()

	deadline := time.Now().Add(120 * time.Second)
	var out string
	for time.Now().Before(deadline) {
		out, err = m.Exec(ctx, h, []string{"cat", "/tmp/priv.txt"})
		if err == nil && strings.Contains(out, "done") {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	t.Logf("execution output:\n%s", out)

	if strings.HasPrefix(out, "root") {
		t.Errorf("execution container runs as root; the CLI refuses to bypass permissions there, got:\n%s", out)
	}
	if strings.Contains(out, rootRefusal) {
		t.Errorf("the CLI refused to run on root grounds, got:\n%s", out)
	}
}
