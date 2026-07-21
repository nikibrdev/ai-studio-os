# TASK-036: Реализация домен-модуля Executor

## Тип

feature

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 2 (Реализация)

## Цель

Пакет `internal/domain/executor` строго по утверждённой спецификации ([docs/specifications/domain/executor.md](../../docs/specifications/domain/executor.md)) — доменная сущность Executor: запись реестра технических бэкендов с идентичностью, набором исполняемых Role и статусом доступности. Не путать с платформенным контрактом адаптера `internal/platform.Executor` (ADR-005) — он не затрагивается.

## Контекст

Третья задача этапа 2 EPIC-003. Спецификация утверждена 2026-07-21 (TASK-031); ключевое решение финального ревью — прямой переход Registered → Retired разрешён. Роли — словарь `internal/domain/shared.Role` (разрешённая зависимость доменного слоя).

## Scope

### Входит

- `internal/domain/executor/` — value-типы (State), сущность Executor, команды Register/Activate/Disable/Retire/GrantRole/RevokeRole, события Registered/Activated/Disabled/Retired.
- Unit-тесты на каждый инвариант; README пакета; обновление `internal/domain/README.md`.

### Не входит

- Изменение `internal/platform/executor.go` (контракт адаптера, ADR-005).
- Критерии активации и автоматического Disable — Open Questions спецификации, Application/Infrastructure.
- Commands/Queries-интерфейсы — нет потребителя.

## Критерии приёмки

- [x] Все три Structural и четыре Behavioral инварианта реализованы проверяемым кодом; набор Role никогда не пуст (RevokeRole последней роли — ошибка).
- [x] Lifecycle: Registered → Active ⇄ Disabled → Retired, плюс прямой Registered → Retired; Retired терминален.
- [x] События Activated/Retired несут состояние-источник (единое событие на целевое состояние — паттерн спецификаций).
- [x] Unit-тесты покрывают успешные и запрещённые сценарии каждой команды; `make verify` — чисто; README создан.

## Затрагиваемые модули и документы

- `internal/domain/executor/` (новый пакет); `internal/domain/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — спецификация Executor утверждена (TASK-031); стиль — artifact/execution

## План реализации

По утверждённому стилю этапа 2: `doc.go`, `types.go` (State + Terminal), `events.go` (Registered/Activated/Disabled/Retired; Activated и Retired несут From), `executor.go` (сущность: id, backend-идентичность, набор `shared.Role`, state; New (Register) + Activate/Disable/Retire/GrantRole/RevokeRole; sentinel-ошибки, включая ErrLastRole и ErrRoleNotGranted; GrantRole уже имеющейся роли — идемпотентный no-op), `executor_test.go` (таблица тестов), README. GrantRole/RevokeRole допустимы в любом нетерминальном состоянии (Behavioral Invariant 3 спецификации).

## История

2026-07-21 — Architect — этап 2 EPIC-003: задача поставлена в очередь (третья по порядку проектирования).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен; идемпотентность GrantRole и явная ошибка RevokeRole для отсутствующей роли — приняты как уточнения реализации, не меняющие спецификацию. Приступать.
2026-07-21 — Claude Code (Developer) — пакет реализован: `doc.go`, `types.go`, `events.go`, `executor.go` (сущность + New/Activate/Disable/Retire/GrantRole/RevokeRole + предикаты AvailableForAssignment/HasRole, дедупликация начального набора ролей), `executor_test.go` (17 тестов, 94.5% покрытия, включая прямой Registered→Retired и запрет отзыва последней роли), `README.md`. `internal/domain/README.md` синхронизирован. `make verify` — чисто.
2026-07-21 — Architect — Code Review: реализация следует спецификации; разделение с платформенным контрактом соблюдено (пакет не импортирует `internal/platform`); предикат AvailableForAssignment корректно выражает Behavioral Invariant 4, не забирая у Application Layer само назначение. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-036 — реализация домен-модуля Executor (EPIC-003, этап 2, третья задача).
2. **Что сделано:** пакет `internal/domain/executor` строго по утверждённой спецификации — реестровая сущность с фиксированной идентичностью бэкенда, набором `shared.Role` (никогда не пуст) и Lifecycle Registered → Active ⇄ Disabled → Retired (включая прямой Registered → Retired из финального ревью спецификации); шесть команд, четыре события.
3. **Изменённые файлы:** `internal/domain/executor/{doc,types,events,executor,executor_test}.go`, `internal/domain/executor/README.md` (новые); `internal/domain/README.md` (обновлён); файл задачи.
4. **Как проверялось:** `go test ./internal/domain/executor/... -cover` — 17 тестов, 94.5% покрытия; `make verify` целиком — чисто.
5. **Обновлённая документация:** `internal/domain/executor/README.md`, `internal/domain/README.md`.
6. **Open Questions:** нет новых; открытые вопросы спецификации (критерии активации, автоматический Disable) — вне домена.
7. **Рекомендации:** следующая задача этапа 2 — TASK-037 (расширение `internal/domain/task`).
