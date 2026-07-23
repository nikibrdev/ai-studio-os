# API: Projects

## Назначение

Создание и активация Project — граница, внутри которой существуют Task и Artifact (`ProjectService`, `internal/application/project.go`, TASK-064).

## Содержание

### Общие сведения

- Версия: v1
- Аутентификация: не требуется ([ADR-012](../adr/ADR-012-identity-and-auth.md))
- Базовый путь: `/projects`

### Порядок вызовов

`Activate` требует хотя бы один подключённый репозиторий (guard домена, спецификация Project, Structural Invariant 1). Порядок обязателен: `POST /projects` → `POST /projects/{id}/repositories` (минимум один раз) → `POST /projects/{id}/activate`.

### Операции

#### Список проектов

**Назначение:** возвращает все проекты — единственный способ для `apps/dashboard` (EPIC-009) узнать, какие проекты вообще существуют (`ProjectService.ListProjects`, `application.ProjectStore.List`, TASK-072).

**Запрос:** `GET /projects` (тело не требуется).

**Ответ:** `200 OK`

```json
[
  { "id": "string", "name": "string", "state": "created|active|archived", "createdAt": "RFC3339" }
]
```

Упорядочен по `id`. Пустой список — `200` с `[]`, не ошибка и не `null`.

**Ошибки:** нет специфичных для этой операции (общая `500` при сбое хранилища).

**События:** нет (операция только читает).

#### Создать проект

**Назначение:** регистрирует Project в состоянии `created`.

**Запрос:** `POST /projects`

```json
{ "id": "string, обязателен", "name": "string, обязателен" }
```

**Ответ:** `201 Created`

```json
{ "id": "string", "name": "string", "state": "created", "createdAt": "RFC3339" }
```

**Ошибки:** `400` — `id` или `name` пусты (`project.ErrMissingField`).

**События:** `ProjectCreated`.

#### Подключить репозиторий

**Назначение:** привязывает репозиторий к проекту — обязательное условие для последующей активации. Повторное подключение уже привязанного репозитория — не ошибка (no-op).

**Запрос:** `POST /projects/{id}/repositories`

```json
{ "repository": "string, обязателен" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — проект не найден (`application.ErrNotFound`); `400` — `repository` пуст (`project.ErrMissingField`); `409` — проект архивирован (`project.ErrArchived`).

**События:** `RepositoryConnected` (не публикуется при повторном подключении того же репозитория).

#### Активировать проект

**Назначение:** переводит Project `created` → `active` (после этого доступно создание Task/Artifact внутри проекта).

**Запрос:** `POST /projects/{id}/activate` (тело не требуется).

**Ответ:** `204 No Content`.

**Ошибки:** `404` — проект не найден; `409` — нет ни одного подключённого репозитория (`project.ErrNoRepository`); `409` — уже активен (`project.ErrAlreadyActive`); `409` — архивирован (`project.ErrArchived`).

**События:** `ProjectActivated`.

### Модели данных

**Project** (представление в ответах): `id` (string), `name` (string), `state` (`created` | `active` | `archived`), `createdAt` (RFC3339).

## Статус

Актуален

## Последнее обновление

2026-07-23
