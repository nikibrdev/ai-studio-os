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
- ADR-001 (принят): лицензия **Apache License 2.0**; полный текст в LICENSE.
- Двухуровневая документация модулей: краткий README в модуле + полная спецификация в `docs/specifications/{domain,application,platform,infrastructure}` (шаблон Specification.md); правило нового пакета: README + Specification + TASK + Acceptance Criteria.
- Уровни проверок: pre-commit (fmt+lint+vet) → pre-push (`make verify`) → GitHub Actions (`make verify`, без исключений); devcontainer подтверждён (минимальный состав); `.gitattributes` с нормализацией окончаний строк.
- Правило необратимости архитектурных решений закреплено в шаблоне ADR и engineering/decisions.

#### Changed

- Канонический жизненный цикл задачи расширен состояниями Testing и Cancelled (`docs/architecture/state-machine.md`); зависимые документы синхронизированы.
- Обновлены по итогам аудита: overview, system-design, components, event-model, workflow (`docs/architecture/`), глоссарий и контекст (`.claude/context/`), CLAUDE.md, README каталогов `tasks/`.
- **Architecture Freeze (2026-07-19)**: приняты ADR-002 (In-Memory Event Bus), ADR-003 (REST API), ADR-004 (PostgreSQL — источник истины задач, `tasks/` — экспорт), ADR-009 (Go 1.24, Next.js 15, pnpm, golangci-lint, gofumpt), ADR-014 (все проходят через Core); документация синхронизирована с решениями.

## Статус

Актуален

## Последнее обновление

2026-07-19
