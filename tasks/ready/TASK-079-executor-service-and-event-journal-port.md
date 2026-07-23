# TASK-079: ExecutorService, ExecutorStore.List, порт EventJournal

## Тип

feature

## Эпик

[EPIC-010 Orchestrator](../../docs/roadmap/EPIC-010-orchestrator.md)

## Цель

Добавить в `internal/application` три недостающих элемента, без которых Orchestrator не может ни зарегистрировать исполнителя, ни найти его, ни узнать о новых событиях из отдельного процесса: `ExecutorService` (Register/Activate), `ExecutorStore.List` (запрос) и новый порт `EventJournal.Since`.

## Контекст

Разбор при открытии EPIC-010 («Контекст» эпика) показал: `internal/domain/executor` полностью реализован (EPIC-003), но ни одного use-case для регистрации/активации исполнителя в Application Layer нет — тесты создают `Executor` напрямую через домен. Отдельно: продуктивная `eventbus.Bus` — только внутрипроцессная (ADR-002), поэтому Orchestrator как отдельный процесс не может подписаться на неё; нужен курсорный опрос журнала событий через новый узкий порт (реализация — TASK-080), а не прямой доступ к `internal/infrastructure` (запрещён `module-boundaries.md`).

## Scope

### Входит

- `internal/application/executor.go` (новый файл) — `ExecutorService{Executors ExecutorStore, Events platform.EventBus}`, по стилю `project.go` (TASK-064): `Register(ctx, params) (*executor.Executor, error)` (оборачивает `executor.New`, публикует `ExecutorRegistered`), `Activate(ctx, id, actor string) error` (оборачивает `Executor.Activate`, публикует `ExecutorActivated`).
- `internal/application/ports.go` — `ExecutorStore.List(ctx) ([]*executor.Executor, error)`, по прецеденту `ProjectStore.List` (EPIC-009); новый порт `EventJournal interface { Since(ctx context.Context, after time.Time) ([]platform.Event, error) }`.
- `internal/application/inmemory` — реализация `List` на фейке `ExecutorStore` (по образцу фейка `ProjectStore`).
- Юнит-тесты: успешные пути `Register`/`Activate`, отказные сценарии домена (`ErrMissingField`, `ErrNoRoles`, `ErrAlreadyActive`, `ErrRetired`), `List` на пустом и непустом хранилище.

### Не входит

- Реализация `EventJournal` (курсорный SQL-запрос) — TASK-080.
- Использование `ExecutorService`/`List`/`EventJournal` из `apps/orchestrator` — TASK-081/082.
- HTTP-эндпоинты для исполнителей — не входит в эпик вовсе (см. «Не входит» EPIC-010).

## Критерии приёмки

- [ ] `ExecutorService.Register`/`Activate` реализованы, стиль идентичен `ProjectService`/`TaskPlanningService`.
- [ ] `ExecutorStore.List` объявлен и реализован в `inmemory`-фейке.
- [ ] Порт `EventJournal` объявлен в `ports.go` (без реализации — она в TASK-080).
- [ ] Юнит-тесты покрывают успешные и отказные пути.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/executor.go`, `internal/application/executor_test.go`, `internal/application/ports.go`, `internal/application/inmemory/store.go` (или аналог), `internal/application/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — использует уже принятый `internal/domain/executor`, не меняет его

## План реализации

<Заполняется исполнителем до начала работы; реализация начинается только после утверждения плана.>

## История

2026-07-23 — Architect — EPIC-010 открыт; задача поставлена в очередь первой (остальные задачи эпика зависят от неё).

## Отчёт о выполнении

<Заполняется исполнителем после завершения.>
