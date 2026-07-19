# Review: EPIC-002 Foundation (контракты ядра)

## Назначение

Запись code review результатов EPIC-002 (TASK-001…TASK-011), проведённого архитектором проекта 2026-07-19, и принятых по нему исправлений.

## Содержание

### Контекст

Ревьюились: Go-модуль, контракты `internal/core` и интерфейсные пакеты `internal/events`, `internal/workflow`, `internal/tasks`, `internal/project` (только интерфейсы, без логики).

### Замечания и решения

| № | Замечание | Решение | Статус |
|---|---|---|---|
| 1 | Контракт Agent (Assignment/Result с полями) предвосхищает непринятый ADR-005 | `Agent.Execute(ctx, Request) (Response, error)`; Request/Response — абстрактные типы (`any`) до утверждения ADR-005; Assignment, Result, ExecutionStatus, ID/Provider/Roles удалены из контракта | Исправлено |
| 2 | Имя компонента «Engine» преждевременно: устройство пути записи (возможен Command → Event → Projection) не решено | Интерфейсы переименованы: Engine → `Commands`, Reader → `Queries`; в документации пакета зафиксировано, что механизм записи контрактами не фиксируется | Исправлено |
| 3 | Нет доменного слоя: `internal/tasks` создан раньше `internal/domain`; всё должно строиться вокруг домена | Введена структура `internal/domain/` (task, project, event, workflow), `internal/application/`, `internal/infrastructure/`; старые пакеты `internal/{events,workflow,tasks,project}` удалены | Исправлено |
| 4 | Нужна единая команда проверки | Добавлена цель `make verify`: gofumpt → golangci-lint → go vet → go test → markdownlint → проверка документации/Mermaid (`scripts/verify-docs.sh`) | Исправлено |
| 5 | Усилить Documentation First: README в каждом модуле (назначение, зависимости, события, ответственность) | README добавлены во все модули и слои (`internal/core`, `internal/domain` + 4 модуля, `internal/application`, `internal/infrastructure`); правило закреплено в docs/development/documentation.md | Исправлено |

### Процессные выводы (зафиксированы отдельно)

- Переход на процесс «план → утверждение → код → ревью → исправления → merge» — [decisions/2026-07-19-plan-first-process.md](../decisions/2026-07-19-plan-first-process.md).
- Разработка приостановлена до создания инженерной платформы (EPIC-002.5).

### Открытые вопросы ревью

- Размещение кросс-доменных контрактов: оставлен `internal/core` (EventBus, Agent, Tool, Workflow, RepositoryProvider, MemoryProvider, словари Role/TaskState) — ревьюер структуру слоёв задал, место портов явно не указал; требуется подтверждение или указание перенести (например, в domain/application).
- **Дополнение 2026-07-19: вопрос решён архитектором** ([ADR-015](../../docs/adr/ADR-015-internal-layering.md)): `internal/core` упразднён; платформенные абстракции (EventBus, Agent, Tool, MemoryProvider, RepositoryProvider) → `internal/platform`; язык домена (Role, TaskState) → `internal/domain/shared`; контракт Workflow → тип `Rules` в `internal/domain/workflow`. Перенос выполнен, проверки пройдены.

## Статус

Актуален

## Последнее обновление

2026-07-19
