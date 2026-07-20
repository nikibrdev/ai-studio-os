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

- [ ] Пять спецификаций написаны по шаблону Specification-Domain.md, все 20 разделов в каждой, тремя PR (фундамент → поведение → завершение). ([Artifact](../specifications/domain/artifact.md) — done, статус Reference.)
- [ ] Каждая спецификация прошла три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md) (Internal Consistency, Cross-domain Consistency с Delta Review начиная с TASK-030, Future-proof Review).
- [ ] Каждая спецификация непротиворечива с уже принятыми ADR-005/ADR-016/domain-model.md и друг с другом (например: `ExecutorTask`/`Artifact`/`ExecutionStatus` в спецификации Execution/Executor не противоречат абстрактным типам `internal/platform`).
- [ ] Все пять спецификаций явно утверждены архитектором (статус «Утверждена», не «Черновик»).
- [ ] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Декомпозиция

| Задача | Содержание | Статус |
| --- | --- | --- |
| TASK-029 | Спецификация `Artifact` | done — статус **Reference** (первая доменная спецификация проекта, [решение](../../engineering/decisions/2026-07-20-domain-specification-reference-status.md)) |
| TASK-030 | Спецификация `Execution` | ready |
| TASK-031 | Спецификация `Executor` | ready |
| TASK-032 | Спецификация `Task` | ready |
| TASK-033 | Спецификация `Project` | ready |

## Риски и зависимости

- Спецификации Task/Project пишутся поверх уже существующего кода ([internal/domain/task](../../internal/domain/task/), [internal/domain/project](../../internal/domain/project/)) — риск обнаружить расхождение между кодом и новым порядком/решениями (ADR-016 и т.д.); если найдено — фиксируется как Open Question, код в рамках этапа 1 не трогается.
- Спецификации Artifact/Execution/Executor взаимозависимы (Execution ссылается на Artifact и использует Executor) — утверждаются по отдельности, но проверяются на согласованность друг с другом перед утверждением каждой следующей.
- Одна задача (например, TASK-029) может закрываться несколькими PR, если по ходу работы возникают вопросы, требующие решения архитектора, — сознательно принятый риск против скорости ([решение](../../engineering/decisions/2026-07-20-domain-specification-three-pass-review.md)).
- Начиная с TASK-030, каждая спецификация обязана пройти **Delta Review** относительно уже утверждённых/Reference спецификаций (сейчас — только Artifact): не требует ли пересмотра уже принятого, использует ли принятые понятия единообразно, не вводит ли дублирующее понятие под другим именем ([решение](../../engineering/decisions/2026-07-20-domain-specification-reference-status.md)). Утверждённые/Reference спецификации после этого изменяются только через новый ADR, отдельный Domain Revision PR либо обоснованное изменение с тем же Delta Review — не напрямую.

## Статус

В работе (этап 1 открыт 2026-07-20)

## Последнее обновление

2026-07-20
