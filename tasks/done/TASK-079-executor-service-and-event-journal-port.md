# TASK-079: ExecutorService, ExecutorStore.List, порт EventJournal

## Тип

feature

## Эпик

[EPIC-010 Orchestrator](../../docs/roadmap/EPIC-010-orchestrator.md)

## Цель

Добавить в `internal/application` три недостающих элемента, без которых Orchestrator не может ни зарегистрировать исполнителя, ни найти его, ни узнать о новых событиях из отдельного процесса: `ExecutorService` (Register/Activate), `ExecutorStore.List` (запрос) и новый порт `EventJournal.Since`.

## Контекст

Разбор при открытии EPIC-010 («Контекст» эпика) показал: `internal/domain/executor` полностью реализован (EPIC-003), но ни одного use-case для регистрации/активации исполнителя в Application Layer нет — тесты создают `Executor` напрямую через домен. Отдельно: продуктивная `eventbus.Bus` — только внутрипроцессная (ADR-002), поэтому Orchestrator как отдельный процесс не может подписаться на неё; нужен курсорный опрос журнала событий через новый узкий порт (реализация — TASK-080), а не прямой доступ к `internal/infrastructure` (запрещён `module-boundaries.md`).

## Scope

### Входит

- `internal/application/executor.go` (новый файл) — `ExecutorService{Executors ExecutorStore, Events platform.EventBus}`, по стилю `project.go` (TASK-064): `Register(ctx, params) (*executor.Executor, error)` (оборачивает `executor.New`, публикует `ExecutorRegistered`), `Activate(ctx, id, actor string) error` (оборачивает `Executor.Activate`, публикует `ExecutorActivated`).
- `internal/application/ports.go` — `ExecutorStore.List(ctx) ([]*executor.Executor, error)`, по прецеденту `ProjectStore.List` (EPIC-009); новый порт `EventJournal interface { Since(ctx context.Context, after time.Time) ([]platform.Event, error) }`.
- `internal/application/inmemory` — реализация `List` на фейке `ExecutorStore` (по образцу фейка `ProjectStore`).
- Юнит-тесты: успешные пути `Register`/`Activate`, отказные сценарии домена (`ErrMissingField`, `ErrNoRoles`, `ErrAlreadyActive`, `ErrRetired`), `List` на пустом и непустом хранилище.

### Не входит

- Реализация `EventJournal` (курсорный SQL-запрос) — TASK-080.
- Использование `ExecutorService`/`List`/`EventJournal` из `apps/orchestrator` — TASK-081/082.
- HTTP-эндпоинты для исполнителей — не входит в эпик вовсе (см. «Не входит» EPIC-010).

## Критерии приёмки

- [x] `ExecutorService.Register`/`Activate` реализованы, стиль идентичен `ProjectService`/`TaskPlanningService`.
- [x] `ExecutorStore.List` объявлен и реализован в `inmemory`-фейке (плюс в `postgres.ExecutorStore` — вынужденное дополнение, см. Отчёт).
- [x] Порт `EventJournal` объявлен в `ports.go` (без реализации — она в TASK-080).
- [x] Юнит-тесты покрывают успешные и отказные пути.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/executor.go`, `internal/application/executor_test.go`, `internal/application/ports.go`, `internal/application/inmemory/store.go` (или аналог), `internal/application/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — использует уже принятый `internal/domain/executor`, не меняет его

## План реализации

1. `internal/application/ports.go`:
   - Добавить `List(ctx context.Context) ([]*executor.Executor, error)` в интерфейс `ExecutorStore`.
   - Добавить импорт `"time"` и `"ai-studio-os/internal/platform"`; объявить `EventJournal interface { Since(ctx context.Context, after time.Time) ([]platform.Event, error) }` — без реализации (TASK-080).
2. `internal/application/executor.go` (новый файл, по стилю `project.go`):
   - `ExecutorService{Executors ExecutorStore, Events platform.EventBus}`.
   - `RegisterExecutorParams{ID, Backend string, Roles []shared.Role, Actor string}`.
   - `Register(ctx, p RegisterExecutorParams) (*executor.Executor, error)` — оборачивает `executor.New`, сохраняет, публикует `event.ExecutorRegistered` (source `"executor"`).
   - `Activate(ctx, id, actor string) error` — `Get` → `Executor.Activate()` → `Save` → публикует `event.ExecutorActivated`.
   - Приватный `publish`-хелпер, как в `project.go`.
3. `internal/application/inmemory` — фейк `ExecutorStore` уже возвращается как `*Store[executor.Executor]` (`stores.go`), а generic `Store[T].List` уже реализован (использован для `ProjectStore.List`, EPIC-009) — добавления интерфейсного метода в `ports.go` достаточно, отдельного кода фейка не требуется. Проверить сборку (`go build ./...`) сразу после правки ports.go, чтобы убедиться в этом до написания сервиса.
4. `internal/infrastructure/postgres/executor_store.go` — добавить `List(ctx) ([]*executor.Executor, error)` (`SELECT ... FROM executors ORDER BY id`, `executor.Restore` на каждой строке) по образцу `ProjectStore.List` (`project_store.go:52`) — необходимо: `var _ application.ExecutorStore = (*ExecutorStore)(nil)` перестанет компилироваться без этого метода, как только интерфейс вырастет (не по исходному описанию задачи, но неизбежное следствие расширения порта — тот же случай, что `ConnectRepository` в TASK-064).
5. Тесты:
   - `internal/application/executor_test.go` (пакет `application_test`, по образцу `project_test.go`): успешные пути `Register`/`Activate`; отказные — `ErrMissingField`, `ErrNoRoles` (из `executor.New`), `ErrAlreadyActive`, `ErrRetired`, `ErrNotFound` на неизвестном ID.
   - `internal/application/inmemory/stores_test.go` — добавить кейс на `List` для `ExecutorStore` (пустое и непустое хранилище, порядок по id), по образцу существующего теста `ProjectStore.List`.
   - `internal/infrastructure/postgres/executor_execution_artifact_store_integration_test.go` — добавить интеграционный тест `List` (`//go:build integration`), по образцу интеграционного теста `ProjectStore.List`.
6. `internal/application/README.md` — строка о новом сервисе в таблице состава (по образцу записи о `ProjectService`).
7. `make verify`; `go test ./internal/application/... ./internal/infrastructure/postgres/... -cover`.

## История

2026-07-23 — Architect — EPIC-010 открыт; задача поставлена в очередь первой (остальные задачи эпика зависят от неё).

2026-07-23 — Developer — план составлен, реализация начата по стандартному режиму сессии (авторизация пользователя на автономное выполнение).

2026-07-23 — Developer — задача реализована и проверена (см. Отчёт), в том числе интеграционным тестом на реальном PostgreSQL.

## Отчёт о выполнении

### Задача

TASK-079 — `ExecutorService`, `ExecutorStore.List`, порт `EventJournal` в `internal/application`.

### Что сделано

- **`ExecutorService`** (`internal/application/executor.go`) — `Register`/`Activate` поверх уже принятого домена `internal/domain/executor`, без изменения домена. Стиль идентичен `ProjectService` (TASK-064): узкий порт `ExecutorStore`, события через `Envelope`/`platform.EventBus`, приватный `publish`-хелпер (`source="executor"`).
- **События Executor публикуются с пустым `ProjectID`** — Executor не принадлежит проекту: доменный агрегат не несёт ссылки на Project, и `docs/architecture/events.md` не перечисляет поле проекта ни для одного из четырёх его событий. Это осознанное решение, а не упущение: подставлять сюда чужой идентификатор было бы выдумыванием факта, которого в домене нет.
- **`ExecutorStore.List`** — добавлен в порт (`ports.go`) по прецеденту `ProjectStore.List` (EPIC-009). Реализация в `inmemory`-фейке **не потребовала ни строки кода**: фейк — это `*Store[executor.Executor]`, а generic `Store[T].List` уже был реализован в EPIC-009 для `ProjectStore`; расширения интерфейса оказалось достаточно (проверено `go build` сразу после правки `ports.go`, как предписывал шаг 3 плана).
- **`postgres.ExecutorStore.List`** — вынужденное дополнение сверх исходного описания задачи, ровно как предвидел шаг 4 плана: `var _ application.ExecutorStore = (*ExecutorStore)(nil)` перестал компилироваться, как только порт вырос. Реализовано по образцу `ProjectStore.List` (`ORDER BY id`, `executor.Restore` на каждой строке). Тот же случай, что `ConnectRepository` в TASK-064 — неизбежное следствие, а не расширение scope «заодно».
- **Порт `EventJournal`** (`ports.go`) — объявлен без реализации (она в TASK-080). Doc-комментарий фиксирует *почему* порт существует: продуктивная `eventbus.Bus` доставляет только подписчикам своего процесса (ADR-002), поэтому отдельный процесс (`apps/orchestrator`) не может использовать `Subscribe` и опрашивает `Since` по курсору; порт нужен, чтобы `apps/orchestrator` зависел только от `internal/application`, а не от `internal/infrastructure` напрямую (запрещено `module-boundaries.md`).
- **9 юнит-тестов** (`executor_test.go`): успешные пути `Register`/`Activate`; отказные — `ErrMissingField`, `ErrNoRoles`, `ErrAlreadyActive`, `ErrRetired`, `ErrNotFound`; `List` на пустом и непустом хранилище с проверкой порядка по id.
- **1 интеграционный тест** (`postgres`, `//go:build integration`) — `List` на реальном PostgreSQL: проверяет относительный порядок двух уникально-суффиксированных исполнителей (общая тестовая БД накапливает строки других тестов, поэтому утверждение — о относительном порядке, а не о всём результате — тот же приём, что в `TestProjectStore_List_ReturnsCreatedProjectsInOrder`) и сохранность набора ролей после `Restore`.

### Изменённые файлы

- `internal/application/executor.go` (новый) — `ExecutorService`.
- `internal/application/executor_test.go` (новый) — 9 юнит-тестов.
- `internal/application/ports.go` — `ExecutorStore.List`, порт `EventJournal`, импорты `time`/`platform`.
- `internal/infrastructure/postgres/executor_store.go` — `List`.
- `internal/infrastructure/postgres/executor_execution_artifact_store_integration_test.go` — интеграционный тест `List`.
- `internal/application/README.md` — `ExecutorService` в таблице состава, `ExecutorStore.List`/`EventJournal` в строке `ports.go`, абзац в «Назначение».

### Как проверялось

- `make verify` — чисто (fmt, golangci-lint `0 issues.`, vet, все тесты, markdownlint `0 issues`, docs-check 1334 ссылки без ошибок).
- `go test ./internal/application/... -cover` — 83.7% (пакет), по новому файлу: `Register` 75.0%, `Activate` 88.9%, `publish` 100% — непокрытыми остались только пути ошибки `Save`, которые in-memory-фейк не может вызвать (так же, как у остальных сервисов слоя).
- **Интеграционный тест на реальном PostgreSQL** (`docker compose up -d postgres`, `postgres:16-alpine`): `TestExecutorStore_SaveThenGet`, `TestExecutorStore_Get_NotFound`, `TestExecutorStore_List_ReturnsCreatedExecutorsInOrder` — все PASS. Живая проверка, а не только фейки.

### Обновлённая документация

- `internal/application/README.md`.

### Open Questions

Нет.

### Рекомендации

- **Отклонение от плана (шаг 5, второй пункт), сознательное:** отдельный тест `List` для `ExecutorStore` в `internal/application/inmemory/stores_test.go` не добавлен. Существующий `TestStore_List_ReturnsAllOrderedByID` уже покрывает generic `Store[T].List` (через `ProjectStore`), а `executor_test.go` покрывает его же через `ExecutorStore` — третий тест того же generic-кода проверял бы только подстановку типового параметра, которую и так гарантирует компилятор. Зафиксировано здесь, чтобы отклонение было видно ревьюеру, а не осталось незамеченным.
- `postgres.ExecutorStore.List` сейчас возвращает всех исполнителей без фильтра. `apps/orchestrator` (TASK-082) будет фильтровать по роли и состоянию в памяти — на масштабе одной доверенной установки (единицы исполнителей) это дешевле, чем параметризованный запрос. Если исполнителей станут десятки — фильтр по роли/состоянию в SQL, отдельной задачей по реальной потребности.
