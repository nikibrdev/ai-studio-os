package claudecode

import (
	"strings"
	"testing"

	"ai-studio-os/internal/platform"
)

func TestBuildPrompt_IncludesTaskContent(t *testing.T) {
	task := platform.ExecutorTask{
		Role: "developer", Title: "Заголовок", Type: "feature",
		Scope: "Сделать нечто полезное", AcceptanceCriteria: []string{"критерий раз", "критерий два"},
	}
	prompt := buildPrompt(task)

	for _, want := range []string{"developer", "Заголовок", "feature", "Сделать нечто полезное", "критерий раз", "критерий два"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestBuildPrompt_OmitsEmptyScopeAndCriteria(t *testing.T) {
	prompt := buildPrompt(platform.ExecutorTask{Role: "developer", Title: "T", Type: "feature"})
	if strings.Contains(prompt, "Цель и объём") {
		t.Errorf("prompt should omit the scope line when Scope is empty:\n%s", prompt)
	}
	if strings.Contains(prompt, "Критерии приёмки") {
		t.Errorf("prompt should omit the criteria section when empty:\n%s", prompt)
	}
}

func TestClaudeCommand_UsesNonInteractiveFlags(t *testing.T) {
	cmd := claudeCommand(platform.ExecutorTask{Title: "T", Type: "feature"})
	joined := strings.Join(cmd, " ")
	if !strings.Contains(joined, "--print") || !strings.Contains(joined, "--permission-mode bypassPermissions") {
		t.Errorf("claudeCommand() = %v, want --print and --permission-mode bypassPermissions", cmd)
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
