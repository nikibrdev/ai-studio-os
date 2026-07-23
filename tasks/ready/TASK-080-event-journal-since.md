# TASK-080: ReadJournalSince и подключение EventJournal в wiring.System

## Тип

feature

## Эпик

[EPIC-010 Orchestrator](../../docs/roadmap/EPIC-010-orchestrator.md)

## Цель

Реализовать курсорный запрос к журналу событий (`ReadJournalSince`) и подключить его как реализацию порта `EventJournal` (TASK-079) в `internal/infrastructure/wiring.System` — единственный способ, которым `apps/orchestrator`, будучи отдельным процессом, может узнавать о новых событиях (см. «Контекст» EPIC-010: продуктивная `eventbus.Bus` — только внутрипроцессная).

## Контекст

`internal/infrastructure/eventbus/journal.go` уже содержит `ReadJournal` (весь журнал целиком, для восстановления проекций) и таблицу `event_journal` (миграция 0004). Нужно добавить выборку «после момента времени», не заменяя существующую функцию (`ReadJournal` продолжает использоваться для полного восстановления).

## Scope

### Входит

- `internal/infrastructure/eventbus/journal.go` — `ReadJournalSince(ctx context.Context, pool *pgxpool.Pool, after time.Time) ([]platform.Event, error)`: тот же запрос, что `ReadJournal`, плюс `WHERE occurred_at > $1 ORDER BY occurred_at`.
- `internal/infrastructure/wiring/wiring.go` — `System.EventJournal` (реализация нового порта `application.EventJournal` из TASK-079), обёртка над `ReadJournalSince` с уже собранным пулом.
- Юнит-тест на fake-`execer`/query builder (по образцу существующих тестов `bus_test.go`), интеграционный тест на реальном PostgreSQL (`//go:build integration`, по образцу `TestReadJournal_ReturnsReconstructedEvents`): публикует несколько событий с разным временем, проверяет, что `ReadJournalSince` возвращает только события после курсора.

### Не входит

- Само использование в `apps/orchestrator` (цикл опроса, курсор в памяти) — TASK-081.
- Устойчивое хранение курсора между перезапусками — сознательно не входит в эпик (см. «Риски» EPIC-010).

## Критерии приёмки

- [ ] `ReadJournalSince` возвращает только события, произошедшие строго после переданного момента, в порядке `occurred_at`.
- [ ] `wiring.System` предоставляет `EventJournal`, реализующий порт `application.EventJournal`.
- [ ] Интеграционный тест на реальном PostgreSQL подтверждает курсорную выборку.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/eventbus/journal.go`, `internal/infrastructure/eventbus/journal_test.go` (новый), `internal/infrastructure/eventbus/bus_integration_test.go` (или новый файл для курсорного теста), `internal/infrastructure/wiring/wiring.go`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-079 (порт `EventJournal` должен быть объявлен)

## План реализации

<Заполняется исполнителем до начала работы; реализация начинается только после утверждения плана.>

## История

2026-07-23 — Architect — EPIC-010 открыт; задача поставлена в очередь, зависит от TASK-079.

## Отчёт о выполнении

<Заполняется исполнителем после завершения.>
