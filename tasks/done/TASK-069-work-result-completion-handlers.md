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

- `POST /projects/{projectId}/tasks/{id}/start` → `WorkService.StartTask` (путь скорректирован BUGFIX-003).
- `POST /artifacts` (черновик) → `ResultService.RecordDraftArtifact`.
- `PATCH /artifacts/{id}` → `ResultService.UpdateArtifactDraft`.
- `POST /artifacts/{id}/publish` → `ResultService.PublishArtifact`.
- `POST /executions/{id}/succeed` → `ResultService.SucceedExecution` (`projectId` в теле — BUGFIX-003).
- `POST /executions/{id}/fail` → `ResultService.FailExecution` (`projectId` в теле — BUGFIX-003).
- `POST /projects/{projectId}/tasks/{id}/request-review` → `CompletionService.RequestReview`.
- `POST /projects/{projectId}/tasks/{id}/complete-review` → `CompletionService.CompleteReview`.
- `POST /projects/{projectId}/tasks/{id}/complete-testing` → `CompletionService.CompleteTesting`.
- Юнит-тесты хендлеров на `httptest` — через реальные `application.*Service` поверх `internal/application/inmemory` (тот же выбор, что TASK-068).

### Не входит

- Хендлеры Projects/Tasks-создание (TASK-068).
- Сквозной интеграционный тест на реальном PostgreSQL (TASK-070).

## Критерии приёмки

- [x] Все девять операций реализованы в точности по спецификациям TASK-066 — коды ошибок совпадают.
- [x] `complete-testing` через HTTP воспроизводит ADR-008 в полном объёме: отказ merge (через `RepositoryProvider`) оставляет задачу в Testing, а не переводит в Done — проверено тестом.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/api/httpapi/work.go`, `apps/api/httpapi/artifacts.go`, `apps/api/httpapi/completion.go` и соответствующие `_test.go`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-066, 067; независима от TASK-068 (может выполняться параллельно)

## План реализации

1. `apps/api/httpapi/work.go` — `registerWorkRoutes`, `handleStartTask` → `WorkService.StartTask`.
2. `apps/api/httpapi/artifacts.go` — `registerArtifactRoutes`, три хендлера → `ResultService.RecordDraftArtifact`/`UpdateArtifactDraft`/`PublishArtifact` (payload — `[]byte`, `encoding/json` уже кодирует/декодирует его как base64 — вручную ничего делать не нужно).
3. `apps/api/httpapi/executions.go` — `registerExecutionRoutes`, `handleSucceedExecution`/`handleFailExecution` → `ResultService.SucceedExecution`/`FailExecution`.
4. `apps/api/httpapi/completion.go` — `registerCompletionRoutes`, три хендлера → `CompletionService.RequestReview`/`CompleteReview`/`CompleteTesting`.
5. `apps/api/httpapi/server.go` — `NewServer` подключает все четыре новые группы маршрутов.
6. Тесты по каждому файлу (`work_test.go`, `artifacts_test.go`, `executions_test.go`, `completion_test.go`) — через реальные сервисы поверх `internal/application/inmemory`, включая `TestCompleteTesting_MergeFailureKeepsTaskInTesting` (реальный `inmemory.RepositoryProvider.MergeErr`).
7. **Обнаружено при реализации** (детали и полное решение — [BUGFIX-003](../done/BUGFIX-003-task-project-scoped-key.md), исправлено в этой же сессии до завершения этой задачи): живая HTTP-проверка вскрыла, что `TASK-NNN` неуникален глобально между проектами, а `ResultService.SucceedExecution`/`FailExecution` и `CompletionService`'s методы искали задачу только по голому ID. Эта задача реализована СРАЗУ с исправленными сигнатурами/маршрутами (`/projects/{projectId}/tasks/...`, `projectId` в теле для операций над Execution) — отдельной переделки не потребовалось.
8. `make verify`, затем `go test ./apps/api/... -cover`.

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

2026-07-22 — Developer — задача взята в работу; в процессе реализации обнаружен и исправлен блокирующий баг (BUGFIX-003, составной ключ Task); эта задача завершена с учётом исправления (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-069 — хендлеры Work/Result/Completion (вторая половина golden path через HTTP).

### Что сделано

- `apps/api/httpapi/work.go` — `POST /projects/{projectId}/tasks/{id}/start`.
- `apps/api/httpapi/artifacts.go` — `POST /artifacts`, `PATCH /artifacts/{id}`, `POST /artifacts/{id}/publish`.
- `apps/api/httpapi/executions.go` — `POST /executions/{id}/succeed`/`fail` (`projectId` в теле — BUGFIX-003).
- `apps/api/httpapi/completion.go` — `POST /projects/{projectId}/tasks/{id}/request-review`/`complete-review`/`complete-testing`.
- `apps/api/httpapi/server.go` — все девять операций подключены к `NewServer`.
- В процессе реализации (живая HTTP-проверка) обнаружен и исправлен реальный блокирующий баг — см. [BUGFIX-003](../done/BUGFIX-003-task-project-scoped-key.md): `TASK-NNN` уникален только в рамках Project, а не глобально; без исправления два разных проекта портили данные друг друга. Эта задача реализована уже с исправленными сигнатурами (`projectID` в `StartTaskParams`/`SucceedExecution`/`FailExecution`/`RequestReview`/`CompleteReview`/`CompleteTestingParams`) и маршрутами (`/projects/{projectId}/tasks/...`).
- Тесты по каждому файлу, включая `TestCompleteTesting_MergeFailureKeepsTaskInTesting` — ADR-008 воспроизведён через HTTP целиком: отказ merge (`inmemory.RepositoryProvider.MergeErr`) оставляет задачу в `testing`, не в `done`.

### Изменённые файлы

- `apps/api/httpapi/{work,artifacts,executions,completion}.go` (новые).
- `apps/api/httpapi/{work,artifacts,executions,completion}_test.go` (новые).
- `apps/api/httpapi/server.go` (дополнен маршрутами).

### Как проверялось

- `go test ./apps/api/... -v -cover` — все тесты зелёные, 84.8% покрытия `httpapi`.
- `make verify` — чисто.
- Живая проверка на реальном PostgreSQL — см. отчёт [BUGFIX-003](../done/BUGFIX-003-task-project-scoped-key.md): полный golden path (create → connect-repository → activate → create task → plan → start → record/update/publish artifact → succeed execution → request-review → complete-review → complete-testing) пройден через `curl` на настоящем `apps/api` + PostgreSQL, включая проверку независимости двух разных проектов.

### Обновлённая документация

Нет отдельных изменений сверх того, что уже задокументировано в BUGFIX-003 (docs/api/tasks.md, docs/api/executions.md).

### Open Questions

Нет.

### Рекомендации

TASK-070 (сквозной интеграционный тест) может опираться на уже проверенный вручную сценарий из BUGFIX-003 как на основу — golden path целиком уже пройден через реальный HTTP+PostgreSQL, TASK-070 формализует это как автоматический тест (`-tags=integration`).
