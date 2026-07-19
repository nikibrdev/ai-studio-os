# TASK-007: Интерфейс Memory Provider

## Тип

feature

## Эпик

EPIC-002 ([docs/roadmap/EPIC-002-foundation.md](../../docs/roadmap/EPIC-002-foundation.md))

## Цель

Контракт памяти в `internal/core`: MemoryEntry, MemoryProvider (запись, поиск) — по [interfaces.md](../../docs/architecture/interfaces.md) и [memory.md](../../docs/architecture/memory.md).

## Контекст

Реализация памяти — v0.7 (файловый адаптер, затем Qdrant); контракт фиксируется сейчас, чтобы ядро проектировалось против него. Изоляция проектов и «память — не источник истины» — ограничения контракта.

## Scope

### Входит

- `internal/core/memory.go`: интерфейсы MemoryEntry, MemoryProvider; doc-комментарии с ограничениями.

### Не входит

- Реализации; структура каталога `memory/`; политика записи/устаревания (v0.7).

## Критерии приёмки

- [ ] `go build ./...`, `go vet ./...` — без ошибок; форматирование gofumpt.
- [ ] Только интерфейсы; ограничения зафиксированы в doc-комментариях.

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

1. **Задача:** TASK-007 — интерфейс Memory Provider.
2. **Что сделано:** создан `internal/core/memory.go`: интерфейсы MemoryEntry (ID, ProjectID, Kind, Content, Source, RecordedAt) и MemoryProvider (Record, Search с изоляцией по проекту). Ограничения (память — не источник истины; изоляция проектов; заменяемость файловой реализации на Qdrant без изменения интерфейса) — в doc-комментариях.
3. **Изменённые файлы:** `internal/core/memory.go` (новый).
4. **Как проверялось:** `go build ./...` — OK; `go vet ./...` — OK; `gofumpt -l .` — чисто.
5. **Обновлённая документация:** не требуется.
6. **Open Questions:** нет (таксономия Kind — v0.7).
7. **Рекомендации:** нет.
