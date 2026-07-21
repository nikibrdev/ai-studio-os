# TASK-041: Use-case «Постановка задачи»

## Тип

feature

## Эпик

[EPIC-004 Application Layer](../../docs/roadmap/EPIC-004-application-layer.md)

## Цель

Первый use-case золотого пути: создание задачи в границе активного Project (проверка `AcceptsNewContent`), запись scope/критериев приёмки, перевод Backlog → Ready через `workflow.Machine`; публикация `TaskCreated`/`TaskPlanned` через `EventBus`.

## Контекст

Golden path, шаг «Пользователь создаёт задачу → PM доводит до готовности». Доменная логика готова (task/project/workflow); эта задача — оркестрация поверх портов TASK-040.

## Scope

### Входит

- `internal/application/…`: сервис/функции CreateTask (создание + scope/AC) и PlanTask (Backlog → Ready); загрузка Project из ProjectStore, проверка границы, сохранение Task, публикация событий.
- Тесты: успешные пути; Project не Active; невалидные поля; отказ Rules; идемпотентность не требуется (один вызов — одна команда).

### Не входит

- Назначение исполнителя, запуск работы (TASK-042); HTTP-доставка.

## Критерии приёмки

- [x] Задачу нельзя создать в Project вне состояния Active (Behavioral Invariant 4 спецификации Project — на уровне use-case).
- [x] Переход Backlog → Ready — только через `workflow.Machine`; события `TaskCreated`/`TaskPlanned` публикуются с корректными полями конверта (Source=`task`, ProjectID, SubjectID).
- [x] Покрытие пакета ≥ 85%; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/` (+ README при необходимости).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-040 (порты, конверт)

## План реализации

1. Дополнение `internal/application/inmemory`: `EventBus` — детерминированный синхронный фейк `platform.EventBus` (доставка по порядку регистрации, журнал опубликованных событий для проверок) — пробел TASK-040, обнаруженный при написании тестов этой задачи.
2. `internal/application/task_planning.go` — `TaskPlanningService{Projects, Tasks, Events, Rules}`, методы `CreateTask` (guard `AcceptsNewContent`, New + SetScope/SetAcceptanceCriteria, Save, публикация TaskCreated) и `PlanTask` (Get, Transition(Ready) через Rules, Save, публикация TaskPlanned); `ErrProjectNotActive`.
3. Тесты: happy path (создание + план), Project не Active, Project не найден, отказ Rules при PlanTask, содержимое опубликованных событий (тип/источник/ProjectID/SubjectID).
4. `make verify`, PR, CI, merge.

## Открытый вопрос по ходу работы

TASK-040 не включила фейк EventBus, хотя её Scope говорил «in-memory фейки для тестов эпика» в общем смысле — упущение обнаружено здесь, закрывается в этой задаче без возврата к TASK-040 (задача уже done, точечное дополнение — не искажает её отчёт).

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (вторая).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан; обнаружен пробел TASK-040 (нет фейка EventBus) — включён в план этой задачи.
2026-07-21 — Architect — план одобрен, включая точечное дополнение TASK-040 без пересмотра её отчёта. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: `inmemory/eventbus.go` (детерминированный синхронный EventBus с журналом и отменяемыми подписками — закрывает пробел TASK-040) + тесты; `task_planning.go` (TaskPlanningService.CreateTask/PlanTask, ErrProjectNotActive); `task_planning_test.go` (8 тестов: happy path создания и плана, отказ по неактивному/отсутствующему проекту, пробрасывание доменной ошибки валидации, отказ Rules на повторном PlanTask с проверкой отсутствия лишних событий, отсутствующая задача). Покрытие пакета application поднято точечным тестом до 85.7% (порог достигнут). README дополнен. `make verify` — чисто.
2026-07-21 — Architect — Code Review: use-case не решает допустимость перехода сам — делегирует в Rules (соответствует инварианту 8); guard `AcceptsNewContent` проверяется до создания сущности, не после (нет лишней работы на отказном пути); тест на отсутствие «лишних» событий после отказа Rules — важная проверка, что публикация не происходит раньше валидации. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-041 — use-case «Постановка задачи» (EPIC-004, вторая задача).
2. **Что сделано:** `TaskPlanningService` — CreateTask (в границе Active-проекта, с scope/AC, публикация TaskCreated) и PlanTask (Backlog → Ready через `workflow.Rules`, публикация TaskPlanned). Попутно закрыт пробел TASK-040 — добавлен фейк `EventBus`.
3. **Изменённые файлы:** `internal/application/task_planning.go`, `internal/application/task_planning_test.go`, `internal/application/inmemory/{eventbus,eventbus_test}.go`, `internal/application/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/application/... -cover` — 85.7%/92.9%; `make verify` — чисто.
5. **Обновлённая документация:** README `internal/application`.
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-042 (запуск работы) переиспользует `TaskPlanningService`-подобный паттерн (сервис с портами + Rules + Events) и уже готовый `EventBus`-фейк.
