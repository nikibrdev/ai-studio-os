# TASK-029: Спецификация домен-модуля Artifact

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/artifact.md` по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) (19 обязательных разделов — [Domain Specification Review](../../engineering/decisions/2026-07-20-domain-specification-review.md), [Three-Pass Review](../../engineering/decisions/2026-07-20-domain-specification-three-pass-review.md), [Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)) — техническое задание для будущей реализации `internal/domain/artifact` (этап 2, отдельная задача, не начинается без утверждения этой спецификации). Пишется тремя отдельными PR (фундамент → поведение → завершение), см. «План реализации».

## Контекст

Artifact — первый модуль в порядке проектирования Domain Layer ([domain-model.md](../../docs/architecture/domain-model.md), [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)): самостоятельный Aggregate Root, не часть Execution/Task/Project. Концептуальное описание уже есть в ADR-016 и domain-model.md — задача переводит его в полную спецификацию по всем 19 разделам Specification-Domain.md, а не изобретает решение заново. Как первая спецификация Domain Layer, PR 1 этой задачи задаёт прецедент для TASK-030…033.

## Scope

### Входит

- `docs/specifications/domain/artifact.md`, все 19 обязательных разделов, тремя PR:
  - **PR 1 — фундамент**: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle, Relationships (владение/ссылки/создание/удаление, не определяет сущность), Alternative Interpretations Considered. Ни одного упоминания Go.
  - **PR 2 — поведение**: Domain Events, Commands, Queries, Examples (не код — содержательные примеры конкретных Artifact).
  - **PR 3 — завершение и ревью**: Acceptance Criteria, Future Extensions, Anti-Responsibilities, Non-Goals, Removal Test, Decision Log (ADR-016 и другие решения), Open Questions; три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).
- Явное описание разделения Metadata/Payload (по ADR-016) в разделах Responsibilities/Invariants — что обязано хранить ядро (Metadata), что не интерпретирует (Payload).
- Обновление `internal/domain/README.md` (при необходимости) — ссылка на новую спецификацию, без изменения кода.

### Не входит

- Реализация Go-пакета `internal/domain/artifact` — отдельная задача этапа 2, после утверждения.
- Спецификации Execution/Executor/Task/Project — отдельные задачи (TASK-030…033).

## Критерии приёмки

- [ ] Спецификация содержит все 19 обязательных разделов Specification-Domain.md.
- [ ] Пройдены три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md) (Internal Consistency, Cross-domain Consistency, Future-proof Review).
- [ ] Непротиворечива с ADR-016, ADR-005, domain-model.md.
- [ ] Статус спецификации — «Утверждена» (после явного подтверждения архитектора).
- [ ] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Затрагиваемые модули и документы

`docs/specifications/domain/artifact.md` (новый).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## План реализации

<Заполняется при взятии задачи в работу.>

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (первая по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
2026-07-20 — Architect — введён Three-Pass Review (+4 раздела до 16, три прохода проверки, сознательный темп — несколько PR допустимы); задача официально открыта к работе.
2026-07-20 — Architect — утверждён трёхэтапный процесс (PR 1/2/3); введён Model First — PR 1 начинается с One Sentence/Identity, Invariants разделены на Structural/Behavioral, добавлен Alternative Interpretations Considered (итого 19 разделов); сформулирован главный принцип EPIC-003 — цель не закончить спецификацию, а исключить неправильные трактовки.
