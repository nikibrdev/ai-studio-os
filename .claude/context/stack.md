# Контекст: Технологический стек

## Назначение

Утверждённый технологический стек проекта. Использование технологий вне этого списка запрещено без ADR.

## Содержание

### Стек

| Назначение | Технология | Примечание |
| --- | --- | --- |
| Backend | **Go 1.24** | единый `go.mod` в корне ([ADR-009](../../docs/adr/ADR-009-toolchain.md)) |
| Frontend | **Next.js 15** | `apps/dashboard` |
| Package manager (frontend) | **pnpm** | [ADR-009](../../docs/adr/ADR-009-toolchain.md) |
| Linter (Go) | **golangci-lint** | [ADR-009](../../docs/adr/ADR-009-toolchain.md) |
| Formatter (Go) | **gofumpt** | строже gofmt |
| API | **REST** | Dashboard → REST → Go Core ([ADR-003](../../docs/adr/ADR-003-api-protocol.md)) |
| События | **In-Memory Event Bus** | интерфейс неизменен; будущая замена: Redis Streams / NATS ([ADR-002](../../docs/adr/ADR-002-event-delivery.md)) |
| База данных | PostgreSQL | **источник истины задач** ([ADR-004](../../docs/adr/ADR-004-task-storage.md)); подключение с v0.2 |
| Кэш | Redis | с v0.2; не используется для доставки событий в MVP |
| Векторный поиск | Qdrant | с v0.7 (Memory) |
| Тестирование e2e | Playwright | с v0.4 (QA Engine) |
| Git-хостинг | GitHub | ветки, PR, ревью |
| Контейнеры | Docker | Compose — после снятия ограничений Foundation |
| AI Developer | Claude Code | исполнитель роли Developer по умолчанию |

### Правила

1. Новая библиотека или сервис — только через ADR.
2. Версия Node.js, конфигурация ESLint/Prettier и unit-фреймворк frontend фиксируются при создании `apps/dashboard` (см. [ADR-009](../../docs/adr/ADR-009-toolchain.md)).
3. Агенты используют технологии только из этого списка и из явных указаний задачи.

## Статус

Актуален (архитектура заморожена — [overview.md](../../docs/architecture/overview.md))

## Последнее обновление

2026-07-19
