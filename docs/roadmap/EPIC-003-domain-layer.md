# EPIC-003: Domain Layer — предметная область

## Цель

Реализовать доменную логику AI Studio OS (v0.3, [ROADMAP.md](../../ROADMAP.md)) поверх контрактов, зафиксированных в EPIC-002/ADR-015. Отличие от типового «пишем модули по очереди»: реализация начинается не с `Task`, а с того, что система производит — по решению архитектора ([domain-model.md](../architecture/domain-model.md), раздел «Порядок проектирования Domain Layer»; [ADR-016](../adr/ADR-016-artifact-aggregate-root.md)).

## Контекст

Перед этим эпиком приняты все блокирующие решения: ADR-005 (Executor Contract), доменные основания TASK-027 (Execution — не Bounded Context, четыре контекста Planning/Development/Review/Knowledge), ADR-016 (Artifact — самостоятельный Aggregate Root). Архитектор явно определил порядок и режим старта (2026-07-20): **Domain Specifications First** — эпик открывается без единой строки Go, полными спецификациями пяти модулей, и переходит к реализации только после их утверждения.

## Scope

### Этап 1 — Domain Specifications (текущий; без кода)

Написать и утвердить полные спецификации по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) — 20 обязательных разделов ([Domain Specification Review](../../engineering/decisions/2026-07-20-domain-specification-review.md), [Three-Pass Review](../../engineering/decisions/2026-07-20-domain-specification-three-pass-review.md), [Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md), [Reference Status](../../engineering/decisions/2026-07-20-domain-specification-reference-status.md), расширяют [базовое требование](../../engineering/decisions/2026-07-20-domain-layer-specification-requirement.md)), написанные тремя отдельными PR на модуль:

- **PR 1 — фундамент** (One Sentence, Identity, Purpose, Responsibilities, Invariants [Structural/Behavioral], Lifecycle, Relationships, Alternative Interpretations Considered) — ни одного упоминания Go; сущность определяется сама по себе, до того как названо, кто на неё ссылается.
- **PR 2 — поведение** (Domain Events, Commands, Queries, Examples).
- **PR 3 — завершение и ревью** (Acceptance Criteria, Future Extensions, Anti-Responsibilities, Non-Goals, Removal Test, Decision Log, Open Questions), завершается [тремя проходами проверки](../../.claude/checklists/DomainSpecificationReview.md) (Internal Consistency, Cross-domain Consistency, Future-proof Review).

Главный принцип: не стремиться закончить спецификацию — стремиться исключить неправильные трактовки сущности. Ожидаемый темп — сознательно небольшой: одна спецификация законно занимает несколько PR, если по ходу работы возникают вопросы о жизненном цикле, инвариантах или связях — это признак процесса, а не задержка. Порядок написания — по порядку проектирования из ADR-016/domain-model.md:

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

- [x] Пять спецификаций написаны по шаблону Specification-Domain.md, все 20 разделов в каждой, тремя PR (фундамент → поведение → завершение). ([Artifact](../specifications/domain/artifact.md) — статус Reference; [Execution](../specifications/domain/execution.md), [Executor](../specifications/domain/executor.md), [Task](../specifications/domain/task.md), [Project](../specifications/domain/project.md) — статус Утверждена.)
- [x] Каждая спецификация прошла три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md) (Internal Consistency, Cross-domain Consistency с Delta Review начиная с TASK-030, Future-proof Review).
- [x] Каждая спецификация непротиворечива с уже принятыми ADR-005/ADR-016/domain-model.md и друг с другом (например: `ExecutorTask`/`Artifact`/`ExecutionStatus` в спецификации Execution/Executor не противоречат абстрактным типам `internal/platform`).
- [x] Все пять спецификаций явно утверждены архитектором (статус «Утверждена», не «Черновик»).
- [x] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Критерии завершения (этап 2)

- [x] Пять доменных сущностей реализованы строго по утверждённым спецификациям, инварианты — проверяемый код: `artifact`, `execution`, `executor`, `task` (включая решённые расширения контракта — `epicID`, SetScope/SetAcceptanceCriteria), `project` (включая явную команду `Activate`).
- [x] Каноническая state machine реализована (`workflow.Machine`, контракт `Rules`) строго по [state-machine.md](../architecture/state-machine.md) — 20 переходов, исчерпывающий перебор всех 81 пары в тестах; решение о реализации по каноническому источнику без отдельной спецификации — [зафиксировано](../../engineering/decisions/2026-07-21-workflow-rules-canonical-source.md).
- [x] Каждый пакет покрыт unit-тестами на все Structural/Behavioral инварианты (18/17/17/11/11/6 тестов; покрытие 81.8–100%).
- [x] Сквозной сценарий уровня слоя ([internal/domain/goldenpath_test.go](../../internal/domain/goldenpath_test.go)): Task проходит все девять канонических состояний через реальную `workflow.Machine`, порождая Execution и опубликованный Artifact — результат v0.3 из ROADMAP.md подтверждён кодом.
- [x] `make verify` — чисто на каждом PR; merge только при зелёном обязательном статус-чеке CI.

## Декомпозиция

| Задача | Содержание | Статус |
| --- | --- | --- |
| TASK-029 | Спецификация `Artifact` | done — статус **Reference** (первая доменная спецификация проекта, [решение](../../engineering/decisions/2026-07-20-domain-specification-reference-status.md)) |
| TASK-030 | Спецификация `Execution` | done — статус **Утверждена** |
| TASK-031 | Спецификация `Executor` | done — статус **Утверждена** |
| TASK-032 | Спецификация `Task` | done — статус **Утверждена** (расхождения с `contracts.go` — решённое направление расширения на этапе 2) |
| TASK-033 | Спецификация `Project` | done — статус **Утверждена** (Activate — решённое направление расширения `registry.go` на этапе 2) |
| TASK-034 | Реализация `internal/domain/artifact` | done (PR #42) |
| TASK-035 | Реализация `internal/domain/execution` | done (PR #43) |
| TASK-036 | Реализация `internal/domain/executor` | done (PR #44) |
| TASK-037 | Сущность Task + расширение контракта (`epicID`, scope/AC) | done (PR #45) |
| TASK-038 | Сущность Project + команда `Activate` | done (PR #46) |
| TASK-039 | `workflow.Machine` — каноническая state machine | done (PR #47) |

## Риски и зависимости

- Спецификации Task/Project писались поверх уже существующего кода ([internal/domain/task](../../internal/domain/task/), [internal/domain/project](../../internal/domain/project/)) — расхождения найдены (`epicID`/scope-AC у Task; отсутствие `Activate` у Project) и разрешены финальным ревью как решённое направление расширения контрактов на этапе 2, а не как правки кода в рамках этапа 1.
- Спецификации Artifact/Execution/Executor/Task/Project взаимозависимы — утверждались по порядку, каждая следующая проверена на согласованность с уже утверждёнными перед своим утверждением (Delta Review).
- Все пять спецификаций закрыты за один день каждая (как и планировалось [решением](../../engineering/decisions/2026-07-20-domain-specification-three-pass-review.md) — сознательный темп, не в ущерб раундам ревью).
- Начиная с TASK-030, каждая спецификация прошла обязательный **Delta Review** относительно уже утверждённых/Reference спецификаций ([решение](../../engineering/decisions/2026-07-20-domain-specification-reference-status.md)). Утверждённые/Reference спецификации теперь изменяются только через новый ADR, отдельный Domain Revision PR либо обоснованное изменение с тем же Delta Review — не напрямую.

## Статус

**Эпик закрыт** (2026-07-21): этап 1 — все пять спецификаций утверждены; этап 2 — все шесть задач реализации выполнены, сквозной сценарий слоя подтверждает результат v0.3. Следующий эпик — Application Layer (v0.4).

## Последнее обновление

2026-07-21
