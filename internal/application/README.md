# Слой: internal/application

## Назначение

Application Layer (v0.4, [EPIC-004](../../docs/roadmap/EPIC-004-application-layer.md)): use-case'ы поверх завершённого Domain Layer, не завязанные на конкретную инфраструктуру. Каждый use-case зависит от узкого порта (интерфейса), а не от технологии хранения — реализации портов появляются в EPIC-005 (v0.5).

## Содержание

### Состав

| Файл/пакет | Ответственность |
| --- | --- |
| `ports.go` | Пять узких портов хранения агрегатов: `ProjectStore`, `TaskStore`, `ExecutorStore`, `ExecutionStore`, `ArtifactStore` (Get/Save); `ErrNotFound` |
| `event.go` | `Envelope` — оборачивает данные доменных событий в контракт `platform.Event` (ADR-002) перед публикацией |
| `inmemory/` | Детерминированные фейки портов и `EventBus` для тестов этого эпика — не инфраструктурный адаптер |
| `task_planning.go` | `TaskPlanningService` (TASK-041) — «Постановка задачи»: `CreateTask` (в границе Active-проекта, с scope/AC), `PlanTask` (Backlog → Ready через `workflow.Rules`) |
| `work.go` | `WorkService` (TASK-042) — «Запуск работы»: `StartTask` (Ready → In Progress, guard доступности Executor, порождение и немедленный Accept Execution) |
| `result.go` | `ResultService` (TASK-043) — «Производство результата»: `RecordDraftArtifact`/`UpdateArtifactDraft`/`PublishArtifact`, `SucceedExecution`/`FailExecution` |
| `id.go` | `NewID()` — общий генератор идентификаторов (`crypto/rand`, без внешней UUID-зависимости) для сущностей, порождаемых как побочный эффект use-case (Execution, здесь же переиспользуется), а не именованных явной командой |

Остальные use-case'ы (TASK-044…045: завершение задачи, проекция чтения) добавляются отдельными файлами по мере реализации.

### Известное ограничение: нет межагрегатной транзакции

`WorkService.StartTask` и `ResultService.RecordDraftArtifact` сохраняют несколько агрегатов последовательно, не атомарно: если второе сохранение откажет после того, как первое уже прошло и событие опубликовано, отката не происходит (проверено тестом `TestStartTask_PropagatesExecutionStoreFailure`). С in-memory фейками этого эпика это не проявляется (фейки не отказывают); при реализации PostgreSQL-адаптера (EPIC-005) потребуется либо единая транзакция на несколько агрегатов, либо saga/outbox — решение архитектора, не принимается здесь.

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

2026-07-21
