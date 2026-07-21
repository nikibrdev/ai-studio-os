package container

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// commandRunner executes one external command and returns its trimmed
// stdout. Narrowed so tests can inject a fake instead of shelling out to
// a real docker binary.
type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) (stdout string, err error)
}

// execRunner runs commands via os/exec — the production implementation.
type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("container: %s %s: %w: %s", name, strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}
