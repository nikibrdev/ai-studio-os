package container

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// publicNetwork is Docker's default bridge network, which has a real
// route to the internet — only the proxy sidecar is ever connected to
// it; the execution container is not.
const publicNetwork = "bridge"

// Status is a point-in-time read of an Execution container's state.
type Status struct {
	Running  bool
	ExitCode int
}

// StartParams describes one Execution's sandbox to start.
type StartParams struct {
	// ExecutionID names this Execution's containers and network; must be
	// safe as a Docker name component (letters, digits, - and _).
	ExecutionID string

	// Repository and Branch are the git coordinates cloned into the
	// container's working copy at start (ADR-006).
	Repository string
	Branch     string

	// GitToken and ProviderAPIKey are short-lived secrets injected as
	// container environment variables only — never written to the
	// image, never logged, never present in argv visible via
	// `docker inspect`/`ps` (see cloneAndRunScript's use of GIT_ASKPASS).
	GitToken       string
	ProviderAPIKey string

	// Allowlist is the set of hostnames the container may reach over the
	// network, in addition to github.com/api.github.com (always allowed,
	// needed for the clone itself).
	Allowlist []string

	// Command is executed inside the container after the clone succeeds
	// (e.g. the Claude Code invocation — TASK-055 supplies it; this
	// package does not know what a Claude Code invocation looks like).
	Command []string
}

// Handle references one running Execution's sandbox — the container,
// proxy and network Start created, needed to Status/Exec/Stop it later.
type Handle struct {
	containerName string
	proxyName     string
	networkName   string
}

// Manager starts and tears down Execution sandboxes (ADR-006) by
// shelling out to the docker CLI.
type Manager struct {
	run   commandRunner
	image string
}

// NewManager creates a Manager that runs the given execution image
// (docker/execution, TASK-053) via the real docker CLI.
func NewManager(image string) *Manager {
	return &Manager{run: execRunner{}, image: image}
}

// Start clones p.Repository/p.Branch and runs p.Command inside a fresh,
// network-restricted container: a private network with no route to the
// internet, a Squid proxy sidecar (the only member of that network also
// connected to the internet) enforcing the host allowlist, and secrets
// passed only as environment variables.
func (m *Manager) Start(ctx context.Context, p StartParams) (*Handle, error) {
	h := &Handle{
		containerName: "ai-studio-os-exec-" + p.ExecutionID,
		proxyName:     "ai-studio-os-proxy-" + p.ExecutionID,
		networkName:   "ai-studio-os-net-" + p.ExecutionID,
	}

	allowlist := append([]string{"github.com", "api.github.com"}, p.Allowlist...)

	if err := ensureNetwork(ctx, m.run, h.networkName, true); err != nil {
		return nil, err
	}
	if err := startProxy(ctx, m.run, h.proxyName, h.networkName, publicNetwork, allowlist); err != nil {
		_ = removeNetwork(ctx, m.run, h.networkName)
		return nil, err
	}

	script := cloneAndRunScript(p.Repository, p.Branch, p.Command)
	proxyURL := fmt.Sprintf("http://%s:%s", h.proxyName, proxyPort)

	args := []string{
		"run", "-d", "--name", h.containerName,
		"--network", h.networkName,
		"-e", "GIT_TOKEN=" + p.GitToken,
		"-e", "HTTP_PROXY=" + proxyURL,
		"-e", "HTTPS_PROXY=" + proxyURL,
	}
	if p.ProviderAPIKey != "" {
		args = append(args, "-e", "ANTHROPIC_API_KEY="+p.ProviderAPIKey)
	}
	args = append(args, "--entrypoint", "sh", m.image, "-c", script)

	if _, err := m.run.Run(ctx, "docker", args...); err != nil {
		_ = removeContainer(ctx, m.run, h.proxyName)
		_ = removeNetwork(ctx, m.run, h.networkName)
		return nil, fmt.Errorf("container: start execution %s: %w", h.containerName, err)
	}
	return h, nil
}

// Status reports whether the execution container is still running and,
// once it has exited, its exit code.
func (m *Manager) Status(ctx context.Context, h *Handle) (Status, error) {
	out, err := m.run.Run(ctx, "docker", "inspect", "--format", "{{.State.Running}} {{.State.ExitCode}}", h.containerName)
	if err != nil {
		return Status{}, fmt.Errorf("container: status %s: %w", h.containerName, err)
	}

	fields := strings.Fields(out)
	if len(fields) != 2 {
		return Status{}, fmt.Errorf("container: unexpected inspect output for %s: %q", h.containerName, out)
	}
	exitCode, _ := strconv.Atoi(fields[1])
	return Status{Running: fields[0] == "true", ExitCode: exitCode}, nil
}

// Exec runs cmd inside the running execution container and returns its
// output — used to poll progress (e.g. `git log`, `git diff`) without
// needing a dedicated port or API from the container itself.
func (m *Manager) Exec(ctx context.Context, h *Handle, cmd []string) (string, error) {
	args := append([]string{"exec", h.containerName}, cmd...)
	out, err := m.run.Run(ctx, "docker", args...)
	if err != nil {
		return "", fmt.Errorf("container: exec in %s: %w", h.containerName, err)
	}
	return out, nil
}

// Stop tears down the execution container, the proxy and the network —
// the ephemeral working copy dies with the container (ADR-006). Safe to
// call more than once; every step is individually idempotent.
func (m *Manager) Stop(ctx context.Context, h *Handle) error {
	if err := removeContainer(ctx, m.run, h.containerName); err != nil {
		return err
	}
	if err := removeContainer(ctx, m.run, h.proxyName); err != nil {
		return err
	}
	return removeNetwork(ctx, m.run, h.networkName)
}
