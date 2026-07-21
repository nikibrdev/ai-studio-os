# TASK-031: Спецификация домен-модуля Executor

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/executor.md` по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) (Domain Specification Review — 12 обязательных разделов, [решение](../../engineering/decisions/2026-07-20-domain-specification-review.md)) — техническое задание для будущей реализации `internal/domain/executor` (этап 2, не начинается без утверждения).

## Контекст

Executor — третий модуль в порядке проектирования: реестр исполнителей (реальных технических бэкендов — Claude Code, Codex, OpenHands, человек), их возможности и статус ([core.md](../../docs/architecture/core.md)). Не путать с `internal/platform.Executor` (контракт адаптера, [ADR-005](../../docs/adr/ADR-005-executor-contract.md)) — доменный модуль `executor` описывает домен-сущность Executor (кто зарегистрирован, какие роли может исполнять), платформенный контракт — как к нему обращается ядро. Зависит от TASK-030 (Execution использует Executor).

## Scope

### Входит

- `docs/specifications/domain/executor.md`, все 12 обязательных разделов: Purpose, Responsibilities, Invariants, Lifecycle (Registered → Active ⇄ Disabled → Retired), Relationships, Domain Events, Commands, Queries, Acceptance Criteria, Future Extensions, Anti-Responsibilities (явно: не выполняет работу сам — это платформенный адаптер через `internal/platform.Executor`, домен-модуль только реестр/состояние), Decision Log (ADR-005 и соотношение с понятием Agent — [ubiquitous-language.md](../../docs/domain/ubiquitous-language.md)).

### Не входит

- Реализация Go-пакета `internal/domain/executor`.
- Изменение `internal/platform/executor.go` (уже принят, ADR-005/TASK-026).
- Спецификации Artifact/Execution/Task/Project.

## Критерии приёмки

- [ ] Спецификация содержит все 19 обязательных разделов Specification-Domain.md, тремя PR (фундамент → поведение → завершение, [Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)).
- [ ] Пройдены три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).
- [ ] Непротиворечива с ADR-005, `internal/platform/executor.go`, domain-model.md и утверждённой спецификацией Execution (TASK-030).
- [ ] Статус спецификации — «Утверждена».
- [ ] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Затрагиваемые модули и документы

`docs/specifications/domain/executor.md` (новый).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от утверждённой спецификации Execution (TASK-030)

## План реализации

Тот же процесс, что и в TASK-029/TASK-030: три отдельных PR, порядок Model First, 20 разделов [Specification-Domain.md](../../.claude/templates/Specification-Domain.md).

- **PR 1 — фундамент** (сегодня): One Sentence → Identity → Purpose → Responsibilities → Invariants (Structural/Behavioral) → Lifecycle (Registered → Active ⇄ Disabled → Retired) → Relationships → Alternative Interpretations Considered. Ни одного упоминания Go. Источники: [ADR-005](../../docs/adr/ADR-005-executor-contract.md), [ubiquitous-language.md](../../docs/domain/ubiquitous-language.md) (разделение Agent/Executor), [core.md](../../docs/architecture/core.md) (домен-модуль `executor` — реестр, отдельно от платформенного контракта `internal/platform.Executor`), [domain-model.md](../../docs/architecture/domain-model.md). Ключевая граница, которую держит вся спецификация: доменная сущность Executor (реестр — кто зарегистрирован, какие роли способен исполнять, какой статус) — не то же самое, что платформенный контракт адаптера (как ядро технически вызывает бэкенд). Delta Review — относительно Artifact (Reference) и черновика Execution (согласованность связи Execution↔Executor с обеих сторон).
- **PR 2 — поведение**: Domain Events, Commands, Queries, Examples.
- **PR 3 — завершение и ревью**: Acceptance Criteria, Future Extensions, Anti-Responsibilities (явно — не выполняет работу сам), Non-Goals, Removal Test, Decision Log, Open Questions, Stability Assessment; три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).

Открытые по ходу PR 1 вопросы (активация после регистрации, прямой переход Registered → Retired, критерии автоматического Disabled) фиксируются в Open Questions, не решаются самостоятельно.

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (третья по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
2026-07-20 — Architect — введён Three-Pass Review (+4 раздела до 16, три прохода проверки, сознательный темп).
2026-07-20 — Architect — введён Model First (+3 раздела до 19: One Sentence, Identity, Alternative Interpretations Considered; Invariants разделены на Structural/Behavioral; спецификация пишется тремя PR — фундамент/поведение/завершение).
2026-07-20 — Architect — введён Reference Status (+1 раздел Stability Assessment, итого 20; Delta Review обязателен, начиная с TASK-030).
2026-07-21 — Claude Code (Developer) — задача взята в работу, переведена в `tasks/in-progress/`; план PR 1 записан.
2026-07-21 — Claude Code (Developer) — PR 1 (`docs/specifications/domain/executor.md`, черновик) написан: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle (Registered → Active ⇄ Disabled → Retired), Relationships, Alternative Interpretations Considered. Ключевая граница спецификации: доменная сущность Executor (реестр) явно отделена от платформенного контракта адаптера (`internal/platform.Executor`, ADR-005) — контракт описывает «как вызывается», спецификация — «кто зарегистрирован и в каком состоянии». Delta Review проведён относительно Artifact (Reference) и черновика Execution: связь Execution↔Executor согласована с обеих сторон (Execution фиксирует Executor при создании, Executor не владеет своими Execution). Три вопроса зафиксированы в Open Questions, не решены самостоятельно. Содержание требует реального ревью, не самоапрува; merge не выполняется до обратной связи.
