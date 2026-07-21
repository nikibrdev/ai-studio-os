# TASK-037: Расширение и реализация домен-модуля Task

## Тип

feature

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 2 (Реализация)

## Цель

Закрыть оба решённых финальным ревью спецификации ([docs/specifications/domain/task.md](../../docs/specifications/domain/task.md), Decision Log) пробела контракта — `epicID` при создании и команды записи scope/критериев приёмки — и реализовать сущность Task с инвариантами в коде, с валидацией переходов через контракт `workflow.Rules` (ADR-014, инвариант 8 state-machine.md: «проверку допустимости перехода выполняет модуль workflow; состояние изменяет модуль task»).

## Контекст

Четвёртая задача этапа 2 EPIC-003. Спецификация Task утверждена 2026-07-21 (TASK-032) со статусом Stability: Provisional именно из-за этих двух пробелов — их закрытие было зафиксировано как «решённое направление расширения контракта на этапе 2», не как открытый вопрос. Зависимость task → workflow.Rules уже санкционирована контрактом (`contracts.go`: «validated … via the workflow rules contract») и README модуля. Идентификаторы остаются `string` до ADR-011.

## Scope

### Входит

- Расширение `contracts.go`: `Create` получает параметр `epicID` (пустая строка — без Epic, Structural Invariant 2); новые команды `SetScope` и `SetAcceptanceCriteria`.
- Сущность Task (`task.go`): id, projectID, epicID (0..1), title, тип, scope, критерии приёмки, состояние из канонических девяти; методы New/AttachToEpic/SetScope/SetAcceptanceCriteria/Transition.
- Правила сущности: scope/AC/привязка к Epic редактируются только в Backlog (Ready означает выполненный DoR; изменение требований позже — через возврат Ready → Backlog, как и задумано state machine: «требования изменились, DoR нарушен»); переход в Blocked/Cancelled требует непустой причины (инвариант 3 state-machine.md); допустимость перехода решает переданный `workflow.Rules`, не сама сущность.
- События `Created`/`Transitioned` как значения; отображение Transitioned в 15 канонических имён событий (events.md) — ответственность публикатора (Application Layer), не сущности.
- Unit-тесты со стабом Rules; README модуля обновляется.

### Не входит

- Реализация `workflow.Rules` — TASK-039 (сущность зависит только от интерфейса).
- Реализация `Commands`/`Queries`/`Exporter` поверх персистентности — v0.5 (PostgreSQL, ADR-004; формат ID — ADR-011).
- Сущность Epic — вне scope спецификации Task (Non-Goals).

## Критерии приёмки

- [x] `Create` контракта принимает `epicID`; команды `SetScope`/`SetAcceptanceCriteria` добавлены в контракт.
- [x] Сущность Task: все четыре Structural и четыре Behavioral инварианта спецификации в коде; переходы — только через `workflow.Rules`.
- [x] Blocked/Cancelled без причины — ошибка; правки scope/AC/Epic вне Backlog — ошибка.
- [x] Unit-тесты детерминированы (стаб Rules), покрывают успешные и запрещённые сценарии; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/domain/task/` (расширение контракта + новые файлы сущности); `internal/domain/task/README.md`, `internal/domain/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — спецификация Task утверждена (TASK-032); направление расширения контракта решено её Decision Log

## План реализации

1. `contracts.go` — `Create(ctx, projectID, epicID, title, taskType)` (breaking change допустим: реализаций контракта ещё нет); `SetScope(ctx, taskID, scope)`; `SetAcceptanceCriteria(ctx, taskID, criteria)`.
2. `events.go` — `Created`, `Transitioned` (From/To/Reason/At).
3. `task.go` — сущность + New/AttachToEpic/SetScope/SetAcceptanceCriteria/Transition(to, reason, rules); sentinel-ошибки (ErrMissingField, ErrNilRules, ErrNotBacklog, ErrReasonRequired).
4. `task_test.go` — стаб Rules (разрешает по таблице/запрещает всё), тесты инвариантов.
5. README модуля, `internal/domain/README.md`, `make verify`, PR, CI, merge.

## История

2026-07-21 — Architect — этап 2 EPIC-003: задача поставлена в очередь (четвёртая по порядку проектирования); направление расширения контракта уже решено Decision Log спецификации Task.
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен; выбор «epicID параметром Create» (а не отдельной командой привязки) подтверждён — проще и точно выражает опциональность связи при создании; правило «правки только в Backlog» согласуется с семантикой Ready→Backlog канонической state machine. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: `contracts.go` расширен (Create с epicID, SetScope, SetAcceptanceCriteria; Exporter сохранён без изменений), `events.go` (Created/Transitioned — отображение в 15 канонических имён событий оставлено публикатору), `task.go` (сущность + New/AttachToEpic/SetScope/SetAcceptanceCriteria/Transition с делегированием допустимости в workflow.Rules), `task_test.go` (11 тестов со стабами allowAll/denyAll, 81.8% покрытия — непокрытое: простые аксессоры). README модуля переписан, `internal/domain/README.md` синхронизирован. `make verify` — чисто.
2026-07-21 — Architect — Code Review: делегирование допустимости переходов в Rules — точно по инварианту 8 state-machine.md, сущность не дублирует таблицу переходов; ошибка Rules пробрасывается без обёртки — правильный выбор (источник правила виден вызывающему); ErrReasonRequired до вызова Rules — верный порядок (сущность проверяет то, что знает только она). Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-037 — расширение и реализация домен-модуля Task (EPIC-003, этап 2, четвёртая задача).
2. **Что сделано:** оба решённых Decision Log спецификации пробела контракта закрыты (`epicID` в Create; SetScope/SetAcceptanceCriteria); реализована сущность Task с инвариантами в коде — правки scope/AC/Epic только в Backlog, причина обязательна для Blocked/Cancelled, допустимость переходов решает исключительно переданный `workflow.Rules` (инвариант 8 state-machine.md).
3. **Изменённые файлы:** `internal/domain/task/{contracts,events,task,task_test}.go`, `internal/domain/task/README.md`; `internal/domain/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/domain/task/... -cover` — 11 тестов, 81.8%; `make verify` — чисто.
5. **Обновлённая документация:** README модуля task, README слоя domain.
6. **Open Questions:** формат идентификаторов — ADR-011 (Decision Required, зафиксирован в фазе B плана); приоритезация Backlog/Ready — future work спецификации.
7. **Рекомендации:** TASK-038 (project) — тот же паттерн; TASK-039 (workflow.Rules) сделает стабы тестов задач заменяемыми на реальную таблицу.
