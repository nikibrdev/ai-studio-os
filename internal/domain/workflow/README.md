# Модуль: internal/domain/workflow

## Назначение

Правила процесса и определения workflow: реализация канонической state machine задачи (`Machine`, контракт `Rules`) и контракты версионируемых определений процесса (`Definition`, `Step`) — [domain-model.md](../../../docs/architecture/domain-model.md), «Workflow» и «Workflow Step».

## Содержание

### Ответственность

- `Machine` — реализация контракта `Rules` (TASK-039): полная таблица переходов [state-machine.md](../../../docs/architecture/state-machine.md) (20 разрешённых переходов, 9 состояний) и таблица ролей по стадиям [workflow.md](../../../docs/architecture/workflow.md); без состояния и I/O, детерминирована по построению. Реализована напрямую по каноническим документам без отдельной 20-раздельной спецификации — [решение архитектора](../../../engineering/decisions/2026-07-21-workflow-rules-canonical-source.md).
- `Rules` — контракт применения правил: решает, но не действует (состояние меняет модуль `task`, события публикуют владельцы — ADR-014).
- `Definition`/`Step` — контракты определений процесса; опубликованная версия неизменяема (Draft → Published → Deprecated); реализация — при появлении потребителя (Application Layer, v0.4).
- Владение данными: Workflow, WorkflowStep, Role (словарь Role физически — в `internal/domain/shared`, ADR-015).

### Зависимости

- Разрешено: stdlib, `internal/domain/shared` (Role, TaskState).
- Запрещено: другие доменные модули, application, infrastructure; SQL запрещён жёстко (Workflow → SQL, [ADR-014](../../../docs/adr/ADR-014-module-interaction.md)) — персистентность только через порты.

### События

Реагирует на события жизненного цикла задач (через Orchestrator); собственные события определений (публикация версии) — при реализации Definition.

### Ограничения

Изменение таблицы переходов или ролей — изменение замороженной архитектуры: только через ADR ([решение](../../../engineering/decisions/2026-07-21-workflow-rules-canonical-source.md)).

## Статус

Актуален

## Последнее обновление

2026-07-21
