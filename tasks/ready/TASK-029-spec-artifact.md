# TASK-029: Спецификация домен-модуля Artifact

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/artifact.md` по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) (Domain Specification Review — 12 обязательных разделов, [решение](../../engineering/decisions/2026-07-20-domain-specification-review.md)) — техническое задание для будущей реализации `internal/domain/artifact` (этап 2, отдельная задача, не начинается без утверждения этой спецификации).

## Контекст

Artifact — первый модуль в порядке проектирования Domain Layer ([domain-model.md](../../docs/architecture/domain-model.md), [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)): самостоятельный Aggregate Root, не часть Execution/Task/Project. Концептуальное описание уже есть в ADR-016 и domain-model.md — задача переводит его в полную спецификацию по всем 12 разделам Specification-Domain.md, а не изобретает решение заново.

## Scope

### Входит

- `docs/specifications/domain/artifact.md`, все 12 обязательных разделов: Purpose, Responsibilities, Invariants, Lifecycle (жизненный цикл, обязательно — модуль с состояниями), Relationships (владение/ссылки/создание/удаление), Domain Events, Commands, Queries, Acceptance Criteria, Future Extensions, Anti-Responsibilities, Decision Log (ADR-016 и другие решения, на которых основана спецификация).
- Явное описание разделения Metadata/Payload (по ADR-016) в разделах Responsibilities/Invariants — что обязано хранить ядро (Metadata), что не интерпретирует (Payload).
- Обновление `internal/domain/README.md` (при необходимости) — ссылка на новую спецификацию, без изменения кода.

### Не входит

- Реализация Go-пакета `internal/domain/artifact` — отдельная задача этапа 2, после утверждения.
- Спецификации Execution/Executor/Task/Project — отдельные задачи (TASK-030…033).

## Критерии приёмки

- [ ] Спецификация содержит все 16 обязательных разделов Specification-Domain.md.
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
