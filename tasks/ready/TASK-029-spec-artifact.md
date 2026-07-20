# TASK-029: Спецификация домен-модуля Artifact

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/artifact.md` по шаблону [Specification.md](../../.claude/templates/Specification.md) — техническое задание для будущей реализации `internal/domain/artifact` (этап 2, отдельная задача, не начинается без утверждения этой спецификации).

## Контекст

Artifact — первый модуль в порядке проектирования Domain Layer ([domain-model.md](../../docs/architecture/domain-model.md), [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)): самостоятельный Aggregate Root, не часть Execution/Task/Project. Концептуальное описание уже есть в ADR-016 и domain-model.md — задача переводит его в полную спецификацию (требования, сценарии, инварианты, состояния, события, ограничения, Acceptance Criteria), а не изобретает решение заново.

## Scope

### Входит

- `docs/specifications/domain/artifact.md`: назначение; требования (SPEC-ARTIFACT-1, ...); чего модуль НЕ делает; сценарии использования (создание Artifact исполнением, чтение метаданных, архивирование и т.п.); инварианты; допустимые состояния (Created → Stored → Archived, обязательно по требованию Domain Layer); события (публикуемые/потребляемые); ограничения (архитектурные — ADR-016, ADR-015; технологические); будущие расширения; Acceptance Criteria.
- Явное описание разделения Metadata/Payload (по ADR-016) на уровне требований — что обязано хранить ядро (Metadata), что не интерпретирует (Payload).
- Обновление `internal/domain/README.md` (при необходимости) — ссылка на новую спецификацию, без изменения кода.

### Не входит

- Реализация Go-пакета `internal/domain/artifact` — отдельная задача этапа 2, после утверждения.
- Спецификации Execution/Executor/Task/Project — отдельные задачи (TASK-030…033).

## Критерии приёмки

- [ ] Спецификация содержит все обязательные разделы шаблона, включая «Допустимые состояния».
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
