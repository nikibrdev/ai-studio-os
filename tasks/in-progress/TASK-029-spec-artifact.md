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

PR 1 («фундамент») пишется в порядке One Sentence → Identity → Purpose → Responsibilities → Invariants (Structural/Behavioral) → Lifecycle → Relationships → Alternative Interpretations Considered ([Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)). Ни одного упоминания Go/структур/реализации. Источники: [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md) (Aggregate Root, Metadata/Payload, невладение со стороны Execution) и [domain-model.md](../../docs/architecture/domain-model.md) — используются как материал, а не переписываются дословно; там, где ADR-016 не даёт прямого ответа (точный набор состояний Lifecycle, полный список инвариантов, устройство Metadata/Payload на уровне спецификации), формулировка предлагается и явно помечается как черновая, для утверждения архитектором, а не выдаётся за решённый вопрос.

Статус документа после PR 1 — «Черновик» (полный набор разделов появится только к PR 3). Открытые вопросы, если возникнут в процессе (например, по поводу Lifecycle или конкретных инвариантов), фиксируются как Open Questions прямо в PR, а не решаются самостоятельно.

После утверждения PR 1 — PR 2 (Domain Events, Commands, Queries, Examples), затем PR 3 (Acceptance Criteria, Future Extensions, Anti-Responsibilities, Non-Goals, Removal Test, Decision Log, Open Questions + три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md)).

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (первая по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
2026-07-20 — Architect — введён Three-Pass Review (+4 раздела до 16, три прохода проверки, сознательный темп — несколько PR допустимы); задача официально открыта к работе.
2026-07-20 — Architect — утверждён трёхэтапный процесс (PR 1/2/3); введён Model First — PR 1 начинается с One Sentence/Identity, Invariants разделены на Structural/Behavioral, добавлен Alternative Interpretations Considered (итого 19 разделов); сформулирован главный принцип EPIC-003 — цель не закончить спецификацию, а исключить неправильные трактовки.
2026-07-20 — Claude Code (Developer) — задача взята в работу; план PR 1 записан; начинается написание содержания.
2026-07-20 — Claude Code (Developer) — PR 1 (`docs/specifications/domain/artifact.md`, черновик) написан: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle, Relationships, Alternative Interpretations Considered. Локальная верификация пройдена. В отличие от предыдущих PR этой сессии, содержание не продиктовано архитектором дословно — это самостоятельная доменная модель, требующая реального ревью, а не самоапрува; PR отправлен на рассмотрение, merge не выполняется до обратной связи.
2026-07-20 — Architect — первое содержательное ревью PR 1: **Changes Requested** (не отклонение — направление признано сильным). Замечания: (1) One Sentence слишком завязан на «содержание», нужен акцент на «долговременный инженерный результат»; (2) Identity — критерий «можно открыть и прочитать» не универсален (бинарники, архивы, Figma); (3) Metadata как инвариант нужно уточнить минимально обязательным набором полей; (4) Author и Origin — разные понятия, развести; (5) Lifecycle — проверить, является ли Produced состоянием или событием (тест: может ли длиться часы/дни/недели); (6) Relationships — без замечаний; (7) Alternative Interpretations — усилить вариантом «Artifact как производное Event»; (8) Open Questions — добавить кардинальность Execution↔Artifact; (9) главное — не хватает явного ответа, зачем предметной области вообще нужен Artifact (переиспользование, ревью, релизы, накопление знаний, трассируемость).
2026-07-20 — Claude Code (Developer) — все девять замечаний обработаны: One Sentence и Identity переписаны без «контента» и без критерия читаемости; Structural Invariants уточнены минимальным набором полей (Identifier/Type/CreatedAt/Author/Origin), Author и Origin разведены как отдельные понятия; Lifecycle упрощён до Draft → Published → Archived — Produced исключён с явным объяснением через тест состояние-или-событие; в Purpose добавлено объяснение практической ценности (переиспользование/ревью/релизы/знания/трассируемость); Alternative Interpretations дополнен вариантом «производное Event»; кардинальность Execution↔Artifact зафиксирована как Behavioral Invariant (не просто открытый вопрос) с пояснением вывода. Локальная верификация пройдена повторно, PR обновлён новым коммитом (не amend).
2026-07-20 — Architect — второе ревью: **Approve with one non-blocking architectural note**. Согласие с исключением Produced, но по более глубокой причине (Produced описывал степень доверия процесса к артефакту, а не состояние самого Artifact — это ответственность Review/Workflow/Task, не Artifact). Уточнено значение Draft («не признан частью инженерной истории», не «не закончен») и потребован явный инвариант «Archived ≠ удалён». Единственное замечание к смержу: инвариант «не более одного producing Execution» — не делать жёстким доменным инвариантом v1, так как это может закрыть путь к будущей модели истории изменений (например, если один Artifact дополняют несколько разных Execution/Executor последовательно); смягчить до «ссылка на первое вводящее Execution» и зафиксировать кандидата на будущую сущность Revision. Разрешение смержить после смягчения.
2026-07-20 — Claude Code (Developer) — инвариант о producing Execution смягчён (ссылка только на первое вводящее Execution, вопрос истории нескольких Execution явно оставлен открытым, не запрещён); добавлен явный инвариант «Archived не означает удалённый»; Draft переопределён через «не признан частью инженерной истории проекта»; будущий кандидат Revision зафиксирован в Open Questions. Локальная верификация пройдена, PR обновлён.
