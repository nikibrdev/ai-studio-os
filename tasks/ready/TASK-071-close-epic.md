# TASK-071: README apps/api, синхронизация документации, закрытие EPIC-008

## Тип

docs

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Закрыть EPIC-008: README `apps/api`, синхронизация архитектурной документации, обновление PROJECT_MANIFEST/PROJECT_HEALTH/ROADMAP/CHANGELOG.

## Контекст

Последняя задача эпика — по аналогии с TASK-051/057/063, закрывавшими предыдущие эпики.

## Scope

### Входит

- `apps/api/README.md` — назначение, зависимости (`internal/application`, `internal/infrastructure/wiring`), запуск локально, ссылка на `docs/api/`.
- `docs/architecture/module-boundaries.md` — сверка с фактической реализацией (пакет `httpapi`, точка входа `main.go`).
- `docs/architecture/system-design.md` — если описание API-слоя там расходится с реализацией.
- ROADMAP.md (v0.9 — Завершено, с честным описанием ограничений: без auth, без списковых проекций), PROJECT_MANIFEST.md, PROJECT_HEALTH.md, CHANGELOG.md.

### Не входит

- Открытие эпика Dashboard (v0.8) — отдельная задача после этой.

## Критерии приёмки

- [ ] `apps/api/README.md` написан по стандарту README модуля.
- [ ] Архитектурная документация, упоминающая API-слой, сверена с фактической реализацией.
- [ ] ROADMAP/PROJECT_MANIFEST/PROJECT_HEALTH/CHANGELOG отражают фактический результат, включая честные ограничения (без auth, без списковых проекций сверх `TaskProjection.Get`).
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/api/README.md`, `docs/architecture/module-boundaries.md`, `docs/architecture/system-design.md`, `ROADMAP.md`, `PROJECT_MANIFEST.md`, `PROJECT_HEALTH.md`, `CHANGELOG.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-064…070

## План реализации

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь (последняя, закрывает эпик).

## Отчёт о выполнении
