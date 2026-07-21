package inmemory

import (
	"context"
	"errors"
	"sync"

	"ai-studio-os/internal/platform"
)

// RepositoryProvider is a deterministic fake of platform.RepositoryProvider
// for this epic's tests: it records calls and reports pull request state
// from an in-memory map, with no real git hosting behind it. The GitHub
// adapter arrives in EPIC-005 (v0.5).
type RepositoryProvider struct {
	mu         sync.Mutex
	prState    map[string]platform.PullRequestState // key: repo+"/"+prID
	MergeErr   error                                // if set, MergePullRequest returns this instead of merging
	MergeCalls []string                             // repo+"/"+prID for every MergePullRequest call, in order
}

// NewRepositoryProvider creates an empty RepositoryProvider fake.
func NewRepositoryProvider() *RepositoryProvider {
	return &RepositoryProvider{prState: make(map[string]platform.PullRequestState)}
}

func key(repo, prID string) string { return repo + "/" + prID }

// CreateBranch implements platform.RepositoryProvider; it only records
// that the call happened, no branches actually exist.
func (r *RepositoryProvider) CreateBranch(_ context.Context, _, _, _ string) error {
	return nil
}

// OpenPullRequest implements platform.RepositoryProvider, returning a
// deterministic fake pull request identifier in the Open state.
func (r *RepositoryProvider) OpenPullRequest(_ context.Context, repo, branch, _, _ string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	prID := branch
	r.prState[key(repo, prID)] = platform.PullRequestOpen
	return prID, nil
}

// RequestReview implements platform.RepositoryProvider; it only records
// that the call happened.
func (r *RepositoryProvider) RequestReview(_ context.Context, _, _ string) error {
	return nil
}

// MergePullRequest implements platform.RepositoryProvider. If MergeErr is
// set, it returns that error instead of merging — the way this fake lets
// tests exercise ADR-008's guard (Task must not reach Done unless merge
// actually succeeded).
func (r *RepositoryProvider) MergePullRequest(_ context.Context, repo, prID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.MergeCalls = append(r.MergeCalls, key(repo, prID))
	if r.MergeErr != nil {
		return r.MergeErr
	}
	r.prState[key(repo, prID)] = platform.PullRequestMerged
	return nil
}

// ClosePullRequest implements platform.RepositoryProvider.
func (r *RepositoryProvider) ClosePullRequest(_ context.Context, repo, prID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prState[key(repo, prID)] = platform.PullRequestClosed
	return nil
}

// PullRequestState implements platform.RepositoryProvider.
func (r *RepositoryProvider) PullRequestState(_ context.Context, repo, prID string) (platform.PullRequestState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.prState[key(repo, prID)]
	if !ok {
		return "", errors.New("inmemory: unknown pull request")
	}
	return s, nil
}
