# AI Studio OS — Makefile
#
# Toolchain: ADR-009 (Go 1.24, golangci-lint, gofumpt).
# make verify — полная проверка одной командой:
#   gofumpt -> golangci-lint -> go vet -> go test -> markdownlint -> docs (ссылки + Mermaid)

.PHONY: help build vet fmt fmt-check lint test md-lint docs-check verify

help: ## Показать список целей
	@echo "AI Studio OS"
	@echo ""
	@echo "  make build      - go build ./..."
	@echo "  make vet        - go vet ./..."
	@echo "  make fmt        - gofumpt -l -w ."
	@echo "  make fmt-check  - проверка форматирования (без записи)"
	@echo "  make lint       - golangci-lint run"
	@echo "  make test       - go test ./..."
	@echo "  make md-lint    - markdownlint (npx; в CI обязателен)"
	@echo "  make docs-check - ссылки и Mermaid в документации"
	@echo "  make verify     - все проверки одной командой"

build:
	go build ./...

vet:
	go vet ./...

fmt:
	gofumpt -l -w .

fmt-check:
	@out="$$(gofumpt -l .)"; if [ -n "$$out" ]; then echo "gofumpt: файлы не отформатированы:"; echo "$$out"; exit 1; fi; echo "fmt-check: OK"

lint:
	golangci-lint run ./...

test:
	go test ./...

md-lint:
	@if command -v npx >/dev/null 2>&1; then npx --yes markdownlint-cli2 "**/*.md" "#node_modules"; else echo "md-lint: SKIPPED (npx не найден); в CI проверка обязательна"; fi

docs-check:
	@bash scripts/verify-docs.sh

verify: fmt-check lint vet test md-lint docs-check
	@echo ""
	@echo "verify: все проверки пройдены"
