# TASK-003: Интерфейс Agent

## Тип

feature

## Эпик

EPIC-002 ([docs/roadmap/EPIC-002-foundation.md](../../docs/roadmap/EPIC-002-foundation.md))

## Цель

Контракт агента в `internal/core`: Agent, Assignment, Result — минимальный состав по [interfaces.md](../../docs/architecture/interfaces.md), не предвосхищающий нерешённые [ADR-005](../../docs/adr/ADR-005-agent-adapter-contract.md)/[ADR-006](../../docs/adr/ADR-006-agent-execution-environment.md).

## Контекст

Роль отделена от исполнителя ([agents.md](../../docs/architecture/agents.md)); формат обмена и среда выполнения — Decision Required, поэтому контракт фиксирует только независимый от этих решений минимум. Требуются общие типы Role (словарь ролей) — вводятся этой же задачей в `internal/core`.

## Scope

### Входит

- `internal/core/types.go`: тип Role и константы ролей (PM, Developer, QA, Reviewer, Architect).
- `internal/core/agent.go`: интерфейсы Agent, Assignment, Result; тип ExecutionStatus с константами.

### Не входит

- Формат обмена (ADR-005), среда запуска (ADR-006), реализация адаптеров.

## Критерии приёмки

- [ ] `go build ./...`, `go vet ./...` — без ошибок; форматирование gofumpt.
- [ ] Только интерфейсы/типы/константы.
- [ ] В doc-комментариях отмечено, что детализация — ADR-005/006.

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
2026-07-19 — Architect — code review, замечание № 1: Assignment/Result предвосхищают ADR-005; требование — Execute(ctx, Request) (Response, error) с абстрактными типами ([запись ревью](../../engineering/reviews/2026-07-19-epic-002-code-review.md)).
2026-07-19 — Claude Code (Developer) — исправлено: Request/Response — абстрактные типы (`any`) до ADR-005; Assignment, Result, ExecutionStatus, ID/Provider/Roles удалены. Проверки verify пройдены. Повторное ревью.

## Отчёт о выполнении

1. **Задача:** TASK-003 — интерфейс Agent.
2. **Что сделано:** создан `internal/core/types.go` со словарём Role (5 ролей MVP); создан `internal/core/agent.go`: интерфейсы Agent (ID, Provider, Roles, Execute), Assignment (минимум: TaskID, Role), Result (Status, Report, OpenQuestions, ArtifactIDs), тип ExecutionStatus (succeeded/failed/aborted). Запрет Agent → Database и минимальность контракта до ADR-005/006 зафиксированы в doc-комментариях.
3. **Изменённые файлы:** `internal/core/types.go`, `internal/core/agent.go` (новые).
4. **Как проверялось:** `go build ./...` — OK; `go vet ./...` — OK; `gofumpt -l .` — чисто.
5. **Обновлённая документация:** не требуется.
6. **Open Questions:** нет (расширение Assignment/streaming хода работы — по ADR-005).
7. **Рекомендации:** нет. Примечание процесса: файл `types.go` также содержит TaskState — формально scope TASK-005; типы были записаны одним файлом, оба входят в замороженный словарь core.
