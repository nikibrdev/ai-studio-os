# TASK-066: Спецификации docs/api/*.md (Documentation First)

## Тип

docs

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Описать весь REST-контракт `apps/api` в `docs/api/` до реализации хендлеров — **Documentation First**, тот же принцип, что Domain Specifications First (EPIC-003), приведённый к масштабу этого слоя (шаблон [API.md](../../.claude/templates/API.md) уже существует и предполагает именно это: «Заполняется до реализации»).

## Контекст

Операции берутся один в один из уже существующего Application Layer (EPIC-004) плюс `ProjectService` (TASK-064) — эта задача не придумывает новые возможности, только описывает HTTP-форму уже принятых use-case-методов: `TaskPlanningService.CreateTask/PlanTask`, `WorkService.StartTask`, `ResultService.RecordDraftArtifact/UpdateArtifactDraft/PublishArtifact/SucceedExecution/FailExecution`, `CompletionService.RequestReview/CompleteReview/CompleteTesting`, `TaskProjection.Get`, `ProjectService.CreateProject/ConnectRepository/Activate` (TASK-064 — `ConnectRepository` обязателен перед `Activate`, иначе guard домена «≥1 Repository» всегда отказывает).

## Scope

### Входит

- `docs/api/projects.md` — `POST /projects`, `POST /projects/{id}/repositories` (`ConnectRepository` — обязателен перед `activate`: guard домена «≥1 Repository», TASK-064), `POST /projects/{id}/activate`.
- `docs/api/tasks.md` — `POST /tasks`, `POST /tasks/{id}/plan`, `GET /tasks/{id}` (через `TaskProjection`), `POST /tasks/{id}/start`, `POST /tasks/{id}/request-review`, `POST /tasks/{id}/complete-review`, `POST /tasks/{id}/complete-testing`.
- `docs/api/artifacts.md` — `POST /artifacts` (черновик), `PATCH /artifacts/{id}`, `POST /artifacts/{id}/publish`.
- `docs/api/executions.md` — `POST /executions/{id}/succeed`, `POST /executions/{id}/fail`.
- Каждая операция по шаблону API.md: назначение, запрос (метод/путь/тело с типами), ответ, ошибки (коды и условия), события (если публикуются).
- `docs/api/README.md` — обновить (сейчас заглушка), оглавление по новым файлам.

### Не входит

- Аутентификация — не описывается (ADR-012, Вариант 1: её нет в этой версии).
- Реализация хендлеров (TASK-067…069) — используют уже написанные здесь спецификации как источник требований, не наоборот.
- OpenAPI-документ — вне scope эпика (см. EPIC-008 «Не входит»).

## Критерии приёмки

- [ ] Каждая HTTP-операция, которую будет реализовывать TASK-067…069, описана в `docs/api/` до начала их реализации.
- [ ] Коды ошибок для каждой операции перечислены исходя из реальных sentinel-ошибок Application/Domain Layer (`ErrProjectNotActive` и т.п.), не придуманы заново.
- [ ] `make verify` (markdownlint, docs-check) — чисто.

## Затрагиваемые модули и документы

- `docs/api/README.md`, `docs/api/projects.md`, `docs/api/tasks.md`, `docs/api/artifacts.md`, `docs/api/executions.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-064 (ProjectService, для раздела Projects)

## План реализации

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

## Отчёт о выполнении
