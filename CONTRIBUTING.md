# Вклад в AI Studio OS

## Назначение

Пошаговая инструкция для нового участника проекта — человека или AI-агента: как получить рабочее окружение и внести изменение по принятому процессу. Правила процесса — источник истины [CLAUDE.md](CLAUDE.md), [CONSTITUTION.md](CONSTITUTION.md), [docs/development/](docs/development/); этот документ — практический маршрут по ним.

## Содержание

### Быстрый старт

**Вариант A — Dev Container (рекомендуется, минимум ручных шагов):**

1. Установите [Docker](https://www.docker.com/) и расширение VS Code [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) (или используйте `devcontainer` CLI).
2. Клонируйте репозиторий и откройте его в VS Code → «Reopen in Container» (либо `devcontainer up --workspace-folder .`).
3. Готово: Go 1.24, Node LTS, pnpm, `gofumpt`, `golangci-lint`, `make` и git-хуки уже настроены — `postCreateCommand` делает это автоматически ([.devcontainer/](.devcontainer/)).

**Вариант B — локально:**

1. Установите Go 1.24, Node.js (LTS) и Git.
2. Клонируйте репозиторий:

   ```sh
   git clone https://github.com/nikibrdev/ai-studio-os.git
   cd ai-studio-os
   ```

3. Одна команда — рабочее окружение готово:

   ```sh
   make install-hooks
   ```

   Это настраивает git-хуки: `pre-commit` (быстрые проверки) и `pre-push` (полный `make verify`) — [engineering/decisions/2026-07-19-quality-gates.md](engineering/decisions/2026-07-19-quality-gates.md).

4. Проверьте, что всё работает:

   ```sh
   make verify
   ```

### Инструменты и их версии

Точные версии зафиксированы в [ADR-009](docs/adr/ADR-009-toolchain.md) и `.github/workflows/verify.yml` — тот же `golangci-lint`/`gofumpt`, что в CI и Dev Container, во избежание расхождений ([engineering/decisions/2026-07-20-pin-ci-tool-versions.md](engineering/decisions/2026-07-20-pin-ci-tool-versions.md)).

| Команда | Что делает |
| --- | --- |
| `make help` | Список всех целей |
| `make build` | `go build ./...` |
| `make fmt` / `make fmt-check` | Форматирование `gofumpt` |
| `make lint` | `golangci-lint run` |
| `make test` | `go test ./...` |
| `make verify` | Все проверки разом — тот же набор, что в CI |
| `make install-hooks` | Установка git-хуков |

### Процесс внесения изменений

1. Задача берётся из `tasks/ready/` (или ставится вами) по шаблону [.claude/templates/Task.md](.claude/templates/Task.md).
2. Ветка — по правилам [git-workflow.md](docs/development/git-workflow.md): `feature/`, `bugfix/`, `docs/`, `refactor/` + `<task-id>-<short-name>`.
3. Коммиты — [Conventional Commits](docs/development/git-workflow.md#коммиты--conventional-commits); проверяются в CI (`scripts/check-commits.sh`).
4. Для новых пакетов — README + Specification обязательны до кода ([docs/specifications/README.md](docs/specifications/README.md)).
5. PR — по шаблону [.github/PULL_REQUEST_TEMPLATE.md](.github/PULL_REQUEST_TEMPLATE.md); чек-лист [.claude/checklists/PR.md](.claude/checklists/PR.md) перед запросом ревью.
6. `main` защищена: прямые коммиты и push невозможны — только через PR с зелёным `verify` ([docs/roadmap/EPIC-002.5-engineering-platform.md](docs/roadmap/EPIC-002.5-engineering-platform.md)).

### Архитектура и правила проекта

- [CONSTITUTION.md](CONSTITUTION.md) — принципы и правила проекта.
- [docs/architecture/overview.md](docs/architecture/overview.md) — обзор архитектуры (заморожена, [ADR](docs/adr/DECISIONS_INDEX.md)).
- [PROJECT_MANIFEST.md](PROJECT_MANIFEST.md) — текущее состояние проекта одним взглядом.
- [ROADMAP.md](ROADMAP.md) — релизные вехи.

### Если вы — AI-агент

Дополнительно прочитайте [CLAUDE.md](CLAUDE.md) — инструкция по роли Developer, обязательному чтению перед задачей, порядку выполнения и запретам.

## Статус

Актуален

## Последнее обновление

2026-07-20
