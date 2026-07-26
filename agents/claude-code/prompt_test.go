package claudecode

import (
	"errors"
	"strings"
	"testing"

	"ai-studio-os/agents/claude-code/container"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// mustPrompt builds a prompt the test expects to succeed.
func mustPrompt(t *testing.T, task platform.ExecutorTask) string {
	t.Helper()
	prompt, err := buildPrompt(task)
	if err != nil {
		t.Fatalf("buildPrompt: %v", err)
	}
	return prompt
}

func TestBuildPrompt_IncludesTaskContent(t *testing.T) {
	task := platform.ExecutorTask{
		Role: "developer", Title: "Заголовок", Type: "feature",
		Scope: "Сделать нечто полезное", AcceptanceCriteria: []string{"критерий раз", "критерий два"},
	}
	prompt := mustPrompt(t, task)

	for _, want := range []string{"developer", "Заголовок", "feature", "Сделать нечто полезное", "критерий раз", "критерий два"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestBuildPrompt_OmitsEmptyScopeAndCriteria(t *testing.T) {
	prompt := mustPrompt(t, platform.ExecutorTask{Role: "developer", Title: "T", Type: "feature"})
	if strings.Contains(prompt, "Цель и объём") {
		t.Errorf("prompt should omit the scope line when Scope is empty:\n%s", prompt)
	}
	if strings.Contains(prompt, "Критерии приёмки") {
		t.Errorf("prompt should omit the criteria section when empty:\n%s", prompt)
	}
}

// The whole point of TASK-087: one adapter, different instructions per role.
func TestBuildPrompt_InstructionsDifferByRole(t *testing.T) {
	dev := mustPrompt(t, platform.ExecutorTask{Role: string(shared.RoleDeveloper), Title: "T", Type: "feature"})
	rev := mustPrompt(t, platform.ExecutorTask{Role: string(shared.RoleReviewer), Title: "T", Type: "feature"})

	if dev == rev {
		t.Fatal("Developer and Reviewer got identical prompts; the role must change the instructions")
	}

	// A reviewer must not be told to commit — it produces a decision, not code.
	if strings.Contains(rev, "Закоммить") {
		t.Errorf("reviewer prompt tells the agent to commit:\n%s", rev)
	}
	if !strings.Contains(rev, container.VerdictFile) {
		t.Errorf("reviewer prompt does not say where to write the verdict:\n%s", rev)
	}
	if !strings.Contains(rev, verdictApproved) || !strings.Contains(rev, verdictChangesRequested) {
		t.Errorf("reviewer prompt does not state both accepted words:\n%s", rev)
	}
	// The base is origin/main, not the clone's HEAD: for a reviewer the clone
	// already contains the developer's commits.
	if !strings.Contains(rev, "origin/"+baseBranchName) {
		t.Errorf("reviewer prompt does not compare against origin/%s:\n%s", baseBranchName, rev)
	}

	// A developer must not be handed the verdict convention.
	if strings.Contains(dev, container.VerdictFile) {
		t.Errorf("developer prompt mentions the verdict file:\n%s", dev)
	}
	if !strings.Contains(dev, "Закоммить") {
		t.Errorf("developer prompt does not tell the agent to commit:\n%s", dev)
	}
}

// PM and QA are deliberately not dispatched yet, so they have no
// instructions — and must not silently inherit the Developer block.
func TestBuildPrompt_UndispatchedAndUnknownRolesAreRejected(t *testing.T) {
	for _, role := range []string{
		string(shared.RoleProjectManager),
		string(shared.RoleQA),
		string(shared.RoleArchitect),
		"",
		"nonsense",
	} {
		_, err := buildPrompt(platform.ExecutorTask{Role: role, Title: "T", Type: "feature"})
		if !errors.Is(err, ErrUnsupportedRole) {
			t.Errorf("buildPrompt(role=%q) error = %v, want %v", role, err, ErrUnsupportedRole)
		}
	}
}

func TestClaudeCommand_UsesNonInteractiveFlags(t *testing.T) {
	cmd, err := claudeCommand(platform.ExecutorTask{Role: "developer", Title: "T", Type: "feature"})
	if err != nil {
		t.Fatalf("claudeCommand: %v", err)
	}
	joined := strings.Join(cmd, " ")
	if !strings.Contains(joined, "--print") || !strings.Contains(joined, "--permission-mode bypassPermissions") {
		t.Errorf("claudeCommand() = %v, want --print and --permission-mode bypassPermissions", cmd)
	}
}

func TestClaudeCommand_PropagatesRoleError(t *testing.T) {
	if _, err := claudeCommand(platform.ExecutorTask{Role: "qa", Title: "T"}); !errors.Is(err, ErrUnsupportedRole) {
		t.Fatalf("claudeCommand() error = %v, want %v", err, ErrUnsupportedRole)
	}
}

func TestParseCommitArtifacts_Empty(t *testing.T) {
	if got := parseCommitArtifacts("  \n  "); got != nil {
		t.Errorf("parseCommitArtifacts(blank) = %v, want nil", got)
	}
}

func TestParseCommitArtifacts_IgnoresTrailingIncompleteLine(t *testing.T) {
	got := parseCommitArtifacts("abc123\nfeat: x\ndangling-hash-with-no-subject")
	if len(got) != 1 {
		t.Fatalf("parseCommitArtifacts() = %d entries, want 1: %+v", len(got), got)
	}
}
