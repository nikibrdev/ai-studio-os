package claudecode

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/agents/claude-code/container"
	"ai-studio-os/internal/platform"
)

func TestParseVerdict_Approved(t *testing.T) {
	v, err := parseVerdict("approved\nВсё в порядке, критерии выполнены.\n")
	if err != nil {
		t.Fatalf("parseVerdict: %v", err)
	}
	if !v.Approved {
		t.Error("Approved = false, want true")
	}
	if v.Comment != "Всё в порядке, критерии выполнены." {
		t.Errorf("Comment = %q, want the explanation", v.Comment)
	}
}

func TestParseVerdict_ChangesRequested(t *testing.T) {
	v, err := parseVerdict("changes-requested\nНет тестов на отказной путь.")
	if err != nil {
		t.Fatalf("parseVerdict: %v", err)
	}
	if v.Approved {
		t.Error("Approved = true, want false")
	}
	if v.Comment != "Нет тестов на отказной путь." {
		t.Errorf("Comment = %q, want the explanation", v.Comment)
	}
}

// Tolerant about shape: whitespace and case vary harmlessly between runs.
func TestParseVerdict_NormalisesWhitespaceAndCase(t *testing.T) {
	for _, raw := range []string{"approved", "  approved  \n", "APPROVED\n", "Approved"} {
		v, err := parseVerdict(raw)
		if err != nil {
			t.Fatalf("parseVerdict(%q): %v", raw, err)
		}
		if !v.Approved {
			t.Errorf("parseVerdict(%q).Approved = false, want true", raw)
		}
		if v.Comment != "" {
			t.Errorf("parseVerdict(%q).Comment = %q, want empty", raw, v.Comment)
		}
	}
}

// Strict about meaning: an unexpected word must not be guessed at, because
// the answer decides whether code advances toward being merged.
func TestParseVerdict_UnrecognizedIsRejected(t *testing.T) {
	for _, raw := range []string{
		"looks good to me",
		"yes",
		"approve",
		"lgtm\napproved",
		"я одобряю",
	} {
		if _, err := parseVerdict(raw); !errors.Is(err, ErrUnrecognizedVerdict) {
			t.Errorf("parseVerdict(%q) error = %v, want %v", raw, err, ErrUnrecognizedVerdict)
		}
	}
}

func TestParseVerdict_EmptyIsNoVerdict(t *testing.T) {
	for _, raw := range []string{"", "   ", "\n\n", "\t\n "} {
		if _, err := parseVerdict(raw); !errors.Is(err, ErrNoVerdict) {
			t.Errorf("parseVerdict(%q) error = %v, want %v", raw, err, ErrNoVerdict)
		}
	}
}

func TestVerdict_BeforeAccept_ReturnsError(t *testing.T) {
	e := newTestExecutor(&fakeSandbox{})
	if _, err := e.Verdict(context.Background()); !errors.Is(err, ErrNotAccepted) {
		t.Fatalf("Verdict() before Accept: error = %v, want ErrNotAccepted", err)
	}
}

func TestVerdict_ReadsVerdictFileFromSandbox(t *testing.T) {
	sb := &fakeSandbox{execFunc: func(cmd []string) (string, error) {
		if len(cmd) == 2 && cmd[0] == "cat" && cmd[1] == container.VerdictFile {
			return "changes-requested\nдоработать\n", nil
		}
		return "", errors.New("unexpected command")
	}}
	e := newTestExecutor(sb)
	if err := e.Accept(context.Background(), platform.ExecutorTask{Role: "reviewer", TaskID: "task-1"}); err != nil {
		t.Fatalf("Accept: %v", err)
	}

	v, err := e.Verdict(context.Background())
	if err != nil {
		t.Fatalf("Verdict: %v", err)
	}
	if v.Approved || v.Comment != "доработать" {
		t.Errorf("Verdict() = %+v, want changes-requested with the comment", v)
	}
}

// A missing file surfaces as a failing `cat` — the ordinary case of "the
// agent never wrote one", reported as ErrNoVerdict rather than as an
// infrastructure fault.
func TestVerdict_MissingFileIsNoVerdict(t *testing.T) {
	sb := &fakeSandbox{execErr: errors.New("cat: /tmp/ai-studio-os-verdict: No such file or directory")}
	e := newTestExecutor(sb)
	if err := e.Accept(context.Background(), platform.ExecutorTask{Role: "reviewer", TaskID: "task-1"}); err != nil {
		t.Fatalf("Accept: %v", err)
	}

	if _, err := e.Verdict(context.Background()); !errors.Is(err, ErrNoVerdict) {
		t.Fatalf("Verdict() error = %v, want %v", err, ErrNoVerdict)
	}
}
