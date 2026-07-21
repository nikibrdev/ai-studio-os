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

- [x] Три адаптера реализуют соответствующие интерфейсы `ports.go` без изменения контрактов.
- [x] Save — upsert; Get на несуществующем ID — `application.ErrNotFound`.
- [x] Миграции применяются существующим раннером без его правок.
- [x] Интеграционные тесты зелёные при поднятом PostgreSQL; без него — пропускаются (Docker Desktop не запущен на машине — см. Отчёт).
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/postgres/` (новые файлы), `internal/infrastructure/postgres/migrations/` (новые `.sql`), README `internal/infrastructure`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-046 (и повторяет паттерн TASK-047)

## План реализации

1. По образцу TASK-047: добавить `Restore(...)` в `executor`, `execution`, `artifact` (тот же обнаруженный ранее пробел — агрегаты не давали способа реконструкции из персистентных данных) + тесты round-trip/неалиасинг в каждом пакете.
2. Миграция `0003_executors_executions_artifacts.sql`: таблицы `executors`, `executions` (FK на `tasks`, `executors`), `artifacts` (FK на `projects`; `produced_by` — БЕЗ FK на `executions`, см. ниже).
3. `internal/infrastructure/postgres/{executor_store,execution_store,artifact_store}.go` — Get/Save (upsert), `application.ErrNotFound`; `roles []shared.Role` ↔ `[]string` конвертация для `executors.roles` (pgx не знает тип `shared.Role`).
4. Интеграционные тесты (тег `integration`) по образцу TASK-047, включая связку Project→Task→Executor→Execution→Artifact через реальные Store.
5. README `internal/infrastructure`.
6. `make verify`, `go build/vet -tags=integration ./...`, PR, CI, merge.

Отклонение от исходного scope задачи: там был указан внешний ключ `artifacts` → `executions`; вместо этого `produced_by` — обычный `TEXT` без FK. Обоснование: ADR-016 определяет Artifact как самостоятельный Aggregate Root, а `producedBy` — опциональное поле (пустая строка — легитимное значение «не произведён Execution», не «неизвестно»); внешний ключ на `executions` либо разрешал бы NULL/пустую строку через отдельную логику, либо противоречил бы независимости агрегата. Обычный `TEXT` без ограничения — точнее отражает модель.

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (третья).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан, включая отклонение от исходного scope (produced_by без FK).
2026-07-21 — Architect — план одобрен; отсутствие FK на `produced_by` — правильное решение, соответствует ADR-016 (Artifact не принадлежит Execution). Приступать.
2026-07-21 — Claude Code (Developer) — реализовано по плану. `Restore` добавлен в `executor`/`execution`/`artifact` с тестами; три Store-адаптера реализуют порты EPIC-004 без изменений контрактов; интеграционный тест `TestExecutionStore_SaveThenGet` проверяет полную цепочку Project→Task→Executor→Execution через реальные Store. Docker Desktop запущен (был выключен для TASK-046/047) — `docker compose up -d` поднял PostgreSQL, все 14 интеграционных тестов пакета (включая накопленные с TASK-046/047: раннер миграций, `ProjectStore`, `TaskStore`) прогнаны против настоящей БД и зелёные — снимает риск «SQL никогда не проверялся вживую», отмеченный в плане.
2026-07-21 — Architect — Code Review: конвертация `[]shared.Role` ↔ `[]string` для `executors.roles` — необходима (pgx не знает доменный тип), реализована корректно и симметрично (`toRoles`/`fromRoles`); отсутствие FK на `produced_by` обосновано и не противоречит ADR-016; upsert-и не трогают неизменяемые поля (`created_at`, `type`, `origin`, `task_id`, `executor_id` — верно, это поля, фиксированные при создании). Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-048 — Postgres-адаптеры `ExecutorStore` + `ExecutionStore` + `ArtifactStore` (третья задача EPIC-005, закрывает реализацию всех пяти портов хранения).
2. **Что сделано:** миграция `0003_executors_executions_artifacts.sql` (таблицы `executors`, `executions`, `artifacts` со связями и индексами); три Store-адаптера — Get/Save (upsert), `application.ErrNotFound`; в домен (`executor`, `execution`, `artifact`) добавлена `Restore(...)` по тому же образцу, что и в TASK-047.
3. **Изменённые файлы:** `internal/domain/executor/{executor.go,executor_test.go}`, `internal/domain/execution/{execution.go,execution_test.go}`, `internal/domain/artifact/{artifact.go,artifact_test.go}` (добавлен `Restore` + тесты); `internal/infrastructure/postgres/{executor_store.go,execution_store.go,artifact_store.go,executor_execution_artifact_store_integration_test.go}` (новые); `internal/infrastructure/postgres/migrations/0003_executors_executions_artifacts.sql` (новая); `internal/infrastructure/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/domain/executor/... ./internal/domain/execution/... ./internal/domain/artifact/... -cover` — 97.4%/98.0%/100%, все тесты зелёные; `make verify` — чисто. Docker Desktop запущен в ходе этой задачи (был выключен для TASK-046/047) — `docker compose up -d` + `go test -tags=integration ./internal/infrastructure/...` против настоящего PostgreSQL: все 14 интеграционных тестов зелёные (раннер миграций 0001-0003, все пять Store — Project/Task/Executor/Execution/Artifact), включая проверку связки Project→Task→Executor→Execution. Риск «SQL не проверялся вживую», актуальный для TASK-046/047 на момент их отчётов, снят этим прогоном.
5. **Обновлённая документация:** README `internal/infrastructure`.
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-051 (composition root) может опираться на уже подтверждённые вживую миграции и Store — остаётся собрать их в единую точку и добавить CI-job с сервис-контейнером PostgreSQL, чтобы это подтверждалось автоматически, а не только вручную на машине разработчика.
