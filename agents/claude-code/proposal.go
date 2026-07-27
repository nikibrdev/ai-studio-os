package claudecode

import (
	"errors"
	"strings"
)

// Line prefixes a Project Manager agent uses in container.ProposalFile.
//
// Line-prefixed rather than JSON, for the same reason the review verdict is a
// single lowercase word: it is the simplest shape an agent is unlikely to
// garble. An agent asked for JSON readily wraps it in markdown fences or breaks
// a quote, and then the whole proposal has to be rejected.
const (
	proposalScopePrefix     = "scope:"
	proposalCriterionPrefix = "criterion:"
)

// Errors distinguishing the two ways a proposal can be unusable — the same
// split as the review verdict, because they call for different responses:
// "the agent never got that far" versus "the agent wrote something we cannot
// read".
var (
	// ErrNoProposal means the file is missing or empty.
	ErrNoProposal = errors.New("claudecode: no Definition of Ready proposal was written")

	// ErrUnrecognizedProposal means the file has content but not a single
	// recognisable line — the agent wrote prose instead of following the
	// convention, and guessing which part is the scope would be inventing it.
	ErrUnrecognizedProposal = errors.New("claudecode: Definition of Ready proposal is not recognized")
)

// Proposal is what a Project Manager agent prepares for a task in Backlog: a
// sharper scope and acceptance criteria, for a human to accept or not.
//
// Either field may be empty — a proposal that only sharpens the scope is
// legitimate, and the platform applies exactly what was proposed (TASK-096:
// a partial refinement must not erase the other field).
type Proposal struct {
	Scope              string
	AcceptanceCriteria []string
}

// parseProposal reads the proposal file.
//
// Tolerant about layout, strict about meaning: a line without a known prefix
// continues the scope, which forgives wrapped text without introducing
// ambiguity — there is nothing else a stray line could belong to. A file with
// no recognisable prefix at all is rejected rather than treated as one big
// scope: that shape means the agent ignored the convention, and accepting it
// would put arbitrary prose into a task's scope.
func parseProposal(raw string) (Proposal, error) {
	if strings.TrimSpace(raw) == "" {
		return Proposal{}, ErrNoProposal
	}

	var (
		scopeLines []string
		criteria   []string
		recognised bool
	)

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case hasPrefixFold(trimmed, proposalScopePrefix):
			recognised = true
			if rest := strings.TrimSpace(trimmed[len(proposalScopePrefix):]); rest != "" {
				scopeLines = append(scopeLines, rest)
			}
		case hasPrefixFold(trimmed, proposalCriterionPrefix):
			recognised = true
			if rest := strings.TrimSpace(trimmed[len(proposalCriterionPrefix):]); rest != "" {
				criteria = append(criteria, rest)
			}
		case recognised:
			// A continuation of the scope: only meaningful once a prefix has
			// been seen, otherwise the file is not following the convention.
			scopeLines = append(scopeLines, trimmed)
		}
	}

	if !recognised {
		return Proposal{}, ErrUnrecognizedProposal
	}
	return Proposal{Scope: strings.Join(scopeLines, " "), AcceptanceCriteria: criteria}, nil
}

// hasPrefixFold reports whether s starts with prefix, ignoring case — an agent
// writing "Scope:" instead of "scope:" is following the convention.
func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}
