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

- [ ] Задачу нельзя создать в Project вне состояния Active (Behavioral Invariant 4 спецификации Project — на уровне use-case).
- [ ] Переход Backlog → Ready — только через `workflow.Machine`; события `TaskCreated`/`TaskPlanned` публикуются с корректными полями конверта (Source=`task`, ProjectID, SubjectID).
- [ ] Покрытие пакета ≥ 85%; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/` (+ README при необходимости).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-040 (порты, конверт)

## План реализации

<Заполняется при взятии задачи в работу.>

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (вторая).
