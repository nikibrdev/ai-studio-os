# Модуль: internal/domain/workflow

## Назначение

Контракты определений процесса: Definition (версионируемое описание workflow) и Step (шаг с единственной ответственной ролью) — [domain-model.md](../../../docs/architecture/domain-model.md), «Workflow» и «Workflow Step».

## Содержание

### Ответственность

- Структура определений процесса; опубликованная версия неизменяема (Draft → Published → Deprecated).
- Применение правил переходов — контракт `core.Workflow`; таблица переходов реализуется в Domain Layer строго по [state-machine.md](../../../docs/architecture/state-machine.md).
- Владение данными: Workflow, WorkflowStep, Role (словарь Role физически — в `internal/core`).

### Зависимости

- Разрешено: stdlib, `internal/core` (Role, TaskState).
- Запрещено: другие доменные модули, application, infrastructure; SQL запрещён жёстко (Workflow → SQL, [ADR-014](../../../docs/adr/ADR-014-module-interaction.md)) — персистентность только через порты.

### События

Реагирует на события жизненного цикла задач (через Orchestrator); собственные события определений (публикация версии) — при реализации.

## Статус

Актуален

## Последнее обновление

2026-07-19
