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
