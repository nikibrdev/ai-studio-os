# Контекст: Архитектура

## Назначение

Сжатая выжимка архитектуры для быстрой загрузки контекста AI-агентом. Источник истины — [docs/architecture/](../../docs/architecture/).

## Содержание

### Главное

1. **Modular Monolith**: один деплоймент, жёсткие границы модулей ([module-boundaries.md](../../docs/architecture/module-boundaries.md)).
2. **Event-Driven**: междоменное взаимодействие — только через события ([event-model.md](../../docs/architecture/event-model.md), каталог — [events.md](../../docs/architecture/events.md)).
3. **Clean Architecture**: ядро (`internal/`, 10 модулей — [core.md](../../docs/architecture/core.md)) не зависит от инфраструктуры; подключение — адаптерами через порты.
4. **Agent-agnostic**: ядро не знает о конкретных AI-моделях; агенты — адаптеры в `agents/` по контракту Agent ([interfaces.md](../../docs/architecture/interfaces.md)).
5. **Расширяемость без изменения ядра**: агенты, инструменты, подписчики событий добавляются без правок ядра.

### Части системы

- `apps/api/` — backend API (Go), без доменной логики.
- `apps/dashboard/` — веб-интерфейс (Next.js), работает только через API.
- `apps/orchestrator/` — координация процесса и запуск агентов; без доменных правил.
- `internal/` — ядро по слоям (ADR-015): `domain/` (предметная область: `shared` — Role/TaskState; модули task, project, event, workflow), `application/`, `platform/` (абстракции: EventBus, Agent, Tool, MemoryProvider, RepositoryProvider; домен-агностичен), `infrastructure/`; `pkg/` — утилиты без домена.
- `agents/`, `tools/`, `memory/` — адаптеры агентов, инструменты, знания.
- `tasks/` — файловый жизненный цикл задач; канонические состояния — [state-machine.md](../../docs/architecture/state-machine.md) (Backlog → Ready → In Progress → Review → Testing → Done; Blocked, Cancelled, Archived).

### Чего делать нельзя

- Принимать архитектурные решения без ADR (агент — вообще не принимает, фиксирует Open Questions).
- Нарушать направление зависимостей и границы модулей ([module-boundaries.md](../../docs/architecture/module-boundaries.md)).
- Добавлять каталоги верхнего уровня и технологии вне стека.
- Реализовывать что-либо, зависящее от ADR со статусом Decision Required, до принятия решения.

### Architecture Freeze

Архитектура заморожена (2026-07-19). Приняты: ADR-002 (In-Memory Event Bus, интерфейс неизменен), ADR-003 (REST), ADR-004 (PostgreSQL — источник истины задач; `tasks/` — экспорт; переходный период — файлы), ADR-009 (Go 1.24, Next.js 15, pnpm, golangci-lint, gofumpt, единый `go.mod`), ADR-014 (все проходят через Core; запрещены Tool → Core, Agent → Database, Workflow → SQL). Остальные ADR — Decision Required, реализация, зависящая от них, не начинается. Изменение замороженной архитектуры — только новым ADR.

## Статус

Актуален

## Последнее обновление

2026-07-19
