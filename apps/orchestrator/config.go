package main

import (
	"fmt"
	"os"
	"time"

	"ai-studio-os/internal/infrastructure/github"
	"ai-studio-os/internal/infrastructure/postgres"
)

// Environment variables specific to the orchestrator. The storage and
// GitHub credentials reuse the names their own packages already define
// (postgres.DatabaseURLEnv, github.TokenEnv) rather than inventing
// parallel ones.
const (
	executionImageEnv = "EXECUTION_IMAGE"
	providerAPIKeyEnv = "ANTHROPIC_API_KEY"
	pollIntervalEnv   = "POLL_INTERVAL"
	qdrantURLEnv      = "QDRANT_URL"
)

// Defaults chosen so a local run needs only DATABASE_URL: the image name
// is the one agents/claude-code/README.md documents building, and five
// seconds is short enough that a planned task starts promptly while
// keeping the journal query rate negligible.
const (
	defaultExecutionImage = "ai-studio-os-execution"
	defaultPollInterval   = 5 * time.Second
)

// Config is everything the orchestrator process needs from its
// environment.
type Config struct {
	// DatabaseURL is the PostgreSQL DSN; required.
	DatabaseURL string

	// QdrantURL is optional — empty leaves wiring.System.Memory nil, the
	// same opt-in behaviour apps/api relies on.
	QdrantURL string

	// ExecutionImage is the Docker image agents run inside (ADR-006).
	ExecutionImage string

	// GitToken authenticates git operations inside the sandbox.
	GitToken string

	// ProviderAPIKey authenticates the AI provider. Empty is valid:
	// claudecode.New accepts it and still starts the sandbox, which is how
	// the container lifecycle is exercised without a real provider call
	// (TASK-056's Open Question).
	ProviderAPIKey string

	// PollInterval is how often the event journal is polled.
	PollInterval time.Duration
}

// FromEnv reads Config from the process environment, applying defaults.
// It fails only on values that are present but unusable — a missing
// DATABASE_URL is left to wiring.New, which already reports it precisely
// (postgres.ErrDatabaseURLNotSet).
func FromEnv() (Config, error) {
	cfg := Config{
		DatabaseURL:    os.Getenv(postgres.DatabaseURLEnv),
		QdrantURL:      os.Getenv(qdrantURLEnv),
		ExecutionImage: os.Getenv(executionImageEnv),
		GitToken:       os.Getenv(github.TokenEnv),
		ProviderAPIKey: os.Getenv(providerAPIKeyEnv),
		PollInterval:   defaultPollInterval,
	}

	if cfg.ExecutionImage == "" {
		cfg.ExecutionImage = defaultExecutionImage
	}

	if raw := os.Getenv(pollIntervalEnv); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("orchestrator: parse %s=%q: %w", pollIntervalEnv, raw, err)
		}
		if d <= 0 {
			return Config{}, fmt.Errorf("orchestrator: %s must be positive, got %s", pollIntervalEnv, d)
		}
		cfg.PollInterval = d
	}

	return cfg, nil
}
