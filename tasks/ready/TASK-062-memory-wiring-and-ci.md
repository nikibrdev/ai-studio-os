# TASK-062: Композиция Memory Provider, CI-интеграция, README

## Тип

feature

## Эпик

[EPIC-007 Memory System](../../docs/roadmap/EPIC-007-memory-system.md)

## Цель

Подключить `internal/infrastructure/memory.Provider` к composition root (`internal/infrastructure/wiring`), расширить CI-job `integration` сервис-контейнером Qdrant, задокументировать слой.

## Контекст

`internal/infrastructure/wiring.System` (EPIC-005) уже собирает Postgres-адаптеры, EventBus и best-effort `RepositoryProvider`. `platform.MemoryProvider` — ещё один порт того же уровня; добавляется тем же образом.

## Scope

### Входит

- `wiring.System.Memory platform.MemoryProvider` — собирается в `wiring.New`, использует тот же `DATABASE_URL`-независимый путь (Qdrant — отдельный DSN/адрес, переменная окружения по аналогии с `DatabaseURLEnv`).
- `.github/workflows/verify.yml` — job `integration` дополняется сервис-контейнером Qdrant.
- README `internal/infrastructure` — раздел про Memory Provider (аналогично разделам про postgres/eventbus/github).
- `agents/README.md`/`agents/claude-code/README.md` — если уместно (Memory используется агентами) — точечное упоминание, не переработка.

### Не входит

- Автоматический вызов `Record` из `agents/claude-code` или application-сервисов — вне scope эпика (см. EPIC-007 Scope «Не входит»).

## Критерии приёмки

- [ ] `wiring.System.Memory` собирается и доступен наравне с остальными адаптерами.
- [ ] CI-job `integration` включает Qdrant, интеграционные тесты TASK-061 проходят в CI.
- [ ] README обновлены.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/wiring/wiring.go`, `.github/workflows/verify.yml`, `internal/infrastructure/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-061

## План реализации

## История

2026-07-22 — Architect — EPIC-007 открыт; задача поставлена в очередь (пятая, после TASK-061).

## Отчёт о выполнении
