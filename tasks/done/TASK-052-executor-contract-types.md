# TASK-052: Конкретизация типов Executor Contract

## Тип

feature

## Эпик

[EPIC-006 AI Agent Runtime](../../docs/roadmap/EPIC-006-ai-agent-runtime.md)

## Цель

Конкретизировать `platform.ExecutorTask`, `platform.Artifact`, `platform.ExecutionStatus` — сейчас это псевдонимы `any` (ADR-005: «форма — задача Domain Layer, когда появятся реальные Task/Artifact/Execution»). Domain Layer (EPIC-003) готов — пора дать этим типам реальную форму, не меняя контракт `Executor` (Accept/Artifacts/Status/Finish) и не нарушая домен-агностичность `internal/platform` (ADR-015).

## Контекст

`internal/platform` не может импортировать `internal/domain` (ADR-015). Решение: конкретные структуры с примитивными полями (строки, срезы, `time.Time`), не псевдонимы доменных типов. Вызывающая сторона (`internal/application`, уже импортирует домен) собирает эти структуры из геттеров `task.Task`/`artifact.Artifact`/`execution.Execution`; адаптер (`agents/`, вне `internal/`) получает и возвращает только эти плоские структуры, не доменные типы.

## Scope

### Входит

- `internal/platform/executor.go`: `ExecutorTask` — структура (минимум: TaskID, ProjectID, Title, Type, Scope, AcceptanceCriteria, Role — по факту того, что реально нужно исполнителю для начала работы); `Artifact` — структура (минимум: ID, Type, Origin, Author, Payload); `ExecutionStatus` — структура или enum-подобный тип (минимум: State, сообщение/детали).
- Компиляционная проверка, что контракт `Executor` не меняется (Accept/Artifacts/Status/Finish — те же четыре метода).
- Юнит-тесты на сами структуры, если есть нетривиальная логика (конструкторы, валидация) — если это просто структуры данных, тесты не требуются искусственно.

### Не входит

- Сам адаптер `agents/claude-code` (TASK-055).
- Изменение доменных пакетов — только чтение их геттеров на стороне `internal/application` при конструировании этих структур (не в этой задаче — конструирование появится там, где типы реально используются, начиная с TASK-055).

## Критерии приёмки

- [x] `ExecutorTask`/`Artifact`/`ExecutionStatus` — конкретные структуры с примитивными полями, не `any`, не псевдонимы доменных типов.
- [x] `internal/platform` не импортирует `internal/domain` — `executor.go` импортирует только `context` (проверено чтением файла и `go build`/`go vet`, которые ловят реальный нарушающий импорт как ошибку компиляции — здесь его физически нет).
- [x] Контракт `Executor` (четыре метода) не изменён.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/platform/executor.go`, README `internal/platform` (если описывает эти типы).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — EPIC-005 закрыт; ADR-005/ADR-015 приняты

## План реализации

1. Проверить, что ничего в кодовой базе не полагается на `ExecutorTask`/`Artifact`/`ExecutionStatus` как на `any` (grep) — подтверждено: только doc-комментарий в `internal/domain/executor/doc.go` упоминает контракт текстом, кода нет.
2. `ExecutorTask` — структура: `TaskID`, `ProjectID`, `Role`, `Title`, `Type`, `Scope`, `AcceptanceCriteria []string` (планирование, из `task.Task`), плюс `Repository`, `Branch` — по ADR-006 Executor клонирует рабочую копию ветки задачи, значит эти координаты обязаны быть в задании; в исходном scope задачи не были прямо перечислены, но необходимы буквально по ADR-006 — добавлены с обоснованием здесь, а не домыслены произвольно.
3. `Artifact` — структура: `ID`, `Type`, `Origin`, `Author`, `Payload []byte` (тот же словарь, что `domain/artifact.Artifact`, без импорта пакета).
4. `ExecutionStatus` — структура: `State string` (собственная оценка Executor'а, не авторитетное состояние — это явно прокомментировано: авторитетное состояние `Execution` выставляется тем application-сервисом, который вызывает `Finish`, не выводится автоматически из `State`), `Message string`.
5. Тесты не добавлены — это плоские структуры данных без логики (конструкторов/валидации нет), задача явно не требует их для проформы.
6. README `internal/platform` — обновлена строка про `Executor`, дата обновления.

## История

2026-07-21 — Architect — EPIC-006 открыт; задача поставлена в очередь (первая).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан, включая добавление `Repository`/`Branch` в `ExecutorTask` сверх изначально перечисленного в scope минимума (обосновано ADR-006).
2026-07-21 — Architect — план одобрен; `Repository`/`Branch` — обоснованное дополнение, не домысел (ADR-006 прямо требует клонирование ветки); `ExecutionStatus.State` как «собственная неавторитетная оценка Executor'а» — важное разграничение, которое стоит сохранить как явный комментарий в коде (уже сделано). Приступать.
2026-07-21 — Claude Code (Developer) — реализовано по плану. `go build`/`make verify` подтверждают отсутствие импорта `internal/domain` в `internal/platform` и отсутствие регрессий во всей кодовой базе.
2026-07-21 — Architect — Code Review: структуры точно отражают то, что реально нужно для ADR-006 (клонирование ветки) и словарь `domain/artifact`, не более и не менее; разграничение `ExecutionStatus.State` (неавторитетно) от реального состояния `Execution` — правильное архитектурное уточнение, предотвращающее будущую путаницу, когда адаптер появится (TASK-055). Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-052 — конкретизация типов `platform.Executor` (первая задача EPIC-006).
2. **Что сделано:** `ExecutorTask`/`Artifact`/`ExecutionStatus` в `internal/platform/executor.go` заменены с `any` на конкретные структуры с примитивными полями; контракт `Executor` (Accept/Artifacts/Status/Finish) не изменён; домен-агностичность `internal/platform` подтверждена (единственный импорт — `context`).
3. **Изменённые файлы:** `internal/platform/executor.go`, `internal/platform/README.md`; файл задачи.
4. **Как проверялось:** `go build ./...`, `make verify` — чисто; grep по всей кодовой базе подтвердил отсутствие кода, полагающегося на прежний `any`.
5. **Обновлённая документация:** README `internal/platform`.
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-053/054 могут проектировать образ и жизненный цикл контейнера уже относительно конкретной формы `ExecutorTask` (в частности, полей `Repository`/`Branch`).
