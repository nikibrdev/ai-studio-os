# Документация эксплуатации

## Назначение

Каталог для документации по развёртыванию, настройке и эксплуатации AI Studio OS.

## Содержание

### Образ исполнения Execution (`docker/execution/`)

Docker-образ для контейнера Execution ([ADR-006](../adr/ADR-006-agent-execution-environment.md), [EPIC-006](../roadmap/EPIC-006-ai-agent-runtime.md) TASK-053) — целевая среда, в которой Executor (Claude Code) выполняет назначенную задачу: одна рабочая копия ветки задачи на Execution, воспроизводимый набор инструментов, без интерактивных инструментов разработки.

**База** — `golang:1.24-bookworm`, тот же образ, что и Dev Container разработки (`.devcontainer/devcontainer.json`): тот же закреплённый тулчейн Go ([ADR-009](../adr/ADR-009-toolchain.md)), git уже в базовом образе. Отличие от Dev Container: без VS Code server, без фичи Node через devcontainer-features (Node/npm ставятся напрямую через `apt`) — образ предназначен для автоматического запуска, не для интерактивной разработки.

**Установлено сверх базового образа**:

- `nodejs`/`npm` (apt) — для Claude Code CLI и `markdownlint` (шаг `md-lint` в `make verify` использует `npx`).
- `@anthropic-ai/claude-code` (npm, глобально) — сам исполнитель.
- `gofumpt@v0.9.2`, `golangci-lint@v2.8.0` — те же версии, что закреплены в CI (`.github/workflows/verify.yml`) и Dev Container (`.devcontainer/setup.sh`); Executor должен иметь возможность прогнать `make verify` перед завершением работы, как и человек-разработчик (тот же процесс, [CLAUDE.md](../../CLAUDE.md)).

**Сборка и смоук-проверка**:

```bash
docker build -t ai-studio-os-execution -f docker/execution/Dockerfile .
docker run --rm ai-studio-os-execution bash -c "git --version && go version && claude --version && gofumpt --version && golangci-lint --version"
```

**Что этот образ НЕ включает** (см. TASK-054/055 и риски EPIC-006):

- Клонирование рабочей копии, сетевой allowlist, инъекцию секретов — это рантайм-часть запуска контейнера (жизненный цикл Execution), не содержимое образа.
- Учётные данные любого рода — образ не содержит секретов; git-токен и ключ AI-провайдера передаются контейнеру при старте переменными окружения (ADR-006) и никогда не сохраняются в образе или слоях.

## Статус

Актуален (EPIC-006)

## Последнее обновление

2026-07-21
