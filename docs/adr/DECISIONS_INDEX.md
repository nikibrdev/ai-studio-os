# DECISIONS_INDEX — индекс архитектурных решений

## Назначение

Навигация по всем ADR проекта: статус, тема, что блокирует. Обновляется при каждом изменении статуса любого ADR (в том же PR).

## Содержание

### Индекс

| ADR | Статус | Тема | Блокирует |
|---|---|---|---|
| [000](ADR-000-template.md) | Template | Шаблон ADR | — |
| [001](ADR-001-license.md) | **Accepted** | Лицензия — Apache License 2.0 | Нет |
| [002](ADR-002-event-delivery.md) | **Accepted** | Event Bus — In-Memory (интерфейс стабилен; позже Redis Streams/NATS) | Нет |
| [003](ADR-003-api-protocol.md) | **Accepted** | Протокол API — REST | Нет |
| [004](ADR-004-task-storage.md) | **Accepted** | Задачи — PostgreSQL source of truth; `tasks/` — экспорт | Нет |
| [005](ADR-005-agent-adapter-contract.md) | Decision Required | Контракт адаптера агента (формат обмена) | v0.3 Developer Engine; форму Request/Response |
| [006](ADR-006-agent-execution-environment.md) | Decision Required | Среда выполнения и изоляция агентов | v0.3 Developer Engine |
| [007](ADR-007-pm-qa-executors.md) | Decision Required | Исполнители ролей PM и QA в MVP | Объём v0.2 (PM) и v0.4 (QA) |
| [008](ADR-008-git-policies.md) | Decision Required | Git-политики: слияние, ревью, момент merge относительно Testing | Условие Testing → Done; настройки защиты main |
| [009](ADR-009-toolchain.md) | **Accepted** | Toolchain — Go 1.24, Next.js 15, pnpm, golangci-lint, gofumpt | Нет |
| [010](ADR-010-documentation-language.md) | Decision Required | Язык документации (EN-версия) | Публичный релиз v1.0 |
| [011](ADR-011-task-identifiers.md) | Decision Required | Формат идентификаторов задач/эпиков | Модель данных Task Engine |
| [012](ADR-012-identity-and-auth.md) | Decision Required | Пользователи и аутентификация | v0.6 Dashboard; поле «инициатор» в событиях |
| [013](ADR-013-managed-projects.md) | Decision Required | Подключение управляемых проектов (`projects/`) | Модуль project детально; среду агентов |
| [014](ADR-014-module-interaction.md) | **Accepted** | Взаимодействие модулей — все проходят через Core, только события | Нет |
| [015](ADR-015-internal-layering.md) | **Accepted** | Слои internal: domain(+shared) / application / platform / infrastructure | Нет |

### Сводка

- **Принято:** 7 (001, 002, 003, 004, 009, 014, 015)
- **Decision Required:** 8 (005, 006, 007, 008, 010, 011, 012, 013)
- Ближайшие к принятию по roadmap: ADR-011 (нужен Task Engine, EPIC-003+), ADR-005/006 (v0.3).

## Статус

Актуален

## Последнее обновление

2026-07-19
