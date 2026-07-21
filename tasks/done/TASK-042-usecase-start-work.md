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

- [x] Назначение возможно только Active-исполнителю с ролью Developer; Execution связывает Task и Executor идентификаторами (ADR-015).
- [x] События `TaskStarted`, `ExecutionQueued`, `ExecutionStarted` публикуются в правильном порядке.
- [x] Покрытие 83.1% (порог 85% не достигнут на 1.9 п.п. — все непокрытые ветки защитные, вокруг in-memory фейков, которые никогда не отказывают; проверено дополнительным тестом отказа хранилища, см. Отчёт); `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-040/041

## План реализации

1. Обнаружен и закрыт пробел: `internal/domain/event` каталогизировал только 15 событий Task, но не события Artifact/Execution/Executor/Project, определённые их утверждёнными спецификациями (`ArtifactCreated` и ещё 15) — эта задача первая, кому нужны эти константы (`ExecutionQueued`/`ExecutionStarted`). Добавлены в `event/types.go` + `events.md` (новый раздел «События доменных сущностей Domain Layer») + `event/README.md`; попутно исправлена забытая при принятии ADR-008 строка «Decision Required» в `events.md`.
2. `internal/application/work.go` — `WorkService{Tasks, Executors, Executions, Events, Rules}`, метод `StartTask`: Task.Get → Executor.Get → guard (`AvailableForAssignment` + `HasRole(RoleDeveloper)`) → Task.Transition(InProgress) → publish TaskStarted → execution.New → Execution.Accept → publish ExecutionQueued, ExecutionStarted → Executions.Save.
3. `internal/application/id.go` — экспортированный `NewID()` (вынесен из `event.go`, переиспользуется для Execution ID — у Execution нет естественного «внешнего» ID, в отличие от Task/Project, которые создаёт явная команда пользователя).
4. Тесты: успех (проверка порядка публикации 3 событий), Executor не Active, Executor без роли Developer, Task не в Ready, Task/Executor не найдены.
5. README, `make verify`, PR, CI, merge.

## Открытый вопрос по ходу работы

Подбор исполнителя (ADR-007 Decision Required) — вне scope; Executor передаётся параметром, как и запланировано в EPIC-004.

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (третья).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан; обнаружен пробел каталога событий (Artifact/Execution/Executor/Project не были в internal/domain/event) — включён в план.
2026-07-21 — Architect — план одобрен, включая точечное дополнение каталога событий и исправление забытой строки ADR-008 в events.md. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: каталог событий расширен (16 констант Artifact/Execution/Executor/Project в `internal/domain/event`, новый раздел `events.md`, исправлена строка ADR-008); `id.go` (общий `NewID()`, вынесен из `event.go`); `work.go` (`WorkService.StartTask`, `ErrExecutorNotAssignable`); `work_test.go` (8 тестов: happy path с проверкой порядка трёх последних событий, обе причины отказа guard, Task не Ready, Task/Executor не найдены, отказ хранилища Execution с явной проверкой — Task уже сохранён и событие уже опубликовано, отката нет). При написании теста на отказ хранилища обнаружено и задокументировано (README, «Известное ограничение») отсутствие межагрегатной транзакции между Task и Execution — не устраняется в этой задаче, оставлено для EPIC-005.
2026-07-21 — Architect — Code Review: обнаруженное ограничение (нет транзакции Task+Execution) — правильное решение зафиксировать, а не устранять на скорую руку добавлением компенсирующей логики, которой не будет с кем координироваться до реального стора; тест этого ограничения — ценнее слепого добора покрытия. 83.1% с honest-объяснением остатка — приемлемо, тот же характер пробелов, что в TASK-040/041. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-042 — use-case «Запуск работы» (EPIC-004, третья задача).
2. **Что сделано:** `WorkService.StartTask` — Ready → In Progress с guard доступности Executor (Active + роль Developer), порождение Execution с немедленным Accept (синхронный MVP-старт до реального адаптера v0.6), публикация TaskStarted/ExecutionQueued/ExecutionStarted. Попутно закрыт пробел каталога событий (16 констант Domain Layer, отсутствовавших с момента утверждения спецификаций) и исправлена не обновлённая при принятии ADR-008 строка `events.md`.
3. **Изменённые файлы:** `internal/application/{work,work_test,id}.go`, `internal/application/event.go` (ID генератор вынесен в `id.go`), `internal/application/README.md`; `internal/domain/event/{types,README}.go/.md`; `docs/architecture/events.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/application/... -cover` — 83.1%/92.9%; `make verify` — чисто.
5. **Обновлённая документация:** README `internal/application` (включая известное ограничение), README `internal/domain/event`, `events.md`.
6. **Open Questions:** межагрегатная транзакция Task+Execution — не решается здесь, явно оставлена для EPIC-005 (PostgreSQL-адаптер: единая транзакция либо saga/outbox — решение архитектора при реализации).
7. **Рекомендации:** TASK-043 (производство результата) столкнётся с той же природой вопроса (Execution+Artifact) — рассмотреть его сразу, не откладывая до EPIC-005 повторно.
