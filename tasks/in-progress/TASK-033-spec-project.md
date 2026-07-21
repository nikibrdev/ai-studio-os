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

- [ ] Спецификация содержит все 19 обязательных разделов Specification-Domain.md, тремя PR (фундамент → поведение → завершение, [Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)).
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

Тот же процесс, что и в TASK-029/030/031/032: три PR, порядок Model First, 20 разделов [Specification-Domain.md](../../.claude/templates/Specification-Domain.md). Как и Task — документирование поверх уже принятого контракта ([internal/domain/project/registry.go](../../internal/domain/project/registry.go)), последняя спецификация этапа 1.

- **PR 1 — фундамент** (сегодня): One Sentence → Identity → Purpose → Responsibilities → Invariants → Lifecycle (Created → Active → Archived) → Relationships (владение Task/Artifact, `Project ├── Task └── Artifact`) → Alternative Interpretations Considered. Delta Review — против Artifact (Reference), Execution, Executor и Task (черновики). Сверка с `registry.go` выявила: контракт не содержит явной команды `Activate` — гипотеза (переход Created→Active при подключении первого Repository) зафиксирована как предложение, не факт; отсутствует операция отключения Repository.
- **PR 2 — поведение**: Domain Events, Commands/Queries (согласованные с уже принятым `Registry`), Examples.
- **PR 3 — завершение и ревью**: Acceptance Criteria, Future Extensions, Anti-Responsibilities, Non-Goals, Removal Test, Decision Log (ADR-013), Open Questions, Stability Assessment; три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md). Это последняя из пяти спецификаций этапа 1 — после утверждения всех пяти EPIC-003 переходит к критериям завершения этапа.

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (пятая по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
2026-07-20 — Architect — введён Three-Pass Review (+4 раздела до 16, три прохода проверки, сознательный темп).
2026-07-20 — Architect — введён Model First (+3 раздела до 19: One Sentence, Identity, Alternative Interpretations Considered; Invariants разделены на Structural/Behavioral; спецификация пишется тремя PR — фундамент/поведение/завершение).
2026-07-20 — Architect — введён Reference Status (+1 раздел Stability Assessment, итого 20; Delta Review обязателен).
2026-07-21 — Claude Code (Developer) — задача взята в работу, переведена в `tasks/in-progress/`; план PR 1 записан.
2026-07-21 — Claude Code (Developer) — PR 1 (`docs/specifications/domain/project.md`, черновик) написан: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle (Created → Active → Archived), Relationships (владение Epic/Task/Artifact), Alternative Interpretations Considered. Delta Review — против Artifact (Reference), Execution, Executor и Task (черновики) — без расхождений. Сверка с `internal/domain/project/registry.go` выявила: контракт не содержит явной команды `Activate` — гипотеза о переходе Created→Active при подключении первого Repository зафиксирована как предложение, требующее подтверждения, не как решённый факт; операция отключения Repository в контракте отсутствует. Оба зафиксированы как Open Questions, код не изменён. Содержание требует реального ревью, не самоапрува. Ветка `feature/TASK-033-spec-project` — последняя в стеке TASK-030 → ... → TASK-033; коммит `fa13309`.
2026-07-21 — Claude Code (Developer) — PR 2 написан: Domain Events (три события на переходах Lifecycle плюс `RepositoryConnected`, обоснованный как значимый за пределами Project операционный факт), Commands (согласованы с уже принятым `registry.go`: Create/ConnectRepository/Archive, новых команд не введено), Queries (концептуальный минимум — контракта пока нет в коде), Examples (четыре сценария, включая ещё не активный проект и проект с несколькими репозиториями). Коммит `7abd8cf`.
2026-07-21 — Claude Code (Developer) — PR 3 написан: Acceptance Criteria (8 критериев), Future Extensions (ожидаемые: Activate/контракт назначений ролей/отключение Repository/ADR-013; возможные: квоты/иерархия/настройки), Anti-Responsibilities (5 пунктов), Non-Goals (4 пункта), Removal Test, Decision Log (7 строк), Open Questions (4 вопроса) плюс письменные ответы на три диагностических вопроса Three-Pass Review, включая Delta Review против всех четырёх предыдущих спецификаций (без пересмотра, понятия единообразны, дублирующих понятий нет). Stability Assessment: **Provisional**, Confidence Medium — из-за неподтверждённого механизма Created→Active и ожидающего ADR-013. Статус документа оставлен «Черновик». Три PR завершены за один день — все пять спецификаций этапа 1 EPIC-003 теперь имеют полный черновик (Artifact — Reference; Execution/Executor/Task/Project — Черновик, ждут ревью). Готово к реальному ревью архитектора.
