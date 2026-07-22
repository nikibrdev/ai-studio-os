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

- [ ] `GET /projects` возвращает все существующие проекты; пустой список — не ошибка, а `200` с `[]`.
- [ ] `GET /projects/{id}/tasks` возвращает все задачи, о которых знает `TaskProjection` для данного проекта; задачи другого проекта не попадают в список (реальная проверка изоляции — прямое продолжение BUGFIX-003).
- [ ] `ProjectStore.List` проверен на реальном PostgreSQL.
- [ ] `make verify` — чисто.

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

## История

2026-07-22 — Architect — EPIC-009 открыт; задача поставлена в очередь (первая — остальные задачи Dashboard зависят от неё).

## Отчёт о выполнении
