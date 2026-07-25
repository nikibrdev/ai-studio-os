package container

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeRepository(t *testing.T) {
	tests := []struct{ name, in, want string }{
		// The form the platform actually stores — the one that produced
		// https://github.com/github.com/owner/name.git (BUGFIX-006).
		{name: "stored form with host", in: "github.com/nikibrdev/ai-studio-os", want: "nikibrdev/ai-studio-os"},
		{name: "already short", in: "org/repo", want: "org/repo"},
		{name: "https url", in: "https://github.com/org/repo", want: "org/repo"},
		{name: "url with .git", in: "https://github.com/org/repo.git", want: "org/repo"},
		{name: "trailing slash", in: "github.com/org/repo/", want: "org/repo"},
		{name: "mixed case host", in: "GitHub.com/org/repo", want: "org/repo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeRepository(tt.in)
			if err != nil {
				t.Fatalf("normalizeRepository(%q) error = %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("normalizeRepository(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeRepository_Invalid(t *testing.T) {
	for _, in := range []string{"", "repo", "github.com/repo", "org/", "/repo", "github.com/org/repo/extra"} {
		if got, err := normalizeRepository(in); !errors.Is(err, ErrInvalidRepository) {
			t.Errorf("normalizeRepository(%q) = (%q, %v), want ErrInvalidRepository", in, got, err)
		}
	}
}

// TestCloneAndRunScript_StoredFormYieldsSingleHost is the regression test for
// BUGFIX-006. Every other script test passes the short "org/repo" form,
// which is exactly why the doubled host went unnoticed until the first live
// end-to-end run.
func TestCloneAndRunScript_StoredFormYieldsSingleHost(t *testing.T) {
	script, err := cloneAndRunScript("github.com/nikibrdev/ai-studio-os", "feature/TASK-001", []string{"true"})
	if err != nil {
		t.Fatalf("cloneAndRunScript: %v", err)
	}

	if n := strings.Count(script, "github.com"); n != 1 {
		t.Errorf("clone URL mentions github.com %d times, want exactly 1:\n%s", n, script)
	}
	want := "https://x-access-token@github.com/nikibrdev/ai-studio-os.git"
	if !strings.Contains(script, want) {
		t.Errorf("script does not contain %q:\n%s", want, script)
	}
}

// A bad identifier must be rejected while building the script, so Start
// fails before creating a network and a proxy it would have to clean up.
func TestCloneAndRunScript_RejectsInvalidRepository(t *testing.T) {
	if _, err := cloneAndRunScript("not-a-repo", "main", []string{"true"}); !errors.Is(err, ErrInvalidRepository) {
		t.Fatalf("cloneAndRunScript() error = %v, want ErrInvalidRepository", err)
	}
}
