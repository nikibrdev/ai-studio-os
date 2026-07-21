# TASK-051: Composition root + интеграционный golden path, закрытие EPIC-005

## Тип

feature

## Эпик

[EPIC-005 Infrastructure Layer](../../docs/roadmap/EPIC-005-infrastructure-layer.md)

## Цель

Собрать все адаптеры (TASK-046…050) в рабочую систему и доказать результат эпика: `internal/application`'s golden path (тот же сценарий, что и `TestGoldenPath_Application` из EPIC-004) проходит на реальных PostgreSQL-адаптерах и производственном EventBus, без единой строки изменений в `internal/application`/`internal/domain`. Закрыть EPIC-005.

## Контекст

Последняя задача эпика — по аналогии с TASK-045, закрывшей EPIC-004. Требует Docker (PostgreSQL) — см. риски EPIC-005 (daemon не запущен на машине разработчика на момент открытия эпика; тест пишется и работает в CI сервис-контейнером, локальный прогон — по готовности Docker Desktop у человека).

## Scope

### Входит

- Composition root (`internal/infrastructure/wiring` или аналог — уточнить в плане): функция/тип, собирающий `pgxpool.Pool` → пять Postgres-Store'ов, производственный `EventBus`, `RepositoryProvider`, применяющий миграции при старте.
- Интеграционный тест (`//go:build integration`) — golden path через `TaskPlanningService`/`WorkService`/`ResultService`/`CompletionService`/`TaskProjection` на реальных адаптерах (кроме `RepositoryProvider` — где уместно, заменяется на GitHub-адаптер против тестового репозитория либо остаётся тестовым дублем с явной пометкой; решение фиксируется в плане).
- CI: новый job (или шаг) с сервис-контейнером `postgres`, запускающий `go test -tags=integration ./...` — не блокирует обычный job `verify`.
- Обновление README `internal/infrastructure` (итоговая схема слоя, как запускать интеграционные тесты локально и в CI).
- Закрытие EPIC-005: критерии завершения, ROADMAP (v0.5 — Завершено), PROJECT_MANIFEST, PROJECT_HEALTH, CHANGELOG.

### Не входит

- HTTP/REST API (v0.9) — composition root не поднимает сервер.
- Решение о хранении GitHub-секрета в CI (см. риски эпика) — если не будет готово, интеграционный тест использует GitHub-адаптер только юнит-тестами (TASK-050), а golden-path интеграционный тест — с тестовым `RepositoryProvider` (как в EPIC-004), с явной пометкой почему.

## Критерии приёмки

- [x] Composition root собирает систему из реальных адаптеров + применяет миграции.
- [x] Интеграционный golden-path тест зелёный (локально, три прогона подряд против настоящего PostgreSQL); `internal/application`/`internal/domain` не изменены. CI-прогон (сервис-контейнер) подтверждается на PR перед merge.
- [x] Новый CI job/шаг не влияет на обязательный статус-чек `verify` для PR, не требующих Docker (отдельный job `integration`, не `needs: verify`).
- [x] EPIC-005 закрыт: все критерии завершения отмечены, ROADMAP/PROJECT_MANIFEST/PROJECT_HEALTH/CHANGELOG обновлены.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/` (composition root, новый тест), `.github/workflows/`, README `internal/infrastructure`, `docs/roadmap/EPIC-005-infrastructure-layer.md`, `ROADMAP.md`, `PROJECT_MANIFEST.md`, `PROJECT_HEALTH.md`, `CHANGELOG.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-046…050

## План реализации

1. `eventbus.ReadJournal(ctx, pool)` (`internal/infrastructure/eventbus/journal.go`) — читает `event_journal`, восстанавливает `platform.Event` (включая `dataCarrier`-данные через тот же структурный интерфейс, что и `Bus.journal`). Не было в исходном scope этой задачи явно, но необходимо: без него нечем наполнить `TaskProjection.Rebuild` из реальной БД в интеграционном тесте — журнал существует именно для этого (ADR-002/event-model.md).
2. `internal/infrastructure/wiring.System` + `wiring.New(ctx, dsn)` — подключение, `Migrate`, пять Postgres-Store, `eventbus.New(pool)`; `RepositoryProvider` — `github.New()` best-effort (nil при отсутствии `GITHUB_TOKEN`, не фатально: GitHub-адаптер не зависит от Postgres).
3. Решение по `RepositoryProvider` в golden-path тесте: реального `GITHUB_TOKEN` в этой сессии нет (TASK-050 Open Question) → используется тот же in-memory фейк EPIC-004 (`internal/application/inmemory.NewRepositoryProvider()`), с явной пометкой в тесте и README почему. Всё остальное (пять Store, EventBus) — реальные адаптеры.
4. `internal/infrastructure/wiring/golden_path_integration_test.go` (тег `integration`) — тот же сценарий, что `TestGoldenPath_Application` (TASK-045): создание задачи → план → работа → результат → review (с одним "changes requested") → testing (с одним "failed") → Done; состояние — только через `TaskProjection`; в конце — `Rebuild` из `eventbus.ReadJournal` (реальная БД), не из живой шины.
5. `.github/workflows/verify.yml` — новый job `integration` (сервис-контейнер `postgres:16-alpine`, `go test -tags=integration ./...`); отдельный job (не `needs: verify`, не блокирует обязательный статус-чек).
6. README `internal/infrastructure` — секции `wiring`, `ReadJournal`, обновлённая секция CI.
7. Закрытие EPIC-005: критерии эпика, ROADMAP (v0.5 — Завершено), PROJECT_MANIFEST, PROJECT_HEALTH, CHANGELOG.
8. `make verify`, локальный прогон интеграционных тестов (Docker уже поднят) трижды подряд для проверки стабильности, PR, проверка нового CI job `integration` вживую на самом PR, merge.

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (последняя, закрывает эпик).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан, включая обнаруженную по ходу необходимость `ReadJournal` (не в исходном scope, но требуется для содержательного Rebuild-теста против реальной БД) и явное решение про `RepositoryProvider`-фейк в golden-path тесте.
2026-07-21 — Architect — план одобрен; `ReadJournal` — оправданное дополнение (без него журнал не выполняет свою декларированную в ADR-002 функцию); фейковый `RepositoryProvider` в golden-path тесте — честное и задокументированное решение, не маскировка пробела. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано по плану. `TestGoldenPath_Infrastructure` зелёный локально три прогона подряд против настоящего PostgreSQL (Docker поднят). 170 unit-тестов во всём проекте (было 136 на закрытии EPIC-004).
2026-07-21 — Architect — Code Review: `wiring.System.Repository` как best-effort nil-able поле — корректный компромисс (GitHub-адаптер не должен быть жёсткой зависимостью для Postgres-ориентированной сборки); `ReadJournal` переиспользует тот же `dataCarrier`, что и `Bus.journal` — согласованно; golden-path тест дословно повторяет сценарий TASK-045 на реальных адаптерах, отличие только в `RepositoryProvider` и явно им помечено. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-051 — composition root + интеграционный golden-path тест на реальной инфраструктуре; закрытие EPIC-005 (последняя задача эпика).
2. **Что сделано:** `internal/infrastructure/wiring.System`/`New` собирает все реальные адаптеры (пять Postgres-Store, производственный `EventBus`, best-effort `RepositoryProvider`), применяя миграции при старте; `eventbus.ReadJournal` — точечное дополнение TASK-049, доводящее декларированное в ADR-002 назначение журнала («перестроение проекций») до работающего кода; `TestGoldenPath_Infrastructure` — тот же сценарий, что `TestGoldenPath_Application` (EPIC-004), на реальных PostgreSQL-адаптерах и реальном EventBus (кроме `RepositoryProvider` — честно задокументированный in-memory фейк, реального GitHub-токена в сессии нет); новый CI-job `integration` (сервис-контейнер PostgreSQL), не входящий в обязательный статус-чек. EPIC-005 закрыт целиком.
3. **Изменённые файлы:** `internal/infrastructure/eventbus/journal.go` (новый), `internal/infrastructure/eventbus/bus_integration_test.go` (дополнен тестом `ReadJournal`), `internal/infrastructure/wiring/{doc.go,wiring.go,golden_path_integration_test.go}` (новые), `.github/workflows/verify.yml` (новый job `integration`), `internal/infrastructure/README.md`, `docs/roadmap/EPIC-005-infrastructure-layer.md`, `ROADMAP.md`, `PROJECT_MANIFEST.md`, `PROJECT_HEALTH.md`, `CHANGELOG.md`, файл задачи.
4. **Как проверялось:** `go test ./internal/... -v` — 170 unit-тестов, все зелёные; интеграционные тесты (`-tags=integration`, Docker поднят) — три прогона подряд, все зелёные, включая новый `TestGoldenPath_Infrastructure`; `make verify` — чисто; новый CI-job `integration` проверен на самом PR перед merge (см. историю).
5. **Обновлённая документация:** README `internal/infrastructure` (финальная схема слоя); `docs/roadmap/EPIC-005-infrastructure-layer.md`, `ROADMAP.md`, `PROJECT_MANIFEST.md`, `PROJECT_HEALTH.md`, `CHANGELOG.md` — закрытие эпика.
6. **Open Questions:** ручная проверка `RepositoryProvider` против настоящего GitHub API по-прежнему не выполнена (см. TASK-050) — не блокирует закрытие эпика, поскольку golden path использует фейк для этой части намеренно и явно.
7. **Рекомендации:** Phase D плана «дальнейшего выполнения проекта» (v0.5 Infrastructure Layer) завершена. Следующий шаг — Phase E, v0.6 AI Agent Runtime (EPIC-006): первый реальный адаптер Executor (Claude Code) по контракту ADR-005/ADR-006. Отдельно стоит решение о хранении GitHub-секрета для реального прогона `RepositoryProvider` — не блокирует v0.6, но стоит держать в поле зрения при подключении реального git-цикла к платформе.
