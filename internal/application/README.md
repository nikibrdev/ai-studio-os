# Слой: internal/application

## Назначение

Application Layer (v0.4, [EPIC-004](../../docs/roadmap/EPIC-004-application-layer.md)): use-case'ы поверх завершённого Domain Layer, не завязанные на конкретную инфраструктуру. Каждый use-case зависит от узкого порта (интерфейса), а не от технологии хранения — реализации портов появляются в EPIC-005 (v0.5). Дополнен `ProjectService` в [EPIC-008](../../docs/roadmap/EPIC-008-api-layer.md) (v0.9, TASK-064) — точечное добавление, обоснованное тем, что без него `apps/api` не может создавать проекты. Дополнен списковыми операциями (`ProjectStore.List`/`ProjectService.ListProjects`/`TaskProjection.ListByProject`) в [EPIC-009](../../docs/roadmap/EPIC-009-dashboard.md) (v0.8, TASK-072) — без них `apps/dashboard` не может показать даже список проектов. Дополнен `ExecutorService`, `ExecutorStore.List` и портом `EventJournal` в [EPIC-010](../../docs/roadmap/EPIC-010-orchestrator.md) (v1.0, TASK-079) — без них `apps/orchestrator` не может ни зарегистрировать исполнителя, ни найти его, ни узнать о событиях из отдельного процесса.

## Содержание

### Состав

| Файл/пакет | Ответственность |
| --- | --- |
| `ports.go` | Пять узких портов хранения агрегатов: `ProjectStore` (Get/Save/`List` — TASK-072, EPIC-009), `TaskStore` (`Get(ctx, projectID, id)` — BUGFIX-003), `ExecutorStore` (Get/Save/`List` — TASK-079, EPIC-010), `ExecutionStore`, `ArtifactStore` (Get/Save); `ErrNotFound`; `TaskIDGenerator` (TASK-065, EPIC-008); `EventJournal`/`JournalEntry` (TASK-079/080, EPIC-010) — чтение уже опубликованных событий по курсору, единственный доступный отдельному процессу способ узнать о них (продуктивная `eventbus.Bus` доставляет только внутри своего процесса, ADR-002); курсор — монотонная последовательность вставки (`seq`), не время: `OccurredAt` штампуется доменом **до** записи строки, поэтому курсор по времени бесшумно перешагивает строки, ставшие видимыми «в прошлом» (TASK-080, миграция 0007) |
| `event.go` | `Envelope` — оборачивает данные доменных событий в контракт `platform.Event` (ADR-002) перед публикацией |
| `inmemory/` | Детерминированные фейки портов, `EventBus` и `RepositoryProvider` для тестов этого эпика — не инфраструктурный адаптер |
| `project.go` | `ProjectService` (TASK-064, EPIC-008) — жизненный цикл Project: `CreateProject`, `ConnectRepository`, `Activate` (guard «≥1 Repository» — целиком в домене); `ListProjects` (TASK-072, EPIC-009) — тонкая обёртка над `ProjectStore.List` |
| `executor.go` | `ExecutorService` (TASK-079, EPIC-010) — жизненный цикл реестра Executor до Active: `Register`, `Activate`; тот же пробел, что `ProjectService` закрыл для Project — домен `executor` готов с EPIC-003, но пути к нему через Application Layer не было. События `ExecutorRegistered`/`ExecutorActivated` публикуются без `ProjectID`: Executor не принадлежит проекту (`docs/architecture/events.md` не перечисляет поле проекта для его событий) |
| `task_planning.go` | `TaskPlanningService` (TASK-041) — «Постановка задачи»: `CreateTask` (в границе Active-проекта, с scope/AC), `PlanTask` (Backlog → Ready через `workflow.Rules`); опциональный порт `IDs TaskIDGenerator` (TASK-065, EPIC-008) — генерирует `TASK-NNN` (ADR-011), если `CreateTaskParams.ID` не задан вызывающим; `CreateTask` прикрепляет title/type/scope/acceptanceCriteria к `TaskCreated` через `Envelope.WithData` (TASK-076, EPIC-009) — единственный способ донести их до `TaskProjection` |
| `task_planning.go` (продолжение) | `RefineTask` (TASK-096, EPIC-013) — уточнение scope/критериев задачи **в Backlog**: то, что готовит роль Project Manager перед приёмом Definition of Ready человеком. Не переход: задача остаётся в Backlog. Guard «только Backlog» целиком доменный (`ErrNotBacklog`), сервис его пробрасывает. Событие `TaskRefined` несёт **только фактически изменённые** поля — отсутствие ключа значит «не уточнялось», а не «очищено» (домен запрещает очистку, поэтому трактовка однозначна), иначе уточнение одного поля стирало бы соседнее. Пустое уточнение — ошибка, а не no-op: событие, ничего не меняющее, было бы ложным сигналом «PM что-то сделал» |
| `work.go` | `WorkService` (TASK-042) — «Запуск работы»: `StartTask` (Ready → In Progress, guard доступности Executor, порождение и немедленный Accept Execution); `StartTaskParams.ProjectID` — BUGFIX-003 |
| `result.go` | `ResultService` (TASK-043) — «Производство результата»: `RecordDraftArtifact`/`UpdateArtifactDraft`/`PublishArtifact`, `SucceedExecution`/`FailExecution` (оба принимают `projectID` — BUGFIX-003) |
| `completion.go` | `CompletionService` (TASK-044) — «Завершение задачи»: `RequestReview`, `CompleteReview`, `CompleteTesting` (все принимают/несут `projectID` — BUGFIX-003) — реализует ADR-008 (merge — код-гейт перед Done, порядок TestsPassed → MergeCompleted → TaskCompleted) |
| `projection.go` | `TaskProjection` (TASK-045) — read-модель статуса задачи, построенная только из событий (ADR-014); ключ — пара (ProjectID, ID), не голый ID (BUGFIX-003); `Rebuild` доказывает пересобираемость с нуля из журнала; `ListByProject` (TASK-072, EPIC-009) — линейный проход `views` с фильтром по `ProjectID`, без перестройки ключа карты; `TaskView.Title/Type/Scope/AcceptanceCriteria` (TASK-076, EPIC-009) — заполняются один раз из данных `TaskCreated`, не меняются последующими событиями (единственный путь чтения этих полей, ADR-014 — `apps/api` не читает `TaskStore` напрямую) |
| `decision.go` | `DecisionKind`/`AwaitingDecision` и `TaskProjection.ListAwaitingDecision` (TASK-092, EPIC-012) — read-модель контрольных точек человека. Соответствие «состояние → тип решения» задано **явным перечислением двух** состояний, отражающим нормативную таблицу [workflow.md](../../docs/architecture/workflow.md#контрольные-точки-человека), а не выведено из `workflow.Rules.NextRole`: точки заданы ADR-007 по имени, и вывод из «того, что пока не автоматизировано» молча опустошил бы список при автоматизации PM/QA. `DecisionFor` экспортирована, чтобы слой доставки размечал отдельную задачу, не дублируя правило в UI (запрещено `module-boundaries.md`) |
| `id.go` | `NewID()` — общий генератор идентификаторов (`crypto/rand`, без внешней UUID-зависимости) для сущностей, порождаемых как побочный эффект use-case (Execution, здесь же переиспользуется), а не именованных явной командой |
| `e2e_test.go` | Сквозной тест golden path целиком (`docs/architecture/golden-path.md`) через все четыре сервиса, включая ветки «changes requested» и «tests failed» — состояние проверяется только через `TaskProjection` |

Декомпозиция EPIC-004 завершена всеми шестью задачами (TASK-040…045).

### Envelope.WithData — данные, специфичные для события

`platform.Event` (ADR-002) несёт только общие поля. Когда одному имени события соответствуют разные исходы (`ReviewCompleted` → Testing или обратно в In Progress), `CompletionService` прикрепляет исход через `Envelope.WithData(map[string]string{"to": ...})` — метод сверх контракта `platform.Event`, не изменение самого контракта.

Читается обратно **структурно** — через приватный интерфейс `dataCarrier` (`Data() map[string]string`), а не приведением к конкретному типу `Envelope` ([BUGFIX-004](../../tasks/done/BUGFIX-004-projection-rebuild-loses-envelope-data.md)). Это важно именно для пересборки: события, восстановленные из журнала (`eventbus.ReadJournal`), имеют собственный тип пакета `eventbus`, а не `Envelope`, — приведение к конкретному типу молча теряло все данные `WithData` при replay (описательные поля оказывались пустыми, `ReviewCompleted` не двигал состояние), обессмысливая `Rebuild`. Пакет `eventbus` на своей стороне той же границы читает эти данные так же структурно. Тот же механизм переиспользован для описательных полей задачи: `TaskPlanningService.CreateTask` прикрепляет `title`/`type`/`scope`/`acceptanceCriteria` (последнее — JSON-строкой, `WithData` несёт только `map[string]string`) к `TaskCreated`, `TaskProjection.Handle` читает их обратно только для этого типа события (TASK-076, EPIC-009).

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

2026-07-23
