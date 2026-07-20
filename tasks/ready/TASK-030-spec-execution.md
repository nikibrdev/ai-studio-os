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

- [ ] Спецификация содержит все 12 обязательных разделов Specification-Domain.md.
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

<Заполняется при взятии задачи в работу.>

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (вторая по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
