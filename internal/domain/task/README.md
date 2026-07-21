# Модуль: internal/domain/task

## Назначение

Сущность Task и контракты работы с задачами: единый путь записи состояния, чтение для слоя доставки и markdown-экспорт в `tasks/` ([полная спецификация](../../../docs/specifications/domain/task.md), статус Утверждена; [ADR-004](../../../docs/adr/ADR-004-task-storage.md)).

## Содержание

### Ответственность

- Тип `Task` — сущность с инвариантами в коде (TASK-037): принадлежность Project (обязательна), Epic (опциональна, `0..1`), scope/критерии приёмки (редактируются только в Backlog — Ready означает выполненный DoR, изменение требований идёт через возврат Ready → Backlog), состояние из девяти канонических.
- Допустимость перехода решает контракт [`workflow.Rules`](../workflow/README.md), не сама сущность (инвариант 8 [state-machine.md](../../../docs/architecture/state-machine.md)); сущность применяет только то, что знает сама: причина обязательна для Blocked/Cancelled (инвариант 3).
- `Commands` — единственный путь изменения состояния задачи; расширен по Decision Log спецификации (этап 2): `Create` принимает `epicID`, добавлены `SetScope`/`SetAcceptanceCriteria`.
- `Queries` — чтение состояния для приложений (не межмодульный контракт, ADR-014).
- `Exporter` — генерация markdown-представления в `tasks/` (экспорт — не источник истины).

Внутреннее устройство пути записи (обычная персистентность или Command → Event → Projection) контрактами намеренно не зафиксировано; реализация поверх PostgreSQL — v0.5 (ADR-004, ADR-011).

### Зависимости

- Разрешено: stdlib, `internal/domain/shared` (TaskState, Role), `internal/domain/workflow` (контракт Rules — санкционированная зависимость, инвариант 8 state-machine.md).
- Запрещено: другие доменные модули (внутренности), application, infrastructure, SQL/драйверы.

### События

`Created`, `Transitioned` — возвращаются командами сущности как значения; отображение в 15 канонических имён событий ([events.md](../../../docs/architecture/events.md)) — ответственность публикатора (Application Layer).

### Владение данными

Task, Epic ([domain-model.md](../../../docs/architecture/domain-model.md)); Epic как самостоятельная сущность не специфицирован (Non-Goals спецификации Task).

## Статус

Актуален

## Последнее обновление

2026-07-21
