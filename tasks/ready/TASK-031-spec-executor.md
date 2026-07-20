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

- [ ] Спецификация содержит все 12 обязательных разделов Specification-Domain.md.
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

<Заполняется при взятии задачи в работу.>

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (третья по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
