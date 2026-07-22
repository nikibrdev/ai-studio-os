# TASK-070: Сквозной golden path через реальные HTTP-вызовы

## Тип

test

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Подтвердить, что `apps/api` реально проводит задачу через весь golden path — не юнит-тестами хендлеров на фейках, а настоящими HTTP-запросами к поднятому серверу поверх настоящего PostgreSQL (тот же принцип живой проверки, что `TestGoldenPath_Infrastructure`, EPIC-005, и все интеграционные тесты этой сессии).

## Контекст

Golden path — [golden-path.md](../../docs/architecture/golden-path.md): создание проекта → создание и планирование задачи → запуск работы → черновик и публикация артефакта → результат исполнения → ревью → тестирование → завершение. TASK-064…069 реализуют каждый шаг по отдельности; эта задача — сквозная проверка их совместной работы через реальный сетевой протокол, а не только прямые вызовы Go-функций.

## Scope

### Входит

- `apps/api/httpapi/golden_path_integration_test.go` (тег `integration`, `TEST_DATABASE_URL` — тот же опт-ин паттерн, что везде в проекте): поднимает `apps/api` на `httptest.Server` (или реальном порту) поверх настоящего `wiring.System`, проводит весь сценарий через `net/http.Client`, проверяет промежуточные состояния через `GET /tasks/{id}`.
- Прогон вживую минимум трижды подряд (`-count=1`), фиксация результата в отчёте задачи.

### Не входит

- Нагрузочное или конкурентное тестирование сверх того, что уже покрыто TASK-065 (последовательность ID).

## Критерии приёмки

- [x] Сценарий проходит целиком через HTTP на реальном PostgreSQL, включая ветки «changes requested» и «tests failed» (по аналогии с `internal/application/e2e_test.go`, EPIC-004).
- [x] Три прогона подряд (`-count=1`) — зелёные.
- [x] `make verify` — чисто; тест — реальный PostgreSQL, пропускается без него.

## Затрагиваемые модули и документы

- `apps/api/httpapi/golden_path_integration_test.go`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-064…069

## План реализации

1. `apps/api/httpapi/golden_path_integration_test.go` (тег `integration`, package `httpapi` — переиспользует уже существующие неэкспортированные типы запросов/ответов пакета): `wiring.New(dsn, qdrantURL)` — реальный `wiring.System`; `Completion.Repositories` — `inmemory.NewRepositoryProvider()` (не `sys.Repository`), тот же принцип, что уже применён в `TestGoldenPath_Infrastructure` (EPIC-005): токена GitHub нет во всех окружениях, где это выполняется, а не «нужно ещё что-то придумать».
2. `httptest.NewServer(NewServer(deps))` — реальный TCP-листенер на loopback, `server.Client()` + собственный `httpDo`-хелпер (`net/http.NewRequest`/`client.Do`) — намеренно НЕ переиспользует `doRequest` из `deps_test.go` (тот вызывает `ServeHTTP` в процессе, не по-настоящему через сеть — не то, что должен проверить этот тест).
3. Полный сценарий golden path через реальные HTTP-запросы: создание/подключение репозитория/активация проекта → создание и планирование задачи → (исполнитель сохраняется напрямую в `sys.Executors` — нет HTTP-маршрута регистрации, вне scope эпика) → запуск работы → черновик/доработка/публикация артефакта → результат исполнения → ревью (первый круг — правки запрошены, второй — одобрено) → тестирование (первый прогон — провален, второй — успешен с реальным merge через фейк `RepositoryProvider`) → `Done`. Промежуточные состояния проверяются через `GET /projects/{projectId}/tasks/{id}`.
4. `make verify`, затем живой прогон трижды подряд (`-count=1`) против `docker compose up -d postgres`.

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

2026-07-22 — Developer — задача взята в работу, реализована и проверена вживую (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-070 — сквозной golden path через реальные HTTP-вызовы.

### Что сделано

- `apps/api/httpapi/golden_path_integration_test.go` — `TestGoldenPath_HTTP`: полный golden path (`docs/architecture/golden-path.md`) проведён через настоящие HTTP-запросы (`httptest.Server`, реальный TCP-листенер, `net/http.Client`) к серверу, собранному на реальном `wiring.System` (PostgreSQL), включая обе ветки TASK-045 (changes requested, tests failed).
- `RepositoryProvider` — `inmemory.NewRepositoryProvider()` вместо `sys.Repository`, тот же принцип, что уже применён в `TestGoldenPath_Infrastructure` (EPIC-005): токена GitHub нет во всех окружениях, где это выполняется.
- Исполнитель сохранён напрямую в `sys.Executors` (реальный PostgreSQL) — нет HTTP-маршрута регистрации исполнителя (вне scope эпика, ADR-007).
- Собственный HTTP-хелпер (`httpDo`) через `net/http.Client` — намеренно не переиспользует `doRequest` из `deps_test.go` (тот вызывает `ServeHTTP` в процессе, не через настоящую сеть).

### Изменённые файлы

- `apps/api/httpapi/golden_path_integration_test.go` (новый).

### Как проверялось

- `make verify` — чисто.
- Живая проверка: `docker compose up -d postgres`, `TEST_DATABASE_URL=... go test -tags=integration -count=1 ./apps/api/httpapi/... -run TestGoldenPath_HTTP -v` прогнан трижды подряд без кеша — все три прогона зелёные (0.36–0.56s каждый), включая проверку, что `MergeCalls` фейкового `RepositoryProvider` зафиксировал ровно один вызов для `github.com/org/repo/pr-1`. `docker compose down -v` после проверки — чисто.
- `go test -tags=integration ./...` — весь проект компилируется и проходит с тегом `integration`.

### Обновлённая документация

Нет отдельных изменений документации сверх кода — эта задача целиком тестовая, спецификации (`docs/api/*.md`) уже описывают проверяемое поведение.

### Open Questions

Нет.

### Рекомендации

TASK-071 (закрытие эпика) может явно сослаться на этот тест как на прямое доказательство критерия завершения EPIC-008 «`apps/api` реализует весь golden path через HTTP».
