# TASK-032: Спецификация домен-модуля Task

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/task.md` по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) (Domain Specification Review — 12 обязательных разделов, [решение](../../engineering/decisions/2026-07-20-domain-specification-review.md)) — техническое задание для реализации `internal/domain/task`, чьи контракты уже частично существуют ([internal/domain/task/contracts.go](../../internal/domain/task/contracts.go), EPIC-002) без полной спецификации.

## Контекст

Task — четвёртый модуль в порядке проектирования (не первый, по решению архитектора: Task — способ организовать работу, не самоцель, [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)). Контракты `Commands`/`Queries`/`Exporter` уже приняты ([ADR-004](../../docs/adr/ADR-004-task-storage.md)); задача — задокументировать полную спецификацию поверх них, зафиксировать связь с Execution (Task создаёт Execution, TASK-030) и Artifact (Task не владеет Artifact напрямую — им владеет Project, [domain-model.md](../../docs/architecture/domain-model.md)).

## Scope

### Входит

- `docs/specifications/domain/task.md`, все 12 обязательных разделов: Purpose, Responsibilities, Invariants, Lifecycle (ссылка на канонический [state-machine.md](../../docs/architecture/state-machine.md), 9 состояний, не дублируется), Relationships, Domain Events (15 событий жизненного цикла из [events.md](../../docs/architecture/events.md)), Commands/Queries (согласованные с уже принятыми `Commands`/`Queries`/`Exporter`), Acceptance Criteria, Future Extensions, Anti-Responsibilities, Decision Log (ADR-004, ADR-011 — формат идентификаторов, Decision Required, зафиксировать как ограничение).
- Сверка с текущим кодом `internal/domain/task/contracts.go` — если спецификация выявит расхождение с уже принятыми решениями, расхождение фиксируется как Open Question, код не меняется в рамках этой задачи.

### Не входит

- Изменение `internal/domain/task/contracts.go` — только документирование.
- Спецификации Artifact/Execution/Executor/Project.

## Критерии приёмки

- [x] Спецификация содержит все 20 обязательных разделов Specification-Domain.md, тремя PR (фундамент → поведение → завершение, [Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)).
- [x] Пройдены три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).
- [x] Согласована с уже принятым кодом `internal/domain/task` — расхождения зафиксированы и разрешены как решённое направление расширения контракта на этапе 2 (Decision Log), не решены явочным порядком в этой задаче.
- [x] Непротиворечива с ADR-004, ADR-011, state-machine.md, утверждённой спецификацией Execution (TASK-030).
- [x] Статус спецификации — «Утверждена».
- [x] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Затрагиваемые модули и документы

`docs/specifications/domain/task.md` (новый); `internal/domain/task/` (только сверка, без правок).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от утверждённой спецификации Execution (TASK-030)

## План реализации

Тот же процесс, что и в TASK-029/030/031: три PR, порядок Model First, 20 разделов [Specification-Domain.md](../../.claude/templates/Specification-Domain.md). Отличие от предыдущих трёх задач: Task — не новая сущность, а документирование поверх уже принятых контрактов ([internal/domain/task/contracts.go](../../internal/domain/task/contracts.go)) и уже канонического жизненного цикла ([state-machine.md](../../docs/architecture/state-machine.md), 9 состояний — не дублируется, только ссылка).

- **PR 1 — фундамент** (сегодня): One Sentence → Identity → Purpose → Responsibilities → Invariants → Lifecycle (ссылка на state-machine.md) → Relationships → Alternative Interpretations Considered. Delta Review — против Artifact (Reference), Execution и Executor (черновики). Сверка с `contracts.go` выявила два расхождения (`Create` не принимает `epicID`; операции записи scope/критериев приёмки отсутствуют в контракте) — зафиксированы как Open Questions, код не меняется.
- **PR 2 — поведение**: Domain Events (15 событий из [events.md](../../docs/architecture/events.md), не изобретаются заново), Commands/Queries (согласованные с уже принятыми `Commands`/`Queries`/`Exporter`), Examples.
- **PR 3 — завершение и ревью**: Acceptance Criteria, Future Extensions, Anti-Responsibilities, Non-Goals, Removal Test, Decision Log (ADR-004, ADR-011), Open Questions, Stability Assessment; три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (четвёртая по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
2026-07-20 — Architect — введён Three-Pass Review (+4 раздела до 16, три прохода проверки, сознательный темп).
2026-07-20 — Architect — введён Model First (+3 раздела до 19: One Sentence, Identity, Alternative Interpretations Considered; Invariants разделены на Structural/Behavioral; спецификация пишется тремя PR — фундамент/поведение/завершение).
2026-07-20 — Architect — введён Reference Status (+1 раздел Stability Assessment, итого 20; Delta Review обязателен).
2026-07-21 — Claude Code (Developer) — задача взята в работу, переведена в `tasks/in-progress/`; план PR 1 записан.
2026-07-21 — Claude Code (Developer) — PR 1 (`docs/specifications/domain/task.md`, черновик) написан: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle (ссылка на канонический state-machine.md, не дублируется), Relationships, Alternative Interpretations Considered. Delta Review — против Artifact (Reference), Execution и Executor (черновики) — без расхождений. Сверка с `internal/domain/task/contracts.go` выявила два расхождения: (1) `Commands.Create` не принимает `epicID` — связь Task↔Epic не выражена в контракте; (2) контракт не содержит операций записи scope/критериев приёмки отдельно от Create — в переходный период эти данные живут в файле задачи, не в контракте. Оба зафиксированы как Open Questions, код не изменён (вне scope этапа 1). Содержание требует реального ревью, не самоапрува. Ветка `feature/TASK-032-spec-task` — часть стека TASK-030 → TASK-031 → TASK-032; коммит `e417e52`.
2026-07-21 — Claude Code (Developer) — PR 2 написан: Domain Events (таблица-сверка всех 15 событий с уже принятым каталогом events.md, без переопределения), Commands/Queries (согласованы с уже принятым `contracts.go` — Create/Transition/State — новых команд не введено), Examples (пять сценариев, включая задачу без Epic и заблокированную задачу). Коммит `b732f8f`.
2026-07-21 — Claude Code (Developer) — PR 3 написан: Acceptance Criteria (8 критериев), Future Extensions (ожидаемые: расширенные Queries/операции scope-AC/формат идентификаторов; возможные: полная спецификация Epic/приоритезация/SLA), Anti-Responsibilities (5 пунктов), Non-Goals (4 пункта), Removal Test, Decision Log (7 строк), Open Questions (4 вопроса, включая оба пробела контракта) плюс письменные ответы на три диагностических вопроса Three-Pass Review, включая Delta Review против Artifact/Execution/Executor (без пересмотра, понятия единообразны, дублирующих понятий нет). Stability Assessment: **Provisional** (не Stable for v1) — сознательно ниже, чем у Execution/Executor, из-за двух нерешённых пробелов контракта. Статус документа оставлен «Черновик». Три PR завершены за один день. Готово к реальному ревью архитектора, включая явное решение по обоим расхождениям с contracts.go.
2026-07-21 — Architect — Final Architecture Review: оба расхождения с `contracts.go` (`epicID`, scope/AC) — не ошибки спецификации, а корректно спроектированная целевая модель, которую контракт этапа 2 должен догнать (Model First: спецификация описывает предметную область, не текущий код буквально). Формат идентификатора (ADR-011) и приоритезация — верно оставлены открытыми, не блокируют. Решение: зафиксировать направление расширения контракта в Decision Log вместо Open Questions; Approve.
2026-07-21 — Claude Code (Developer) — Decision Log дополнен решением о расширении контракта на этапе 2 (`epicID`, отдельные команды scope/AC); Open Questions сокращён до двух подлинно открытых, некритичных для утверждения пунктов (ADR-011, приоритезация); формулировки Commands/Future Extensions/Non-Goals синхронизированы. Локальная верификация пройдена.
2026-07-21 — Architect — Approve. Статус спецификации выставлен «Утверждена». Задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-032 — спецификация домен-модуля Task, четвёртая спецификация Domain Layer (EPIC-003, этап 1).
2. **Что сделано:** написана и утверждена полная спецификация `docs/specifications/domain/task.md` — 20 разделов, тремя PR, с обязательным Delta Review против Artifact/Execution/Executor. Два реальных расхождения с уже принятым `internal/domain/task/contracts.go` (`epicID` в Create, операции scope/AC) не устранялись правкой кода (вне scope этапа 1), а получили явное архитектурное решение: контракт целенаправленно расширяется вслед за спецификацией на этапе 2. Итоговый статус — **Утверждена**.
3. **Изменённые файлы:** `docs/specifications/domain/task.md` (новый), файл задачи. `internal/domain/task/contracts.go` не менялся.
4. **Как проверялось:** на каждом PR — `gofumpt`/`golangci-lint`/`go vet` (чисто, Go-код не менялся), `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто; финально — три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md) с письменными ответами.
5. **Обновлённая документация:** см. «Изменённые файлы» — реализация/расширение `internal/domain/task` — отдельная задача этапа 2.
6. **Open Questions:** два подлинно открытых вопроса остаются (формат идентификатора — ADR-011, Decision Required; приоритезация Backlog/Ready — future work) — не блокируют утверждение.
7. **Рекомендации:** этап 2 для `internal/domain/task` должен явно включать расширение `Create` (`epicID`) и новые команды записи scope/AC — уже решённое направление, не заново открытый вопрос при реализации.
