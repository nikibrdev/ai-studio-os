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

- [x] Все шесть операций реализованы в точности по `docs/api/projects.md`/`docs/api/tasks.md` (TASK-066) — коды ошибок совпадают со спецификацией.
- [x] `POST /tasks` использует генератор последовательности (TASK-065), не принимает ID в теле запроса.
- [x] `GET /tasks/{id}` для неизвестного ID возвращает 404, не 200 с пустым телом.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/api/httpapi/projects.go`, `apps/api/httpapi/tasks.go` и соответствующие `_test.go`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-064, 065, 066, 067

## План реализации

1. `apps/api/httpapi/json.go` — дополнить `decodeOptionalJSON` (пустое тело → нулевое значение, не ошибка — большинство операций `docs/api/*.md` имеют полностью опциональное тело или только опциональный `actor`) и `writeInvalidBody` (400 для реально нечитаемого JSON, отдельно от таблицы доменных ошибок `statusFor`).
2. `apps/api/httpapi/projects.go` — три хендлера, по одному на операцию `docs/api/projects.md`; `registerProjectRoutes`.
3. `apps/api/httpapi/tasks.go` — три хендлера (create/plan/get) из шести операций `docs/api/tasks.md` (остальные три — TASK-069); `registerTaskCreationRoutes`.
4. `apps/api/httpapi/server.go` — `NewServer` вызывает обе функции регистрации маршрутов (убран `_` у параметра `deps`).
5. Тесты: `deps_test.go` — общий `testDeps()`, реальные `application.*Service` поверх `internal/application/inmemory` (тот же выбор, что и тесты `internal/application` самого — `Deps` хранит конкретные типы сервисов, не интерфейсы, поэтому нет иного естественного шва для подмены на этом уровне) + `sequentialTaskIDGenerator` (детерминированный фейк `TaskIDGenerator` для тестов — настоящий генератор, TASK-065, требует базы). `projects_test.go`/`tasks_test.go` — по HTTP целиком (`httptest`), включая полную последовательность create→connect→activate→create-task→plan→get.
6. `make verify`, затем живая проверка: `docker compose up -d postgres`, реальный `go run ./apps/api`, `curl` через весь путь (create project → connect repository → activate → create task ×2 → plan → get → get неизвестного) — подтверждает, что `TASK-001`/`TASK-002` берутся из настоящего генератора БД (TASK-065), а не из тестового фейка.

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

2026-07-22 — Developer — задача взята в работу, реализована и проверена вживую (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-068 — хендлеры Projects/Tasks (первая половина golden path через HTTP).

### Что сделано

- `apps/api/httpapi/projects.go` — `POST /projects`, `POST /projects/{id}/repositories`, `POST /projects/{id}/activate`.
- `apps/api/httpapi/tasks.go` — `POST /tasks` (ID генерируется платформой, не читается из тела), `POST /tasks/{id}/plan`, `GET /tasks/{id}` (из `TaskProjection`, 404 если проекция не видела события по ID).
- `apps/api/httpapi/json.go` — `decodeOptionalJSON`/`writeInvalidBody`, использованы во всех новых хендлерах.
- `apps/api/httpapi/server.go` — `NewServer` подключает обе группы маршрутов.
- Тесты: реальные `application.*Service` поверх `internal/application/inmemory` (не фейки на уровне HTTP — `Deps` хранит конкретные типы, интерфейсного шва на этом уровне нет и вводить его ради тестируемости было бы избыточно); 16 тестов покрывают успешные пути, все три класса ошибок (400/404/409) и полную последовательность создания проекта.

### Изменённые файлы

- `apps/api/httpapi/projects.go`, `apps/api/httpapi/tasks.go` (новые).
- `apps/api/httpapi/json.go`, `apps/api/httpapi/server.go` (дополнены).
- `apps/api/httpapi/deps_test.go`, `apps/api/httpapi/projects_test.go`, `apps/api/httpapi/tasks_test.go` (новые).

### Как проверялось

- `go test ./apps/api/... -v -cover` — все тесты зелёные, 90.8% покрытия `httpapi`.
- `make verify` — чисто.
- Живая проверка: `docker compose up -d postgres`, `DATABASE_URL=... PORT=8082 go run ./apps/api` — реальный запущенный процесс; через `curl` пройден весь путь: создание проекта (201), подключение репозитория (204), активация (204), создание задачи дважды подряд (201, ID `TASK-001` и `TASK-002` — настоящий генератор БД, TASK-065, не тестовый фейк), планирование (204), чтение состояния (200, `state: ready`), чтение неизвестной задачи (404). Процесс и `docker compose down` — чисто.

### Обновлённая документация

Нет отдельных изменений документации сверх кода (спецификации уже написаны в TASK-066).

### Open Questions

Нет.

### Рекомендации

TASK-069 продолжает тот же файл `server.go` (добавление вызова `registerWorkResultCompletionRoutes` или аналогичного) и может переиспользовать `testDeps()`/`jsonBody`/`doRequest`/`createActiveProject` из `deps_test.go` — не создавать копии этих хелперов.
