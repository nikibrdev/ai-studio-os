# TASK-002: Интерфейс EventBus

## Тип

feature

## Эпик

EPIC-002 ([docs/roadmap/EPIC-002-foundation.md](../../docs/roadmap/EPIC-002-foundation.md))

## Цель

Контракт шины событий в `internal/core`: Event, EventHandler, Subscription, EventBus — строго по [interfaces.md](../../docs/architecture/interfaces.md) и [ADR-002](../../docs/adr/ADR-002-event-delivery.md).

## Контекст

ADR-002 (принят): In-Memory Event Bus в MVP; интерфейс неизменен при замене реализации. Общие поля событий — [events.md](../../docs/architecture/events.md). Только интерфейсы, без реализации.

## Scope

### Входит

- `internal/core/eventbus.go`: интерфейсы Event, Subscription, EventBus; тип EventHandler; doc-комментарии с ограничениями контракта (неизменяемость, идемпотентность обработчиков).

### Не входит

- Реализация шины; журнал; конкретные типы событий (TASK-008).

## Критерии приёмки

- [ ] `go build ./...`, `go vet ./...` — без ошибок; форматирование gofumpt.
- [ ] Только интерфейсы и типы; исполняемой логики нет.
- [ ] Контракт соответствует interfaces.md (ответственность, входы/выходы, ограничения — в doc-комментариях).

## Затрагиваемые модули и документы

- `internal/core/`

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-19 — Architect — задача определена в составе EPIC-002.
2026-07-19 — Claude Code (Developer) — задача оформлена в ready.
2026-07-19 — Claude Code (Developer) — выполнена, переведена в review.

## Отчёт о выполнении

1. **Задача:** TASK-002 — интерфейс EventBus.
2. **Что сделано:** созданы `internal/core/doc.go` (документация пакета контрактов) и `internal/core/eventbus.go`: интерфейсы Event (8 общих полей по events.md), Subscription, EventBus (Publish/Subscribe), тип EventHandler; ограничения контракта (неизменяемость событий, идемпотентность обработчиков, независимость подписчиков, стабильность интерфейса при замене In-Memory → Redis Streams/NATS) зафиксированы в doc-комментариях.
3. **Изменённые файлы:** `internal/core/doc.go`, `internal/core/eventbus.go` (новые).
4. **Как проверялось:** `go build ./...` — OK; `go vet ./...` — OK; `gofumpt -l .` — чисто.
5. **Обновлённая документация:** не требуется (контракт следует interfaces.md/ADR-002 дословно).
6. **Open Questions:** нет.
7. **Рекомендации:** In-Memory реализацию шины делать отдельной задачей этапа Infrastructure с unit-тестами на идемпотентную повторную доставку.
