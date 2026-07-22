# TASK-064: ProjectService (CreateProject/Activate)

## Тип

feature

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Добавить в `internal/application` минимальный use-case-сервис для создания и активации Project — сейчас Application Layer (EPIC-004) не содержит такого сценария (все четыре существующих сервиса действуют уже внутри Active-проекта); без него `apps/api` не сможет создавать проекты.

## Контекст

По образцу уже существующих сервисов EPIC-004 (`TaskPlanningService`, `internal/application/task_planning.go`): узкий порт хранения (`ProjectStore`, уже объявлен в `internal/application/ports.go`), события — через `platform.EventBus` и `Envelope`. Доменная модель уже готова: `internal/domain/project` (`project.New`, `project.Activate`, guard «≥1 Repository»). Решение открыть эту задачу — раздел «Контекст» [EPIC-008](../../docs/roadmap/EPIC-008-api-layer.md).

## Scope

### Входит

- `internal/application/project.go` — `ProjectService{Projects ProjectStore, Events platform.EventBus}`.
- `CreateProjectParams` + `CreateProject(ctx, params) (*project.Project, error)` — оборачивает `project.New`, сохраняет, публикует событие создания.
- `ConnectRepository(ctx, projectID, repo, actor string) error` — оборачивает `project.ConnectRepository`; **добавлено при реализации, не было в исходном описании задачи**: `Activate` требует ≥1 подключённого репозитория (guard в домене), а `CreateProject` его не подключает — без этого метода `Activate` не мог бы успешно сработать ни разу.
- `Activate(ctx, projectID, actor string) error` — оборачивает `project.Activate` (guard «≥1 Repository» уже в домене), сохраняет, публикует событие активации.
- Юнит-тесты на `inmemory`-фейке `ProjectStore`/`EventBus` (тот же паттерн, что тесты остальных сервисов EPIC-004).

### Не входит

- HTTP-хендлеры (TASK-068).
- Изменение доменного пакета `internal/domain/project` — используется как есть.

## Критерии приёмки

- [x] `ProjectService.CreateProject`/`Activate` реализованы, стиль идентичен остальным сервисам `internal/application` (узкий порт, `Envelope`, sentinel-ошибки).
- [x] Guard «≥1 Repository» перед активацией — проверен тестом (ошибка домена пробрасывается, а не дублируется в Application Layer).
- [x] Юнит-тесты покрывают успешный путь и минимум по одному отказному сценарию на метод.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/project.go`, `internal/application/project_test.go`, `internal/application/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — использует уже принятый `internal/domain/project` и `internal/application/ports.go`

## План реализации

1. Прочитать `internal/domain/project` (`New`/`ConnectRepository`/`Activate`, guard «≥1 Repository», события `Created`/`RepositoryConnected`/`Activated`) и `internal/application/ports.go` (`ProjectStore` уже объявлен) — убедиться, что домен не меняется, только оборачивается.
2. `internal/application/project.go` — `ProjectService{Projects, Events}` по стилю `task_planning.go`: `CreateProject`, `ConnectRepository` (обнаружено как необходимое дополнение — без него `Activate` не может быть вызван успешно, поскольку `CreateProject` не подключает репозиторий), `Activate`; приватный `publish`-хелпер (`source="project"`), как в остальных сервисах.
3. `internal/application/project_test.go` (пакет `application_test`, по образцу `task_planning_test.go`): успешные пути всех трёх методов, guard-ошибка `Activate` без подключённого репозитория, ошибка `ErrAlreadyActive`, `ErrNotFound` на неизвестном проекте, no-op повторного `ConnectRepository` без лишнего события.
4. `internal/application/README.md` — строка в таблице состава, абзац в «Назначение» о точечном дополнении из EPIC-008.
5. `make verify`, проверка покрытия (`go test ./internal/application/... -cover`).

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь (первая — остальные хендлеры зависят от неё).

2026-07-22 — Developer — задача взята в работу, реализована и проверена (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-064 — ProjectService (CreateProject/Activate) в `internal/application`.

### Что сделано

- `internal/application/project.go` — `ProjectService` реализует `CreateProject`, `ConnectRepository`, `Activate` поверх уже принятого домена `internal/domain/project`, без изменения домена. Стиль идентичен остальным сервисам EPIC-004: узкий порт `ProjectStore`, события через `Envelope`/`platform.EventBus`, приватный `publish`-хелпер.
- `ConnectRepository` добавлен сверх изначального описания задачи (только CreateProject/Activate) — при реализации обнаружено, что `Activate` требует ≥1 подключённого репозитория (guard в домене, `ErrNoRepository`), а `CreateProject` репозиторий не подключает; без `ConnectRepository` `Activate` не мог бы успешно сработать ни разу, делая саму задачу бессмысленной.
- `ConnectRepository` не публикует событие повторно для уже подключённого репозитория — домен уже трактует это как no-op (`changed=false`), Application Layer это уважает, а не публикует факт, которого не произошло.
- 8 юнит-тестов: успешные пути всех трёх методов, guard `ErrNoRepository`, `ErrAlreadyActive`, `ErrMissingField`, `ErrNotFound` на неизвестном проекте, no-op повторного подключения репозитория без лишнего события.

### Изменённые файлы

- `internal/application/project.go` — реализация `ProjectService`.
- `internal/application/project_test.go` — юнит-тесты.
- `internal/application/README.md` — состав пакета, назначение.

### Как проверялось

- `go test ./internal/application/... -run "TestCreateProject|TestConnectRepository|TestActivate" -v` — все 8 тестов зелёные.
- `go test ./internal/application/... -cover` — 82.7% (было 83.1% до задачи; несущественное изменение, тот же порядок величины, что и остальные сервисы).
- `make verify` — чисто (fmt, lint, vet, тесты, markdownlint, docs-check).

### Обновлённая документация

- `internal/application/README.md`.

### Open Questions

Нет.

### Рекомендации

TASK-068 (хендлеры Projects/Tasks) должен вызвать и `CreateProject`, и `ConnectRepository`, и `Activate` — без первых двух шагов `Activate` через API всегда будет возвращать `ErrNoRepository`; спецификация `docs/api/projects.md` (TASK-066) должна явно описывать порядок вызовов (create → connect-repository → activate), а не только конечные состояния.
