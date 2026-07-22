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

- [x] Каждая HTTP-операция, которую будет реализовывать TASK-067…069, описана в `docs/api/` до начала их реализации.
- [x] Коды ошибок для каждой операции перечислены исходя из реальных sentinel-ошибок Application/Domain Layer (`ErrProjectNotActive` и т.п.), не придуманы заново.
- [x] `make verify` (markdownlint, docs-check) — чисто.

## Затрагиваемые модули и документы

- `docs/api/README.md`, `docs/api/projects.md`, `docs/api/tasks.md`, `docs/api/artifacts.md`, `docs/api/executions.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-064 (ProjectService, для раздела Projects)

## План реализации

1. Перечитать сигнатуры всех методов `TaskPlanningService`/`WorkService`/`ResultService`/`CompletionService`/`ProjectService`/`TaskProjection.Get`, а также sentinel-ошибки соответствующих доменных пакетов (`project`, `task`, `artifact`, `execution`, `workflow`) — источник требований, ничего не придумывать.
2. `docs/api/README.md` — переписать заглушку: общие сведения, единая конвенция отображения ошибок в HTTP-коды (таблица, используется всеми последующими задачами), оглавление по ресурсам.
3. `docs/api/projects.md` — три операции (create/connect-repository/activate), явно зафиксирован обязательный порядок вызовов (обнаружено в TASK-064: без connect-repository activate всегда отказывает).
4. `docs/api/tasks.md` — семь операций (create/plan/get/start/request-review/complete-review/complete-testing), включая точный порядок событий ADR-008 для complete-testing.
5. `docs/api/artifacts.md` — три операции (create draft/update draft/publish).
6. `docs/api/executions.md` — две операции (succeed/fail), с явной пометкой, что создание Execution происходит не здесь.
7. `make verify` (markdownlint споткнулся об MD031 при первом проходе — блоки ```json внутри пунктов списка требуют пустых строк вокруг себя; исправлено переходом на плоский формат **Запрос:**/**Ответ:** вместо вложенных списков).

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

2026-07-22 — Developer — задача взята в работу, спецификации написаны и проверены (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-066 — спецификации `docs/api/*.md` (Documentation First).

### Что сделано

- `docs/api/README.md` — общие сведения (без auth — ADR-012; JSON; `apps/api` без бизнес-логики) и единая таблица отображения sentinel-ошибок в HTTP-коды (404/400/409/500), используемая одинаково всеми последующими операциями — не решается заново в каждом файле.
- `docs/api/projects.md`, `docs/api/tasks.md`, `docs/api/artifacts.md`, `docs/api/executions.md` — 15 операций суммарно, один в один с уже существующими методами Application Layer (TASK-064/065 + EPIC-004), включая точный порядок обязательных вызовов (`create` → `connect-repository` → `activate`) и точный порядок событий ADR-008 (`TestsPassed` → `MergeCompleted` → `TaskCompleted`).
- Формат кода в спецификациях: плоские **Запрос:**/**Ответ:** вместо вложенных Markdown-списков — markdownlint (MD031) требует пустых строк вокруг ```-блоков, что плохо сочетается со вложенностью внутри пункта списка; выбран более простой, одинаково читаемый формат.

### Изменённые файлы

- `docs/api/README.md`, `docs/api/projects.md`, `docs/api/tasks.md`, `docs/api/artifacts.md`, `docs/api/executions.md`.

### Как проверялось

- `make verify` — чисто (markdownlint, docs-check — 1272 ссылки проверены, 0 ошибок).
- Ручная сверка: каждая операция сопоставлена с реальной сигнатурой use-case-метода и реальными sentinel-ошибками (grep по `errors.New` в соответствующих доменных пакетах) — ни одна ошибка/операция не придумана.

### Обновлённая документация

Вся документация этой задачи и есть основной результат — см. «Изменённые файлы».

### Open Questions

Нет.

### Рекомендации

TASK-067 должен реализовать таблицу отображения ошибок из `docs/api/README.md` одной функцией (не по хендлеру) — этот выбор уже отражён в плане TASK-067.
