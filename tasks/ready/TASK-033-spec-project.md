# TASK-033: Спецификация домен-модуля Project

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/project.md` по шаблону [Specification.md](../../.claude/templates/Specification.md) — техническое задание для реализации `internal/domain/project`, чьи контракты уже частично существуют ([internal/domain/project/registry.go](../../internal/domain/project/registry.go), EPIC-002) без полной спецификации.

## Контекст

Project — пятый, последний модуль в порядке проектирования этапа 1: граница, внутри которой существуют Task и Artifact (`Project ├── Task └── Artifact`, [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)). `Registry` (Created → Active → Archived) уже принят; задача — полная спецификация поверх него, с явным описанием владения Task и Artifact.

## Scope

### Входит

- `docs/specifications/domain/project.md`: назначение; требования (согласованные с уже принятым `Registry`); чего модуль НЕ делает; сценарии использования (создание, подключение репозитория, архивирование); инварианты (минимум один Repository, архив неизменяем); допустимые состояния (Created → Active → Archived); события; ограничения (формат подключения репозиториев — [ADR-013](../../docs/adr/ADR-013-managed-projects.md), Decision Required — зафиксировать как ограничение, не решать здесь); будущие расширения; Acceptance Criteria.
- Сверка с текущим кодом `internal/domain/project/registry.go` — расхождения фиксируются как Open Questions, код не меняется в рамках этой задачи.

### Не входит

- Изменение `internal/domain/project/registry.go` — только документирование.
- Спецификации Artifact/Execution/Executor/Task.

## Критерии приёмки

- [ ] Спецификация содержит все обязательные разделы шаблона.
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
