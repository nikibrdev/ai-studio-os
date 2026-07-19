# TASK-005: Интерфейс Workflow

## Тип

feature

## Эпик

EPIC-002 ([docs/roadmap/EPIC-002-foundation.md](../../docs/roadmap/EPIC-002-foundation.md))

## Цель

Контракт применения правил процесса в `internal/core`: интерфейс Workflow (проверка допустимости перехода, роль следующего шага) и тип TaskState с девятью каноническими состояниями — по [interfaces.md](../../docs/architecture/interfaces.md) и [state-machine.md](../../docs/architecture/state-machine.md).

## Контекст

Правила переходов (таблица, guard-условия) — логика Domain Layer (следующий эпик); здесь фиксируется только контракт и словарь состояний. По [ADR-014](../../docs/adr/ADR-014-module-interaction.md) Workflow решает, но не действует.

## Scope

### Входит

- `internal/core/types.go` (дополнение): тип TaskState и константы девяти состояний.
- `internal/core/workflow.go`: интерфейс Workflow (CanTransition, NextRole); doc-комментарии (детерминированность, отсутствие побочных действий, Workflow → SQL запрещён).

### Не входит

- Реализация таблицы переходов; определения процессов (TASK-009).

## Критерии приёмки

- [ ] `go build ./...`, `go vet ./...` — без ошибок; форматирование gofumpt.
- [ ] Девять состояний соответствуют state-machine.md дословно.
- [ ] Только интерфейсы/типы/константы.

## Затрагиваемые модули и документы

- `internal/core/`

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-19 — Architect — задача определена в составе EPIC-002.
2026-07-19 — Claude Code (Developer) — задача оформлена в ready.
2026-07-19 — Claude Code (Developer) — выполнена, переведена в review.

## Отчёт о выполнении

1. **Задача:** TASK-005 — интерфейс Workflow.
2. **Что сделано:** в `internal/core/types.go` добавлен тип TaskState с девятью каноническими состояниями (дословно по state-machine.md); создан `internal/core/workflow.go`: интерфейс Workflow (CanTransition, NextRole) с ограничениями в doc-комментариях: детерминированность, «решает — не действует», запрет Workflow → SQL (ADR-014).
3. **Изменённые файлы:** `internal/core/types.go` (дополнен), `internal/core/workflow.go` (новый).
4. **Как проверялось:** `go build ./...` — OK; `go vet ./...` — OK; `gofumpt -l .` — чисто.
5. **Обновлённая документация:** не требуется.
6. **Open Questions:** нет.
7. **Рекомендации:** таблицу переходов в Domain Layer генерировать из одной структуры данных, чтобы соответствие state-machine.md проверялось тестом.
