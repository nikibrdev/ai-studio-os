package claudecode

import (
	"errors"
	"fmt"
	"strings"

	"ai-studio-os/agents/claude-code/container"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// ErrUnsupportedRole is returned when a task names a role this adapter has
// no instructions for.
//
// An error rather than a neutral fallback (TASK-087): PM and QA are
// deliberately not dispatched yet — their outputs are exactly the human
// checkpoints ADR-007 names, so they wait for the confirmation mechanism.
// Inventing their instructions now would build something untestable, and
// quietly reusing the Developer block would tell an agent to commit code
// when it was asked to plan or verify. Reported before the sandbox starts,
// so an accidental dispatch fails loudly instead of spending a container.
var ErrUnsupportedRole = errors.New("claudecode: no prompt instructions for this role")

// baseBranchName is the branch a task branch is cut from and compared
// against (docs/development/git-workflow.md). The orchestrator uses the same
// value when creating the branch; the adapter needs it only to phrase the
// diff for a reviewer.
const baseBranchName = "main"

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
func claudeCommand(task platform.ExecutorTask) ([]string, error) {
	prompt, err := buildPrompt(task)
	if err != nil {
		return nil, err
	}
	return []string{"claude", "--print", "--permission-mode", "bypassPermissions", prompt}, nil
}

// buildPrompt assembles the prompt: the task's content, which is the same
// for every role, followed by instructions specific to the role.
//
// One adapter and one image serve all roles, differing only here (ADR-007) —
// not separate adapters or images per role.
func buildPrompt(task platform.ExecutorTask) (string, error) {
	instructions, err := roleInstructions(task.Role)
	if err != nil {
		return "", err
	}

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
	b.WriteString(instructions)
	return b.String(), nil
}

// roleInstructions returns what the agent is being asked to do. Only roles
// the platform actually dispatches have instructions; see ErrUnsupportedRole
// on why the rest are an error rather than a default.
func roleInstructions(role string) (string, error) {
	switch shared.Role(role) {
	case shared.RoleDeveloper:
		return "Работай в текущей директории — это уже клонированный репозиторий на нужной ветке. " +
			"Закоммить изменения по завершении работы.\n", nil

	case shared.RoleReviewer:
		// origin/main, not the clone's HEAD: for a reviewer the clone already
		// contains the developer's commits, so HEAD as a base would compare
		// the branch against itself. The clone is not --single-branch, so
		// origin/main is present (see container.cloneAndRunScript).
		return "Ты ревьюер. Изучи изменения этой ветки относительно основной: " +
			"`git diff origin/" + baseBranchName + "...HEAD` и `git log origin/" + baseBranchName + "..HEAD`. " +
			"Оцени соответствие критериям приёмки выше, корректность и качество.\n" +
			"Ничего не коммить и не изменять в репозитории.\n" +
			"Запиши вердикт в файл " + container.VerdictFile + ": первая строка — ровно одно слово, " +
			verdictApproved + " (изменения можно принять) или " + verdictChangesRequested + " (нужны правки); " +
			"со второй строки — пояснение для человека.\n", nil

	case shared.RoleProjectManager:
		// Prepares, never decides: accepting Definition of Ready is a human
		// checkpoint (docs/architecture/workflow.md), and the platform applies
		// the proposal itself once a human accepts it.
		return "Ты Project Manager. Задача ещё не принята в работу — твоя цель подготовить её к приёму: " +
			"уточнить цель и объём (scope) и предложить проверяемые критерии приёмки.\n" +
			"Изучи репозиторий в текущей директории, чтобы предложение опиралось на реальный код, а не на догадки.\n" +
			"Ничего не коммить, не менять файлы репозитория и не менять состояние задачи — решение принимает человек.\n" +
			"Запиши предложение в файл " + container.ProposalFile + " строками с префиксами:\n" +
			proposalScopePrefix + " <цель и объём; можно продолжить на следующих строках>\n" +
			proposalCriterionPrefix + " <один критерий приёмки>\n" +
			proposalCriterionPrefix + " <ещё один критерий>\n", nil

	case shared.RoleQA:
		// Same base as the reviewer, and for the same reason: the clone already
		// contains the developer's commits, so HEAD as a base would compare the
		// branch against itself.
		return "Ты QA-инженер. Проверь изменения этой ветки относительно основной: " +
			"`git diff origin/" + baseBranchName + "...HEAD` и `git log origin/" + baseBranchName + "..HEAD`. " +
			"Проверь соответствие критериям приёмки выше и запусти проверки проекта, если они есть.\n" +
			"Ничего не коммить, не менять файлы репозитория, не сливать ветку и не менять состояние задачи — " +
			"приёмочное решение принимает человек.\n" +
			"Запиши отчёт в файл " + container.ReportFile + ": что проверено, что нашлось, что осталось сомнительным. " +
			"Формат свободный — отчёт читает человек.\n", nil

	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedRole, role)
	}
}
