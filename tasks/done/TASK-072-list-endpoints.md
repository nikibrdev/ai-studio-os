# TASK-072: Списковые операции — GET /projects, GET /projects/{id}/tasks

## Тип

feature

## Эпик

[EPIC-009 Dashboard](../../docs/roadmap/EPIC-009-dashboard.md)

## Цель

Добавить единственный недостающий класс операций в API — списки: сейчас `ProjectStore.Get`/`TaskProjection.Get` отдают ровно одну запись по точному идентификатору, и Dashboard не может показать даже список проектов.

## Контекст

Раздел «Контекст» [EPIC-009](../../docs/roadmap/EPIC-009-dashboard.md): пробел уже был отмечен как «самый заметный практический предел» в рисках EPIC-008. Project читается напрямую из `ProjectStore` (не через проекцию) уже сейчас (`TaskPlanningService.CreateTask` вызывает `s.Projects.Get` напрямую) — список проектов продолжает этот же паттерн (`ProjectStore.List`), не требует новой проекции. Task читается только через `TaskProjection` (ADR-014) — список задач проекта продолжает этот же паттерн (`TaskProjection.ListByProject`), не через `TaskStore` напрямую.

## Scope

### Входит

- `application.ProjectStore.List(ctx) ([]*project.Project, error)` — новый метод порта; `postgres.ProjectStore.List` (реальная реализация, `ORDER BY id` для детерминированности); generic `inmemory.Store[T].List` (используется фейком Project/Executor/Execution/Artifact).
- `application.ProjectService.ListProjects(ctx) ([]*project.Project, error)` — тонкая обёртка над `ProjectStore.List`.
- `application.TaskProjection.ListByProject(projectID string) []TaskView` — внутренняя карта `TaskProjection` перестраивается на `map[string]map[string]TaskView` (projectID → id → view) для тривиальной и эффективной выборки по проекту, взамен плоской карты с составным строковым ключом.
- `apps/api/httpapi`: `GET /projects` → `ProjectService.ListProjects`; `GET /projects/{id}/tasks` → `TaskProjection.ListByProject`.
- Юнит-тесты (реальные сервисы поверх `internal/application/inmemory`, тот же паттерн, что TASK-068/069) + интеграционный тест `ProjectStore.List` на реальном PostgreSQL.

### Не входит

- Пагинация, фильтрация, сортировка сверх `ORDER BY id` — решение по реальной потребности объёма данных, не сейчас.
- Спецификации `docs/api/*.md` для новых операций — TASK-073.

## Критерии приёмки

- [x] `GET /projects` возвращает все существующие проекты; пустой список — не ошибка, а `200` с `[]`.
- [x] `GET /projects/{id}/tasks` возвращает все задачи, о которых знает `TaskProjection` для данного проекта; задачи другого проекта не попадают в список (реальная проверка изоляции — прямое продолжение BUGFIX-003).
- [x] `ProjectStore.List` проверен на реальном PostgreSQL.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/{ports.go,project.go,projection.go}` и тесты.
- `internal/application/inmemory/{store.go,stores_test.go}`.
- `internal/infrastructure/postgres/{project_store.go,project_task_store_integration_test.go}`.
- `apps/api/httpapi/{projects.go,tasks.go}` и тесты.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — блокирует TASK-075/076 (страницам Dashboard нужны эти данные)

## План реализации

1. `application.ProjectStore.List(ctx) ([]*project.Project, error)` — новый метод порта (`ports.go`).
2. `postgres.ProjectStore.List` — `SELECT ... ORDER BY id`; `inmemory.Store[T].List` — общий для всех пяти generic-фейков (сортировка по `idOf`).
3. `application.ProjectService.ListProjects` — тонкая обёртка над `Projects.List`.
4. `application.TaskProjection.ListByProject(projectID)` — линейный проход `views` с фильтром `v.ProjectID == projectID`; восстановление полной нормализации ключа карты (обсуждавшаяся при планировании перестройка на `map[string]map[string]TaskView`) не понадобилось — `TaskView` уже несёт `ProjectID` как поле, простой фильтр даёт тот же результат при значительно меньшем риске регрессии для уже работающего `Handle`/`Get`/`Rebuild`.
5. `apps/api/httpapi`: `GET /projects` (`projects.go`), `GET /projects/{projectId}/tasks` (`tasks.go`, координирует с уже существующими `POST /projects/{projectId}/tasks` и `GET /projects/{projectId}/tasks/{id}` — `net/http.ServeMux` разруливает по числу сегментов пути без конфликтов).
6. Тесты: юнит (`internal/application`, `internal/application/inmemory`, `apps/api/httpapi`) — упорядоченность, пустой список не ошибка, изоляция между проектами (прямое продолжение BUGFIX-003); интеграционный (`internal/infrastructure/postgres`) — `ProjectStore.List` на реальном PostgreSQL, устойчивый к параллельным данным от других тестов (уникальный суффикс + проверка относительного порядка, не точного количества).
7. `make verify`, затем живой прогон интеграционных тестов и полного набора (`-tags=integration`) на реальном PostgreSQL, включая регрессию `TestGoldenPath_HTTP`.

## История

2026-07-22 — Architect — EPIC-009 открыт; задача поставлена в очередь (первая — остальные задачи Dashboard зависят от неё).

2026-07-22 — Developer — задача взята в работу, реализована и проверена вживую (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-072 — списковые операции: `GET /projects`, `GET /projects/{id}/tasks`.

### Что сделано

- `application.ProjectStore.List` (порт) + `postgres.ProjectStore.List` (реализация, `ORDER BY id`) + `inmemory.Store[T].List` (общий фейк для всех пяти агрегатов).
- `application.ProjectService.ListProjects` — тонкая обёртка.
- `application.TaskProjection.ListByProject` — фильтр по полю `ProjectID`, без перестройки внутренней структуры карты (более простое решение, чем изначально предполагалось при планировании).
- `apps/api/httpapi`: `GET /projects` → `ListProjects`; `GET /projects/{projectId}/tasks` → `ListByProject`. Оба пустых списка возвращают `200` с `[]`, не `null`.
- Юнит-тесты во всех трёх слоях (Application/inmemory/httpapi), включая явную проверку изоляции между проектами для списка задач.
- Интеграционный тест `ProjectStore.List` на реальном PostgreSQL.

### Изменённые файлы

- `internal/application/{ports.go,project.go,projection.go}` и тесты (`project_test.go`, `projection_test.go`).
- `internal/application/inmemory/{store.go,stores_test.go}`.
- `internal/infrastructure/postgres/{project_store.go,project_task_store_integration_test.go}`.
- `apps/api/httpapi/{projects.go,projects_test.go,tasks.go,tasks_test.go}`.
- `internal/application/README.md`, `internal/infrastructure/README.md`, `apps/api/README.md`.

### Как проверялось

- `go test ./... -cover` — все пакеты зелёные (application 83.4%, inmemory 88.4%, httpapi 85.1%).
- `make verify` — чисто.
- Живая проверка: `docker compose up -d postgres`, `TEST_DATABASE_URL` выставлен, `go test -tags=integration -count=1 ./internal/infrastructure/postgres/... -run TestProjectStore_List -v` — три прогона подряд без кеша, все зелёные. Дополнительно прогнан весь интеграционный набор (`postgres`, `wiring`, `apps/api/httpapi`) с тегом `integration` — включая `TestGoldenPath_HTTP` — всё зелёное, регрессий нет. `docker compose down -v` — чисто.

### Обновлённая документация

- `internal/application/README.md`, `internal/infrastructure/README.md`, `apps/api/README.md`.

### Open Questions

Нет.

### Рекомендации

TASK-075/076 (страницы Dashboard) должны трактовать `null`/отсутствие `state` как невозможное — оба списковых эндпоинта гарантированно возвращают `[]`, не `null`, для пустого результата.
