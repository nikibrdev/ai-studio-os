# TASK-048: Postgres-адаптеры ExecutorStore + ExecutionStore + ArtifactStore

## Тип

feature

## Эпик

[EPIC-005 Infrastructure Layer](../../docs/roadmap/EPIC-005-infrastructure-layer.md)

## Цель

Реализовать оставшиеся три порта хранения (`application.ExecutorStore`, `application.ExecutionStore`, `application.ArtifactStore`) на PostgreSQL — тем же образцом, что и TASK-047 (Project/Task).

## Контекст

Опирается на TASK-046 (подключение, раннер) и повторяет паттерн TASK-047. Execution ссылается на Task (TaskID), Artifact — на Execution (ADR-016: Artifact — самостоятельный Aggregate Root со своим циклом, не часть Execution/Task/Project, но со ссылкой на породившую Execution).

## Scope

### Входит

- Миграции: таблицы `executors`, `executions` (внешний ключ на `tasks`), `artifacts` (ссылка на `executions`).
- `internal/infrastructure/postgres/{executor_store,execution_store,artifact_store}.go` — Get/Save, `application.ErrNotFound`.
- Компиляционные проверки соответствия портам.
- Интеграционные тесты (`//go:build integration`) по образцу TASK-047.

### Не входит

- EventBus/GitHub-адаптеры (TASK-049, TASK-050).
- Изменение контрактов `ports.go`.

## Критерии приёмки

- [ ] Три адаптера реализуют соответствующие интерфейсы `ports.go` без изменения контрактов.
- [ ] Save — upsert; Get на несуществующем ID — `application.ErrNotFound`.
- [ ] Миграции применяются существующим раннером без его правок.
- [ ] Интеграционные тесты зелёные при поднятом PostgreSQL; без него — пропускаются.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/postgres/` (новые файлы), `internal/infrastructure/postgres/migrations/` (новые `.sql`), README `internal/infrastructure`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-046 (и повторяет паттерн TASK-047)

## План реализации

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (третья).

## Отчёт о выполнении
