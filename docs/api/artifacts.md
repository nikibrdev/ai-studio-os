# API: Artifacts

## Назначение

Запись, доработка и публикация Artifact — результата работы Execution (`ResultService`, `internal/application/result.go`).

## Содержание

### Общие сведения

- Версия: v1
- Аутентификация: не требуется ([ADR-012](../adr/ADR-012-identity-and-auth.md))
- Базовый путь: `/artifacts`

### Операции

#### Прочитать артефакт

**Назначение:** отдаёт артефакт целиком, включая содержимое (`ResultService.Artifact`). Введена в TASK-100 ради конкретной потребности: человек, принимающий приёмочное решение, должен прочитать отчёт QA, на который ссылается `qaReportId` в представлении задачи. Общего обзора артефактов (списки, фильтры по типу/автору, которые перечисляет спецификация домена) сознательно нет — они появятся при реальной потребности.

**Запрос:** `GET /artifacts/{id}`

**Ответ:** `200 OK`, модель **Artifact** (см. ниже). `payload` — строкой, не base64: артефакты, которые эта операция обслуживает, текстовые, и кодирование только затруднило бы чтение.

**Ошибки:** `404` — артефакт не найден.

**События:** нет (операция чтения).

#### Создать черновик артефакта

**Назначение:** создаёт Artifact в состоянии `draft` внутри границы Active-проекта и связывает его с производящим Execution с обеих сторон (`ResultService.RecordDraftArtifact`). Execution должен быть в состоянии `running`.

**Запрос:** `POST /artifacts`

```json
{
  "id": "string, обязателен",
  "projectId": "string, обязателен",
  "executionId": "string, обязателен",
  "type": "string, обязателен",
  "origin": "string, обязателен",
  "author": "string, обязателен",
  "payload": "string (base64), опционален"
}
```

**Ответ:** `201 Created`

```json
{ "id": "string", "projectId": "string", "type": "string", "origin": "string", "author": "string", "state": "draft" }
```

**Ошибки:** `404` — проект или execution не найдены; `409` — проект не Active (`application.ErrProjectNotActive`); `409` — execution не в состоянии running (`application.ErrExecutionNotRunning`); `400` — обязательные поля пусты (`artifact.ErrMissingField`).

**События:** `ArtifactCreated`.

#### Обновить черновик

**Назначение:** обновляет payload и/или автора черновика (`ResultService.UpdateArtifactDraft`) — доступно только пока Artifact в `draft`.

**Запрос:** `PATCH /artifacts/{id}`

```json
{ "payload": "string (base64), опционален", "author": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — артефакт не найден; `409` — артефакт уже опубликован (`artifact.ErrPublished`) или архивирован (`artifact.ErrArchived`).

**События:** нет (только `Publish` публикует событие).

#### Опубликовать артефакт

**Назначение:** переводит Artifact `draft` → `published` (`ResultService.PublishArtifact`) — требует непустой payload.

**Запрос:** `POST /artifacts/{id}/publish`

```json
{ "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — артефакт не найден; `400` — payload пуст (`artifact.ErrPayloadRequired`); `409` — уже опубликован/архивирован.

**События:** `ArtifactPublished`.

### Модели данных

**Artifact** (представление в ответах): `id`, `projectId`, `type`, `origin`, `author`, `state` (`draft` | `published` | `archived`), а в ответе `GET /artifacts/{id}` дополнительно `payload` (строкой) и `createdAt`. В ответах на команды создания/изменения `payload` не возвращается — вызывающий его только что передал.

## Статус

Актуален

## Последнее обновление

2026-07-27
