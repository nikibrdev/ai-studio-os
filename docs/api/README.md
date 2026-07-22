# Документация API

## Назначение

Спецификации REST API платформы AI Studio OS (`apps/api`, [ADR-003](../adr/ADR-003-api-protocol.md), [EPIC-008](../roadmap/EPIC-008-api-layer.md)) — **Documentation First**: каждая операция описана здесь до реализации соответствующего HTTP-хендлера (тот же принцип, что Domain Specifications First, EPIC-003, приведённый к масштабу этого слоя).

## Содержание

### Общие сведения

- **Версия:** v1 (единственная, без версионирования пути — вводится при появлении реальной необходимости).
- **Протокол:** REST поверх HTTP, JSON-тела запросов/ответов ([ADR-003](../adr/ADR-003-api-protocol.md)).
- **Аутентификация:** не требуется — доверенная однопользовательская установка ([ADR-012](../adr/ADR-012-identity-and-auth.md), Вариант 1). Пересмотр при появлении внешнего потребителя.
- **`apps/api` не содержит бизнес-логики** ([module-boundaries.md](../architecture/module-boundaries.md)) — каждый хендлер вызывает ровно один use-case-метод `internal/application` и отображает его результат/ошибку в HTTP-ответ.

### Отображение ошибок в HTTP-коды

Единая конвенция для всех операций ниже (реализуется TASK-067 одной функцией, не дублируется по хендлерам):

| Класс ошибки | Код | Примеры sentinel-ошибок |
| --- | --- | --- |
| Не найдено | 404 | `application.ErrNotFound` |
| Некорректные данные запроса | 400 | `project.ErrMissingField`, `task.ErrMissingField`, `artifact.ErrMissingField`, `execution.ErrMissingField`, `artifact.ErrPayloadRequired` |
| Конфликт состояния (guard/переход недопустим) | 409 | `project.ErrArchived/ErrAlreadyActive/ErrNoRepository`, `application.ErrProjectNotActive/ErrExecutorNotAssignable/ErrExecutionNotRunning`, `workflow.ErrTransitionNotAllowed`, `artifact.ErrArchived/ErrPublished`, `execution.ErrNotQueued/ErrNotRunning/ErrTerminal` |
| Прочее (инфраструктура, сеть) | 500 | обёрнутые ошибки `internal/infrastructure` (например, отказ `RepositoryProvider.MergePullRequest`) |

Успешное создание — 201 с телом созданного ресурса; успешное действие без возвращаемого ресурса — 204 без тела.

### Спецификации по ресурсам

| Ресурс | Файл | Use-case-сервис |
| --- | --- | --- |
| Projects | [projects.md](projects.md) | `ProjectService` |
| Tasks | [tasks.md](tasks.md) | `TaskPlanningService`, `WorkService`, `CompletionService`, `TaskProjection` |
| Artifacts | [artifacts.md](artifacts.md) | `ResultService` |
| Executions | [executions.md](executions.md) | `ResultService` |

## Статус

Актуален

## Последнее обновление

2026-07-22
