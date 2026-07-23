# DECISIONS_INDEX — индекс архитектурных решений

## Назначение

Навигация по всем ADR проекта: статус, тема, что блокирует. Обновляется при каждом изменении статуса любого ADR (в том же PR).

## Содержание

### Индекс

| ADR | Статус | Тема | Блокирует |
| --- | --- | --- | --- |
| [000](ADR-000-template.md) | Template | Шаблон ADR | — |
| [001](ADR-001-license.md) | **Accepted** | Лицензия — Apache License 2.0 | Нет |
| [002](ADR-002-event-delivery.md) | **Accepted** | Event Bus — In-Memory (интерфейс стабилен; позже Redis Streams/NATS) | Нет |
| [003](ADR-003-api-protocol.md) | **Accepted** | Протокол API — REST | Нет |
| [004](ADR-004-task-storage.md) | **Accepted** | Задачи — PostgreSQL source of truth; `tasks/` — экспорт | Нет |
| [005](ADR-005-executor-contract.md) | **Accepted** | Executor Contract — четыре возможности (Accept/Artifacts/Status/Finish) | Нет |
| [006](ADR-006-agent-execution-environment.md) | **Accepted** | Среда агентов — Docker-контейнер на Execution (сеть по allowlist, короткоживущие секреты); до v0.6 — только локально под надзором человека | Нет |
| [007](ADR-007-pm-qa-executors.md) | **Accepted** | Исполнители ролей PM и QA в MVP — обе роли агенты (Вариант 2), человек подтверждает на контрольных точках; технически — один и тот же адаптер `agents/claude-code`, роль различает промпт | Нет |
| [008](ADR-008-git-policies.md) | **Accepted** | Git-политики: merge commit; слияние после Testing (TestsPassed → MergeCompleted → TaskCompleted); 1 ревьюер, агент допустим | Нет |
| [009](ADR-009-toolchain.md) | **Accepted** | Toolchain — Go 1.24, Next.js 15, pnpm, golangci-lint, gofumpt | Нет |
| [010](ADR-010-documentation-language.md) | **Accepted** | Язык документации — только русский для v1.0 (Вариант 1); перевод — по факту реальной потребности | Нет |
| [011](ADR-011-task-identifiers.md) | **Accepted** | Идентификаторы — `TASK-NNN`/`EPIC-NNN`, последовательные в рамках Project; суррогатный ключ в БД; выдача — модуль `task` (последовательность на проект, v0.5) | Нет |
| [012](ADR-012-identity-and-auth.md) | **Accepted** | Пользователи и аутентификация — отложены (Вариант 1): доверенная однопользовательская установка без auth в v0.9 API/v0.8 Dashboard; пересмотр при появлении внешнего потребителя | Нет |
| [013](ADR-013-managed-projects.md) | **Accepted** | Управляемые проекты — метаданные в PostgreSQL на агрегате `Project` (не файлы); рабочие копии эфемерны (клон на Execution); `projects/` не используется — ретроспективная формализация уже сложившейся реализации | Нет |
| [014](ADR-014-module-interaction.md) | **Accepted** | Взаимодействие модулей — все проходят через Core, только события | Нет |
| [015](ADR-015-internal-layering.md) | **Accepted** | Слои internal: domain(+shared) / application / platform / infrastructure | Нет |
| [016](ADR-016-artifact-aggregate-root.md) | **Accepted** | Artifact — самостоятельный Aggregate Root, не часть Execution/Task/Project | Нет |
| [017](ADR-017-postgresql-driver.md) | **Accepted** | Драйвер PostgreSQL — `pgx/v5` (нативный интерфейс, `pgxpool`); миграции — самописный раннер по `.sql`-файлам | Нет |
| [018](ADR-018-memory-embeddings-and-qdrant-schema.md) | **Accepted** | Эмбеддинги Memory — наивный локальный feature hashing (256 измерений, без внешних зависимостей); схема Qdrant — одна коллекция `memory_entries` | Нет |

### Сводка

- **Принято:** 18 (001, 002, 003, 004, 005, 006, 007, 008, 009, 010, 011, 012, 013, 014, 015, 016, 017, 018)
- **Decision Required:** 0
- Все архитектурные решения проекта приняты по состоянию на 2026-07-23 (открытие v1.0 First Public MVP).

## Статус

Актуален

## Последнее обновление

2026-07-23
