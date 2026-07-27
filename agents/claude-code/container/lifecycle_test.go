package container

import (
	"context"
	"errors"
	"os"
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

	// beforeRun observes a call while it is being made — needed for state
	// that does not outlive it, such as the secrets env file, which Start
	// removes as soon as `docker run` returns.
	beforeRun func(name string, args ...string)
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{responses: map[string]string{}, errors: map[string]error{}}
}

func (f *fakeRunner) key(name string, args ...string) string {
	return name + " " + strings.Join(args, " ")
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	f.calls = append(f.calls, fakeCall{name: name, args: args})
	if f.beforeRun != nil {
		f.beforeRun(name, args...)
	}
	key := f.key(name, args...)
	if err, ok := f.errors[key]; ok {
		return "", err
	}
	return f.responses[key], nil
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

// Secret values must never appear in the arguments of any command the
// package runs. Two paths depend on it: the host's process list, which
// any local user can read, and execRunner's error text, which embeds the
// whole argument list and is what ends up in logs (TASK-106).
//
// The assertion is deliberately made over *every* call rather than the
// container start alone: a secret leaking through the proxy or network
// commands would be the same defect somewhere else.
func TestStart_SecretValuesNeverAppearInArgv(t *testing.T) {
	run := newFakeRunner()
	m := newTestManager(run)

	const (
		token = "super-secret-token"
		key   = "super-secret-key"
	)

	if _, err := m.Start(context.Background(), StartParams{
		ExecutionID: "exec-1", Repository: "org/repo", Branch: "main",
		GitToken: token, ProviderAPIKey: key,
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}

	for _, call := range run.calls {
		joined := strings.Join(call.args, " ")
		if strings.Contains(joined, token) {
			t.Errorf("git token leaked into argv of %q: %v", call.name, call.args)
		}
		if strings.Contains(joined, key) {
			t.Errorf("provider key leaked into argv of %q: %v", call.name, call.args)
		}
	}
}

// The secrets still have to reach the container — the point is the
// delivery mechanism, not dropping them.
func TestStart_SecretsDeliveredThroughEnvFile(t *testing.T) {
	run := newFakeRunner()
	m := newTestManager(run)

	var envFile string
	run.beforeRun = func(name string, args ...string) {
		if name != "docker" || len(args) < 2 || args[0] != "run" {
			return
		}
		for i, a := range args {
			if a == "--env-file" && i+1 < len(args) {
				// Read while the file still exists: Start removes it as
				// soon as `docker run` returns.
				envFile = args[i+1]
				content, err := os.ReadFile(envFile)
				if err != nil {
					t.Errorf("read env file: %v", err)
					return
				}
				got := string(content)
				if !strings.Contains(got, "GIT_TOKEN=super-secret-token\n") {
					t.Errorf("env file missing git token, got:\n%s", got)
				}
				if !strings.Contains(got, "ANTHROPIC_API_KEY=super-secret-key\n") {
					t.Errorf("env file missing provider key, got:\n%s", got)
				}
			}
		}
	}

	if _, err := m.Start(context.Background(), StartParams{
		ExecutionID: "exec-1", Repository: "org/repo", Branch: "main",
		GitToken: "super-secret-token", ProviderAPIKey: "super-secret-key",
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if envFile == "" {
		t.Fatal("expected the container start to pass --env-file")
	}
	if _, err := os.Stat(envFile); !os.IsNotExist(err) {
		t.Errorf("env file %q must not outlive the container start (stat err = %v)", envFile, err)
	}
}

// An absent provider key is how the sandbox reports "no key configured"
// (TASK-056); writing it as an empty value would look like a configured
// one.
func TestStart_EmptyProviderKeyIsOmittedFromEnvFile(t *testing.T) {
	run := newFakeRunner()
	m := newTestManager(run)

	run.beforeRun = func(name string, args ...string) {
		if name != "docker" || len(args) < 2 || args[0] != "run" {
			return
		}
		for i, a := range args {
			if a == "--env-file" && i+1 < len(args) {
				content, err := os.ReadFile(args[i+1])
				if err != nil {
					t.Errorf("read env file: %v", err)
					return
				}
				if strings.Contains(string(content), "ANTHROPIC_API_KEY") {
					t.Errorf("empty provider key must be omitted, got:\n%s", content)
				}
			}
		}
	}

	if _, err := m.Start(context.Background(), StartParams{
		ExecutionID: "exec-1", Repository: "org/repo", Branch: "main",
		GitToken: "tok",
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
}

// A line break would silently split one credential into a second,
// malformed entry in docker's --env-file format.
func TestWriteSecretsEnvFile_RejectsLineBreakInValue(t *testing.T) {
	if _, err := writeSecretsEnvFile("probe", map[string]string{"GIT_TOKEN": "abc\ndef"}); err == nil {
		t.Fatal("expected a secret containing a line break to be refused")
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
