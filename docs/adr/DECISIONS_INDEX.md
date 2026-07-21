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
| [007](ADR-007-pm-qa-executors.md) | Decision Required | Исполнители ролей PM и QA в MVP | Объём v0.2 (PM) и v0.4 (QA) |
| [008](ADR-008-git-policies.md) | **Accepted** | Git-политики: merge commit; слияние после Testing (TestsPassed → MergeCompleted → TaskCompleted); 1 ревьюер, агент допустим | Нет |
| [009](ADR-009-toolchain.md) | **Accepted** | Toolchain — Go 1.24, Next.js 15, pnpm, golangci-lint, gofumpt | Нет |
| [010](ADR-010-documentation-language.md) | Decision Required | Язык документации (EN-версия) | Публичный релиз v1.0 |
| [011](ADR-011-task-identifiers.md) | **Accepted** | Идентификаторы — `TASK-NNN`/`EPIC-NNN`, последовательные в рамках Project; суррогатный ключ в БД; выдача — модуль `task` (последовательность на проект, v0.5) | Нет |
| [012](ADR-012-identity-and-auth.md) | Decision Required | Пользователи и аутентификация | v0.6 Dashboard; поле «инициатор» в событиях |
| [013](ADR-013-managed-projects.md) | Decision Required | Подключение управляемых проектов (`projects/`) | Модуль project детально; среду агентов |
| [014](ADR-014-module-interaction.md) | **Accepted** | Взаимодействие модулей — все проходят через Core, только события | Нет |
| [015](ADR-015-internal-layering.md) | **Accepted** | Слои internal: domain(+shared) / application / platform / infrastructure | Нет |
| [016](ADR-016-artifact-aggregate-root.md) | **Accepted** | Artifact — самостоятельный Aggregate Root, не часть Execution/Task/Project | Нет |

### Сводка

- **Принято:** 12 (001, 002, 003, 004, 005, 006, 008, 009, 011, 014, 015, 016)
- **Decision Required:** 4 (007, 010, 012, 013)
- Ближайшие к принятию по roadmap: ADR-007 (исполнители PM/QA — v0.4), ADR-013 (подключение управляемых проектов — v0.5/v0.6).

## Статус

Актуален

## Последнее обновление

2026-07-21
