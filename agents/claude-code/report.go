package claudecode

import (
	"errors"
	"strings"
)

// ErrNoReport means the QA agent wrote no report, or an empty one.
//
// There is deliberately no ErrUnrecognizedReport counterpart. A report has no
// format to violate: it is read by a human, not parsed by the platform, so
// any non-empty text is a valid report. Adding a symmetric error for the sake
// of symmetry would declare a condition that can never occur.
var ErrNoReport = errors.New("claudecode: no QA report was written")

// parseReport reads the QA report: the whole file, trimmed.
//
// No structure is imposed. The report exists so a human can decide whether to
// accept the task, and prescribing a layout would only add ways for the agent
// to get it wrong without making the text more useful to the reader.
func parseReport(raw string) (string, error) {
	report := strings.TrimSpace(raw)
	if report == "" {
		return "", ErrNoReport
	}
	return report, nil
}
