# Слой: internal/application

## Назначение

Application Layer (v0.4, [EPIC-004](../../docs/roadmap/EPIC-004-application-layer.md)): use-case'ы поверх завершённого Domain Layer, не завязанные на конкретную инфраструктуру. Каждый use-case зависит от узкого порта (интерфейса), а не от технологии хранения — реализации портов появляются в EPIC-005 (v0.5). Дополнен `ProjectService` в [EPIC-008](../../docs/roadmap/EPIC-008-api-layer.md) (v0.9, TASK-064) — точечное добавление, обоснованное тем, что без него `apps/api` не может создавать проекты. Дополнен списковыми операциями (`ProjectStore.List`/`ProjectService.ListProjects`/`TaskProjection.ListByProject`) в [EPIC-009](../../docs/roadmap/EPIC-009-dashboard.md) (v0.8, TASK-072) — без них `apps/dashboard` не может показать даже список проектов.

## Содержание

### Состав

| Файл/пакет | Ответственность |
| --- | --- |
| `ports.go` | Пять узких портов хранения агрегатов: `ProjectStore` (Get/Save/`List` — TASK-072, EPIC-009), `TaskStore` (`Get(ctx, projectID, id)` — BUGFIX-003), `ExecutorStore`, `ExecutionStore`, `ArtifactStore` (Get/Save); `ErrNotFound`; `TaskIDGenerator` (TASK-065, EPIC-008) |
| `event.go` | `Envelope` — оборачивает данные доменных событий в контракт `platform.Event` (ADR-002) перед публикацией |
| `inmemory/` | Детерминированные фейки портов, `EventBus` и `RepositoryProvider` для тестов этого эпика — не инфраструктурный адаптер |
| `project.go` | `ProjectService` (TASK-064, EPIC-008) — жизненный цикл Project: `CreateProject`, `ConnectRepository`, `Activate` (guard «≥1 Repository» — целиком в домене); `ListProjects` (TASK-072, EPIC-009) — тонкая обёртка над `ProjectStore.List` |
| `task_planning.go` | `TaskPlanningService` (TASK-041) — «Постановка задачи»: `CreateTask` (в границе Active-проекта, с scope/AC), `PlanTask` (Backlog → Ready через `workflow.Rules`); опциональный порт `IDs TaskIDGenerator` (TASK-065, EPIC-008) — генерирует `TASK-NNN` (ADR-011), если `CreateTaskParams.ID` не задан вызывающим |
| `work.go` | `WorkService` (TASK-042) — «Запуск работы»: `StartTask` (Ready → In Progress, guard доступности Executor, порождение и немедленный Accept Execution); `StartTaskParams.ProjectID` — BUGFIX-003 |
| `result.go` | `ResultService` (TASK-043) — «Производство результата»: `RecordDraftArtifact`/`UpdateArtifactDraft`/`PublishArtifact`, `SucceedExecution`/`FailExecution` (оба принимают `projectID` — BUGFIX-003) |
| `completion.go` | `CompletionService` (TASK-044) — «Завершение задачи»: `RequestReview`, `CompleteReview`, `CompleteTesting` (все принимают/несут `projectID` — BUGFIX-003) — реализует ADR-008 (merge — код-гейт перед Done, порядок TestsPassed → MergeCompleted → TaskCompleted) |
| `projection.go` | `TaskProjection` (TASK-045) — read-модель статуса задачи, построенная только из событий (ADR-014); ключ — пара (ProjectID, ID), не голый ID (BUGFIX-003); `Rebuild` доказывает пересобираемость с нуля из журнала; `ListByProject` (TASK-072, EPIC-009) — линейный проход `views` с фильтром по `ProjectID`, без перестройки ключа карты |
| `id.go` | `NewID()` — общий генератор идентификаторов (`crypto/rand`, без внешней UUID-зависимости) для сущностей, порождаемых как побочный эффект use-case (Execution, здесь же переиспользуется), а не именованных явной командой |
| `e2e_test.go` | Сквозной тест golden path целиком (`docs/architecture/golden-path.md`) через все четыре сервиса, включая ветки «changes requested» и «tests failed» — состояние проверяется только через `TaskProjection` |

Декомпозиция EPIC-004 завершена всеми шестью задачами (TASK-040…045).

### Envelope.WithData — данные, специфичные для события

`platform.Event` (ADR-002) несёт только общие поля. Когда одному имени события соответствуют разные исходы (`ReviewCompleted` → Testing или обратно в In Progress), `CompletionService` прикрепляет исход через `Envelope.WithData(map[string]string{"to": ...})` — метод сверх контракта `platform.Event`, не изменение самого контракта; читается обратно только через type assertion на конкретный тип `Envelope` (см. `projection.go`, `targetState`).

### ADR-008 в коде

`CompletionService.CompleteTesting` кодирует решение ADR-008 не только комментарием: при успехе тестов сначала публикуется `TestsPassed`, затем вызывается `RepositoryProvider.MergePullRequest`, затем `MergeCompleted` — и только после успешного merge задача переходит в Done с `TaskCompleted`. Если merge вернёт ошибку, задача остаётся в Testing и `TaskCompleted` не публикуется — проверено тестом `TestCompleteTesting_MergeFailure_BlocksDone`.

### Известное ограничение: нет межагрегатной транзакции

`WorkService.StartTask` и `ResultService.RecordDraftArtifact` сохраняют несколько агрегатов последовательно, не атомарно: если второе сохранение откажет после того, как первое уже прошло и событие опубликовано, отката не происходит (проверено тестом `TestStartTask_PropagatesExecutionStoreFailure`). С in-memory фейками этого эпика это не проявляется (фейки не отказывают); при реализации PostgreSQL-адаптера (EPIC-005) потребуется либо единая транзакция на несколько агрегатов, либо saga/outbox — решение архитектора, не принимается здесь.

### BUGFIX-003 — TASK-NNN уникален только в рамках Project

Живая проверка EPIC-008/TASK-069 вскрыла реальный баг: `TaskStore.Get` принимал только `id`, но публичный `TASK-NNN` (ADR-011) уникален лишь в рамках Project — два разных проекта неизбежно получают одинаковый `TASK-001` (TASK-065 генерирует номер отдельно на проект), и без `projectID` в ключе поиска задача одного проекта могла быть перепутана или молча испорчена операциями над задачей другого. Исправлено по всему стеку: `TaskStore.Get(ctx, projectID, id)`; `TaskPlanningService.PlanTask`, `CompletionService.RequestReview`/`CompleteReview`, `WorkService.StartTaskParams`, `CompletionService.CompleteTestingParams` — все принимают/несут `projectID`; `ResultService.SucceedExecution`/`FailExecution` тоже принимают `projectID` явным параметром (Execution сам не хранит ссылку на Project — ADR-015, домен не меняли) вместо попытки вывести его из голого `TaskID`; `TaskProjection` — внутренняя карта ключится парой (ProjectID, ID), `Get(projectID, id)`. `apps/api` (TASK-068/069) вкладывает задаче-специфичные маршруты под `/projects/{projectId}/tasks/...`.

### Почему порты здесь, а не в internal/platform

`internal/platform` домен-независим ([ADR-015](../../docs/adr/ADR-015-internal-layering.md)); порты хранения оперируют конкретными доменными типами (`*task.Task` и т.д.) — размещение в Application Layer, рядом с использующими их use-case'ами, а не в платформенном слое. Подробности — [решение](../../engineering/decisions/2026-07-21-application-ports-placement.md).

### Зависимости

- Разрешено: stdlib, все пакеты `internal/domain/*`, `internal/platform` (контракты `EventBus`, `RepositoryProvider` и т.д. — use-case'ы работают против них, не против конкретных адаптеров).
- Запрещено: `internal/infrastructure`, `apps/`, конкретные технологии хранения/доставки.

### События

Use-case'ы оборачивают доменные события (`Created`, `Transitioned` и т.д. — значения из доменных пакетов) в `Envelope` и публикуют через порт `platform.EventBus`; канонические имена типов — `internal/domain/event`.

## Статус

Актуален

## Последнее обновление

2026-07-22
