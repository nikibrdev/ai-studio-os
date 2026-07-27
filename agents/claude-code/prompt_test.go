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

// Roles without instructions must not silently inherit another role's block.
// PM and QA gained instructions in EPIC-013 (TASK-097) and are no longer here;
// Architect still has none, and neither does anything unrecognised.
func TestBuildPrompt_UnknownRolesAreRejected(t *testing.T) {
	for _, role := range []string{
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

// Every dispatched role must get its own instructions — and none of them may be
// told to do another role's job. The checkpoint roles in particular must never
// be told to change a task's state (docs/architecture/workflow.md).
func TestBuildPrompt_EachDispatchedRoleGetsOwnInstructions(t *testing.T) {
	cases := []struct {
		role        shared.Role
		mustContain []string
		mustNotHave []string
	}{
		{
			role:        shared.RoleDeveloper,
			mustContain: []string{"Закоммить"},
			mustNotHave: []string{container.ProposalFile, container.ReportFile, container.VerdictFile},
		},
		{
			role:        shared.RoleReviewer,
			mustContain: []string{container.VerdictFile, "Ничего не коммить"},
			mustNotHave: []string{container.ProposalFile, container.ReportFile},
		},
		{
			role:        shared.RoleProjectManager,
			mustContain: []string{container.ProposalFile, "не менять состояние задачи"},
			mustNotHave: []string{"Закоммить", container.VerdictFile, container.ReportFile},
		},
		{
			role:        shared.RoleQA,
			mustContain: []string{container.ReportFile, "не менять состояние задачи"},
			mustNotHave: []string{"Закоммить", container.VerdictFile, container.ProposalFile},
		},
	}

	for _, tc := range cases {
		t.Run(string(tc.role), func(t *testing.T) {
			prompt, err := buildPrompt(platform.ExecutorTask{
				Role: string(tc.role), Title: "T", Type: "feature",
			})
			if err != nil {
				t.Fatalf("buildPrompt: %v", err)
			}
			for _, want := range tc.mustContain {
				if !strings.Contains(prompt, want) {
					t.Errorf("prompt for %s does not contain %q", tc.role, want)
				}
			}
			for _, unwanted := range tc.mustNotHave {
				if strings.Contains(prompt, unwanted) {
					t.Errorf("prompt for %s contains %q, which belongs to another role", tc.role, unwanted)
				}
			}
			// No role talks to the platform's API: that is what keeps the human
			// checkpoints structurally out of an agent's reach (EPIC-013).
			if strings.Contains(prompt, "apps/api") || strings.Contains(prompt, "localhost:8080") {
				t.Errorf("prompt for %s mentions the platform API; agents must not reach it", tc.role)
			}
		})
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
	if _, err := claudeCommand(platform.ExecutorTask{Role: "nonsense", Title: "T"}); !errors.Is(err, ErrUnsupportedRole) {
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
