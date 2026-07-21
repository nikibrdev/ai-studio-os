# TASK-053: Docker-образ исполнения на базе Dev Container

## Тип

feature

## Эпик

[EPIC-006 AI Agent Runtime](../../docs/roadmap/EPIC-006-ai-agent-runtime.md)

## Цель

Собрать воспроизводимый Docker-образ для контейнера Execution (ADR-006, целевая модель): рабочая среда для Claude Code внутри контейнера — git, инструменты сборки, сам Claude Code CLI. Базой служит уже проверенный Dev Container (`golang:1.24-bookworm`), а не образ с нуля.

## Контекст

`.devcontainer/devcontainer.json` использует `golang:1.24-bookworm` напрямую (без отдельного Dockerfile) + фичу Node LTS + `setup.sh` (corepack/pnpm, gofumpt, golangci-lint, git-хуки). git и make уже есть в базовом образе. Образ исполнения — не тот же контейнер, что Dev Container разработки: он должен содержать Claude Code CLI и минимум, необходимый для клонирования/сборки/тестов внутри Execution, без интерактивных инструментов разработки (VS Code server и т.п., которые ставит devcontainer runtime, а не сам образ).

## Scope

### Входит

- `docker/execution/Dockerfile` (путь уточнить в плане) — на базе `golang:1.24-bookworm`; установка Claude Code CLI; git уже есть в базовом образе (не переустанавливать).
- Документация: как собрать образ локально (`docker build`), какой тег использовать.
- Смоук-проверка: контейнер стартует, `claude --version` (или эквивалент) и `git --version` отвечают.

### Не входит

- Сетевой allowlist и инъекция секретов при запуске (TASK-054 — это про рантайм запуска контейнера, не про содержимое образа).
- Публикация образа в реестр — образ собирается локально/в CI по необходимости, публикация вне scope MVP.

## Критерии приёмки

- [x] Образ собирается командой из README без ошибок (проверено вживую, Docker поднят).
- [x] Внутри образа доступны: git, Claude Code CLI, инструменты сборки Go (совместимые с версией из `go.mod`).
- [x] Смоук-проверка (ручная) подтверждает старт контейнера и доступность инструментов — вывод см. в Отчёте.
- [x] README задокументирован; `make verify` — чисто.

## Затрагиваемые модули и документы

- `docker/execution/` (новое), README (новый или расширение `internal/infrastructure`/отдельный README рядом с Dockerfile — уточнить в плане).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — независима от TASK-052, может выполняться параллельно; требует Docker на машине исполнителя задачи

## План реализации

1. `docker/execution/Dockerfile` — на базе `golang:1.24-bookworm` (тот же, что Dev Container); `docker/` — уже существующий scaffolded каталог именно для этого (`project-structure.md`, README.md корня).
2. Node.js/npm — для Claude Code CLI и `markdownlint` (`npx` в `make verify`).
3. Пины `gofumpt`/`golangci-lint` — те же версии, что CI/Dev Container (BUGFIX-001/002).
4. `npm install -g @anthropic-ai/claude-code`.
5. Документация — `docs/operations/README.md` (был плейсхолдером с 2026-07-19; ADR-006 явно называет `docker/` и `docs/operations/` местом для этого материала).
6. Смоук-проверка: `docker build` + `docker run ... git/go/claude/gofumpt/golangci-lint --version`.

## История

2026-07-21 — Architect — EPIC-006 открыт; задача поставлена в очередь (вторая, не зависит от TASK-052).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен. Приступать.
2026-07-21 — Claude Code (Developer) — первая сборка (`apt-get install nodejs npm`, версия из Debian bookworm — v18.20.4) прошла с предупреждением `npm WARN EBADENGINE`: `@anthropic-ai/claude-code@2.1.216` требует Node `>=22.0.0`. Исправлено — Node ставится из NodeSource (`setup_22.x`), не из apt. Также при первых прогонах `go install`/`npm install` ловились транзитные сетевые сбои (TLS handshake timeout, unexpected EOF к `proxy.golang.org`/`sum.golang.org`) — не системная проблема версий (те же версии стабильно ставятся в CI каждый прогон), а сетевая нестабильность локальной машины/Docker Desktop; добавлены ретраи (5 попыток с паузой) в оба `RUN`-шага — общее, а не разовое усиление устойчивости сборки, не только для этой машины. После обоих исправлений — чистая сборка и смоук-проверка: `git 2.39.5`, `go1.24.13`, `node v22.23.1`, `claude 2.1.216`, `gofumpt v0.9.2`, `golangci-lint v2.8.0` — все версии верны и без предупреждений.
2026-07-21 — Architect — Code Review: переход на NodeSource обоснован конкретным зафиксированным предупреждением (EBADENGINE), не гипотетически; ретраи в Dockerfile — общее улучшение устойчивости, уместное независимо от причины, которая их выявила. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-053 — Docker-образ исполнения на базе Dev Container (вторая задача EPIC-006).
2. **Что сделано:** `docker/execution/Dockerfile` — образ на базе `golang:1.24-bookworm` с Node 22.x (NodeSource, не apt — Debian bookworm's Node 18 не удовлетворяет требованию Claude Code CLI `>=22`), `@anthropic-ai/claude-code`, `gofumpt`/`golangci-lint` тех же версий, что в CI. Ретраи добавлены в оба сетевых `RUN`-шага. `docs/operations/README.md` переписан (был плейсхолдером) — описание образа, что установлено и почему, команды сборки и смоук-проверки, явная граница «что НЕ входит в образ» (секреты, allowlist — TASK-054).
3. **Изменённые файлы:** `docker/execution/Dockerfile` (новый), `docs/operations/README.md` (переписан); файл задачи.
4. **Как проверялось:** `docker build -t ai-studio-os-execution -f docker/execution/Dockerfile .` — успешно (после исправления версии Node); смоук-проверка `docker run --rm ai-studio-os-execution bash -c "git --version && go version && node --version && claude --version && gofumpt --version && golangci-lint --version"` — все шесть команд отвечают ожидаемыми версиями; `make verify` — чисто (Dockerfile/README не участвуют в Go-тестах, но docs-check прошёл по новым ссылкам).
5. **Обновлённая документация:** `docs/operations/README.md`.
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-054 (жизненный цикл контейнера) может опираться на этот образ как на уже проверенный вживую артефакт; тег `ai-studio-os-execution:latest` собран локально — вопрос версионирования/публикации образа не решён и не требуется для MVP (см. Scope «Не входит»).
