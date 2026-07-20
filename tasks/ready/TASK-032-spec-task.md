# TASK-032: Спецификация домен-модуля Task

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/task.md` по шаблону [Specification.md](../../.claude/templates/Specification.md) — техническое задание для реализации `internal/domain/task`, чьи контракты уже частично существуют ([internal/domain/task/contracts.go](../../internal/domain/task/contracts.go), EPIC-002) без полной спецификации.

## Контекст

Task — четвёртый модуль в порядке проектирования (не первый, по решению архитектора: Task — способ организовать работу, не самоцель, [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)). Контракты `Commands`/`Queries`/`Exporter` уже приняты ([ADR-004](../../docs/adr/ADR-004-task-storage.md)); задача — задокументировать полную спецификацию поверх них, зафиксировать связь с Execution (Task создаёт Execution, TASK-030) и Artifact (Task не владеет Artifact напрямую — им владеет Project, [domain-model.md](../../docs/architecture/domain-model.md)).

## Scope

### Входит

- `docs/specifications/domain/task.md`: назначение; требования (согласованные с уже принятыми `Commands`/`Queries`/`Exporter`); чего модуль НЕ делает; сценарии использования; инварианты; допустимые состояния — ссылка на канонический [state-machine.md](../../docs/architecture/state-machine.md) (9 состояний, не дублируется); события — 15 событий жизненного цикла из [events.md](../../docs/architecture/events.md); ограничения (ADR-004, ADR-011 формат идентификаторов — Decision Required, зафиксировать как ограничение); будущие расширения; Acceptance Criteria.
- Сверка с текущим кодом `internal/domain/task/contracts.go` — если спецификация выявит расхождение с уже принятыми решениями, расхождение фиксируется как Open Question, код не меняется в рамках этой задачи.

### Не входит

- Изменение `internal/domain/task/contracts.go` — только документирование.
- Спецификации Artifact/Execution/Executor/Project.

## Критерии приёмки

- [ ] Спецификация содержит все обязательные разделы шаблона.
- [ ] Согласована с уже принятым кодом `internal/domain/task` — расхождения (если есть) зафиксированы как Open Questions, не решены явочным порядком.
- [ ] Непротиворечива с ADR-004, ADR-011, state-machine.md, утверждённой спецификацией Execution (TASK-030).
- [ ] Статус спецификации — «Утверждена».
- [ ] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Затрагиваемые модули и документы

`docs/specifications/domain/task.md` (новый); `internal/domain/task/` (только сверка, без правок).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от утверждённой спецификации Execution (TASK-030)

## План реализации

<Заполняется при взятии задачи в работу.>

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (четвёртая по порядку проектирования).
