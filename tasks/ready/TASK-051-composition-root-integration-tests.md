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

- [ ] Composition root собирает систему из реальных адаптеров + применяет миграции.
- [ ] Интеграционный golden-path тест зелёный в CI (сервис-контейнер PostgreSQL); `internal/application`/`internal/domain` не изменены.
- [ ] Новый CI job/шаг не влияет на обязательный статус-чек `verify` для PR, не требующих Docker.
- [ ] EPIC-005 закрыт: все критерии завершения отмечены, ROADMAP/PROJECT_MANIFEST/PROJECT_HEALTH/CHANGELOG обновлены.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/` (composition root, новый тест), `.github/workflows/`, README `internal/infrastructure`, `docs/roadmap/EPIC-005-infrastructure-layer.md`, `ROADMAP.md`, `PROJECT_MANIFEST.md`, `PROJECT_HEALTH.md`, `CHANGELOG.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-046…050

## План реализации

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (последняя, закрывает эпик).

## Отчёт о выполнении
