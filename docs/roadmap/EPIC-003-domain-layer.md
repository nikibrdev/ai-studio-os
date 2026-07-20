# EPIC-003: Domain Layer — предметная область

## Цель

Реализовать доменную логику AI Studio OS (v0.3, [ROADMAP.md](../../ROADMAP.md)) поверх контрактов, зафиксированных в EPIC-002/ADR-015. Отличие от типового «пишем модули по очереди»: реализация начинается не с `Task`, а с того, что система производит — по решению архитектора ([domain-model.md](../architecture/domain-model.md), раздел «Порядок проектирования Domain Layer»; [ADR-016](../adr/ADR-016-artifact-aggregate-root.md)).

## Контекст

Перед этим эпиком приняты все блокирующие решения: ADR-005 (Executor Contract), доменные основания TASK-027 (Execution — не Bounded Context, четыре контекста Planning/Development/Review/Knowledge), ADR-016 (Artifact — самостоятельный Aggregate Root). Архитектор явно определил порядок и режим старта (2026-07-20): **Domain Specifications First** — эпик открывается без единой строки Go, полными спецификациями пяти модулей, и переходит к реализации только после их утверждения.

## Scope

### Этап 1 — Domain Specifications (текущий; без кода)

Написать и утвердить полные спецификации по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) — Domain Specification Review, 12 обязательных разделов ([решение](../../engineering/decisions/2026-07-20-domain-specification-review.md), расширяет [базовое требование](../../engineering/decisions/2026-07-20-domain-layer-specification-requirement.md)): Purpose, Responsibilities, Invariants, Lifecycle, Relationships, Domain Events, Commands, Queries, Acceptance Criteria, Future Extensions, Anti-Responsibilities, Decision Log. Порядок написания — по порядку проектирования из ADR-016/domain-model.md:

1. `Artifact` (`docs/specifications/domain/artifact.md`)
2. `Execution` (`docs/specifications/domain/execution.md`)
3. `Executor` (`docs/specifications/domain/executor.md`)
4. `Task` (`docs/specifications/domain/task.md`) — переработка: контракты уже существуют ([internal/domain/task](../../internal/domain/task/README.md)), но полной спецификации нет.
5. `Project` (`docs/specifications/domain/project.md`) — переработка: контракты уже существуют ([internal/domain/project](../../internal/domain/project/README.md)), но полной спецификации нет.

### Не входит (этап 1)

- Любая реализация на Go (пакеты `internal/domain/artifact`, `execution`, `executor` не создаются).
- Модули `workflow`, `tool`, `event`, `memory`, `git`, `identity` — не в объёме пяти названных архитектором спецификаций; их спецификации — по решению архитектора отдельно, после этапа 1.
- Изменение уже принятых контрактов `internal/domain/task`, `internal/domain/project` — только документирование их текущего поведения полной спецификацией, без правки кода в рамках этапа 1.

### Этап 2 — Реализация (не начинается без утверждения этапа 1)

Реализация пяти модулей на Go, в том же порядке. Открывается отдельными задачами только после утверждения архитектором всех пяти спецификаций этапа 1.

## Критерии завершения (этап 1)

- [ ] Пять спецификаций написаны по шаблону Specification-Domain.md, все 12 разделов Domain Specification Review в каждой.
- [ ] Каждая спецификация непротиворечива с уже принятыми ADR-005/ADR-016/domain-model.md и друг с другом (например: `ExecutorTask`/`Artifact`/`ExecutionStatus` в спецификации Execution/Executor не противоречат абстрактным типам `internal/platform`).
- [ ] Все пять спецификаций явно утверждены архитектором (статус «Утверждена», не «Черновик»).
- [ ] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Декомпозиция

| Задача | Содержание | Статус |
| --- | --- | --- |
| TASK-029 | Спецификация `Artifact` | ready |
| TASK-030 | Спецификация `Execution` | ready |
| TASK-031 | Спецификация `Executor` | ready |
| TASK-032 | Спецификация `Task` | ready |
| TASK-033 | Спецификация `Project` | ready |

## Риски и зависимости

- Спецификации Task/Project пишутся поверх уже существующего кода ([internal/domain/task](../../internal/domain/task/), [internal/domain/project](../../internal/domain/project/)) — риск обнаружить расхождение между кодом и новым порядком/решениями (ADR-016 и т.д.); если найдено — фиксируется как Open Question, код в рамках этапа 1 не трогается.
- Спецификации Artifact/Execution/Executor взаимозависимы (Execution ссылается на Artifact и использует Executor) — пишутся и утверждаются по отдельности (один PR — одна задача), но проверяются на согласованность друг с другом перед утверждением каждой следующей.

## Статус

В работе (этап 1 открыт 2026-07-20)

## Последнее обновление

2026-07-20
