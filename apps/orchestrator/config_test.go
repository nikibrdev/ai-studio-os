package main

import (
	"testing"
	"time"
)

func TestFromEnv_DefaultsWhenOnlyDatabaseURLIsSet(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("EXECUTION_IMAGE", "")
	t.Setenv("POLL_INTERVAL", "")
	t.Setenv("QDRANT_URL", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("ANTHROPIC_API_KEY", "")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}

	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("DatabaseURL = %q, want the value from the environment", cfg.DatabaseURL)
	}
	if cfg.ExecutionImage != defaultExecutionImage {
		t.Errorf("ExecutionImage = %q, want default %q", cfg.ExecutionImage, defaultExecutionImage)
	}
	if cfg.PollInterval != defaultPollInterval {
		t.Errorf("PollInterval = %s, want default %s", cfg.PollInterval, defaultPollInterval)
	}
}

func TestFromEnv_OverridesFromEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("EXECUTION_IMAGE", "custom-image")
	t.Setenv("POLL_INTERVAL", "250ms")
	t.Setenv("QDRANT_URL", "http://localhost:6333")
	t.Setenv("GITHUB_TOKEN", "gh-token")
	t.Setenv("ANTHROPIC_API_KEY", "provider-key")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}

	if cfg.ExecutionImage != "custom-image" {
		t.Errorf("ExecutionImage = %q, want custom-image", cfg.ExecutionImage)
	}
	if cfg.PollInterval != 250*time.Millisecond {
		t.Errorf("PollInterval = %s, want 250ms", cfg.PollInterval)
	}
	if cfg.QdrantURL != "http://localhost:6333" {
		t.Errorf("QdrantURL = %q, want the value from the environment", cfg.QdrantURL)
	}
	if cfg.GitToken != "gh-token" || cfg.ProviderAPIKey != "provider-key" {
		t.Errorf("credentials = %q/%q, want gh-token/provider-key", cfg.GitToken, cfg.ProviderAPIKey)
	}
}

// An empty ANTHROPIC_API_KEY must stay valid: claudecode.New accepts it and
// still starts the sandbox, which is how the container lifecycle is
// exercised without a real provider call (TASK-056's Open Question).
func TestFromEnv_EmptyProviderKeyIsValid(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("ANTHROPIC_API_KEY", "")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if cfg.ProviderAPIKey != "" {
		t.Errorf("ProviderAPIKey = %q, want empty", cfg.ProviderAPIKey)
	}
}

func TestFromEnv_RejectsUnparsablePollInterval(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("POLL_INTERVAL", "soon")

	if _, err := FromEnv(); err == nil {
		t.Fatal("FromEnv() error = nil, want an error for an unparsable POLL_INTERVAL")
	}
}

func TestFromEnv_RejectsNonPositivePollInterval(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")

	for _, raw := range []string{"0s", "-5s"} {
		t.Setenv("POLL_INTERVAL", raw)
		if _, err := FromEnv(); err == nil {
			t.Errorf("FromEnv() error = nil for POLL_INTERVAL=%q, want an error", raw)
		}
	}
}
