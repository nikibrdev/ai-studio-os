# TASK-034: Реализация домен-модуля Artifact

## Тип

feature

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 2 (Реализация)

## Цель

Первая реализация реальной доменной логики Domain Layer: пакет `internal/domain/artifact` строго по утверждённой спецификации ([docs/specifications/domain/artifact.md](../../docs/specifications/domain/artifact.md), статус Reference) — сущность Artifact, её инварианты как проверяемый код, а не только интерфейсы (в отличие от `internal/domain/{task,project,workflow}`, где логика пока не реализована).

## Контекст

Этап 1 EPIC-003 закрыт: все пять доменных спецификаций утверждены (Artifact — Reference; Execution/Executor/Task/Project — Утверждена). Этап 2 открывается отдельными задачами в том же порядке проектирования; эта задача — первая из них, поскольку Artifact уже Reference и не имеет открытых расширений контракта (в отличие от Task/Project, где Final Architecture Review решил расширить существующие контракты). Критическая бизнес-логика покрывается unit-тестами обязательно ([CONSTITUTION.md](../../CONSTITUTION.md), [testing.md](../../docs/development/testing.md)) — это не привязано к v0.4 QA Engine, а действует с момента появления реальной доменной логики, то есть с этой задачи.

## Scope

### Входит

- `internal/domain/artifact/` — value-типы (Type, Origin, Author, State), сущность Artifact с инвариантами, применяемыми в коде, доменные события как структуры данных (Created, Published, Archived), конструктор и методы команд (New/UpdateDraft/Publish/Archive), каждый — прямое отражение соответствующего раздела спецификации (Invariants, Lifecycle, Commands, Domain Events).
- Unit-тесты (`go test`, без внешних фреймворков — не входят в [stack.md](../../.claude/context/stack.md)), покрывающие каждый Structural/Behavioral инвариант и каждый пункт Acceptance Criteria спецификации.
- README пакета по образцу уже существующих доменных модулей.

### Не входит

- Commands/Queries-интерфейсы в стиле `internal/domain/{task,project}` — преждевременная абстракция без конкретного потребителя (Infrastructure/Application для Artifact не начаты); при появлении реального потребителя — отдельная задача.
- Публикация событий через Event Bus (`internal/platform`) — платформенная проводка, вне Domain Layer.
- Реализация Execution/Executor/Task/Project — отдельные задачи этапа 2, в том же порядке проектирования.

## Критерии приёмки

- [x] Пакет `internal/domain/artifact` реализует все пять Structural и четыре Behavioral инварианта спецификации, каждый — проверяемым кодом, а не комментарием.
- [x] Lifecycle (Draft → Published → Archived, включая прямой Draft → Archived) реализован с запретом недопустимых переходов.
- [x] Все четыре команды (Create/UpdateDraft/Publish/Archive — как New/UpdateDraft/Publish/Archive в коде) реализованы и покрыты тестами на успешный и на каждый запрещённый сценарий.
- [x] Unit-тесты покрывают каждый инвариант и каждый пункт Acceptance Criteria спецификации; тесты детерминированы, без зависимости от времени/сети/порядка выполнения ([testing.md](../../docs/development/testing.md)).
- [x] `make verify` — чисто (fmt, lint, vet, test, markdownlint, docs-check).
- [x] README пакета создан по образцу существующих модулей.

## Затрагиваемые модули и документы

- `internal/domain/artifact/` (новый пакет).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от утверждённой спецификации Artifact (TASK-029, Reference); этап 1 EPIC-003 закрыт

## План реализации

Прямой перевод утверждённой спецификации ([artifact.md](../../docs/specifications/domain/artifact.md)) в Go-код, без Commands/Queries-интерфейсов (см. «Не входит» — преждевременная абстракция без потребителя).

1. `internal/domain/artifact/doc.go` — package doc, ссылка на спецификацию, отметка «первый пакет Domain Layer с реальной логикой».
2. `internal/domain/artifact/types.go` — value-типы: `Type` (string), `Origin` (string + три именованные константы из Examples: produced/imported/uploaded — полный перечень остаётся открытым вопросом спецификации, не решается здесь), `Author` (string + `AuthorUnknown`), `State` (Draft/Published/Archived).
3. `internal/domain/artifact/events.go` — структуры данных `Created`, `Published`, `Archived` — прямое отражение раздела Domain Events.
4. `internal/domain/artifact/artifact.go` — тип `Artifact` (Identifier, ProjectID, Type, Origin, Author, CreatedAt, ProducedBy, Payload, State — Structural Invariant 2 + Relationships «Project владеет»), конструктор `New` и методы `UpdateDraft`/`Publish`/`Archive`, каждый — прямое отражение Commands и обеих групп Invariants; сигнатурные ошибки как sentinel-значения (`errors.New`, без внешних зависимостей).
5. `internal/domain/artifact/artifact_test.go` — таблица тестов на каждый Structural/Behavioral инвариант и каждый пункт Acceptance Criteria: успешный путь Draft→Published→Archived, прямой Draft→Archived, все запрещённые переходы, неизменность Type/Origin, Payload после Archive («не означает удалён»), корректность данных каждого события.
6. `internal/domain/artifact/README.md` — по образцу `internal/domain/{task,project,workflow}/README.md`.
7. `make verify` перед сдачей.

Открытый в спецификации вопрос (полный перечень значений Origin) не решается — используются только три примера, названные в самой спецификации (Examples), тип `Origin` остаётся открытым (`string`), не закрытым перечислением.

## История

2026-07-21 — Architect — этап 2 EPIC-003 открыт; задача поставлена в очередь (первая по порядку проектирования, поскольку Artifact уже Reference и не требует решений по расширению контракта, в отличие от Task/Project).
2026-07-21 — Claude Code (Developer) — задача взята в работу, переведена в `tasks/in-progress/`; план записан («План реализации»).
2026-07-21 — Architect — план одобрен без замечаний: прямой перевод спецификации в код, без преждевременных Commands/Queries-интерфейсов — согласуется с «Не входит» задачи. Приступать к реализации.
2026-07-21 — Claude Code (Developer) — пакет реализован: `doc.go`, `types.go` (Type/Origin/Author/State), `events.go` (Created/Published/Archived), `artifact.go` (сущность + New/UpdateDraft/Publish/Archive), `artifact_test.go` (18 тестов, 97.1% покрытия), `README.md`. `internal/domain/README.md` синхронизирован (добавлена строка `artifact/`, статус этапа 1 обновлён). `make verify` — чисто.
2026-07-21 — Architect — Code Review: реализация точно следует спецификации — инварианты, Lifecycle, обе группы ошибок (терминальность/незаполненный Payload) проверяются в правильном порядке (терминальные состояния — раньше content-специфичных ошибок); `UpdateDraft` не мутирует состояние при отклонённом вызове (проверка состояния — до записи полей); отказ от Commands/Queries-интерфейсов обоснован и соответствует плану. Замечаний нет. Approve.
2026-07-21 — Architect — Задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-034 — первая реализация реальной доменной логики Domain Layer, пакет `internal/domain/artifact` (EPIC-003, этап 2).
2. **Что сделано:** реализована сущность Artifact строго по утверждённой спецификации (Reference) — value-типы (Type/Origin/Author/State), Lifecycle (Draft → Published → Archived, включая прямой Draft → Archived) с запретом недопустимых переходов, четыре команды (New/UpdateDraft/Publish/Archive), три доменных события как значения (Created/Published/Archived). Осознанно не введены Commands/Queries-интерфейсы (нет потребителя). Первый пакет Domain Layer с настоящей, проверяемой в коде логикой — остальные пять доменных пакетов (`task`, `project`, `event`, `workflow`, `shared`) по-прежнему только контракты.
3. **Изменённые файлы:** `internal/domain/artifact/{doc,types,events,artifact,artifact_test}.go`, `internal/domain/artifact/README.md` (новые); `internal/domain/README.md` (обновлён), файл задачи.
4. **Как проверялось:** `make verify` целиком — `gofumpt`, `golangci-lint` (0 issues), `go vet`, `go test ./...` (18 тестов пакета `artifact`, 97.1% покрытия, остальные пакеты без тестов не пострадали), `markdownlint-cli2`, `verify-docs.sh` — все чисто.
5. **Обновлённая документация:** `internal/domain/README.md`, `internal/domain/artifact/README.md`.
6. **Open Questions:** унаследованы из спецификации, не решались в этой задаче — полный перечень значений Origin (тип оставлен открытой строкой); кардинальность истории нескольких Execution на одном Artifact; полный состав Metadata сверх минимума.
7. **Рекомендации:** реализация Execution — следующая задача этапа 2, в том же порядке проектирования; при её выполнении учесть уже установленный в этой задаче стиль (сущность + методы-команды, возвращающие события как значения, без преждевременных Commands/Queries-интерфейсов, пока нет реального потребителя).
