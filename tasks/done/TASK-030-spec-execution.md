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

- [x] Спецификация содержит все 20 обязательных разделов Specification-Domain.md, тремя PR (фундамент → поведение → завершение, [Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)).
- [x] Пройдены три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).
- [x] Непротиворечива с ADR-005, ADR-016, domain-model.md и утверждённой спецификацией Artifact (TASK-029).
- [x] Статус спецификации — «Утверждена».
- [x] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

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
2026-07-21 — Claude Code (Developer) — PR 1 (`docs/specifications/domain/execution.md`, черновик) написан: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle (Queued → Running → Succeeded | Failed | Aborted, плюс Queued → Aborted), Relationships, Alternative Interpretations Considered. Delta Review относительно Artifact проведён: терминология «ссылка, не владение» и кардинальность связи Execution↔Artifact выровнены с уже утверждённой спецификацией, не переопределены заново. Три вопроса, требующие решения архитектора, зафиксированы в Open Questions, не решены самостоятельно. Содержание — самостоятельная доменная модель, требует реального ревью, не самоапрува; merge не выполняется до обратной связи. Локальная верификация (`verify-docs.sh`, `markdownlint-cli2`) пройдена; коммит `dad9256`.
2026-07-21 — Claude Code (Developer) — PR 2 написан: Domain Events (пять, по одному на переход/создание — согласовано с принципом Artifact через Delta Review), Commands (явное соответствие четырём возможностям ADR-005: Accept/RecordArtifact/Succeed+Fail/Abort), Queries, Examples (пять содержательных сценариев, включая человека как Executor и повторную попытку после сбоя). Локальная верификация пройдена; коммит `2e84101`.
2026-07-21 — Claude Code (Developer) — PR 3 написан: Acceptance Criteria (8 критериев с самопроверкой), Future Extensions (ожидаемые: метрики Executor/тайм-аут Queued/политика повторов; возможные: Attempt/параллельные под-исполнения/приоритизация очереди/структура прогресс-лога), Anti-Responsibilities (5 пунктов), Non-Goals (4 пункта), Removal Test, Decision Log (7 строк), Open Questions (3 подлинно открытых вопроса) плюс письменные ответы на три диагностических вопроса Three-Pass Review, включая обязательный Delta Review против Artifact (все три ответа — модель не требует пересмотра, понятия использованы единообразно, дублирующих понятий не введено). Stability Assessment: Provisional-leaning Stable for v1, Confidence Medium (ниже, чем у Artifact, из-за открытой политики повторов, зависящей от будущей спецификации Task). Статус документа оставлен «Черновик» — финальное решение об «Утверждена» не принимается самостоятельно. Локальная верификация пройдена; коммит записывается отдельно. Три PR завершены за один день, как и TASK-029; ветка `feature/TASK-030-spec-execution` готова к реальному ревью архитектора, merge не выполняется до обратной связи.
2026-07-21 — Architect — Final Architecture Review: содержательное замечание к Open Questions PR 3 — граница Failed/Aborted, оставленная открытой, на самом деле разрешима на уровне домена: гонка команд — вопрос порядка доставки (Application/Infrastructure), не отдельное доменное правило, которое стоило бы оставлять неопределённым. Остальные два открытых вопроса (порог Queued, детальная политика повторов) — корректно вне домена/future work, не блокируют утверждение. Решение: добавить Behavioral Invariant, разрешающий гонку порядком выполнения; после этого — Approve.
2026-07-21 — Claude Code (Developer) — добавлен Behavioral Invariant 5 (Execution): гонка Fail/Aborted разрешается порядком выполнения команд в едином пути записи, не отдельной доменной гонкой; Decision Log дополнен; Open Questions сокращён до двух подлинно открытых, некритичных для утверждения пунктов. Локальная верификация пройдена, коммит `<pending>`.
2026-07-21 — Architect — Approve. Статус спецификации выставлен «Утверждена» (не Reference — этот статус остаётся исключительным для Artifact). Задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-030 — спецификация домен-модуля Execution, вторая спецификация Domain Layer (EPIC-003, этап 1).
2. **Что сделано:** написана и утверждена полная спецификация `docs/specifications/domain/execution.md` — 20 разделов, тремя PR, с обязательным Delta Review против Artifact (Reference). Единственное содержательное замечание финального ревью (граница Fail/Aborted) устранено добавлением Behavioral Invariant 5. Итоговый статус — **Утверждена**.
3. **Изменённые файлы:** `docs/specifications/domain/execution.md` (новый), файл задачи.
4. **Как проверялось:** на каждом PR — `gofumpt`/`golangci-lint`/`go vet` (чисто, Go-код не менялся), `bash scripts/verify-docs.sh` (0 ошибок), `npx markdownlint-cli2` (0 issues); финально — три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md) с письменными ответами.
5. **Обновлённая документация:** см. «Изменённые файлы» — реализация `internal/domain/execution` — отдельная задача этапа 2.
6. **Open Questions:** два подлинно открытых вопроса остаются в самой спецификации (порог Queued — Infrastructure/Application; детальная политика повторов — Workflow/Task future work) — не блокируют утверждение, не решены самостоятельно.
7. **Рекомендации:** приступить к оставшимся спецификациям этапа 1 (Executor уже написан параллельно, ждёт этого же ревью); при реализации `internal/domain/execution` (этап 2) учесть оба открытых вопроса как явные конфигурационные решения Application Layer, не как доменные пробелы.
