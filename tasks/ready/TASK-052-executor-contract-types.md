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

- [ ] `ExecutorTask`/`Artifact`/`ExecutionStatus` — конкретные структуры с примитивными полями, не `any`, не псевдонимы доменных типов.
- [ ] `internal/platform` не импортирует `internal/domain` (проверяется `go vet`/сборкой — импорт домена в platform привёл бы к ошибке ADR-015 на уровне архитектурного review, не компилятора; явно проверить и зафиксировать в отчёте).
- [ ] Контракт `Executor` (четыре метода) не изменён.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/platform/executor.go`, README `internal/platform` (если описывает эти типы).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — EPIC-005 закрыт; ADR-005/ADR-015 приняты

## План реализации

## История

2026-07-21 — Architect — EPIC-006 открыт; задача поставлена в очередь (первая).

## Отчёт о выполнении
