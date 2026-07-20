#!/usr/bin/env bash
# Installs the repository's git hooks (TASK-021).
# Usage: bash scripts/install-hooks.sh (or `make install-hooks`)
set -e
cd "$(dirname "$0")/.."

git config core.hooksPath .githooks
chmod +x .githooks/pre-commit .githooks/pre-push

echo "Git hooks installed: core.hooksPath -> .githooks"
echo "  pre-commit: fmt-check, lint, vet (fast)"
echo "  pre-push:   make verify (full)"
