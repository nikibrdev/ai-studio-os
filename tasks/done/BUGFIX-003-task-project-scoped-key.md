# BUGFIX-003: TASK-NNN уникален только в рамках Project — составной ключ (project_id, id)

## Тип

fix

## Эпик

Вне эпика — срочное исправление; обнаружено вживую при разработке TASK-069 (EPIC-008), затрагивает уже закрытые EPIC-004/EPIC-005.

## Цель

Два разных проекта, каждый создающий свою первую задачу, легитимно получают одинаковый `TASK-001` (ADR-011: номер последовательный **в рамках Project**, TASK-065) — и до этого исправления второй проект молча портил данные первого через `ON CONFLICT (id) DO UPDATE`, поскольку `tasks.id` было единственным `PRIMARY KEY`. Задача — сделать так, чтобы любые два проекта могли независимо иметь свои TASK-001, TASK-002 и т.д. без взаимного повреждения данных, по всему стеку от PostgreSQL до HTTP API.

## Контекст

Обнаружено живой проверкой TASK-069 (см. `tasks/done/TASK-069-*`): создал проект `proj-e2e`, его первая задача получила `TASK-001` — но `GET /tasks/TASK-001` вернул данные проекта `proj-live` из предыдущей ручной проверки (TASK-068). Причина: `tasks.id TEXT PRIMARY KEY` (миграция `0002`, TASK-047) — глобальный, не составной; `Save`'s `ON CONFLICT (id) DO UPDATE` при столкновении обновлял `epic_id`/`scope`/`state`, но не `project_id`/`title`/`type` — оставляя гибридную, испорченную запись.

ADR-011 уже явно предвидел эту ситуацию: «Короткая форма ID неуникальна между проектами — любой межпроектный контекст обязан использовать полностью квалифицированную пару (Project, ID); это требование к будущим Dashboard/API». Это не новое архитектурное решение, а исправление пробела в реализации EPIC-005 (TASK-047) относительно уже принятого решения.

Полный корректный фикс оказался больше, чем начальная гипотеза «просто составной PRIMARY KEY»: `application.TaskStore.Get` — единственная точка входа для чтения задачи — требует `projectID`, что каскадом задевает шесть методов Application Layer (`TaskPlanningService.PlanTask`, `WorkService.StartTask`, `CompletionService.RequestReview`/`CompleteReview`/`CompleteTesting`, `ResultService.SucceedExecution`/`FailExecution`) и внутреннюю карту `TaskProjection` (тот же голый-ID-ключ, та же болезнь) — и URL-схему `apps/api` (`/tasks/{id}` → `/projects/{projectId}/tasks/{id}`), включая уже слитые TASK-068 хендлеры.

## Scope

### Входит

- `internal/infrastructure/postgres/migrations/0006_task_project_scoped_key.sql`: `ALTER TABLE tasks ADD PRIMARY KEY (project_id, id)`; `ALTER TABLE executions DROP CONSTRAINT executions_task_id_fkey` (task_id становится неформальной ссылкой — тот же принцип, что уже был у `artifacts.produced_by`, ADR-016).
- `internal/infrastructure/postgres/task_store.go`: `Get(ctx, projectID, id)`, `Save` — `ON CONFLICT (project_id, id)`.
- `internal/application/ports.go`: `TaskStore.Get(ctx, projectID, id)`.
- `internal/application/inmemory`: выделенный `TaskStore` (не generic `Store[T]`), ключ — пара (ProjectID, ID).
- `internal/application/{task_planning,work,completion,result,projection}.go`: `projectID` — параметр/поле там, где раньше был голый `taskID`; `ResultService.SucceedExecution`/`FailExecution` принимают `projectID` явно (Execution не хранит ссылку на Project — домен не менялся, ADR-015) вместо попытки вывести его из `Tasks.Get(ctx, run.TaskID())`.
- `apps/api/httpapi`: `POST /tasks` → `POST /projects/{projectId}/tasks`, аналогично `plan`/`start`/`request-review`/`complete-review`/`complete-testing`/`GET`; `/executions/{id}/succeed`/`fail` не вложены (Execution ID — глобально уникальный `crypto/rand`), но принимают `projectId` в теле.
- Обновлены все существующие тесты (`internal/application/*_test.go`, `internal/infrastructure/wiring/golden_path_integration_test.go`, `apps/api/httpapi/*_test.go`) под новые сигнатуры/маршруты.
- Новые тесты, доказывающие исправление: `TestTaskStore_SameIDDifferentProjectsDoNotCollide` (postgres, реальный), `TestTaskStore_SameIDDifferentProjectsDoNotCollide` (inmemory), `TestCreateTask_SameIDDifferentProjectsDoNotCollide` (HTTP).
- `docs/api/tasks.md`/`docs/api/executions.md`, `internal/application/README.md`, `internal/infrastructure/README.md` — синхронизированы.

### Не входит

- Аналогичная проверка для `EPIC-NNN` — эпики создаются архитектором вручную, не через API (см. TASK-065).
- Изменение домена `internal/domain/execution` (не добавлен `ProjectID` — рассмотрено и отклонено как более инвазивный путь, требующий формального Delta Review уже утверждённой спецификации Execution; вместо этого `projectID` передаётся явным параметром вызывающим кодом, который его уже знает).

## Критерии приёмки

- [x] Два проекта, каждый создающий свою первую задачу, независимо получают `TASK-001` без повреждения данных друг друга — подтверждено на реальном PostgreSQL, на in-memory фейке и через HTTP.
- [x] `make verify` — чисто.
- [x] Живой прогон: `docker compose up -d postgres` (чистый том), полный сценарий через `apps/api` — два проекта, оба с `TASK-001`, независимая эволюция состояния каждого.
- [x] Все существующие golden-path тесты (application/wiring/httpapi) по-прежнему проходят с новыми сигнатурами.

## Затрагиваемые модули и документы

- `internal/infrastructure/postgres/{migrations/0006_task_project_scoped_key.sql,task_store.go,project_task_store_integration_test.go}`.
- `internal/application/{ports.go,task_planning.go,work.go,completion.go,result.go,projection.go}` + все `_test.go` в пакете + `inmemory/{stores.go,task_store.go,stores_test.go}`.
- `internal/infrastructure/wiring/golden_path_integration_test.go`.
- `apps/api/httpapi/{tasks.go,work.go,completion.go,executions.go,server.go}` + все `_test.go`.
- `docs/api/tasks.md`, `docs/api/executions.md`, `internal/application/README.md`, `internal/infrastructure/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## План реализации

1. Миграция `0006`: снять FK `executions.task_id`, сменить PK `tasks` на составной.
2. `TaskStore.Get`/`Save` — обновить под составной ключ; `project_task_store_integration_test.go` — обновить существующие вызовы, добавить `TestTaskStore_SameIDDifferentProjectsDoNotCollide`.
3. `application.TaskStore` — сменить сигнатуру порта; `inmemory.TaskStore` — выделенный тип с составным ключом взамен generic `Store[T]`.
4. Каскадно обновить `PlanTask`/`StartTaskParams`/`RequestReview`/`CompleteReview`/`CompleteTestingParams` — добавить `projectID`; `ResultService.SucceedExecution`/`FailExecution` — `projectID` явным параметром, `publishExecutionEvent` использует его напрямую вместо `Tasks.Get(ctx, run.TaskID())`.
5. `TaskProjection` — составной ключ карты `(ProjectID, ID)`, `Get(projectID, id)`.
6. Обновить все вызовы во всех тестах `internal/application` (`task_planning_test.go`, `work_test.go`, `result_test.go`, `completion_test.go`, `projection_test.go`, `e2e_test.go`) и `internal/infrastructure/wiring/golden_path_integration_test.go` — везде уже используемый `"proj-1"`.
7. `apps/api/httpapi`: маршруты `/tasks/...` → `/projects/{projectId}/tasks/...`; `executions.go` — `projectId` в теле запроса (`executionActionRequest`); обновить все хендлеры и тесты (`tasks.go`, `work.go`, `completion.go`, `executions.go`, `server.go`, `deps_test.go` и все `_test.go`); добавить `TestCreateTask_SameIDDifferentProjectsDoNotCollide`. Попутно исправлен фейк `sequentialTaskIDGenerator` в тестах — считал глобально, а не на проект (сам был подвержен той же категории бага).
8. `docs/api/tasks.md`/`docs/api/executions.md` — синхронизировать пути/тела запросов; `internal/application/README.md`/`internal/infrastructure/README.md` — задокументировать исправление.
9. `make verify`, затем живой прогон: `docker compose down -v` (чистый том, чтобы не путать со старыми ручными проверками), `docker compose up -d postgres`, реальный `go run ./apps/api`, два проекта через `curl`, оба получают `TASK-001`, независимая эволюция состояния — воспроизвести и подтвердить закрытие исходного бага.

## История

2026-07-22 — Developer — обнаружено живой HTTP-проверкой TASK-069; после уточнения полного объёма исправления с пользователем (затрагивает уже слитый TASK-068, требует правки шести сигнатур Application Layer и URL-схемы API) — решено исправить полностью в рамках этой задачи, не откладывая.

## Отчёт о выполнении

### Задача

BUGFIX-003 — составной ключ (project_id, id) для Task по всему стеку: PostgreSQL → Application Layer → apps/api.

### Что сделано

- **PostgreSQL**: `tasks` — `PRIMARY KEY (project_id, id)` вместо `PRIMARY KEY (id)`; `executions.task_id` — FK снят, неформальная ссылка (тот же принцип, что `artifacts.produced_by`, ADR-016); `TaskStore.Get`/`Save` — под составной ключ.
- **Application Layer**: порт `TaskStore.Get(ctx, projectID, id)`; `TaskPlanningService.PlanTask`, `CompletionService.RequestReview`/`CompleteReview`, `WorkService.StartTaskParams`, `CompletionService.CompleteTestingParams` — все принимают/несут `projectID`; `ResultService.SucceedExecution`/`FailExecution` принимают `projectID` явно — `Execution` не менялся (домен, ADR-015; изменение утверждённой спецификации потребовало бы формального Delta Review — рассмотрено и отклонено как несоразмерное); `TaskProjection` — карта ключится парой (ProjectID, ID), `Get(projectID, id)`.
- **apps/api**: задаче-специфичные маршруты вложены под `/projects/{projectId}/tasks/...` (ADR-011 уже предвидел это требование дословно); `/executions/{id}/succeed`/`fail` остались невложенными (Execution ID глобально уникален), но принимают `projectId` в теле.
- Все существующие тесты (`internal/application` — 6 файлов, `internal/infrastructure/wiring`, `apps/api/httpapi` — 6 файлов) обновлены под новые сигнатуры/маршруты; попутно исправлен собственный тестовый фейк `sequentialTaskIDGenerator` (считал глобально вместо на проект — та же категория ошибки, обнаружена новым тестом на пересечение проектов).
- Три новых теста, прямо доказывающих исправление: на реальном PostgreSQL, на in-memory фейке, через HTTP end-to-end — во всех трёх два проекта получают независимый `TASK-001` без повреждения данных.

### Изменённые файлы

- `internal/infrastructure/postgres/migrations/0006_task_project_scoped_key.sql` (новый).
- `internal/infrastructure/postgres/task_store.go`, `project_task_store_integration_test.go`.
- `internal/application/ports.go`, `task_planning.go`, `work.go`, `completion.go`, `result.go`, `projection.go`.
- `internal/application/inmemory/task_store.go` (новый), `stores.go`, `stores_test.go`.
- `internal/application/{task_planning,work,completion,result,projection,e2e}_test.go`.
- `internal/infrastructure/wiring/golden_path_integration_test.go`.
- `apps/api/httpapi/{tasks,work,completion,executions,server}.go` и все соответствующие `_test.go` + `deps_test.go`.
- `docs/api/tasks.md`, `docs/api/executions.md`, `internal/application/README.md`, `internal/infrastructure/README.md`.

### Как проверялось

- `go build ./...`, `go vet ./...`, `go build -tags=integration ./...`, `go vet -tags=integration ./...` — чисто на каждом шаге исправления.
- `go test ./...` — все пакеты зелёные.
- `make verify` — чисто.
- Живая проверка на реальном PostgreSQL (три прогона подряд, `-count=1`): `TestTaskStore_SaveThenGet`, `TestTaskStore_Get_NotFound`, `TestTaskStore_SameIDDifferentProjectsDoNotCollide`, `TestTaskStore_NextID_*` — все зелёные.
- Живая проверка `apps/api` целиком (чистый том `docker compose down -v` → `up -d postgres`, реальный `go run ./apps/api`): создан `proj-live` и `proj-e2e`, оба получили `TASK-001` независимо; после `PlanTask` на `proj-e2e`'s TASK-001 (→ `ready`) `proj-live`'s TASK-001 остался `backlog` — исходный баг воспроизведён и подтверждено его устранение на том же сценарии, что его впервые вскрыл.

### Обновлённая документация

- `docs/api/tasks.md`, `docs/api/executions.md`, `internal/application/README.md`, `internal/infrastructure/README.md`.

### Open Questions

Нет.

### Рекомендации

TASK-069 (хендлеры Work/Result/Completion) продолжается с уже исправленными сигнатурами/маршрутами — сама реализация TASK-069 (сделанная параллельно с обнаружением этого бага) уже учитывает исправление, отдельной доработки не требуется.
