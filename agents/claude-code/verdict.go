package claudecode

import (
	"errors"
	"fmt"
	"strings"
)

// The two words a reviewing agent may write on the first line of
// container.VerdictFile. Deliberately a single lowercase word each: the
// simplest thing an agent is unlikely to get wrong, and unambiguous to
// parse. Anything else is rejected rather than guessed at.
const (
	verdictApproved         = "approved"
	verdictChangesRequested = "changes-requested"
)

// Errors distinguishing the two ways a verdict can be absent. Both leave the
// task in Review (TASK-088), but they are different situations for whoever
// looks into it, so they are not collapsed into one error: "the agent never
// got that far" and "the agent wrote something we cannot read" call for
// different responses.
var (
	// ErrNoVerdict means the file is missing or empty — the agent did not
	// reach a decision (crashed, ran out of turns, was cut off).
	ErrNoVerdict = errors.New("claudecode: no verdict was written")

	// ErrUnrecognizedVerdict means the file has content that is not one of
	// the two accepted words.
	ErrUnrecognizedVerdict = errors.New("claudecode: verdict is not recognized")
)

// Verdict is a reviewing agent's decision about a task's changes.
//
// Not a platform.Artifact and not reported through Artifacts: an artifact is
// evidence of work produced (a commit), while a verdict is a decision about
// work. The platform, not the agent, acts on it — the adapter only carries
// it across (ADR-005/ADR-014: an Executor does not reach into the platform).
type Verdict struct {
	// Approved is true for "approved", false for "changes-requested".
	Approved bool

	// Comment is the agent's explanation, empty if it wrote none.
	Comment string
}

// parseVerdict reads the verdict file's contents.
//
// Tolerant about shape, strict about meaning: surrounding whitespace and
// letter case are normalised, because those vary harmlessly between runs;
// an unexpected word is an error, because guessing what an agent meant is
// exactly what must not happen when the answer decides whether code
// advances toward being merged.
func parseVerdict(raw string) (Verdict, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Verdict{}, ErrNoVerdict
	}

	lines := strings.SplitN(trimmed, "\n", 2)
	word := strings.ToLower(strings.TrimSpace(lines[0]))

	comment := ""
	if len(lines) > 1 {
		comment = strings.TrimSpace(lines[1])
	}

	switch word {
	case verdictApproved:
		return Verdict{Approved: true, Comment: comment}, nil
	case verdictChangesRequested:
		return Verdict{Approved: false, Comment: comment}, nil
	default:
		return Verdict{}, fmt.Errorf("%w: first line was %q", ErrUnrecognizedVerdict, word)
	}
}
