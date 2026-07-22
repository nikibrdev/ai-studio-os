# apps/api — REST-слой платформы

## Назначение

Реализация REST API ([ADR-003](../../docs/adr/ADR-003-api-protocol.md), [EPIC-008](../../docs/roadmap/EPIC-008-api-layer.md)): единственный способ для внешних клиентов и будущего Dashboard (v0.8, строится после этого эпика — [decision](../../engineering/decisions/2026-07-22-api-before-dashboard-build-order.md)) обратиться к платформе. Контракт каждой операции — [docs/api/](../../docs/api/README.md), написан до реализации (Documentation First).

## Содержание

### Структура

| Пакет | Содержимое | Задача |
| --- | --- | --- |
| (корень) | `main.go` — точка входа: сборка зависимостей через `wiring.System`, запуск HTTP-сервера, graceful shutdown | TASK-067 |
| `httpapi/` | Маршрутизация, хендлеры (`projects.go`, `tasks.go`, `work.go`, `artifacts.go`, `executions.go`, `completion.go`), JSON-кодирование, отображение ошибок в HTTP-коды | TASK-067…070 |

### Зависимости (module-boundaries.md)

- **Разрешено:** `internal/application` (use-case-сервисы, `TaskProjection`); `internal/infrastructure/wiring` (только в `main.go`, для сборки `System` — сами хендлеры `System` не видят); `pkg/`; стандартная библиотека (`net/http`).
- **Запрещено:** доменная логика (правила — только в `internal/domain`, вызываемые исключительно через `internal/application`); прямой доступ к хранилищам в обход `internal/application`; зависимость от `apps/orchestrator`, `agents/`, `tools/`.
- **Узкое исключение:** `httpapi/errors.go` импортирует несколько пакетов `internal/domain/*` **только** для сравнения по идентичности с уже публичными sentinel-ошибками (`errors.Is`) при отображении в HTTP-код — не для вызова доменной логики. `internal/application`'s use-case-методы возвращают эти ошибки как есть, не оборачивая (решение EPIC-004), поэтому любой вызывающий, различающий их, неизбежно знает о них.

### `main.go` — точка входа

Читает `DATABASE_URL` (обязателен, [`postgres.DatabaseURLEnv`](../../internal/infrastructure/postgres/config.go)) и `QDRANT_URL` (опционален — пустое значение оставляет `wiring.System.Memory` равным `nil`, тот же принцип, что и `GITHUB_TOKEN`/`Repository`); порт — `PORT`, по умолчанию `8080`. Строит `wiring.System`, подписывает `application.TaskProjection` на `System.Events`, конструирует все пять use-case-сервисов (`ProjectService`, `TaskPlanningService` — с `IDs: sys.Tasks`, тот же `*postgres.TaskStore` уже реализует оба порта, `WorkService`, `ResultService`, `CompletionService`) и запускает `http.Server` с graceful shutdown по `SIGINT`/`SIGTERM`.

### `httpapi` — маршрутизация и хендлеры

`NewServer(deps Deps) http.Handler` строит маршрутизатор на стандартном `net/http.ServeMux` (Go 1.22+ маршрутизация по методу и пути — отдельная библиотека-роутер не нужна, тот же принцип, что и у REST-клиентов `github`/`qdrant`, EPIC-005/007). Единая функция `writeError`/`statusFor` (`errors.go`) отображает sentinel-ошибки Application/Domain Layer в HTTP-коды по конвенции [docs/api/README.md](../../docs/api/README.md) — ни один хендлер не выбирает код самостоятельно. `writeJSON`/`decodeJSON` (`json.go`) — общие хелперы кодирования.

Все 15 операций [docs/api/](../../docs/api/README.md) реализованы (`GET /healthz` + Projects/Tasks/Artifacts/Executions). Задаче-специфичные маршруты вложены под `/projects/{projectId}/tasks/...`, а не `/tasks/{id}/...` (**BUGFIX-003**): публичный `TASK-NNN` уникален только в рамках Project (ADR-011) — живая проверка вскрыла, что при плоском `/tasks/{id}` два разных проекта с одинаковым `TASK-001` молча портили данные друг друга. `/executions/{id}/succeed`/`fail` остались невложенными (Execution ID — глобально уникальный `crypto/rand`), но принимают `projectId` в теле запроса.

### Проверено вживую

`apps/api/httpapi/golden_path_integration_test.go` (TASK-070, тег `integration`) проводит задачу через весь golden path настоящими HTTP-запросами к серверу на реальном PostgreSQL — создание/активация проекта, задача, запуск работы, артефакт, ревью (обе ветки — одобрено/отклонено), тестирование (обе ветки — успех/провал), `Done`. `RepositoryProvider` — фейк EPIC-004 (тот же принцип, что и у `TestGoldenPath_Infrastructure`, EPIC-005): токена GitHub нет во всех окружениях, где это выполняется.

### Запуск локально

```bash
docker compose up -d
export DATABASE_URL="postgres://ai_studio_os:ai_studio_os@localhost:5432/ai_studio_os?sslmode=disable"
export QDRANT_URL="http://localhost:6333"  # опционально
go run ./apps/api
```

## Статус

Завершён (EPIC-008)

## Последнее обновление

2026-07-22
