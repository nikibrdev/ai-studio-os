# API: Executions

## Назначение

Завершение Execution — успехом или отказом (`ResultService`, `internal/application/result.go`). Создание Execution происходит не здесь, а как побочный эффект `POST /tasks/{id}/start` ([tasks.md](tasks.md)) — Execution не создаётся отдельной операцией.

## Содержание

### Общие сведения

- Версия: v1
- Аутентификация: не требуется ([ADR-012](../adr/ADR-012-identity-and-auth.md))
- Базовый путь: `/executions`

### Операции

#### Зафиксировать успех

**Назначение:** переводит Execution `running` → `succeeded` (`ResultService.SucceedExecution`). Если Execution уже завершено (Fail/Abort выиграли гонку), домен отклоняет вызов — use-case не переигрывает решение (spec Execution Behavioral Invariant 5).

**Запрос:** `POST /executions/{id}/succeed`

```json
{ "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — execution не найден; `409` — execution не в состоянии `running` (`execution.ErrNotRunning`) или уже завершён (`execution.ErrTerminal`).

**События:** `ExecutionSucceeded`.

#### Зафиксировать отказ

**Назначение:** переводит Execution `running` → `failed` (`ResultService.FailExecution`), уже произведённые Artifact сохраняются как есть.

**Запрос:** `POST /executions/{id}/fail`

```json
{ "actor": "string, опционален" }
```

**Ответ:** `204 No Content`.

**Ошибки:** `404` — execution не найден; `409` — не в состоянии `running` или уже завершён.

**События:** `ExecutionFailed`.

### Модели данных

**Execution** (представление в ответах — только в ответе `POST /tasks/{id}/start`, [tasks.md](tasks.md)): `executionId`, `taskId`, `executorId`, `state` (`queued` | `running` | `succeeded` | `failed` | `aborted`).

## Статус

Актуален

## Последнее обновление

2026-07-22
