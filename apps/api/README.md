# apps/api — REST-слой платформы

## Назначение

Реализация REST API ([ADR-003](../../docs/adr/ADR-003-api-protocol.md), [EPIC-008](../../docs/roadmap/EPIC-008-api-layer.md)): единственный способ для внешних клиентов и будущего Dashboard (v0.8, строится после этого эпика — [decision](../../engineering/decisions/2026-07-22-api-before-dashboard-build-order.md)) обратиться к платформе. Контракт каждой операции — [docs/api/](../../docs/api/README.md), написан до реализации (Documentation First).

## Содержание

### Структура

| Пакет | Содержимое | Задача |
| --- | --- | --- |
| (корень) | `main.go` — точка входа: сборка зависимостей через `wiring.System`, запуск HTTP-сервера, graceful shutdown | TASK-067 |
| `httpapi/` | Маршрутизация, хендлеры, JSON-кодирование, отображение ошибок в HTTP-коды | TASK-067…069 |

### Зависимости (module-boundaries.md)

- **Разрешено:** `internal/application` (use-case-сервисы, `TaskProjection`); `internal/infrastructure/wiring` (только в `main.go`, для сборки `System` — сами хендлеры `System` не видят); `pkg/`; стандартная библиотека (`net/http`).
- **Запрещено:** доменная логика (правила — только в `internal/domain`, вызываемые исключительно через `internal/application`); прямой доступ к хранилищам в обход `internal/application`; зависимость от `apps/orchestrator`, `agents/`, `tools/`.
- **Узкое исключение:** `httpapi/errors.go` импортирует несколько пакетов `internal/domain/*` **только** для сравнения по идентичности с уже публичными sentinel-ошибками (`errors.Is`) при отображении в HTTP-код — не для вызова доменной логики. `internal/application`'s use-case-методы возвращают эти ошибки как есть, не оборачивая (решение EPIC-004), поэтому любой вызывающий, различающий их, неизбежно знает о них.

### `main.go` — точка входа

Читает `DATABASE_URL` (обязателен, [`postgres.DatabaseURLEnv`](../../internal/infrastructure/postgres/config.go)) и `QDRANT_URL` (опционален — пустое значение оставляет `wiring.System.Memory` равным `nil`, тот же принцип, что и `GITHUB_TOKEN`/`Repository`); порт — `PORT`, по умолчанию `8080`. Строит `wiring.System`, подписывает `application.TaskProjection` на `System.Events`, конструирует все пять use-case-сервисов (`ProjectService`, `TaskPlanningService` — с `IDs: sys.Tasks`, тот же `*postgres.TaskStore` уже реализует оба порта, `WorkService`, `ResultService`, `CompletionService`) и запускает `http.Server` с graceful shutdown по `SIGINT`/`SIGTERM`.

### `httpapi` — маршрутизация и хендлеры

`NewServer(deps Deps) http.Handler` строит маршрутизатор на стандартном `net/http.ServeMux` (Go 1.22+ маршрутизация по методу и пути — отдельная библиотека-роутер не нужна, тот же принцип, что и у REST-клиентов `github`/`qdrant`, EPIC-005/007). Единая функция `writeError`/`statusFor` (`errors.go`) отображает sentinel-ошибки Application/Domain Layer в HTTP-коды по конвенции [docs/api/README.md](../../docs/api/README.md) — ни один хендлер не выбирает код самостоятельно. `writeJSON`/`decodeJSON` (`json.go`) — общие хелперы кодирования.

Сейчас реализован только `GET /healthz` (не обращается ни к одной зависимости) — резолюция ресурсных маршрутов добавляется по мере реализации хендлеров (TASK-068/069), редактированием `NewServer` в том же файле.

### Запуск локально

```bash
docker compose up -d
export DATABASE_URL="postgres://ai_studio_os:ai_studio_os@localhost:5432/ai_studio_os?sslmode=disable"
export QDRANT_URL="http://localhost:6333"  # опционально
go run ./apps/api
```

## Статус

В работе (EPIC-008)

## Последнее обновление

2026-07-22
