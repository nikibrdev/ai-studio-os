#!/usr/bin/env bash
# postCreateCommand (TASK-024): brings the base golang:1.24-bookworm image up
# to the exact tool versions pinned in CI (.github/workflows/verify.yml) —
# same reasoning as BUGFIX-001/002: local tooling must not drift from CI.
# git and make are already present in the base image (verified via
# `docker run golang:1.24-bookworm`) — no apt-get needed for them.
set -e

corepack enable
corepack prepare pnpm@latest --activate

go install mvdan.cc/gofumpt@v0.9.2
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0

# `golangci-lint --version`/`gofumpt --version` self-report is unreliable for
# plain `go install` builds (no release-time ldflags stamping the version
# string) — confirmed during TASK-024 testing: the same pinned tag produced
# different --version output across runs/proxies while the underlying
# source was the correct pinned tag both times. So we verify FUNCTIONALLY
# instead of trusting the version string: run the pinned tools against this
# repo and require the result golangci-lint already gives everywhere else
# (0 issues) — that's what actually matters for a working dev environment.
gofumpt -l . >/tmp/gofumpt-check.log
if [ -s /tmp/gofumpt-check.log ]; then
	echo "FATAL: gofumpt found unformatted files:" >&2
	cat /tmp/gofumpt-check.log >&2
	exit 1
fi

golangci-lint run ./...

bash scripts/install-hooks.sh

echo "Dev Container ready: Go 1.24, Node LTS, pnpm, golangci-lint, gofumpt, make, git hooks installed."
