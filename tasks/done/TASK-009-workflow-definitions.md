# TASK-009: Контракты определений процесса (internal/workflow)

## Тип

feature

## Эпик

EPIC-002 ([docs/roadmap/EPIC-002-foundation.md](../../docs/roadmap/EPIC-002-foundation.md))

## Цель

Пакет `internal/workflow` с контрактами определений процесса: Definition (версионируемое определение workflow) и Step (шаг с ответственной ролью) — по [domain-model.md](../../docs/architecture/domain-model.md) (Workflow, Workflow Step).

## Контекст

Интерфейс применения правил (core.Workflow) создан в TASK-005; здесь — контракты структуры определений. Опубликованное определение неизменяемо; изменение — новой версией. Реализация и стандартный workflow MVP — Domain Layer.

## Scope

### Входит

- `internal/workflow/definition.go`: интерфейсы Definition, Step (импортируют `internal/core` для Role/TaskState).

### Не входит

- Таблица переходов state machine (Domain Layer); хранение определений.

## Критерии приёмки

- [ ] `go build ./...`, `go vet ./...` — без ошибок; форматирование gofumpt.
- [ ] Только интерфейсы; импорт — только `internal/core` (границы соблюдены).

## Затрагиваемые модули и документы

- `internal/workflow/`

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-19 — Architect — задача определена в составе EPIC-002.
2026-07-19 — Claude Code (Developer) — задача оформлена в ready.
2026-07-19 — Claude Code (Developer) — выполнена, переведена в review.
2026-07-19 — Architect — code review, замечание № 3 (структура): пакет перенесён `internal/workflow` → `internal/domain/workflow` ([запись ревью](../../engineering/reviews/2026-07-19-epic-002-code-review.md)).
2026-07-19 — Claude Code (Developer) — перенос выполнен; проверки verify пройдены.

## Отчёт о выполнении

1. **Задача:** TASK-009 — контракты определений процесса (internal/workflow).
2. **Что сделано:** создан пакет `internal/workflow` (`definition.go`): интерфейсы Definition (Name, Version, Steps; версионирование, неизменяемость опубликованной версии) и Step (Name, Role, EntryState; одна ответственная роль). Импорт — только `ai-studio-os/internal/core` (границы соблюдены).
3. **Изменённые файлы:** `internal/workflow/definition.go` (новый).
4. **Как проверялось:** `go build ./...` — OK; `go vet ./...` — OK; `gofumpt -l .` — чисто.
5. **Обновлённая документация:** не требуется.
6. **Open Questions:** нет.
7. **Рекомендации:** стандартный workflow MVP описать декларативно в Domain Layer и валидировать против state-machine.md тестом.
