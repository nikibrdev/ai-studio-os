# TASK-047: Postgres-адаптеры ProjectStore + TaskStore

## Тип

feature

## Эпик

[EPIC-005 Infrastructure Layer](../../docs/roadmap/EPIC-005-infrastructure-layer.md)

## Цель

Реализовать `application.ProjectStore` и `application.TaskStore` на PostgreSQL — источник истины задач по [ADR-004](../../docs/adr/ADR-004-task-storage.md). Первая пара агрегатов, поскольку Task ссылается на Project (внешний ключ) и это самый прямолинейный порядок для проверки миграций на связанных таблицах.

## Контекст

Опирается на TASK-046 (`pgxpool`, раннер миграций). Контракты `ProjectStore`/`TaskStore` (`internal/application/ports.go`) не меняются — только реализация. Сериализация полей агрегатов — по публичным геттерам сущностей `project.Project`/`task.Task` (домен не знает об SQL, слой infrastructure отвечает за маппинг).

## Scope

### Входит

- Миграция(и): таблицы `projects`, `tasks` (внешний ключ на `projects`), поля — по факту требуемых геттеров агрегатов.
- `internal/infrastructure/postgres/project_store.go` — Get/Save, `application.ErrNotFound` при отсутствии строки.
- `internal/infrastructure/postgres/task_store.go` — Get/Save, `application.ErrNotFound` при отсутствии строки.
- Компиляционные проверки `var _ application.ProjectStore = (*ProjectStore)(nil)` (и аналогично для Task) — по образцу in-memory фейков EPIC-004.
- Интеграционные тесты (`//go:build integration`) — Get/Save/ErrNotFound на реальной БД (Docker Compose из TASK-046).

### Не входит

- ExecutorStore/ExecutionStore/ArtifactStore (TASK-048).
- Изменение контрактов `ports.go` — если геттеров агрегата не хватает для сериализации, добавить геттер в домен точечно (в этой же задаче, с обоснованием в отчёте), не расширять контракт use-case'ов.

## Критерии приёмки

- [ ] `ProjectStore`/`TaskStore` реализуют интерфейсы `internal/application/ports.go` без изменения контрактов.
- [ ] Save — upsert (создание и обновление одним методом, как и в in-memory фейке); Get на несуществующем ID — `application.ErrNotFound`.
- [ ] Миграции применяются раннером TASK-046 без правок раннера.
- [ ] Интеграционные тесты зелёные при поднятом `docker-compose` PostgreSQL; без него — пропускаются, не ломают `make verify`.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/postgres/` (новые файлы), `internal/infrastructure/postgres/migrations/` (новые `.sql`), README `internal/infrastructure`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-046

## План реализации

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (вторая, после TASK-046).

## Отчёт о выполнении
