package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"ai-studio-os/internal/platform"
)

// TokenEnv is the environment variable holding the GitHub API token.
const TokenEnv = "GITHUB_TOKEN"

const defaultBaseURL = "https://api.github.com"

// ErrTokenNotSet is returned by New when TokenEnv is empty.
var ErrTokenNotSet = errors.New("github: " + TokenEnv + " is not set")

// Provider implements platform.RepositoryProvider against the GitHub REST
// API. repo is always "owner/name" (the format already fixed by the
// platform.RepositoryProvider contract).
type Provider struct {
	baseURL string
	token   string
	client  *http.Client
}

var _ platform.RepositoryProvider = (*Provider)(nil)

// New creates a Provider using the token from TokenEnv.
func New() (*Provider, error) {
	token := os.Getenv(TokenEnv)
	if token == "" {
		return nil, ErrTokenNotSet
	}
	return NewWithToken(token), nil
}

// NewWithToken creates a Provider with an explicit token.
func NewWithToken(token string) *Provider {
	return &Provider{baseURL: defaultBaseURL, token: token, client: &http.Client{Timeout: 30 * time.Second}}
}

// APIError is returned when the GitHub API responds with a non-2xx
// status; it carries enough context (operation, status, response body)
// to diagnose the failure without a debugger.
type APIError struct {
	Method     string
	Path       string
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("github: %s %s: unexpected status %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// CreateBranch implements platform.RepositoryProvider: it resolves base's
// current commit SHA and creates branch pointing at it.
func (p *Provider) CreateBranch(ctx context.Context, repo, branch, base string) error {
	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := p.do(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/git/ref/heads/%s", repo, base), nil, &ref); err != nil {
		return fmt.Errorf("github: resolve base branch %s: %w", base, err)
	}

	body := map[string]string{"ref": "refs/heads/" + branch, "sha": ref.Object.SHA}
	if err := p.do(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/git/refs", repo), body, nil); err != nil {
		return fmt.Errorf("github: create branch %s: %w", branch, err)
	}
	return nil
}

// OpenPullRequest implements platform.RepositoryProvider: it always
// targets "main" — the contract's own doc comment fixes the target
// ("opens a pull request from the branch into the main branch"), matching
// this project's git-workflow (every branch merges into main).
func (p *Provider) OpenPullRequest(ctx context.Context, repo, branch, title, body string) (string, error) {
	reqBody := map[string]string{"title": title, "body": body, "head": branch, "base": "main"}

	var resp struct {
		Number int `json:"number"`
	}
	if err := p.do(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/pulls", repo), reqBody, &resp); err != nil {
		return "", fmt.Errorf("github: open pull request from %s: %w", branch, err)
	}
	return strconv.Itoa(resp.Number), nil
}

// RequestReview implements platform.RepositoryProvider. The contract
// carries no reviewer identity (ADR-008: review is enforced by the
// platform's own Review workflow stage, not by GitHub's native
// required-reviewers setting — see the ADR's transitional-period note),
// so there is no GitHub reviewer to assign here. This posts a visible PR
// comment marking the request instead of silently doing nothing.
func (p *Provider) RequestReview(ctx context.Context, repo, prID string) error {
	body := map[string]string{"body": "Запрошено ревью."}
	if err := p.do(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/issues/%s/comments", repo, prID), body, nil); err != nil {
		return fmt.Errorf("github: request review on PR %s: %w", prID, err)
	}
	return nil
}

// MergePullRequest implements platform.RepositoryProvider: merge commit,
// per ADR-008 (not squash, not rebase).
func (p *Provider) MergePullRequest(ctx context.Context, repo, prID string) error {
	body := map[string]string{"merge_method": "merge"}
	path := fmt.Sprintf("/repos/%s/pulls/%s/merge", repo, prID)
	if err := p.do(ctx, http.MethodPut, path, body, nil); err != nil {
		return fmt.Errorf("github: merge PR %s: %w", prID, err)
	}
	return nil
}

// ClosePullRequest implements platform.RepositoryProvider.
func (p *Provider) ClosePullRequest(ctx context.Context, repo, prID string) error {
	body := map[string]string{"state": "closed"}
	path := fmt.Sprintf("/repos/%s/pulls/%s", repo, prID)
	if err := p.do(ctx, http.MethodPatch, path, body, nil); err != nil {
		return fmt.Errorf("github: close PR %s: %w", prID, err)
	}
	return nil
}

// PullRequestState implements platform.RepositoryProvider.
func (p *Provider) PullRequestState(ctx context.Context, repo, prID string) (platform.PullRequestState, error) {
	var resp struct {
		State  string `json:"state"`
		Merged bool   `json:"merged"`
	}
	path := fmt.Sprintf("/repos/%s/pulls/%s", repo, prID)
	if err := p.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return "", fmt.Errorf("github: get state of PR %s: %w", prID, err)
	}

	switch {
	case resp.Merged:
		return platform.PullRequestMerged, nil
	case resp.State == "closed":
		return platform.PullRequestClosed, nil
	default:
		return platform.PullRequestOpen, nil
	}
}

// do sends a GitHub REST API request and decodes the JSON response into
// out, if non-nil. body, if non-nil, is marshalled as the JSON request
// body.
func (p *Provider) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("github: marshal request body for %s %s: %w", method, path, err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("github: build request %s %s: %w", method, path, err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("github: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("github: read response body for %s %s: %w", method, path, err)
	}

	if resp.StatusCode >= 300 {
		return &APIError{Method: method, Path: path, StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("github: decode response for %s %s: %w", method, path, err)
		}
	}
	return nil
}
