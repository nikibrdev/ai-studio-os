package container

import (
	"context"
	"fmt"
	"strings"
)

// proxyImage is the forward proxy used to enforce the network allowlist
// (ADR-006). A well-known public image, not a custom build — the same
// "prefer an existing well-known dependency over a homegrown one"
// reasoning as using postgres:16-alpine directly for the database
// (docker-compose.yml) rather than building a custom Postgres image.
const proxyImage = "ubuntu/squid:latest"

const proxyPort = "3128"

// renderSquidConf builds a Squid configuration allowing CONNECT (and
// plain HTTP) only to the given hostnames, denying everything else.
// dstdomain matching the CONNECT target does not require decrypting
// TLS (no ssl_bump) — Squid only needs the hostname the client asked to
// CONNECT to, not the certificate exchanged after tunneling starts.
func renderSquidConf(allowlist []string) string {
	domains := make([]string, len(allowlist))
	for i, host := range allowlist {
		domains[i] = "." + strings.TrimPrefix(host, ".")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "http_port %s\n", proxyPort)
	b.WriteString("acl allowed_ssl_ports port 443\n")
	b.WriteString("acl CONNECT method CONNECT\n")
	fmt.Fprintf(&b, "acl allowed_domains dstdomain %s\n", strings.Join(domains, " "))
	b.WriteString("http_access allow CONNECT allowed_domains allowed_ssl_ports\n")
	b.WriteString("http_access allow allowed_domains\n")
	b.WriteString("http_access deny all\n")
	return b.String()
}

// ensureNetwork creates the given Docker network if it does not already
// exist. internal, when true, gives the network no route to the
// internet (docker network create --internal) — used for the network the
// execution container attaches to, so its only path out is the proxy.
func ensureNetwork(ctx context.Context, run commandRunner, name string, internal bool) error {
	if _, err := run.Run(ctx, "docker", "network", "inspect", name); err == nil {
		return nil
	}

	args := []string{"network", "create"}
	if internal {
		args = append(args, "--internal")
	}
	args = append(args, name)
	if _, err := run.Run(ctx, "docker", args...); err != nil {
		return fmt.Errorf("container: create network %s: %w", name, err)
	}
	return nil
}

// removeNetwork deletes the given Docker network. A missing network is
// not an error — teardown must be idempotent.
func removeNetwork(ctx context.Context, run commandRunner, name string) error {
	if _, err := run.Run(ctx, "docker", "network", "rm", name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("container: remove network %s: %w", name, err)
	}
	return nil
}

// startProxy starts a Squid container on internalNetwork (so the
// execution container can reach it) and also connects it to
// publicNetwork (so it can reach the allowlisted hosts) — the two-network
// sidecar pattern: the execution container itself is never attached to
// publicNetwork and has no other route to the internet.
func startProxy(ctx context.Context, run commandRunner, name, internalNetwork, publicNetwork string, allowlist []string) error {
	conf := renderSquidConf(allowlist)
	script := fmt.Sprintf("cat > /etc/squid/squid.conf <<'SQUIDCONF'\n%sSQUIDCONF\nexec squid -N -f /etc/squid/squid.conf\n", conf)

	args := []string{
		"run", "-d", "--name", name,
		"--network", internalNetwork,
		"--entrypoint", "sh",
		proxyImage, "-c", script,
	}
	if _, err := run.Run(ctx, "docker", args...); err != nil {
		return fmt.Errorf("container: start proxy %s: %w", name, err)
	}

	if _, err := run.Run(ctx, "docker", "network", "connect", publicNetwork, name); err != nil {
		return fmt.Errorf("container: connect proxy %s to %s: %w", name, publicNetwork, err)
	}
	return nil
}

// removeContainer force-removes the named container. A missing container
// is not an error — teardown must be idempotent.
func removeContainer(ctx context.Context, run commandRunner, name string) error {
	if _, err := run.Run(ctx, "docker", "rm", "-f", name); err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return nil
		}
		return fmt.Errorf("container: remove container %s: %w", name, err)
	}
	return nil
}
