# TASK-069: Хендлеры Work/Result/Completion

## Тип

feature

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Реализовать оставшуюся часть golden path через HTTP: запуск работы, черновик/публикация артефакта, результат исполнения, ревью и тестирование.

## Контекст

По спецификациям TASK-066 (`docs/api/tasks.md`, `docs/api/artifacts.md`, `docs/api/executions.md`), поверх каркаса TASK-067, вызывая уже существующие `WorkService`/`ResultService`/`CompletionService` (EPIC-004) — без изменений в них.

## Scope

### Входит

- `POST /tasks/{id}/start` → `WorkService.StartTask`.
- `POST /artifacts` (черновик) → `ResultService.RecordDraftArtifact`.
- `PATCH /artifacts/{id}` → `ResultService.UpdateArtifactDraft`.
- `POST /artifacts/{id}/publish` → `ResultService.PublishArtifact`.
- `POST /executions/{id}/succeed` → `ResultService.SucceedExecution`.
- `POST /executions/{id}/fail` → `ResultService.FailExecution`.
- `POST /tasks/{id}/request-review` → `CompletionService.RequestReview`.
- `POST /tasks/{id}/complete-review` → `CompletionService.CompleteReview`.
- `POST /tasks/{id}/complete-testing` → `CompletionService.CompleteTesting`.
- Юнит-тесты хендлеров на `httptest` с фейковыми сервисами.

### Не входит

- Хендлеры Projects/Tasks-создание (TASK-068).
- Сквозной интеграционный тест на реальном PostgreSQL (TASK-070).

## Критерии приёмки

- [ ] Все девять операций реализованы в точности по спецификациям TASK-066 — коды ошибок совпадают.
- [ ] `complete-testing` через HTTP воспроизводит ADR-008 в полном объёме: отказ merge (через `RepositoryProvider`) оставляет задачу в Testing, а не переводит в Done — проверено тестом.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/api/httpapi/work.go`, `apps/api/httpapi/artifacts.go`, `apps/api/httpapi/completion.go` и соответствующие `_test.go`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-066, 067; независима от TASK-068 (может выполняться параллельно)

## План реализации

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

## Отчёт о выполнении
