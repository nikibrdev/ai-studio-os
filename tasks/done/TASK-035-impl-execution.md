# TASK-035: Реализация домен-модуля Execution

## Тип

feature

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 2 (Реализация)

## Цель

Пакет `internal/domain/execution` строго по утверждённой спецификации ([docs/specifications/domain/execution.md](../../docs/specifications/domain/execution.md)) — сущность Execution с инвариантами, применяемыми в коде: единичный запуск Executor'а, производящий Artifact и несущий статус исполнения.

## Контекст

Вторая задача этапа 2 EPIC-003, в порядке проектирования (Artifact уже реализован — TASK-034, образец стиля). Спецификация Execution утверждена 2026-07-21 (TASK-030); ключевое решение финального ревью — гонка Fail/Abort разрешается порядком выполнения команд (Behavioral Invariant 5), не отдельным доменным правилом. Согласно [ADR-015](../../docs/adr/ADR-015-internal-layering.md) доменные модули не зависят друг от друга — ссылки на Artifact и Task выражаются идентификаторами-строками, не импортом пакетов.

## Scope

### Входит

- `internal/domain/execution/` — value-типы (State), сущность Execution, команды New (Create)/Accept/RecordArtifact/Succeed/Fail/Abort, события Queued/Started/Succeeded/Failed/Aborted как структуры данных.
- Unit-тесты на каждый Structural/Behavioral инвариант и каждый пункт Acceptance Criteria спецификации.
- README пакета; обновление `internal/domain/README.md`.

### Не входит

- Commands/Queries-интерфейсы — нет потребителя (то же решение, что в TASK-034).
- Публикация событий через Event Bus — платформенная проводка.
- Политика повторных попыток, тайм-аут Queued — Open Questions спецификации, вне домена.

## Критерии приёмки

- [x] Все пять Structural и пять Behavioral инвариантов спецификации реализованы проверяемым кодом.
- [x] Lifecycle (Queued → Running → Succeeded | Failed | Aborted, плюс Queued → Aborted) с запретом недопустимых переходов; терминальные состояния необратимы.
- [x] RecordArtifact допустим только в Running (Behavioral Invariant 4); после терминального перехода множество Artifact неизменяемо (Behavioral Invariant 1).
- [x] Гонка Fail/Abort: первый терминальный переход выигрывает, второй получает ошибку (Behavioral Invariant 5).
- [x] Unit-тесты детерминированы, покрывают успешные и все запрещённые сценарии каждой команды.
- [x] `make verify` — чисто; README пакета создан.

## Затрагиваемые модули и документы

- `internal/domain/execution/` (новый пакет); `internal/domain/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — спецификация Execution утверждена (TASK-030); стиль — по образцу artifact (TASK-034)

## План реализации

По образцу `internal/domain/artifact` (утверждённый стиль этапа 2):

1. `doc.go` — package doc со ссылкой на спецификацию.
2. `types.go` — `State` (Queued/Running/Succeeded/Failed/Aborted) с предикатом терминальности.
3. `events.go` — `Queued`, `Started`, `Succeeded`, `Failed`, `Aborted` (данные — по разделу Domain Events спецификации; Succeeded/Failed несут ссылки на произведённые Artifact; Aborted несёт состояние-источник Queued|Running).
4. `execution.go` — сущность (id, taskID, executorID — строки, ADR-015; createdAt; artifactIDs; state), конструктор `New` и методы `Accept`/`RecordArtifact`/`Succeed`/`Fail`/`Abort`; sentinel-ошибки (ErrMissingField, ErrNotQueued, ErrNotRunning, ErrTerminal).
5. `execution_test.go` — таблица тестов: happy path, Queued→Aborted, все запрещённые переходы, RecordArtifact вне Running, неизменяемость множества Artifact после терминала, гонка Fail-затем-Abort и Abort-затем-Fail.
6. README пакета + строка в `internal/domain/README.md`.
7. `make verify`, PR, зелёный CI, merge.

## История

2026-07-21 — Architect — этап 2 EPIC-003: задача поставлена в очередь (вторая по порядку проектирования).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен: стиль artifact, ссылки-строки по ADR-015, без Commands/Queries. Приступать.
2026-07-21 — Claude Code (Developer) — пакет реализован: `doc.go`, `types.go` (State + предикат Terminal), `events.go` (Queued/Started/Succeeded/Failed/Aborted), `execution.go` (сущность + New/Accept/RecordArtifact/Succeed/Fail/Abort, копирование множества Artifact в терминальных событиях и аксессорах), `execution_test.go` (17 тестов, 95.9% покрытия, включая обе стороны гонки Fail/Abort и защиту от мутации слайсов), `README.md`. `internal/domain/README.md` синхронизирован. `make verify` — чисто.
2026-07-21 — Architect — Code Review: реализация следует спецификации; порядок проверок (терминальность → состояние → контент) единообразен с artifact; Failed несёт произведённые до сбоя Artifact (пример TestReport из спецификации покрыт тестом); копирование слайсов защищает Behavioral Invariant 1 от мутации через события. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-035 — реализация домен-модуля Execution (EPIC-003, этап 2, вторая задача).
2. **Что сделано:** пакет `internal/domain/execution` строго по утверждённой спецификации — сущность Execution с фиксированными ссылками (TaskID/ExecutorID — строки, ADR-015), Lifecycle с двумя путями в Aborted, шесть команд, пять событий; гонка Fail/Abort разрешена порядком выполнения (первый терминальный переход выигрывает).
3. **Изменённые файлы:** `internal/domain/execution/{doc,types,events,execution,execution_test}.go`, `internal/domain/execution/README.md` (новые); `internal/domain/README.md` (обновлён); файл задачи.
4. **Как проверялось:** `go test ./internal/domain/execution/... -cover` — 17 тестов, 95.9% покрытия; `make verify` целиком — чисто.
5. **Обновлённая документация:** `internal/domain/execution/README.md`, `internal/domain/README.md`.
6. **Open Questions:** нет новых; открытые вопросы спецификации (тайм-аут Queued, политика повторов) — вне домена, не решались.
7. **Рекомендации:** следующая задача этапа 2 — TASK-036 (`internal/domain/executor`), тот же стиль.
