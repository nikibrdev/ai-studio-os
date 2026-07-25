package platform

import "context"

// PullRequestState is a state of a pull request
// (docs/architecture/domain-model.md, "Pull Request").
type PullRequestState string

// Pull request states.
const (
	PullRequestOpen   PullRequestState = "open"
	PullRequestMerged PullRequestState = "merged"
	PullRequestClosed PullRequestState = "closed"
)

// RepositoryProvider is the single gateway to the git hosting (GitHub) for
// the whole platform (docs/architecture/interfaces.md, "Repository
// Provider"). Agents reach git through tools that use this same contract.
//
// Contract constraints:
//   - the provider executes operations and reports state; it never makes
//     process decisions (when to merge is decided by the workflow process);
//   - merge policies are fixed by ADR-008 (accepted: merge commit, merge
//     after Testing) and belong to the caller, not to this contract;
//   - repo is the platform's canonical stored repository identifier,
//     "<host>/<owner>/<name>" (e.g. "github.com/nikibrdev/ai-studio-os") —
//     what project.Project holds and ProjectService.ConnectRepository
//     records. Keeping the host makes the identifier say which hosting it
//     refers to, so the domain need not assume a single provider;
//     translating it into whatever shape a hosting API wants is each
//     adapter's job (BUGFIX-005 — internal/infrastructure/github reduces it
//     to GitHub's bare "owner/name"). This fixes the format ADR-013 was
//     expected to settle and did not.
type RepositoryProvider interface {
	// CreateBranch creates a branch from the base branch.
	CreateBranch(ctx context.Context, repo, branch, base string) error

	// OpenPullRequest opens a pull request from the branch into the main
	// branch and returns the identifier of the created pull request.
	OpenPullRequest(ctx context.Context, repo, branch, title, body string) (string, error)

	// RequestReview requests a review on the pull request.
	RequestReview(ctx context.Context, repo, prID string) error

	// MergePullRequest merges the pull request into the main branch.
	MergePullRequest(ctx context.Context, repo, prID string) error

	// ClosePullRequest closes the pull request without merging.
	ClosePullRequest(ctx context.Context, repo, prID string) error

	// PullRequestState returns the current state of the pull request.
	PullRequestState(ctx context.Context, repo, prID string) (PullRequestState, error)
}
