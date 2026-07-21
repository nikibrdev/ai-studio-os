# Образ исполнения Execution (ADR-006, EPIC-006 TASK-053).
#
# База - тот же golang:1.24-bookworm, что и Dev Container разработки
# (.devcontainer/devcontainer.json): тот же закреплённый тулчейн
# (ADR-009), git уже в базовом образе. В отличие от Dev Container - без
# VS Code server и интерактивных инструментов разработки: только то, что
# нужно Executor'у внутри Execution (клонировать ветку задачи, писать
# код, прогонять make verify, коммитить).
FROM golang:1.24-bookworm

# Тот же жёсткий пин тулчейна, что в CI и Dev Container (BUGFIX-002).
ENV GOTOOLCHAIN=local

# Node.js/npm - для Claude Code CLI (npm-пакет) и markdownlint (шаг
# md-lint в make verify использует npx). Node из apt в bookworm - v18,
# а Claude Code CLI требует >=22 (обнаружено при сборке: npm ставит
# пакет с EBADENGINE-предупреждением на v18) - ставим Node 22.x из
# NodeSource, не из apt.
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates curl gnupg \
	&& curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
	&& apt-get install -y --no-install-recommends nodejs \
	&& rm -rf /var/lib/apt/lists/*

# Те же версии, что закреплены в CI (.github/workflows/verify.yml) и
# Dev Container (.devcontainer/setup.sh) - не должны расходиться
# (тот же принцип, что BUGFIX-001/002). Ретраи - go-модули скачиваются
# из нескольких внешних хостов (proxy.golang.org, sum.golang.org);
# единичный сетевой сбой на одном из них не должен требовать пересборки
# всего шага с нуля.
RUN for i in 1 2 3 4 5; do \
		go install mvdan.cc/gofumpt@v0.9.2 \
		&& go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0 \
		&& break || { echo "retry $i/5"; sleep 5; }; \
	done

RUN for i in 1 2 3 4 5; do \
		npm install -g @anthropic-ai/claude-code && break || { echo "retry $i/5"; sleep 5; }; \
	done

WORKDIR /workspace
