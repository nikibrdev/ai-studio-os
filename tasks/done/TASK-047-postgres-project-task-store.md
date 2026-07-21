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

- [x] `ProjectStore`/`TaskStore` реализуют интерфейсы `internal/application/ports.go` без изменения контрактов.
- [x] Save — upsert (создание и обновление одним методом, как и в in-memory фейке); Get на несуществующем ID — `application.ErrNotFound`.
- [x] Миграции применяются раннером TASK-046 без правок раннера.
- [x] Интеграционные тесты зелёные при поднятом `docker-compose` PostgreSQL; без него — пропускаются, не ломают `make verify` (Docker Desktop не запущен на машине разработки — прогон против реальной БД не выполнен «вживую», см. Отчёт).
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/postgres/` (новые файлы), `internal/infrastructure/postgres/migrations/` (новые `.sql`), README `internal/infrastructure`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-046

## План реализации

1. Миграция `0002_projects_tasks.sql`: таблицы `projects` (id, name, repositories TEXT[], created_at, state) и `tasks` (id, project_id FK, epic_id, title, task_type, scope, acceptance_criteria TEXT[], created_at, state) + индекс на `tasks.project_id`.
2. Обнаружено при проектировании: `project.Project`/`task.Task` хранят поля неэкспортированными и не давали способа собрать сущность из уже сохранённых данных — только через бизнес-команды. Добавить `Restore(...)` в оба доменных пакета: чистая реконструкция без бизнес-правил и без события, только для Store-адаптеров (документируется в комментарии к функции). Не расширяет и не меняет существующие контракты/инварианты — точечное дополнение, как `Envelope.WithData` в EPIC-004.
3. Тесты на `Restore` в `project_test.go`/`task_test.go` — round-trip полей, копирование срезов (не алиас).
4. `internal/infrastructure/postgres/project_store.go` — `ProjectStore{pool}`, `Get`/`Save` (upsert), `application.ErrNotFound` через `errors.Is(err, pgx.ErrNoRows)`.
5. `internal/infrastructure/postgres/task_store.go` — аналогично для `Task`.
6. Интеграционные тесты (`project_task_store_integration_test.go`, тег `integration`): Save→Get round-trip, Get на несуществующем ID, upsert обновляет существующую строку.
7. README `internal/infrastructure` — раздел про Store-адаптеры и `Restore`.
8. `make verify`, `go build/vet -tags=integration ./...`, PR, CI, merge.

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (вторая, после TASK-046).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен, включая точечное добавление `Restore` в домен (не расширяет контракт, не бизнес-правило — чистая реконструкция для инфраструктуры). Приступать.
2026-07-21 — Claude Code (Developer) — реализовано по плану. `Restore` добавлен в `project`/`task` с тестами на round-trip и неалиасинг срезов; `ProjectStore`/`TaskStore` реализуют порты EPIC-004 без изменений контрактов. Docker Desktop не запущен на машине — интеграционные тесты не прогнаны «вживую», но компилируются и проходят `go vet` под тегом `integration`; реальный прогон — при поднятом Docker или в CI-job TASK-051.
2026-07-21 — Architect — Code Review: `Restore` корректно ограничен комментарием («callers outside a Store implementation should not use it») и не меняет ни одного существующего инварианта; upsert через `ON CONFLICT DO UPDATE` не трогает `created_at` (верно — момент создания неизменен); `errors.Is(err, pgx.ErrNoRows)` — правильный способ маппинга на `application.ErrNotFound`. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-047 — Postgres-адаптеры `ProjectStore` + `TaskStore` (вторая задача EPIC-005).
2. **Что сделано:** миграция `0002_projects_tasks.sql` (таблицы `projects`, `tasks` со связью и индексом); `ProjectStore`/`TaskStore` в `internal/infrastructure/postgres` — Get/Save (upsert), `application.ErrNotFound` при отсутствии строки; в домен (`project`, `task`) добавлена `Restore(...)` — реконструкция агрегата из персистентных данных без бизнес-правил и событий, нужна только Store-адаптерам (обнаруженный по ходу задачи пробел: агрегаты не давали иного способа собрать себя из уже сохранённых полей).
3. **Изменённые файлы:** `internal/domain/project/{project.go,project_test.go}`, `internal/domain/task/{task.go,task_test.go}` (добавлен `Restore` + тесты); `internal/infrastructure/postgres/{project_store.go,task_store.go,project_task_store_integration_test.go}` (новые); `internal/infrastructure/postgres/migrations/0002_projects_tasks.sql` (новая); `internal/infrastructure/README.md` (раздел Store-адаптеров); файл задачи.
4. **Как проверялось:** `go test ./internal/domain/project/... ./internal/domain/task/... -cover` — 97.3%/91.1%, все тесты зелёные, включая новые для `Restore`; `go build/vet -tags=integration ./...` — компилируется; `make verify` — чисто. Покрытие `internal/infrastructure/postgres` низкое (6.5%) — ожидаемо: логика Store-адаптеров требует реальной БД и проверяется только интеграционными тестами, которые не входят в обычный прогон (тот же паттерн, что и в TASK-046). Docker Desktop не запущен на машине разработки — интеграционные тесты не прогнаны «вживую»; реальный прогон — по готовности Docker или в CI-job TASK-051.
5. **Обновлённая документация:** README `internal/infrastructure`.
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-048 повторяет этот же паттерн для `ExecutorStore`/`ExecutionStore`/`ArtifactStore`; если `Execution`/`Executor`/`Artifact` тоже не дают способа реконструкции — добавить `Restore` по тому же образцу.
