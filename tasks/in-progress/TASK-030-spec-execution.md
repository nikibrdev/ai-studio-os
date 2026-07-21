# TASK-030: Спецификация домен-модуля Execution

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/execution.md` по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) (Domain Specification Review — 12 обязательных разделов, [решение](../../engineering/decisions/2026-07-20-domain-specification-review.md)) — техническое задание для будущей реализации `internal/domain/execution` (этап 2, не начинается без утверждения).

## Контекст

Execution — второй модуль в порядке проектирования ([domain-model.md](../../docs/architecture/domain-model.md)): один запуск Executor'а для выполнения задачи или шага workflow; производит Artifact и несёт `ExecutionStatus`, но никогда не владеет произведёнными Artifact ([ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)). Execution — не Bounded Context, а сквозная возможность, координируемая Application Layer ([bounded-contexts.md](../../docs/domain/bounded-contexts.md)); зависит от TASK-029 (Artifact), должна быть согласована с ним по вопросу ссылки Execution → Artifact.

## Scope

### Входит

- `docs/specifications/domain/execution.md`, все 12 обязательных разделов: Purpose, Responsibilities, Invariants, Lifecycle (Queued → Running → Succeeded | Failed | Aborted и правила переходов), Relationships (Task создаёт, Executor используется, Artifact — ссылка без владения), Domain Events, Commands, Queries, Acceptance Criteria, Future Extensions, Anti-Responsibilities, Decision Log (ADR-005, ADR-016).
- Согласованность с [ADR-005](../../docs/adr/ADR-005-executor-contract.md) (`Accept`/`Artifacts`/`Status`/`Finish`) в разделах Commands/Domain Events.

### Не входит

- Реализация Go-пакета `internal/domain/execution`.
- Спецификации Artifact/Executor/Task/Project.

## Критерии приёмки

- [ ] Спецификация содержит все 19 обязательных разделов Specification-Domain.md, тремя PR (фундамент → поведение → завершение, [Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)).
- [ ] Пройдены три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).
- [ ] Непротиворечива с ADR-005, ADR-016, domain-model.md и утверждённой спецификацией Artifact (TASK-029).
- [ ] Статус спецификации — «Утверждена».
- [ ] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Затрагиваемые модули и документы

`docs/specifications/domain/execution.md` (новый).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от утверждённой спецификации Artifact (TASK-029)

## План реализации

Тот же процесс, что и в TASK-029 (Artifact, Reference): три отдельных PR, в порядке Model First ([решение](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)), 20 разделов итогового шаблона [Specification-Domain.md](../../.claude/templates/Specification-Domain.md).

- **PR 1 — фундамент** (сегодня): One Sentence → Identity → Purpose → Responsibilities → Invariants (Structural/Behavioral) → Lifecycle → Relationships → Alternative Interpretations Considered. Ни одного упоминания Go. Источники: [ADR-005](../../docs/adr/ADR-005-executor-contract.md) (Accept/Artifacts/Status/Finish, без Result), [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md) (Execution ссылается на Artifact, не владеет), [domain-model.md](../../docs/architecture/domain-model.md) (раздел Execution и Lifecycle Queued→Running→Succeeded|Failed|Aborted) — используются как материал, не переписываются дословно. Впервые проводится обязательный Delta Review ([решение](../../engineering/decisions/2026-07-20-domain-specification-reference-status.md)) относительно Artifact (Reference): терминология «производит Artifact по ссылке, не владеет» и кардинальность (Execution → 0..* Artifact; Artifact → не более одного породившего Execution) выровнены с уже утверждённой спецификацией Artifact, не переопределяются заново.
- **PR 2 — поведение**: Domain Events, Commands, Queries, Examples — после ревью и merge PR 1.
- **PR 3 — завершение и ревью**: Acceptance Criteria, Future Extensions, Anti-Responsibilities, Non-Goals, Removal Test, Decision Log, Open Questions, Stability Assessment; три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md) (Internal Consistency, Cross-domain Consistency + Delta Review, Future-proof Review) с письменными ответами на диагностические вопросы.

Открытые по ходу PR 1 вопросы (порог ожидания в Queued, граница Failed/Aborted, момент порождения повторной попытки) — не решаются самостоятельно, фиксируются в Open Questions спецификации для решения архитектором.

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (вторая по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
2026-07-20 — Architect — введён Three-Pass Review (+4 раздела до 16, три прохода проверки, сознательный темп).
2026-07-20 — Architect — введён Model First (+3 раздела до 19: One Sentence, Identity, Alternative Interpretations Considered; Invariants разделены на Structural/Behavioral; спецификация пишется тремя PR — фундамент/поведение/завершение).
2026-07-20 — Architect — введён Reference Status (+1 раздел Stability Assessment, итого 20; Artifact — первая спецификация со статусом Reference; Delta Review обязателен начиная с этой задачи).
2026-07-21 — Claude Code (Developer) — задача взята в работу, переведена в `tasks/in-progress/`; план PR 1 записан («План реализации»); начато написание содержания по образцу Artifact (Reference) с обязательным Delta Review.
2026-07-21 — Claude Code (Developer) — PR 1 (`docs/specifications/domain/execution.md`, черновик) написан: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle (Queued → Running → Succeeded | Failed | Aborted, плюс Queued → Aborted), Relationships, Alternative Interpretations Considered. Delta Review относительно Artifact проведён: терминология «ссылка, не владение» и кардинальность связи Execution↔Artifact выровнены с уже утверждённой спецификацией, не переопределены заново. Три вопроса, требующие решения архитектора, зафиксированы в Open Questions, не решены самостоятельно. Содержание — самостоятельная доменная модель, требует реального ревью, не самоапрува; merge не выполняется до обратной связи.
