# TASK-033: Спецификация домен-модуля Project

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/project.md` по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) (Domain Specification Review — 12 обязательных разделов, [решение](../../engineering/decisions/2026-07-20-domain-specification-review.md)) — техническое задание для реализации `internal/domain/project`, чьи контракты уже частично существуют ([internal/domain/project/registry.go](../../internal/domain/project/registry.go), EPIC-002) без полной спецификации.

## Контекст

Project — пятый, последний модуль в порядке проектирования этапа 1: граница, внутри которой существуют Task и Artifact (`Project ├── Task └── Artifact`, [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)). `Registry` (Created → Active → Archived) уже принят; задача — полная спецификация поверх него, с явным описанием владения Task и Artifact.

## Scope

### Входит

- `docs/specifications/domain/project.md`, все 12 обязательных разделов: Purpose, Responsibilities, Invariants (минимум один Repository, архив неизменяем), Lifecycle (Created → Active → Archived), Relationships (владение Task и Artifact — `Project ├── Task └── Artifact`, ADR-016), Domain Events, Commands/Queries (согласованные с уже принятым `Registry`), Acceptance Criteria, Future Extensions, Anti-Responsibilities, Decision Log (ADR-013 — формат подключения репозиториев, Decision Required, зафиксировать как ограничение, не решать здесь).
- Сверка с текущим кодом `internal/domain/project/registry.go` — расхождения фиксируются как Open Questions, код не меняется в рамках этой задачи.

### Не входит

- Изменение `internal/domain/project/registry.go` — только документирование.
- Спецификации Artifact/Execution/Executor/Task.

## Критерии приёмки

- [ ] Спецификация содержит все 16 обязательных разделов Specification-Domain.md.
- [ ] Пройдены три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).
- [ ] Согласована с уже принятым кодом `internal/domain/project` — расхождения (если есть) зафиксированы как Open Questions.
- [ ] Непротиворечива с ADR-013, domain-model.md, утверждёнными спецификациями Artifact (TASK-029) и Task (TASK-032).
- [ ] Статус спецификации — «Утверждена».
- [ ] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Затрагиваемые модули и документы

`docs/specifications/domain/project.md` (новый); `internal/domain/project/` (только сверка, без правок).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от утверждённых спецификаций Artifact (TASK-029) и Task (TASK-032)

## План реализации

<Заполняется при взятии задачи в работу.>

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (пятая по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
2026-07-20 — Architect — введён Three-Pass Review (+4 раздела до 16, три прохода проверки, сознательный темп).
