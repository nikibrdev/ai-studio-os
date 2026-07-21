package claudecode

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/agents/claude-code/container"
	"ai-studio-os/internal/platform"
)

type fakeSandbox struct {
	startParams container.StartParams
	startErr    error
	handle      *container.Handle

	status    container.Status
	statusErr error

	execOut string
	execErr error

	stopErr   error
	stopCalls int
}

func (f *fakeSandbox) Start(_ context.Context, p container.StartParams) (*container.Handle, error) {
	f.startParams = p
	if f.startErr != nil {
		return nil, f.startErr
	}
	if f.handle == nil {
		f.handle = &container.Handle{}
	}
	return f.handle, nil
}

func (f *fakeSandbox) Status(_ context.Context, _ *container.Handle) (container.Status, error) {
	return f.status, f.statusErr
}

func (f *fakeSandbox) Exec(_ context.Context, _ *container.Handle, _ []string) (string, error) {
	return f.execOut, f.execErr
}

func (f *fakeSandbox) Stop(_ context.Context, _ *container.Handle) error {
	f.stopCalls++
	return f.stopErr
}

func newTestExecutor(sb *fakeSandbox) *Executor {
	return &Executor{sandbox: sb, gitToken: "tok", providerAPIKey: "key", executionID: "exec-1"}
}

func TestAccept_StartsSandboxWithTaskAndAllowlist(t *testing.T) {
	sb := &fakeSandbox{}
	e := newTestExecutor(sb)

	task := platform.ExecutorTask{
		TaskID: "task-1", Role: "developer", Title: "Заголовок", Type: "feature",
		Scope: "Сделать нечто полезное", AcceptanceCriteria: []string{"работает"},
		Repository: "org/repo", Branch: "feature/x",
	}
	if err := e.Accept(context.Background(), task); err != nil {
		t.Fatalf("Accept: %v", err)
	}

	if sb.startParams.Repository != "org/repo" || sb.startParams.Branch != "feature/x" {
		t.Errorf("Start() params = %+v, want repo/branch from task", sb.startParams)
	}
	if sb.startParams.GitToken != "tok" || sb.startParams.ProviderAPIKey != "key" {
		t.Errorf("Start() params did not carry the executor's credentials: %+v", sb.startParams)
	}
	found := false
	for _, h := range sb.startParams.Allowlist {
		if h == "api.anthropic.com" {
			found = true
		}
	}
	if !found {
		t.Errorf("Start() allowlist = %v, want api.anthropic.com included", sb.startParams.Allowlist)
	}
	if len(sb.startParams.Command) == 0 || sb.startParams.Command[0] != "claude" {
		t.Errorf("Start() command = %v, want it to invoke claude", sb.startParams.Command)
	}
}

func TestAccept_PropagatesSandboxError(t *testing.T) {
	wantErr := errors.New("docker unavailable")
	sb := &fakeSandbox{startErr: wantErr}
	e := newTestExecutor(sb)

	err := e.Accept(context.Background(), platform.ExecutorTask{TaskID: "task-1"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Accept() error = %v, want wrapping %v", err, wantErr)
	}
}

func TestArtifacts_BeforeAccept_ReturnsError(t *testing.T) {
	e := newTestExecutor(&fakeSandbox{})
	if _, err := e.Artifacts(context.Background()); !errors.Is(err, ErrNotAccepted) {
		t.Fatalf("Artifacts() before Accept: error = %v, want ErrNotAccepted", err)
	}
}

func TestStatus_BeforeAccept_ReturnsError(t *testing.T) {
	e := newTestExecutor(&fakeSandbox{})
	if _, err := e.Status(context.Background()); !errors.Is(err, ErrNotAccepted) {
		t.Fatalf("Status() before Accept: error = %v, want ErrNotAccepted", err)
	}
}

func TestFinish_BeforeAccept_ReturnsError(t *testing.T) {
	e := newTestExecutor(&fakeSandbox{})
	if err := e.Finish(context.Background()); !errors.Is(err, ErrNotAccepted) {
		t.Fatalf("Finish() before Accept: error = %v, want ErrNotAccepted", err)
	}
}

func TestArtifacts_ParsesCommitsFromGitLog(t *testing.T) {
	sb := &fakeSandbox{execOut: "abc123\nfeat: add thing\ndef456\nfix: correct bug\n"}
	e := newTestExecutor(sb)
	if err := e.Accept(context.Background(), platform.ExecutorTask{TaskID: "task-1"}); err != nil {
		t.Fatalf("Accept: %v", err)
	}

	artifacts, err := e.Artifacts(context.Background())
	if err != nil {
		t.Fatalf("Artifacts: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("Artifacts() = %d entries, want 2: %+v", len(artifacts), artifacts)
	}
	if artifacts[0].ID != "abc123" || artifacts[0].Type != "Commit" || string(artifacts[0].Payload) != "feat: add thing" {
		t.Errorf("Artifacts()[0] = %+v", artifacts[0])
	}
}

func TestStatus_MapsRunningSucceededFailed(t *testing.T) {
	tests := []struct {
		name   string
		status container.Status
		want   string
	}{
		{"running", container.Status{Running: true}, "running"},
		{"succeeded", container.Status{Running: false, ExitCode: 0}, "succeeded"},
		{"failed", container.Status{Running: false, ExitCode: 1}, "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &fakeSandbox{status: tt.status}
			e := newTestExecutor(sb)
			if err := e.Accept(context.Background(), platform.ExecutorTask{TaskID: "task-1"}); err != nil {
				t.Fatalf("Accept: %v", err)
			}

			got, err := e.Status(context.Background())
			if err != nil {
				t.Fatalf("Status: %v", err)
			}
			if got.State != tt.want {
				t.Errorf("Status().State = %q, want %q", got.State, tt.want)
			}
		})
	}
}

func TestFinish_StopsSandbox(t *testing.T) {
	sb := &fakeSandbox{}
	e := newTestExecutor(sb)
	if err := e.Accept(context.Background(), platform.ExecutorTask{TaskID: "task-1"}); err != nil {
		t.Fatalf("Accept: %v", err)
	}

	if err := e.Finish(context.Background()); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if sb.stopCalls != 1 {
		t.Errorf("Stop called %d times, want 1", sb.stopCalls)
	}
}

func TestNew_GeneratesUniqueExecutionIDs(t *testing.T) {
	e1, err := New("image", "tok", "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	e2, err := New("image", "tok", "key")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if e1.executionID == "" || e1.executionID == e2.executionID {
		t.Errorf("executionID = %q and %q, want distinct non-empty values", e1.executionID, e2.executionID)
	}
}
