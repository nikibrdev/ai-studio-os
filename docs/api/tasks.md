# API: Tasks

## Назначение

Постановка, планирование, запуск, ревью и завершение задачи — golden path целиком со стороны Task (`TaskPlanningService`, `WorkService`, `CompletionService`, `TaskProjection`, `internal/application`).

## Содержание

### Общие сведения

- Версия: v1
- Аутентификация: не требуется ([ADR-012](../adr/ADR-012-identity-and-auth.md))
- Базовый путь: `/projects/{projectId}/tasks` — все операции вложены под проект (**BUGFIX-003**): публичный `TASK-NNN` уникален только в рамках Project (ADR-011), поэтому голого `/tasks/{id}` недостаточно, чтобы однозначно определить задачу — тот же принцип, что ADR-011 уже предвидел («любой межпроектный контекст обязан использовать полностью квалифицированную пару (Project, ID)»).

### Операции

#### Создать задачу

**Назначение:** регистрирует Task в состоянии `backlog` внутри границы Active-проекта (`TaskPlanningService.CreateTask`).

**Запрос:** `POST /projects/{projectId}/tasks`

```json
{
  "epicId": "string, опционален",
  "title": "string, обязателен",
  "type": "string, обязателен",
  "scope": "string, опционален",
  "acceptanceCriteria": ["string", "..."],
  "actor": "string, опционален"
}
```

`id` и `projectId` в теле запроса не указываются — `projectId` уже есть в пути, `id` платформа генерирует сама (публичный `TASK-NNN`, ADR-011, TASK-065).

**Ответ:** `201 Created`

```json
{ "id": "TASK-NNN", "projectId": "string", "epicId": "string", "title": "string", "type": "string", "scope": "string", "acceptanceCriteria": ["string"], "state": "backlog" }
```

**Ошибки:** `404` — проект не найден; `409` — проект не Active (`application.ErrProjectNotActive`); `400` — `title`/`type` пусты (`task.ErrMissingField`).

**События:** `TaskCreated`.

#### Спланировать задачу

**Назначение:** переводит Task `backlog` → `ready` (Definition of Ready выполнено), решение принимает исключительно `workflow.Rules` (`TaskPlanningService.PlanTask`).

**Запрос:** `POST /projects/{projectId}/tasks/{id}/plan`

```json
{ "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — задача не найдена; `409` — переход недопустим из текущего состояния (`workflow.ErrTransitionNotAllowed`).

**События:** `TaskPlanned`.

#### Получить состояние задачи

**Назначение:** читает текущее состояние задачи из `TaskProjection` — read-модель, построенная только из событий (ADR-014), не из `TaskStore` напрямую.

**Запрос:** `GET /projects/{projectId}/tasks/{id}`

**Ответ:** `200 OK`

```json
{ "id": "string", "projectId": "string", "state": "string", "updatedAt": "RFC3339" }
```

**Ошибки:** `404` — проекция не видела ни одного события по этому (projectId, id) (`TaskProjection.Get` вернул `ok=false`).

**Историческое ограничение (снято в TASK-072, EPIC-009):** до этого возвращала ровно одну запись по (projectId, id) — списковых операций не было (см. EPIC-008 «Риски»). Список всех задач проекта — следующая операция.

#### Список задач проекта

**Назначение:** возвращает все задачи проекта, о которых знает `TaskProjection` — единственный способ для `apps/dashboard` (EPIC-009) показать список задач (`TaskProjection.ListByProject`, TASK-072).

**Запрос:** `GET /projects/{projectId}/tasks` (тело не требуется).

**Ответ:** `200 OK`

```json
[
  { "id": "string", "projectId": "string", "state": "string", "updatedAt": "RFC3339" }
]
```

Упорядочен по `id`. Задачи других проектов никогда не попадают в список (изоляция по `projectId` — то же требование, что и BUGFIX-003 для хранения). Пустой список — `200` с `[]`, не ошибка и не `null`.

**Ошибки:** нет специфичных для этой операции.

**События:** нет (операция только читает).

#### Запустить работу

**Назначение:** переводит Task `ready` → `in-progress` и порождает Execution для указанного Executor (`WorkService.StartTask`). Выбор исполнителя — забота вызывающего (ADR-007, Decision Required, не входит в этот эпик).

**Запрос:** `POST /projects/{projectId}/tasks/{id}/start`

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

**Назначение:** переводит Task `in-progress` → `review` (`CompletionService.RequestReview`).

**Запрос:** `POST /projects/{projectId}/tasks/{id}/request-review`

```json
{ "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — задача не найдена; `409` — переход недопустим.

**События:** `ReviewRequested`.

#### Завершить ревью

**Назначение:** переводит Task из `review` в `testing` (одобрено) или обратно в `in-progress` (запрошены изменения) — `CompletionService.CompleteReview`.

**Запрос:** `POST /projects/{projectId}/tasks/{id}/complete-review`

```json
{ "approved": true, "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — задача не найдена; `409` — переход недопустим.

**События:** `ReviewCompleted`.

#### Завершить тестирование

**Назначение:** конечный шаг golden path (`CompletionService.CompleteTesting`, [ADR-008](../adr/ADR-008-git-policies.md)). При отказе — `testing` → `in-progress`. При успехе — merge пул-реквеста **до** перевода в `done`: если merge отказывает, задача остаётся в `testing`, `TaskCompleted` не публикуется.

**Запрос:** `POST /projects/{projectId}/tasks/{id}/complete-testing`

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

**TaskView** (ответ `GET /projects/{projectId}/tasks/{id}`, из `TaskProjection`): `id`, `projectId`, `state`, `updatedAt`.

## Статус

Актуален

## Последнее обновление

2026-07-23
