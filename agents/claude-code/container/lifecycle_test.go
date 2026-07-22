package container

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeCall struct {
	name string
	args []string
}

// fakeRunner records every call and looks up canned responses by the
// full "name arg1 arg2..." string; a call with no match returns errNoMatch
// unless a default success response is set.
type fakeRunner struct {
	calls     []fakeCall
	responses map[string]string
	errors    map[string]error
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{responses: map[string]string{}, errors: map[string]error{}}
}

func (f *fakeRunner) key(name string, args ...string) string {
	return name + " " + strings.Join(args, " ")
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	f.calls = append(f.calls, fakeCall{name: name, args: args})
	key := f.key(name, args...)
	if err, ok := f.errors[key]; ok {
		return "", err
	}
	return f.responses[key], nil
}

func (f *fakeRunner) lastCall() fakeCall {
	if len(f.calls) == 0 {
		return fakeCall{}
	}
	return f.calls[len(f.calls)-1]
}

func (f *fakeRunner) callCount(name string, argPrefix ...string) int {
	n := 0
	for _, c := range f.calls {
		if c.name != name {
			continue
		}
		if len(c.args) < len(argPrefix) {
			continue
		}
		match := true
		for i, a := range argPrefix {
			if c.args[i] != a {
				match = false
				break
			}
		}
		if match {
			n++
		}
	}
	return n
}

func newTestManager(run *fakeRunner) *Manager {
	return &Manager{run: run, image: "ai-studio-os-execution"}
}

func TestStart_CreatesNetworkProxyAndContainer(t *testing.T) {
	run := newFakeRunner()
	run.errors["docker network inspect ai-studio-os-net-exec-1"] = errors.New("network not found")
	m := newTestManager(run)

	h, err := m.Start(context.Background(), StartParams{
		ExecutionID: "exec-1", Repository: "org/repo", Branch: "feature/x",
		GitToken: "tok", Command: []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if h.containerName != "ai-studio-os-exec-exec-1" {
		t.Errorf("containerName = %q", h.containerName)
	}

	if run.callCount("docker", "network", "create") != 1 {
		t.Errorf("expected exactly one network create call, calls: %+v", run.calls)
	}
	if run.callCount("docker", "run", "-d", "--name", "ai-studio-os-proxy-exec-1") != 1 {
		t.Errorf("expected proxy container start, calls: %+v", run.calls)
	}
	if run.callCount("docker", "network", "connect", "bridge", "ai-studio-os-proxy-exec-1") != 1 {
		t.Errorf("expected proxy connected to public network, calls: %+v", run.calls)
	}
	if run.callCount("docker", "run", "-d", "--name", "ai-studio-os-exec-exec-1") != 1 {
		t.Errorf("expected execution container start, calls: %+v", run.calls)
	}
}

func TestStart_ReusesExistingNetwork(t *testing.T) {
	run := newFakeRunner()
	run.responses["docker network inspect ai-studio-os-net-exec-1"] = "[...]"
	m := newTestManager(run)

	if _, err := m.Start(context.Background(), StartParams{
		ExecutionID: "exec-1", Repository: "org/repo", Branch: "main", GitToken: "tok",
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if run.callCount("docker", "network", "create") != 0 {
		t.Errorf("expected no network create call when network already exists, calls: %+v", run.calls)
	}
}

func TestStart_ExecutionContainerNeverJoinsPublicNetwork(t *testing.T) {
	run := newFakeRunner()
	m := newTestManager(run)

	if _, err := m.Start(context.Background(), StartParams{
		ExecutionID: "exec-1", Repository: "org/repo", Branch: "main", GitToken: "tok",
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}

	for _, c := range run.calls {
		if c.name != "docker" || len(c.args) < 2 || c.args[0] != "network" || c.args[1] != "connect" {
			continue
		}
		if len(c.args) >= 4 && c.args[3] == "ai-studio-os-exec-exec-1" {
			t.Fatalf("execution container must never be connected to the public network directly: %v", c.args)
		}
	}
}

func TestStart_SecretsPassedAsEnvNotArgvValue(t *testing.T) {
	run := newFakeRunner()
	m := newTestManager(run)

	if _, err := m.Start(context.Background(), StartParams{
		ExecutionID: "exec-1", Repository: "org/repo", Branch: "main",
		GitToken: "super-secret-token", ProviderAPIKey: "super-secret-key",
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}

	call := run.lastCall()
	joined := strings.Join(call.args, " ")
	if !strings.Contains(joined, "GIT_TOKEN=super-secret-token") {
		t.Errorf("expected GIT_TOKEN passed via -e, args: %v", call.args)
	}
	if !strings.Contains(joined, "ANTHROPIC_API_KEY=super-secret-key") {
		t.Errorf("expected ANTHROPIC_API_KEY passed via -e, args: %v", call.args)
	}
	// The clone script must read the token from the environment
	// (GIT_ASKPASS), not embed it directly in an argv-visible URL or flag.
	if strings.Contains(joined, "super-secret-token@github.com") {
		t.Errorf("git token must not be embedded directly in the clone URL: %v", call.args)
	}
}

func TestStart_NetworkFailureStopsBeforeStartingContainers(t *testing.T) {
	run := newFakeRunner()
	run.errors["docker network inspect ai-studio-os-net-exec-1"] = errors.New("not found")
	run.errors["docker network create --internal ai-studio-os-net-exec-1"] = errors.New("boom")
	m := newTestManager(run)

	_, err := m.Start(context.Background(), StartParams{ExecutionID: "exec-1", Repository: "org/repo", Branch: "main"})
	if err == nil {
		t.Fatal("expected error when network creation fails")
	}
	if run.callCount("docker", "run") != 0 {
		t.Errorf("no containers should start after network creation fails, calls: %+v", run.calls)
	}
}

func TestStart_ProxyFailureCleansUpNetwork(t *testing.T) {
	run := newFakeRunner()
	run.errors["docker network inspect ai-studio-os-net-exec-1"] = errors.New("not found")
	run.errors["docker network connect bridge ai-studio-os-proxy-exec-1"] = errors.New("boom")
	m := newTestManager(run)

	_, err := m.Start(context.Background(), StartParams{ExecutionID: "exec-1", Repository: "org/repo", Branch: "main"})
	if err == nil {
		t.Fatal("expected error when proxy setup fails")
	}
	if run.callCount("docker", "network", "rm", "ai-studio-os-net-exec-1") != 1 {
		t.Errorf("expected network cleanup after proxy failure, calls: %+v", run.calls)
	}
}

func TestStatus_FinishedReadsExitCodeFile(t *testing.T) {
	run := newFakeRunner()
	run.responses["docker exec ai-studio-os-exec-exec-1 cat "+exitCodeFile] = "1"
	m := newTestManager(run)
	h := &Handle{containerName: "ai-studio-os-exec-exec-1"}

	status, err := m.Status(context.Background(), h)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Running || status.ExitCode != 1 {
		t.Errorf("Status() = %+v, want {Running:false ExitCode:1}", status)
	}
}

func TestStatus_StillRunningWhenExitCodeFileMissing(t *testing.T) {
	run := newFakeRunner()
	run.errors["docker exec ai-studio-os-exec-exec-1 cat "+exitCodeFile] = errors.New(
		"exit status 1: cat: " + exitCodeFile + ": No such file or directory",
	)
	m := newTestManager(run)
	h := &Handle{containerName: "ai-studio-os-exec-exec-1"}

	status, err := m.Status(context.Background(), h)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Running {
		t.Errorf("Status() = %+v, want Running:true while the exit code file does not exist yet", status)
	}
}

func TestStatus_MalformedExitCodeContentIsAnError(t *testing.T) {
	run := newFakeRunner()
	run.responses["docker exec ai-studio-os-exec-exec-1 cat "+exitCodeFile] = "garbage"
	m := newTestManager(run)
	h := &Handle{containerName: "ai-studio-os-exec-exec-1"}

	if _, err := m.Status(context.Background(), h); err == nil {
		t.Fatal("expected error for malformed exit code content")
	}
}

func TestStatus_RealDockerErrorPropagates(t *testing.T) {
	run := newFakeRunner()
	run.errors["docker exec ai-studio-os-exec-exec-1 cat "+exitCodeFile] = errors.New("Error: No such container: ai-studio-os-exec-exec-1")
	m := newTestManager(run)
	h := &Handle{containerName: "ai-studio-os-exec-exec-1"}

	if _, err := m.Status(context.Background(), h); err == nil {
		t.Fatal("expected error to propagate when the container itself is gone")
	}
}

func TestExec_RunsInsideNamedContainer(t *testing.T) {
	run := newFakeRunner()
	run.responses["docker exec --workdir "+workspaceDir+" ai-studio-os-exec-exec-1 git log --oneline"] = "abc123 commit"
	m := newTestManager(run)
	h := &Handle{containerName: "ai-studio-os-exec-exec-1"}

	out, err := m.Exec(context.Background(), h, []string{"git", "log", "--oneline"})
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if out != "abc123 commit" {
		t.Errorf("Exec() = %q", out)
	}
}

func TestStop_IsIdempotent(t *testing.T) {
	run := newFakeRunner()
	run.errors["docker rm -f ai-studio-os-exec-exec-1"] = errors.New("Error: No such container: ai-studio-os-exec-exec-1")
	run.errors["docker rm -f ai-studio-os-proxy-exec-1"] = errors.New("Error: No such container: ai-studio-os-proxy-exec-1")
	run.errors["docker network rm ai-studio-os-net-exec-1"] = errors.New("Error: network not found")
	m := newTestManager(run)
	h := &Handle{
		containerName: "ai-studio-os-exec-exec-1",
		proxyName:     "ai-studio-os-proxy-exec-1",
		networkName:   "ai-studio-os-net-exec-1",
	}

	if err := m.Stop(context.Background(), h); err != nil {
		t.Fatalf("Stop() on already-removed resources should be a no-op, got: %v", err)
	}
}

func TestStop_PropagatesRealErrors(t *testing.T) {
	run := newFakeRunner()
	run.errors["docker rm -f ai-studio-os-exec-exec-1"] = errors.New("permission denied")
	m := newTestManager(run)
	h := &Handle{containerName: "ai-studio-os-exec-exec-1"}

	if err := m.Stop(context.Background(), h); err == nil {
		t.Fatal("expected Stop to propagate a real (non-not-found) error")
	}
}
