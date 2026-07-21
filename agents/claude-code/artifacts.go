package claudecode

import (
	"strings"

	"ai-studio-os/internal/platform"
)

// parseCommitArtifacts turns the output of
// `git log --format=%H%n%s -n 20` (hash and subject on alternating
// lines) into platform.Artifact values.
func parseCommitArtifacts(out string) []platform.Artifact {
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil
	}

	lines := strings.Split(trimmed, "\n")
	artifacts := make([]platform.Artifact, 0, len(lines)/2)
	for i := 0; i+1 < len(lines); i += 2 {
		hash := strings.TrimSpace(lines[i])
		subject := strings.TrimSpace(lines[i+1])
		if hash == "" {
			continue
		}
		artifacts = append(artifacts, platform.Artifact{
			ID:      hash,
			Type:    "Commit",
			Origin:  "produced",
			Author:  "claude-code",
			Payload: []byte(subject),
		})
	}
	return artifacts
}
