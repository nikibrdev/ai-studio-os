package claudecode

import (
	"fmt"
	"strings"

	"ai-studio-os/internal/platform"
)

// claudeCommand builds the Claude Code CLI invocation for task, run
// non-interactively inside the sandbox (container.Manager already placed
// the working copy at the container's working directory before this
// runs — see cloneAndRunScript).
//
// --permission-mode bypassPermissions is required for unattended
// execution (no human to answer a tool-use confirmation prompt inside
// the container); this is the boundary ADR-006's sandbox (network
// allowlist, no platform secrets, ephemeral working copy) exists to make
// safe. Exact CLI flag behavior is confirmed against the real Claude
// Code CLI in TASK-056 — this task's own verification is limited to the
// sandbox itself (TASK-054), not a real AI-provider call (see TASK-056's
// Open Question on credential availability).
func claudeCommand(task platform.ExecutorTask) []string {
	return []string{"claude", "--print", "--permission-mode", "bypassPermissions", buildPrompt(task)}
}

func buildPrompt(task platform.ExecutorTask) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Роль: %s\n", task.Role)
	fmt.Fprintf(&b, "Задача: %s (%s)\n", task.Title, task.Type)
	if task.Scope != "" {
		fmt.Fprintf(&b, "Цель и объём: %s\n", task.Scope)
	}
	if len(task.AcceptanceCriteria) > 0 {
		b.WriteString("Критерии приёмки:\n")
		for _, c := range task.AcceptanceCriteria {
			fmt.Fprintf(&b, "- %s\n", c)
		}
	}
	b.WriteString("Работай в текущей директории — это уже клонированный репозиторий на нужной ветке. Закоммить изменения по завершении работы.\n")
	return b.String()
}
