# Changelog

## Назначение

История заметных изменений проекта. Ведётся по формату [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/); версионирование — по [Semantic Versioning](https://semver.org/lang/ru/).

## Содержание

Типы изменений:

- **Added** — новая функциональность.
- **Changed** — изменения существующей функциональности.
- **Deprecated** — функциональность, которая будет удалена.
- **Removed** — удалённая функциональность.
- **Fixed** — исправления ошибок.
- **Security** — исправления уязвимостей.

### [Unreleased]

#### Added

- Структура репозитория (v0.1 Foundation).
- Базовая документация: README, CONSTITUTION, CLAUDE, ROADMAP, CHANGELOG.
- Архитектурная документация (`docs/architecture/`).
- Документация процессов разработки (`docs/development/`).
- Шаблон ADR (`docs/adr/ADR-000-template.md`).
- Система задач (`tasks/`) с описанием жизненного цикла.
- Роли, команды, шаблоны, чек-листы и контекст для AI-агентов (`.claude/`).
- Architecture v0.2: доменная модель, ядро, интерфейсы, границы модулей, state machine задачи, каталог событий, потоки данных, инженерные принципы (`docs/architecture/`) с Mermaid-диаграммами.
- Заготовки архитектурных решений ADR-001…ADR-014 со статусом Decision Required (`docs/adr/`).
- EPIC-002 Foundation: Go-модуль `ai-studio-os` (go.mod, `.golangci.yml`, Makefile) и контракты ядра — интерфейсы без реализации: `internal/core` (EventBus, Agent, Tool, Workflow, RepositoryProvider, MemoryProvider, словари Role/TaskState) и доменные пакеты `internal/domain/{task,project,event,workflow}` (TASK-001…TASK-011).
- По итогам code review EPIC-002: слоистая структура `internal/{core,domain,application,infrastructure}`; Agent → `Execute(ctx, Request) (Response, error)` с абстрактными Request/Response до ADR-005; Engine/Reader → Commands/Queries; `make verify` (gofumpt → golangci-lint → vet → test → markdownlint → docs/Mermaid, `scripts/verify-docs.sh`); README в каждом модуле; каталог `engineering/` (reviews, retrospective, decisions, metrics); процесс «план → утверждение → код → ревью → merge» закреплён в CLAUDE.md.
- ADR-015 (принят): `internal/core` упразднён — платформенные абстракции → `internal/platform` (EventBus, Agent, Tool, MemoryProvider, RepositoryProvider), язык домена → `internal/domain/shared` (Role, TaskState), контракт Workflow → `workflow.Rules`; слои domain → application → platform → infrastructure.
- Управленческое ревью EPIC-002.5 (2026-07-20): ROADMAP.md пересмотрен под релизные вехи по слоям архитектуры (v0.2 Architecture & Engineering Platform → v1.0 First Public MVP); для Domain Layer (v0.3) введено обязательное утверждение спецификации (назначение, инварианты, допустимые состояния, события, Acceptance Criteria) до реализации; EPIC-002.6 Developer Experience утверждён к исполнению.
- ADR-001 (принят): лицензия **Apache License 2.0**; полный текст в LICENSE.
- Двухуровневая документация модулей: краткий README в модуле + полная спецификация в `docs/specifications/{domain,application,platform,infrastructure}` (шаблон Specification.md); правило нового пакета: README + Specification + TASK + Acceptance Criteria.
- Уровни проверок: pre-commit (fmt+lint+vet) → pre-push (`make verify`) → GitHub Actions (`make verify`, без исключений); devcontainer подтверждён (минимальный состав); `.gitattributes` с нормализацией окончаний строк.
- Правило необратимости архитектурных решений закреплено в шаблоне ADR и engineering/decisions.
- Паспорт проекта PROJECT_MANIFEST.md, прогресс PROJECT_HEALTH.md, индекс решений docs/adr/DECISIONS_INDEX.md.
- Запрет прямых коммитов в `main` без исключений (feature branch → PR → Review → Merge, даже соло; ревью может выполнять Claude).
- Метрики: `make metrics` (scripts/metrics.sh) — снимки в engineering/metrics/; первый снимок 2026-07-20.
- EPIC-002.5 Engineering Platform: репозиторий опубликован на GitHub ([nikibrdev/ai-studio-os](https://github.com/nikibrdev/ai-studio-os), public, Apache-2.0); GitHub Actions `verify` — обязательный статус-чек на каждый PR и push в main (Go 1.24, gofumpt, golangci-lint, markdownlint, docs-check); CODEOWNERS; шаблоны Issue (bug/feature/task) и тип коммита `ci`; защита ветки `main` (обязательный чек, запрет прямого push/force-push/удаления, `enforce_admins`); детальный конфиг markdownlint (MD060 включён, таблицы нормализованы); проверка Conventional Commits в CI (`scripts/check-commits.sh`); Dependabot (gomod, github-actions); релизный процесс — категоризация release notes по типам коммитов (`.github/release.yml`), метки типов в репозитории, раздел «Релизный процесс» в git-workflow.md (TASK-012…020).
- ADR-001 pinned CI tool-version bug (BUGFIX-001/002): `gofumpt`/`golangci-lint` в CI незаметно собирались на Go 1.25 вместо заявленного 1.24 из-за `GOTOOLCHAIN=auto` — обнаружено на Dependabot-обновлении `actions/setup-go`; закреплены совместимые версии (gofumpt v0.9.2, golangci-lint v2.8.0) и `GOTOOLCHAIN: local` на уровне job, чтобы расхождение больше не маскировалось.
- EPIC-002.6 Developer Experience: git-хуки `pre-commit`/`pre-push` (`.githooks/`, `make install-hooks`) — реально протестированы негативными сценариями; `.editorconfig`; рекомендуемые настройки VS Code (`.vscode/extensions.json`, `settings.json`); минимальный Dev Container (`.devcontainer/`) — реально собран и проверен (`devcontainer up`/`exec`, `0 issues.`), базовый образ переведён на Docker Hub `golang:1.24-bookworm` после обнаруженной недоступности `mcr.microsoft.com` из сети разработки; `CONTRIBUTING.md` — проверен сквозным клоном в чистую директорию (TASK-021…025).

#### Changed

- Канонический жизненный цикл задачи расширен состояниями Testing и Cancelled (`docs/architecture/state-machine.md`); зависимые документы синхронизированы.
- Обновлены по итогам аудита: overview, system-design, components, event-model, workflow (`docs/architecture/`), глоссарий и контекст (`.claude/context/`), CLAUDE.md, README каталогов `tasks/`.
- **Architecture Freeze (2026-07-19)**: приняты ADR-002 (In-Memory Event Bus), ADR-003 (REST API), ADR-004 (PostgreSQL — источник истины задач, `tasks/` — экспорт), ADR-009 (Go 1.24, Next.js 15, pnpm, golangci-lint, gofumpt), ADR-014 (все проходят через Core); документация синхронизирована с решениями.
- **ADR-005 принят (Executor Contract, 2026-07-20)**: Agent (логическая роль) и Executor (реальный технический бэкенд) разделены как понятия; контракт `internal/platform` — ровно четыре возможности (`Accept`, `Artifacts`, `Status`, `Finish`); `internal/platform/agent.go` переименован в `executor.go` (`Agent` → `Executor`, `Request`/`Response` → `ExecutorTask`/`Artifact`/`ExecutionStatus`, по-прежнему абстрактны до Domain Layer); терминология синхронизирована во всей документации, ссылавшейся на старый путь ADR или `platform.Agent` (TASK-026).

## Статус

Актуален

## Последнее обновление

2026-07-20
