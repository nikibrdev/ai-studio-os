package claudecode

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ai-studio-os/agents/claude-code/container"
	"ai-studio-os/internal/platform"
)

func TestParseProposal_ScopeAndCriteria(t *testing.T) {
	got, err := parseProposal(strings.Join([]string{
		"scope: переписать модуль авторизации",
		"criterion: старые токены продолжают работать",
		"criterion: покрытие не падает",
	}, "\n"))
	if err != nil {
		t.Fatalf("parseProposal: %v", err)
	}

	if got.Scope != "переписать модуль авторизации" {
		t.Errorf("Scope = %q", got.Scope)
	}
	if len(got.AcceptanceCriteria) != 2 || got.AcceptanceCriteria[1] != "покрытие не падает" {
		t.Errorf("AcceptanceCriteria = %v, want both criteria in order", got.AcceptanceCriteria)
	}
}

// Either half alone is a legitimate proposal: the platform applies exactly what
// was proposed, and a partial refinement must not erase the other field
// (TASK-096).
func TestParseProposal_EitherHalfAlone(t *testing.T) {
	t.Run("scope only", func(t *testing.T) {
		got, err := parseProposal("scope: только уточнение цели")
		if err != nil {
			t.Fatalf("parseProposal: %v", err)
		}
		if got.Scope == "" || len(got.AcceptanceCriteria) != 0 {
			t.Errorf("got = %+v, want scope only", got)
		}
	})

	t.Run("criteria only", func(t *testing.T) {
		got, err := parseProposal("criterion: единственный критерий")
		if err != nil {
			t.Fatalf("parseProposal: %v", err)
		}
		if got.Scope != "" || len(got.AcceptanceCriteria) != 1 {
			t.Errorf("got = %+v, want criteria only", got)
		}
	})
}

// A wrapped scope is the likeliest way an agent formats long text; lines
// without a prefix continue the scope, which is unambiguous because nothing
// else a stray line could belong to.
func TestParseProposal_ContinuesWrappedScope(t *testing.T) {
	got, err := parseProposal(strings.Join([]string{
		"scope: первая строка цели",
		"продолжение цели",
		"criterion: критерий",
		"",
	}, "\n"))
	if err != nil {
		t.Fatalf("parseProposal: %v", err)
	}

	if got.Scope != "первая строка цели продолжение цели" {
		t.Errorf("Scope = %q, want the wrapped lines joined", got.Scope)
	}
	if len(got.AcceptanceCriteria) != 1 {
		t.Errorf("AcceptanceCriteria = %v, want the criterion kept separate", got.AcceptanceCriteria)
	}
}

// An agent writing "Scope:" is following the convention.
func TestParseProposal_PrefixIsCaseInsensitive(t *testing.T) {
	got, err := parseProposal("SCOPE: цель\nCriterion: критерий")
	if err != nil {
		t.Fatalf("parseProposal: %v", err)
	}
	if got.Scope != "цель" || len(got.AcceptanceCriteria) != 1 {
		t.Errorf("got = %+v, want both parsed regardless of case", got)
	}
}

func TestParseProposal_EmptyIsNoProposal(t *testing.T) {
	for _, raw := range []string{"", "   \n\n  "} {
		if _, err := parseProposal(raw); !errors.Is(err, ErrNoProposal) {
			t.Errorf("parseProposal(%q) error = %v, want %v", raw, err, ErrNoProposal)
		}
	}
}

// Prose with no recognisable prefix is rejected rather than swallowed as one
// big scope: that shape means the agent ignored the convention, and accepting
// it would put arbitrary text into a task's scope.
func TestParseProposal_ProseWithoutPrefixIsRejected(t *testing.T) {
	_, err := parseProposal("Я подумал и решил, что задача выглядит нормально.")
	if !errors.Is(err, ErrUnrecognizedProposal) {
		t.Fatalf("parseProposal() error = %v, want %v", err, ErrUnrecognizedProposal)
	}
}

func TestParseReport_ReturnsTrimmedText(t *testing.T) {
	got, err := parseReport("\n  Проверил критерии, всё сходится.  \n")
	if err != nil {
		t.Fatalf("parseReport: %v", err)
	}
	if got != "Проверил критерии, всё сходится." {
		t.Errorf("parseReport() = %q", got)
	}
}

// A report has no format to violate — any non-empty text is valid, so the only
// failure is its absence.
func TestParseReport_EmptyIsNoReport(t *testing.T) {
	for _, raw := range []string{"", "   \n\t "} {
		if _, err := parseReport(raw); !errors.Is(err, ErrNoReport) {
			t.Errorf("parseReport(%q) error = %v, want %v", raw, err, ErrNoReport)
		}
	}
}

func TestProposal_BeforeAccept_ReturnsError(t *testing.T) {
	e := newTestExecutor(&fakeSandbox{})
	if _, err := e.Proposal(context.Background()); !errors.Is(err, ErrNotAccepted) {
		t.Fatalf("Proposal() before Accept: error = %v, want ErrNotAccepted", err)
	}
}

func TestReport_BeforeAccept_ReturnsError(t *testing.T) {
	e := newTestExecutor(&fakeSandbox{})
	if _, err := e.Report(context.Background()); !errors.Is(err, ErrNotAccepted) {
		t.Fatalf("Report() before Accept: error = %v, want ErrNotAccepted", err)
	}
}

// Each role's result comes from its own file: reading the wrong one would give
// another role's output, which the platform would then apply as if it were this
// role's work.
func TestProposalAndReport_ReadTheirOwnFiles(t *testing.T) {
	sb := &fakeSandbox{execFunc: func(cmd []string) (string, error) {
		if len(cmd) != 2 || cmd[0] != "cat" {
			return "", errors.New("unexpected command")
		}
		switch cmd[1] {
		case container.ProposalFile:
			return "scope: подготовленная цель\ncriterion: критерий\n", nil
		case container.ReportFile:
			return "Проверки пройдены.\n", nil
		default:
			return "", errors.New("wrong file: " + cmd[1])
		}
	}}
	e := newTestExecutor(sb)
	if err := e.Accept(context.Background(), platform.ExecutorTask{
		Role: "project-manager", TaskID: "task-1",
	}); err != nil {
		t.Fatalf("Accept: %v", err)
	}

	proposal, err := e.Proposal(context.Background())
	if err != nil {
		t.Fatalf("Proposal: %v", err)
	}
	if proposal.Scope != "подготовленная цель" || len(proposal.AcceptanceCriteria) != 1 {
		t.Errorf("Proposal() = %+v", proposal)
	}

	report, err := e.Report(context.Background())
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	if report != "Проверки пройдены." {
		t.Errorf("Report() = %q", report)
	}
}

// A missing file surfaces as a failing `cat` — the ordinary case of "the agent
// never wrote one", not an infrastructure fault.
func TestProposalAndReport_MissingFilesAreAbsentResults(t *testing.T) {
	sb := &fakeSandbox{execErr: errors.New("cat: No such file or directory")}
	e := newTestExecutor(sb)
	if err := e.Accept(context.Background(), platform.ExecutorTask{Role: "qa", TaskID: "task-1"}); err != nil {
		t.Fatalf("Accept: %v", err)
	}

	if _, err := e.Proposal(context.Background()); !errors.Is(err, ErrNoProposal) {
		t.Errorf("Proposal() error = %v, want %v", err, ErrNoProposal)
	}
	if _, err := e.Report(context.Background()); !errors.Is(err, ErrNoReport) {
		t.Errorf("Report() error = %v, want %v", err, ErrNoReport)
	}
}
