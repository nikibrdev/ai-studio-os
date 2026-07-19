# Слой: internal/platform

## Назначение

Контракты инфраструктурных абстракций платформы: EventBus, Agent, Tool, MemoryProvider, RepositoryProvider ([ADR-015](../../docs/adr/ADR-015-internal-layering.md)). Это не язык предметной области (он в [domain/shared](../domain/shared/README.md)) — это то, на чём платформа работает. Только интерфейсы; реализации — в `internal/infrastructure`, `agents/`, `tools/`.

## Содержание

### Ответственность

- Контракты: EventBus (Event, EventHandler, Subscription), Agent (Request/Response — абстрактны до ADR-005), Tool (ToolDescriptor), MemoryProvider (MemoryEntry), RepositoryProvider (PullRequestState).
- Концептуальный источник истины — [docs/architecture/interfaces.md](../../docs/architecture/interfaces.md).

### Зависимости

- Разрешено: только стандартная библиотека (и `pkg/` при необходимости).
- Запрещено: `internal/domain` (слой домен-агностичен), `internal/application`, `internal/infrastructure`, `apps/`, внешние библиотеки.

### События

Определяет контракт шины и требования к событиям (неизменяемость, версия схемы); сам событий не публикует и не потребляет.

### Ограничения

Изменение опубликованного контракта — только через ADR; интерфейс EventBus неизменен при замене реализации (In-Memory → Redis Streams / NATS, [ADR-002](../../docs/adr/ADR-002-event-delivery.md)).

## Статус

Актуален

## Последнее обновление

2026-07-19
