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

- [x] Порты покрывают потребности use-case'ов TASK-041…045 (Get/Save по агрегату; листинги — только то, что нужно сценариям).
- [x] Конверт событий реализует `platform.Event`, типы — константы `internal/domain/event`; `SchemaVersion` = 1.
- [x] In-memory фейки потокобезопасны не обязаны быть (тесты последовательны), но детерминированы.
- [x] Decision-документ и README созданы; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/` (новый), `engineering/decisions/`, README.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — EPIC-003 закрыт; ADR-002/008/011/014/015 приняты

## План реализации

1. Decision-документ `2026-07-21-application-ports-placement.md`: почему порты живут в `internal/application`, не в `internal/platform` (ADR-015 — платформа домен-независима, а порты хранения оперируют доменными типами агрегатов).
2. `internal/application/ports.go` — пять узких Get/Save интерфейсов (Project/Task/Executor/Execution/Artifact), `ErrNotFound` sentinel.
3. `internal/application/event.go` — `Envelope`, реализующий `platform.Event`; ID — случайный hex через `crypto/rand` (без внешней UUID-зависимости, её нет в stack.md).
4. `internal/application/inmemory/` — обобщённый `Store[T]` (Get/Save по карте с мьютексом) + пять конструкторов-фасадов под интерфейсы TASK 2; отдельный подпакет, поскольку фейки нужны тестам TASK-041…045, не только этой задаче.
5. README `internal/application`.
6. `make verify`, PR, CI, merge.

Обобщённый `Store[T]` — не преждевременная абстракция: пять почти идентичных реализаций Get/Save заменяются одной, конструкторы остаются типобезопасными фасадами.

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (первая).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен; обобщённый Store[T] для пяти почти идентичных in-memory фейков — оправданное упрощение, не преждевременная абстракция. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: decision-документ о размещении портов; `ports.go` (пять интерфейсов + ErrNotFound); `event.go` (Envelope, реализующий platform.Event; ID — crypto/rand hex, без внешней UUID-зависимости); `inmemory/` (обобщённый Store[T] + пять типобезопасных конструкторов); тесты (event_test.go, inmemory/stores_test.go, включая компиляционные проверки соответствия портам). README слоя переписан (был плейсхолдером с 2026-07-19). `make verify` — чисто.
2026-07-21 — Architect — Code Review: размещение портов согласуется с ADR-015 буквально; ID событий через crypto/rand — оправданный выбор без новой зависимости; generic Store[T] — оправданное упрощение, не преждевременная абстракция; компиляционные проверки `var _ application.XStore = ...` в тестах защищают от расхождения фейков с портами при будущих правках сигнатур. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-040 — каркас Application Layer (EPIC-004, первая задача).
2. **Что сделано:** порты хранения пяти агрегатов, конверт доменных событий под контракт `platform.Event`, обобщённые in-memory фейки для тестов эпика, decision-документ о размещении портов, README слоя.
3. **Изменённые файлы:** `internal/application/{doc,ports,event,event_test}.go`, `internal/application/inmemory/{doc,store,stores,stores_test}.go`, `internal/application/README.md` (переписан); `engineering/decisions/2026-07-21-application-ports-placement.md` (новый); файл задачи.
4. **Как проверялось:** `go test ./internal/application/... -cover` — 92.3%/85.7%; `make verify` — чисто.
5. **Обновлённая документация:** README `internal/application`, decision-документ.
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-041 (постановка задачи) — первый реальный use-case на этих портах.
