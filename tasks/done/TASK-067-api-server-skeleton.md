# TASK-067: Каркас apps/api (main.go, wiring, маршрутизация, ошибки)

## Тип

feature

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Создать несущую конструкцию `apps/api`: точку входа, сборку зависимостей через `wiring.System`, маршрутизацию и единообразную обработку ошибок — без бизнес-логики (module-boundaries.md: `apps/api` не содержит доменных правил).

## Контекст

`internal/infrastructure/wiring.System` (EPIC-005/007) уже собирает все адаптеры (`Pool`, пять Store, `Events`, `Repository`, `Memory`); `apps/api` строит поверх него use-case-сервисы `internal/application` и передаёт их HTTP-хендлерам. Маршрутизация — стандартный `net/http.ServeMux` (Go 1.22+ поддерживает метод и путь-паттерны нативно, отдельная библиотека-роутер не нужна — тот же принцип «не добавлять зависимость без необходимости», что REST-клиенты `github`/`qdrant`).

## Scope

### Входит

- `apps/api/main.go` — package main: читает `DATABASE_URL`/`QDRANT_URL` (опционален, как и в `wiring.New`) из окружения, порт HTTP — из окружения с разумным значением по умолчанию, строит `wiring.System`, конструирует use-case-сервисы, запускает сервер, graceful shutdown по сигналу.
- `apps/api/httpapi/` (пакет, тестируемый отдельно от `main`) — маршрутизация, JSON-хелперы кодирования/декодирования запросов и ответов, единая функция отображения ошибок домена/приложения в HTTP-коды (`ErrNotFound` → 404, `ErrProjectNotActive`/доменные sentinel-ошибки о недопустимом переходе → 409/400 — конкретная раскладка кодов фиксируется здесь и используется всеми последующими хендлерами).
- Health-эндпоинт (`GET /healthz`) — не требует БД, для проверки живости процесса.
- Юнит-тесты `httpapi` на `httptest`-сервере (тот же паттерн, что тесты GitHub/Qdrant REST-клиентов, но со стороны сервера).

### Не входит

- Конкретные хендлеры ресурсов (TASK-068/069).
- Аутентификация (ADR-012, Вариант 1).

## Критерии приёмки

- [x] `apps/api/main.go` собирается и запускается локально (реальный `go run`, не только компиляция) против `docker compose up -d`.
- [x] Отображение ошибок в HTTP-коды — единая функция, покрыта тестами на конкретных sentinel-ошибках Application Layer.
- [x] `GET /healthz` отвечает 200 без обращения к БД.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/api/main.go`, `apps/api/httpapi/*.go`, `apps/api/README.md` (черновик, дополняется TASK-071).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — использует `docs/api/*.md` (TASK-066) как источник контракта ошибок/путей

## План реализации

1. `apps/api/httpapi/errors.go` — таблицы `badRequestErrors`/`conflictErrors` (данные, не код-переключатель) по конвенции `docs/api/README.md`; `statusFor`/`writeError`. Узкое, явно задокументированное исключение из module-boundaries.md: импорт нескольких `internal/domain/*` только ради сравнения по идентичности с уже публичными sentinel-ошибками (`internal/application` возвращает их как есть, не оборачивая).
2. `apps/api/httpapi/json.go` — `writeJSON`/`decodeJSON`.
3. `apps/api/httpapi/server.go` — `Deps`, `NewServer` (пока только `GET /healthz`; ресурсные маршруты добавляются TASK-068/069 редактированием этого же файла).
4. `apps/api/main.go` — сборка `wiring.System`, пяти use-case-сервисов (`IDs: sys.Tasks` — тот же `*postgres.TaskStore` уже реализует оба порта, без отдельной сборки), `http.Server` с graceful shutdown.
5. Тесты `httpapi` (`httptest`, без реального Docker/Postgres): все категории `statusFor` (по одной ошибке на каждый sentinel + wrapped-ошибка + неизвестная ошибка → 500), `writeJSON`/`decodeJSON` round-trip, `/healthz` и неизвестный маршрут.
6. `apps/api/README.md` (черновик) — структура, зависимости, узкое исключение по ошибкам, запуск локально.
7. `make verify`, затем живой прогон: `docker compose up -d postgres`, `go run ./apps/api` (реальный процесс, не только `go build`), `curl /healthz` → 200, `curl` неизвестного пути → 404, корректное завершение процесса.

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

2026-07-22 — Developer — задача взята в работу, реализована и проверена вживую (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-067 — каркас `apps/api` (main.go, wiring, маршрутизация, ошибки).

### Что сделано

- `apps/api/httpapi/errors.go` — единая функция `statusFor`/`writeError`, отображающая sentinel-ошибки Application/Domain Layer в HTTP-коды строго по таблице `docs/api/README.md` (404/400/409/500); данные вынесены в две таблицы (`badRequestErrors`/`conflictErrors`), не в цепочку `if`.
- `apps/api/httpapi/json.go` — `writeJSON` (JSON-тело + статус, `nil` → без тела), `decodeJSON`.
- `apps/api/httpapi/server.go` — `Deps` (пять use-case-сервисов + `TaskProjection`), `NewServer` — пока только `GET /healthz`.
- `apps/api/main.go` — реальная точка входа: `wiring.New` из `DATABASE_URL`/`QDRANT_URL` (опционален), подписка `TaskProjection` на `System.Events`, сборка всех пяти сервисов (`TaskPlanningService.IDs: sys.Tasks` — тот же объект уже реализует `TaskIDGenerator`, отдельная сборка не понадобилась, как и предполагалось в рекомендации TASK-065), HTTP-сервер с graceful shutdown по `SIGINT`/`SIGTERM`.
- `apps/api/README.md` — структура, зависимости (включая узкое, явно задокументированное исключение по доменным sentinel-ошибкам), запуск локально.
- 100% покрытие пакета `httpapi` (все ветки `statusFor`, JSON-хелперы, health-эндпоинт, неизвестный маршрут).

### Изменённые файлы

- `apps/api/main.go` (новый).
- `apps/api/httpapi/{doc,server,json,errors}.go` и соответствующие `_test.go` (новые).
- `apps/api/README.md` (новый).

### Как проверялось

- `go test ./apps/api/... -v -cover` — все тесты зелёные, 100.0% покрытия `httpapi` (`apps/api` — `[no test files]`, ожидаемо: `main.go` — тонкая точка входа без бизнес-логики для юнит-теста).
- `make verify` — чисто.
- Живая проверка: `docker compose up -d postgres`; `DATABASE_URL=... PORT=8081 go run ./apps/api` — реальный запущенный процесс (не только `go build`), без `QDRANT_URL` (проверка опционального пути); `curl http://localhost:8081/healthz` → `200`; `curl http://localhost:8081/does-not-exist` → `404`; процесс остановлен, `docker compose down` — чисто, порт освобождён.

### Обновлённая документация

- `apps/api/README.md`.

### Open Questions

Нет.

### Рекомендации

TASK-068/069 редактируют `server.go`: убрать `_` у параметра `deps Deps` в `NewServer`, добавить вызовы `register*Routes(mux, deps)` — каждый в своём файле, по одному на ресурс (Projects+Tasks — TASK-068; Work/Result/Completion — TASK-069), тот же принцип, что и наращивание `wiring.New` по задачам в EPIC-005/007.
