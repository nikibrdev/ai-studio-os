# TASK-042: Use-case «Запуск работы»

## Тип

feature

## Эпик

[EPIC-004 Application Layer](../../docs/roadmap/EPIC-004-application-layer.md)

## Цель

Use-case StartTask: перевод Ready → In Progress с назначением уже выбранного Executor (Active, роль Developer — проверка `AvailableForAssignment`/`HasRole`), порождением Execution (Queued → Accept → Running) и публикацией `TaskStarted` + `ExecutionQueued`/`ExecutionStarted`.

## Контекст

Golden path, шаг «Developer получает работу». Подбор исполнителя — вне scope (ADR-007 Decision Required): Executor передаётся параметром; выбор — будущий Orchestrator.

## Scope

### Входит

- Сервис StartTask: валидации (Task в Ready, Executor годен), переход через Machine, создание Execution, немедленный Accept (в MVP запуск синхронный — платформа фиксирует принятие работы), сохранения, события.
- Тесты: успех; Executor не Active/без роли; Task не в Ready; отказ Rules.

### Не входит

- Реальный вызов бэкенда через `platform.Executor` (v0.6); производство артефактов (TASK-043).

## Критерии приёмки

- [ ] Назначение возможно только Active-исполнителю с ролью Developer; Execution связывает Task и Executor идентификаторами (ADR-015).
- [ ] События `TaskStarted`, `ExecutionQueued`, `ExecutionStarted` публикуются в правильном порядке.
- [ ] Покрытие ≥ 85%; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-040/041

## План реализации

<Заполняется при взятии задачи в работу.>

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (третья).
