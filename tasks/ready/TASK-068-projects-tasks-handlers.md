# TASK-068: Хендлеры Projects/Tasks

## Тип

feature

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Реализовать HTTP-хендлеры для создания/активации проекта и создания/планирования/чтения задачи — первая половина golden path через HTTP.

## Контекст

По спецификациям TASK-066 (`docs/api/projects.md`, `docs/api/tasks.md`), поверх каркаса TASK-067 (маршрутизация, JSON-хелперы, отображение ошибок), вызывая `ProjectService` (TASK-064) и `TaskPlanningService` (уже существует, EPIC-004) — хендлеры не содержат бизнес-правил, только разбор запроса → вызов сервиса → сериализация ответа.

## Scope

### Входит

- `POST /projects` → `ProjectService.CreateProject`.
- `POST /projects/{id}/repositories` → `ProjectService.ConnectRepository` (обязателен перед `activate` — guard домена «≥1 Repository», TASK-064).
- `POST /projects/{id}/activate` → `ProjectService.Activate`.
- `POST /tasks` → `TaskPlanningService.CreateTask` (ID берётся из генератора последовательности, TASK-065, а не из тела запроса).
- `POST /tasks/{id}/plan` → `TaskPlanningService.PlanTask`.
- `GET /tasks/{id}` → `TaskProjection.Get` (404, если проекция ещё не видела ни одного события по этому ID).
- Юнит-тесты хендлеров на `httptest` с фейковыми сервисами (тот же паттерн, что уже используется в `internal/application`'s `inmemory`).

### Не входит

- Хендлеры Work/Result/Completion (TASK-069).
- Сквозной интеграционный тест на реальном PostgreSQL (TASK-070).

## Критерии приёмки

- [ ] Все шесть операций реализованы в точности по `docs/api/projects.md`/`docs/api/tasks.md` (TASK-066) — коды ошибок совпадают со спецификацией.
- [ ] `POST /tasks` использует генератор последовательности (TASK-065), не принимает ID в теле запроса.
- [ ] `GET /tasks/{id}` для неизвестного ID возвращает 404, не 200 с пустым телом.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/api/httpapi/projects.go`, `apps/api/httpapi/tasks.go` и соответствующие `_test.go`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-064, 065, 066, 067

## План реализации

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

## Отчёт о выполнении
