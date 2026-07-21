# Модуль: internal/domain/executor

## Назначение

Доменная сущность Executor — запись реестра технических бэкендов (человек или система), зарегистрированных в платформе и способных исполнять работу от имени одной или нескольких Role ([полная спецификация](../../../docs/specifications/domain/executor.md), статус Утверждена; [ADR-005](../../../docs/adr/ADR-005-executor-contract.md)).

## Содержание

### Ответственность

- Тип `Executor` — идентичность бэкенда (фиксируется при регистрации), набор исполняемых `shared.Role` (никогда не пуст) и статус доступности.
- Lifecycle: `Registered → Active ⇄ Disabled → Retired`, плюс прямой `Registered → Retired`; Retired терминален — вернувшийся бэкенд регистрируется как новый Executor.
- Команды как методы: `New` (Register), `Activate`, `Disable`, `Retire`, `GrantRole`, `RevokeRole` (последняя роль не отзывается — только Retire); события `Registered`/`Activated`/`Disabled`/`Retired` возвращаются значениями.
- Предикат `AvailableForAssignment` — новое Execution назначается только Active-исполнителю (Behavioral Invariant 4); само назначение — ответственность Application Layer.

### Отличие от платформенного контракта

Доменный Executor — «кто зарегистрирован и в каком состоянии»; `internal/platform.Executor` (ADR-005) — «как ядро технически вызывает бэкенд» (Accept/Artifacts/Status/Finish). Связаны идентичностью бэкенда, не владением ([ADR-015](../../../docs/adr/ADR-015-internal-layering.md)).

### Зависимости

- Разрешено: стандартная библиотека, `internal/domain/shared` (словарь Role).
- Запрещено: другие доменные модули, `internal/platform`, application, infrastructure.

### События

`Registered`, `Activated`, `Disabled`, `Retired` — возвращаются командами как значения; Activated/Retired несут состояние-источник (единое событие на целевое состояние).

### Ограничения

Критерии активации и автоматического Disable — Open Questions спецификации, решаются на уровне Application/Infrastructure, не в домене.

## Статус

Актуален

## Последнее обновление

2026-07-21
