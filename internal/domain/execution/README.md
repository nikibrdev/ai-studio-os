# Модуль: internal/domain/execution

## Назначение

Сущность Execution — единичный, ограниченный во времени запуск конкретного Executor'а для выполнения одной Task, производящий Artifact и несущий статус исполнения ([полная спецификация](../../../docs/specifications/domain/execution.md), статус Утверждена; [ADR-005](../../../docs/adr/ADR-005-executor-contract.md), [ADR-016](../../../docs/adr/ADR-016-artifact-aggregate-root.md)).

## Содержание

### Ответственность

- Тип `Execution` — сущность с фиксированными при создании ссылками (TaskID, ExecutorID) и изменяемым состоянием (State, множество произведённых Artifact); инварианты применяются в коде.
- Lifecycle: `Queued → Running → Succeeded | Failed | Aborted`, плюс прямой `Queued → Aborted`; терминальные состояния необратимы, повторная попытка — новый Execution.
- Команды как методы: `New` (Create), `Accept`, `RecordArtifact`, `Succeed`, `Fail`, `Abort` — соответствие четырём возможностям контракта Executor (ADR-005); каждая возвращает доменное событие как значение.
- Гонка Fail/Abort разрешается порядком выполнения: первый терминальный переход выигрывает, второй получает ошибку (Behavioral Invariant 5 спецификации).

### Зависимости

- Разрешено: только стандартная библиотека.
- Запрещено: другие доменные модули (Task/Executor/Artifact — только идентификаторы-строки), `internal/platform`, application, infrastructure ([ADR-015](../../../docs/adr/ADR-015-internal-layering.md)).

### События

`Queued`, `Started`, `Succeeded`, `Failed`, `Aborted` — возвращаются командами как значения; публикация через Event Bus не входит в этот пакет.

### Ограничения

Тайм-аут пребывания в Queued и политика повторных попыток — Open Questions спецификации, вне домена (Application/Workflow).

## Статус

Актуален

## Последнее обновление

2026-07-21
