# TASK-010: Контракты Task Engine (internal/tasks)

## Тип

feature

## Эпик

EPIC-002 ([docs/roadmap/EPIC-002-foundation.md](../../docs/roadmap/EPIC-002-foundation.md))

## Цель

Пакет `internal/tasks` с контрактами Task Engine по [ADR-004](../../docs/adr/ADR-004-task-storage.md): Engine (единственная точка переходов состояния), Reader (чтение для приложений), Exporter (markdown-экспорт в `tasks/`).

## Контекст

ADR-004 (принят): PostgreSQL — источник истины; Task Engine валидирует переходы по [state-machine.md](../../docs/architecture/state-machine.md); `tasks/` — экспорт. Реализация — Domain Layer и далее; здесь только контракты.

## Scope

### Входит

- `internal/tasks/engine.go`: интерфейсы Engine, Reader, Exporter (импортируют `internal/core` для TaskState).

### Не входит

- Реализация, схема БД, импорт существующих файлов задач, формат экспорта.

## Критерии приёмки

- [ ] `go build ./...`, `go vet ./...` — без ошибок; форматирование gofumpt.
- [ ] Только интерфейсы; переходы выражены через core.TaskState.
- [ ] Doc-комментарии фиксируют: Engine — единственный писатель состояния (ADR-004).

## Затрагиваемые модули и документы

- `internal/tasks/`

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-19 — Architect — задача определена в составе EPIC-002.
2026-07-19 — Claude Code (Developer) — задача оформлена в ready.
2026-07-19 — Claude Code (Developer) — выполнена, переведена в review.
2026-07-19 — Architect — code review, замечания № 2 и № 3: имя «Engine» преждевременно (устройство пути записи не решено, возможен Command → Event → Projection); пакет создан раньше доменного слоя ([запись ревью](../../engineering/reviews/2026-07-19-epic-002-code-review.md)).
2026-07-19 — Claude Code (Developer) — исправлено: пакет перенесён `internal/tasks` → `internal/domain/task`; Engine → Commands, Reader → Queries; в doc пакета зафиксировано, что механизм записи контрактами не фиксируется. Проверки verify пройдены. Повторное ревью.

## Отчёт о выполнении

1. **Задача:** TASK-010 — контракты Task Engine (internal/tasks).
2. **Что сделано:** создан пакет `internal/tasks` (`engine.go`): интерфейсы Engine (Create, Transition — единственная точка изменения состояния по ADR-004, с обязательной причиной для Blocked/Cancelled), Reader (State — read-контракт для слоя доставки, с оговоркой ADR-014) и Exporter (Export, ExportAll — markdown-экспорт в `tasks/`; экспорт — представление, не источник истины).
3. **Изменённые файлы:** `internal/tasks/engine.go` (новый).
4. **Как проверялось:** `go build ./...` — OK; `go vet ./...` — OK; `gofumpt -l .` — чисто.
5. **Обновлённая документация:** не требуется (ADR-004 уже описывает целевую модель).
6. **Open Questions:** нет (формат идентификаторов — ADR-011; формат экспорта — задача Domain/Infrastructure).
7. **Рекомендации:** при реализации Engine выполнить разовый импорт существующих файлов `tasks/` (переходный период ADR-004) отдельной задачей.
