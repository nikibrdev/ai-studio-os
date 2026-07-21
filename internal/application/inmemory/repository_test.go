package inmemory_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/platform"
)

var _ platform.RepositoryProvider = inmemory.NewRepositoryProvider()

func TestRepositoryProvider_OpenAndMerge(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewRepositoryProvider()

	prID, err := repo.OpenPullRequest(ctx, "org/repo", "feature/TASK-1", "title", "body")
	if err != nil {
		t.Fatalf("OpenPullRequest: %v", err)
	}
	state, err := repo.PullRequestState(ctx, "org/repo", prID)
	if err != nil || state != platform.PullRequestOpen {
		t.Fatalf("PullRequestState() = (%v, %v), want Open", state, err)
	}

	if err := repo.MergePullRequest(ctx, "org/repo", prID); err != nil {
		t.Fatalf("MergePullRequest: %v", err)
	}
	state, err = repo.PullRequestState(ctx, "org/repo", prID)
	if err != nil || state != platform.PullRequestMerged {
		t.Fatalf("PullRequestState() after merge = (%v, %v), want Merged", state, err)
	}
	if len(repo.MergeCalls) != 1 {
		t.Errorf("MergeCalls = %v, want 1 recorded call", repo.MergeCalls)
	}
}

func TestRepositoryProvider_MergeErrInjection(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewRepositoryProvider()
	wantErr := errors.New("merge conflict")
	repo.MergeErr = wantErr

	if err := repo.MergePullRequest(ctx, "org/repo", "pr-1"); !errors.Is(err, wantErr) {
		t.Errorf("MergePullRequest() error = %v, want %v", err, wantErr)
	}
	if state, err := repo.PullRequestState(ctx, "org/repo", "pr-1"); err == nil {
		t.Errorf("PullRequestState() = %v, want unknown (merge never recorded state)", state)
	}
}

func TestRepositoryProvider_UnknownPullRequest(t *testing.T) {
	repo := inmemory.NewRepositoryProvider()
	if _, err := repo.PullRequestState(context.Background(), "org/repo", "missing"); err == nil {
		t.Error("PullRequestState() for unknown PR error = nil, want an error")
	}
}
