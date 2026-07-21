# TASK-040: Каркас Application Layer — порты, конверт событий

## Тип

feature

## Эпик

[EPIC-004 Application Layer](../../docs/roadmap/EPIC-004-application-layer.md)

## Цель

Фундамент слоя `internal/application`: порты хранения агрегатов (интерфейсы), конверт событий (реализация `platform.Event` поверх доменных событий с каноническими типами из `internal/domain/event`), in-memory фейки для тестов, README слоя. Решение о размещении портов в Application Layer фиксируется decision-документом.

## Контекст

`internal/platform` домен-независим (ADR-015) и не может держать интерфейсы с доменными типами; доменные модули остаются чистыми. Порты хранения объявляет слой, которому они нужны, — Application (hexagonal, driven ports); реализации — EPIC-005 (PostgreSQL, ADR-004/011).

## Scope

### Входит

- `internal/application/ports.go` (или по пакету на use-case — уточнить в плане): интерфейсы `TaskStore`, `ProjectStore`, `ExecutionStore`, `ExecutorStore`, `ArtifactStore` — узкие, по агрегату.
- Конверт событий: тип, реализующий `platform.Event` (ID, Type, SchemaVersion=1, OccurredAt, Source, Actor, ProjectID, SubjectID) + конструктор из доменного события.
- In-memory фейки портов (пакет для тестов эпика).
- `engineering/decisions/2026-07-21-application-ports-placement.md`.
- README `internal/application`.

### Не входит

- Сами use-case'ы (TASK-041…044), проекции (TASK-045), инфраструктурные реализации портов (EPIC-005).

## Критерии приёмки

- [ ] Порты покрывают потребности use-case'ов TASK-041…045 (Get/Save по агрегату; листинги — только то, что нужно сценариям).
- [ ] Конверт событий реализует `platform.Event`, типы — константы `internal/domain/event`; `SchemaVersion` = 1.
- [ ] In-memory фейки потокобезопасны не обязаны быть (тесты последовательны), но детерминированы.
- [ ] Decision-документ и README созданы; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/` (новый), `engineering/decisions/`, README.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — EPIC-003 закрыт; ADR-002/008/011/014/015 приняты

## План реализации

<Заполняется при взятии задачи в работу.>

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (первая).
