# API: Tasks

## Назначение

Постановка, планирование, запуск, ревью и завершение задачи — golden path целиком со стороны Task (`TaskPlanningService`, `WorkService`, `CompletionService`, `TaskProjection`, `internal/application`).

## Содержание

### Общие сведения

- Версия: v1
- Аутентификация: не требуется ([ADR-012](../adr/ADR-012-identity-and-auth.md))
- Базовый путь: `/tasks`

### Операции

#### Создать задачу

**Назначение:** регистрирует Task в состоянии `backlog` внутри границы Active-проекта (`TaskPlanningService.CreateTask`).

**Запрос:** `POST /tasks`

```json
{
  "projectId": "string, обязателен",
  "epicId": "string, опционален",
  "title": "string, обязателен",
  "type": "string, обязателен",
  "scope": "string, опционален",
  "acceptanceCriteria": ["string", "..."],
  "actor": "string, опционален"
}
```

`id` в теле запроса не указывается — платформа сама генерирует публичный `TASK-NNN` (ADR-011, TASK-065).

**Ответ:** `201 Created`

```json
{ "id": "TASK-NNN", "projectId": "string", "epicId": "string", "title": "string", "type": "string", "scope": "string", "acceptanceCriteria": ["string"], "state": "backlog" }
```

**Ошибки:** `404` — проект не найден; `409` — проект не Active (`application.ErrProjectNotActive`); `400` — `title`/`type` пусты (`task.ErrMissingField`).

**События:** `TaskCreated`.

#### Спланировать задачу

**Назначение:** переводит Task `backlog` → `ready` (Definition of Ready выполнено), решение принимает исключительно `workflow.Rules` (`TaskPlanningService.PlanTask`).

**Запрос:** `POST /tasks/{id}/plan`

```json
{ "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — задача не найдена; `409` — переход недопустим из текущего состояния (`workflow.ErrTransitionNotAllowed`).

**События:** `TaskPlanned`.

#### Получить состояние задачи

**Назначение:** читает текущее состояние задачи из `TaskProjection` — read-модель, построенная только из событий (ADR-014), не из `TaskStore` напрямую.

**Запрос:** `GET /tasks/{id}`

**Ответ:** `200 OK`

```json
{ "id": "string", "projectId": "string", "state": "string", "updatedAt": "RFC3339" }
```

**Ошибки:** `404` — проекция не видела ни одного события по этому ID (`TaskProjection.Get` вернул `ok=false`).

**Ограничение:** возвращает ровно одну запись по ID — списковых операций (все задачи проекта и т.п.) нет в этой версии (см. EPIC-008 «Риски»).

#### Запустить работу

**Назначение:** переводит Task `ready` → `in_progress` и порождает Execution для указанного Executor (`WorkService.StartTask`). Выбор исполнителя — забота вызывающего (ADR-007, Decision Required, не входит в этот эпик).

**Запрос:** `POST /tasks/{id}/start`

```json
{ "executorId": "string, обязателен", "actor": "string, опционален" }
```

**Ответ:** `201 Created`

```json
{ "executionId": "string", "taskId": "string", "executorId": "string", "state": "running" }
```

**Ошибки:** `404` — задача или исполнитель не найдены; `409` — исполнитель недоступен или не имеет роли Developer (`application.ErrExecutorNotAssignable`); `409` — переход недопустим (`workflow.ErrTransitionNotAllowed`).

**События:** `TaskStarted`, `ExecutionQueued`, `ExecutionStarted`.

#### Запросить ревью

**Назначение:** переводит Task `in_progress` → `review` (`CompletionService.RequestReview`).

**Запрос:** `POST /tasks/{id}/request-review`

```json
{ "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — задача не найдена; `409` — переход недопустим.

**События:** `ReviewRequested`.

#### Завершить ревью

**Назначение:** переводит Task из `review` в `testing` (одобрено) или обратно в `in_progress` (запрошены изменения) — `CompletionService.CompleteReview`.

**Запрос:** `POST /tasks/{id}/complete-review`

```json
{ "approved": true, "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — задача не найдена; `409` — переход недопустим.

**События:** `ReviewCompleted`.

#### Завершить тестирование

**Назначение:** конечный шаг golden path (`CompletionService.CompleteTesting`, [ADR-008](../adr/ADR-008-git-policies.md)). При отказе — `testing` → `in_progress`. При успехе — merge пул-реквеста **до** перевода в `done`: если merge отказывает, задача остаётся в `testing`, `TaskCompleted` не публикуется.

**Запрос:** `POST /tasks/{id}/complete-testing`

```json
{
  "passed": true,
  "repository": "string, обязателен если passed=true",
  "pullRequestId": "string, обязателен если passed=true",
  "actor": "string, опционален"
}
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — задача не найдена; `409` — переход недопустим; `500` — отказ merge (`RepositoryProvider.MergePullRequest`, задача остаётся в `testing`).

**События:** при отказе — `TestsFailed`; при успехе — `TestsPassed`, затем `MergeCompleted`, затем `TaskCompleted` (в этом порядке, ADR-008).

### Модели данных

**Task** (представление в ответах): `id`, `projectId`, `epicId`, `title`, `type`, `scope`, `acceptanceCriteria` (string[]), `state`.

**TaskView** (ответ `GET /tasks/{id}`, из `TaskProjection`): `id`, `projectId`, `state`, `updatedAt`.

## Статус

Актуален

## Последнее обновление

2026-07-22
