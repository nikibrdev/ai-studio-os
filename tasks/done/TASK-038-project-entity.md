# TASK-038: Расширение и реализация домен-модуля Project

## Тип

feature

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 2 (Реализация)

## Цель

Закрыть решённое финальным ревью спецификации ([docs/specifications/domain/project.md](../../docs/specifications/domain/project.md), Decision Log) расширение контракта — явная команда `Activate` в `Registry` — и реализовать сущность Project с инвариантами в коде: граница инициативы, владеющая Epic/Task/Artifact, с Lifecycle Created → Active → Archived.

## Контекст

Пятая задача этапа 2 EPIC-003. Спецификация Project утверждена 2026-07-21 (TASK-033); ключевое решение финального ревью — переход Created → Active выполняется явной командой Activate с guard-условием «есть хотя бы один подключённый Repository», а не побочным эффектом ConnectRepository. Отсутствие операции отключения Repository — сознательное ограничение v1 (подключение — необратимая часть истории Project).

## Scope

### Входит

- Расширение `registry.go`: команда `Activate(ctx, projectID)` в интерфейсе `Registry`.
- Сущность Project (`project.go`): id, название, набор подключённых Repository (только растёт), состояние Created | Active | Archived; методы New/ConnectRepository/Activate/Archive; предикат допустимости создания контента (Behavioral Invariant 4: новые Epic/Task/Artifact — только в Active).
- События Created/RepositoryConnected/Activated/Archived как значения.
- Unit-тесты; README модуля; `internal/domain/README.md`.

### Не входит

- Отключение Repository — сознательное ограничение v1 (Decision Log спецификации).
- Контракт назначений исполнителей ролей — анонсированная future work, не специфицирована.
- Формат подключения репозиториев — ADR-013 (Decision Required); Repository — строка-ссылка до его принятия.
- Queries-контракт — концептуальный минимум спецификации реализуется на этапе появления потребителя.

## Критерии приёмки

- [x] `Registry` содержит `Activate`; guard «≥1 Repository» применяется в сущности.
- [x] Все три Structural и четыре Behavioral инварианта спецификации в коде; Archived терминален; ConnectRepository недопустим в Archived; набор Repository только растёт.
- [x] Предикат Behavioral Invariant 4 (создание контента только в Active) выражен в коде.
- [x] Unit-тесты покрывают успешные и запрещённые сценарии; `make verify` — чисто; README обновлён.

## Затрагиваемые модули и документы

- `internal/domain/project/` (расширение контракта + новые файлы сущности); `internal/domain/project/README.md`, `internal/domain/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — спецификация Project утверждена (TASK-033); направление расширения решено её Decision Log

## План реализации

1. `registry.go` — добавить `Activate(ctx context.Context, projectID string) error` в `Registry` (breaking change допустим — реализаций нет).
2. `events.go` — `Created`, `RepositoryConnected`, `Activated`, `Archived`.
3. `project.go` — сущность + New/ConnectRepository/Activate/Archive + предикат `AcceptsNewContent()`; sentinel-ошибки (ErrMissingField, ErrArchived, ErrAlreadyActive, ErrNoRepository, ErrNotActive).
4. `project_test.go` — таблица тестов, включая guard Activate без Repository и повторное подключение того же Repository (идемпотентный no-op — тот же паттерн, что GrantRole в executor).
5. README модуля, README слоя, `make verify`, PR, CI, merge.

## История

2026-07-21 — Architect — этап 2 EPIC-003: задача поставлена в очередь (пятая по порядку проектирования); направление расширения контракта решено Decision Log спецификации Project.
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен; идемпотентность повторного ConnectRepository того же репозитория — уточнение реализации по образцу GrantRole (executor), спецификацию не меняет. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: `registry.go` расширен командой Activate (док-комментарий явно фиксирует, что ConnectRepository не меняет состояние сам), `events.go` (Created/RepositoryConnected/Activated/Archived), `project.go` (сущность + New/ConnectRepository/Activate/Archive + предикат AcceptsNewContent; ConnectRepository возвращает признак added для различения no-op), `project_test.go` (11 тестов, 91.7% покрытия, включая guard без Repository и проверку, что первое подключение НЕ переводит состояние). README модуля переписан, `internal/domain/README.md` синхронизирован. `make verify` — чисто.
2026-07-21 — Architect — Code Review: явный Activate с guard-условием реализован точно по решению финального ревью спецификации; тест TestConnectRepository_DoesNotTransitionState прямо защищает это решение от регрессии; сохранение репозиториев после Archive соответствует «архив неизменяем, не удалён». Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-038 — расширение и реализация домен-модуля Project (EPIC-003, этап 2, пятая задача).
2. **Что сделано:** контракт `Registry` расширен явной командой Activate (решение Decision Log спецификации); реализована сущность Project — Lifecycle Created → Active → Archived с guard-условием активации «≥1 Repository», необратимым архивом и предикатом AcceptsNewContent (Behavioral Invariant 4). Все пять доменных сущностей EPIC-003 теперь реализованы.
3. **Изменённые файлы:** `internal/domain/project/{registry,events,project,project_test}.go`, `internal/domain/project/README.md`; `internal/domain/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/domain/project/... -cover` — 11 тестов, 91.7%; `make verify` — чисто.
5. **Обновлённая документация:** README модуля project, README слоя domain.
6. **Open Questions:** формат подключения репозиториев — ADR-013 (Decision Required); контракт назначений ролей — future work.
7. **Рекомендации:** TASK-039 (workflow.Rules) — последняя задача кода этапа 2; после неё — сквозной тест фазы и закрытие эпика.
