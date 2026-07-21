# TASK-046: Каркас Infrastructure Layer — подключение к PostgreSQL, миграции

## Тип

feature

## Эпик

[EPIC-005 Infrastructure Layer](../../docs/roadmap/EPIC-005-infrastructure-layer.md)

## Цель

Фундамент слоя `internal/infrastructure`: конфигурация подключения к PostgreSQL (`pgxpool`, [ADR-017](../../docs/adr/ADR-017-postgresql-driver.md)), самописный раннер миграций по встроенным `.sql`-файлам, схема `schema_migrations`, Docker Compose для локального PostgreSQL, README слоя. Без этой задачи ни один Postgres-адаптер (TASK-047…049) не может быть написан.

## Контекст

`internal/platform`/`internal/application` уже объявляют все нужные порты (EPIC-002/004); EPIC-005 их реализует. ADR-017 зафиксировал драйвер (`pgx/v5`) и решение не вводить библиотеку миграций — раннер пишется здесь. `docker-compose.yml` в корне репозитория — впервые в проекте; ограничения Foundation по Docker Compose сняты (архитектура заморожена, инженерная платформа готова).

## Scope

### Входит

- `internal/infrastructure/postgres/` (или аналогичное имя пакета — уточнить в плане): конфигурация подключения (DSN из переменной окружения, без хардкода секретов), `pgxpool.Pool`, health-check (`Ping`).
- Раннер миграций: `embed.FS` с `.sql`-файлами, таблица `schema_migrations` (версия, применена когда), применение по возрастанию имени файла, идемпотентность (повторный запуск — no-op).
- Первая миграция — только `schema_migrations` (или пустая база) — таблицы агрегатов появляются в TASK-047/048.
- `docker-compose.yml` — сервис `postgres` для локальной разработки/интеграционных тестов.
- README `internal/infrastructure` (структура пакета, как поднять локальный PostgreSQL, как писать новую миграцию).

### Не входит

- Схемы таблиц агрегатов и сами адаптеры (TASK-047, TASK-048).
- EventBus/GitHub-адаптеры (TASK-049, TASK-050).
- Composition root и интеграционные тесты golden path (TASK-051).

## Критерии приёмки

- [x] `pgxpool.Pool` поднимается из DSN в переменной окружения (имя переменной задокументировано), есть health-check.
- [x] Раннер миграций применяет `.sql`-файлы из `embed.FS` по порядку, фиксирует применённые версии в `schema_migrations`, повторный запуск на уже мигрированной БД — no-op.
- [x] `docker-compose.yml` поднимает PostgreSQL, пригодный для раннера и будущих адаптеров (порт, имя БД, креды — для локальной разработки, не для продакшена).
- [x] Юнит-тесты, не требующие реальной БД, — зелёные в `make verify`; тесты, требующие PostgreSQL, — за build-тегом `integration`, пропускаются при отсутствии `TEST_DATABASE_URL` (`t.Skip`), не ломают `go test ./...` без Docker.
- [x] README создан; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/` (новый), `docker-compose.yml` (новый), `go.mod`/`go.sum` (добавление `pgx/v5`).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — EPIC-004 закрыт; ADR-017 принят

## План реализации

1. `go get github.com/jackc/pgx/v5@v5.8.0` (не последняя v5.10.0 — та требует Go 1.25, ADR-009 фиксирует 1.24; v5.8.0 требует ровно go1.24.0, `go.mod` остаётся на `go 1.24`); `go mod tidy` после добавления реального импорта.
2. `internal/infrastructure/doc.go` — package doc слоя (структура, кто что реализует).
3. `internal/infrastructure/postgres/`: `doc.go`, `config.go` (`NewPool`/`NewPoolFromDSN` — DSN из `DATABASE_URL`, `Ping` при создании, `ErrDatabaseURLNotSet`), `migrate.go` (раннер: `embed.FS` миграций, `schema_migrations`, применение по имени файла в отдельной транзакции на файл, идемпотентность через `SELECT EXISTS`).
4. `migrations/0001_init.sql` — маркер без таблиц агрегатов (те — TASK-047/048); раннеру нужен хотя бы один файл для проверки механизма.
5. Тесты: `config_test.go` (без реальной БД — отсутствие `DATABASE_URL`, невалидный DSN); `migrate_integration_test.go` за тегом `integration`, `t.Skip` без `TEST_DATABASE_URL` — обычный `go test ./...`/`make verify` не видят Docker как зависимость.
6. `docker-compose.yml` в корне — сервис `postgres:16-alpine` с healthcheck.
7. README `internal/infrastructure` (было заглушкой с 2026-07-19) — переписан под факт реализации.
8. `make verify`, `go build/vet -tags=integration ./...` (компиляция интеграционного теста без реальной БД), PR, CI, merge.

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (первая).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано по плану. Побочно обнаружено и исправлено: первый `go get` (без явной версии) подтянул pgx v5.10.0 и поднял директиву `go` в `go.mod` до 1.25.0 (её требование) — зафиксировано и исправлено закреплением v5.8.0 (последняя версия pgx/v5, совместимая с зафиксированным ADR-009 Go 1.24); директива `go` вручную возвращена на `1.24` (нормализовалась в `1.24.0`), `go mod tidy` подтвердил отсутствие более высоких транзитивных требований. Docker Desktop не запущен на машине — интеграционный тест `TestMigrate_AppliesAndIsIdempotent` не прогнан локально «вживую», но проверена его компилируемость (`go build/vet -tags=integration ./...`) и корректность `t.Skip` при отсутствии `TEST_DATABASE_URL`; реальный прогон — при поднятом Docker (см. риски EPIC-005) или в CI-job TASK-051.
2026-07-21 — Architect — Code Review: DSN/Ping-подход соответствует контракту; раннер миграций прост и идемпотентен, транзакция на файл — верно; понижение pgx до v5.8.0 вместо повышения Go до 1.25 — правильное решение (toolchain — не в scope этой задачи, ADR-009 не пересматривается точечно); интеграционный тест корректно изолирован тегом и `t.Skip`. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-046 — каркас Infrastructure Layer: подключение к PostgreSQL и раннер миграций (первая задача EPIC-005).
2. **Что сделано:** `internal/infrastructure/postgres` — `pgxpool.Pool` из `DATABASE_URL` с health-check (`NewPool`/`NewPoolFromDSN`); самописный раннер миграций по встроенным `.sql` (без внешней библиотеки, ADR-017) с таблицей `schema_migrations`, идемпотентный, миграция — отдельная транзакция; первая миграция-маркер `0001_init.sql`; `docker-compose.yml` для локального PostgreSQL; README слоя переписан.
3. **Изменённые файлы:** `go.mod`/`go.sum` (добавлена `github.com/jackc/pgx/v5 v5.8.0`); `internal/infrastructure/{doc.go,README.md}` (README переписан); `internal/infrastructure/postgres/{doc,config,config_test,migrate,migrate_integration_test}.go`; `internal/infrastructure/postgres/migrations/0001_init.sql`; `docker-compose.yml` (новый, корень репозитория); файл задачи.
4. **Как проверялось:** `go test ./internal/infrastructure/... -v -cover` — 2/2 unit-теста зелёные, 10.2% (закономерно низко: логика раннера требует реальной БД и покрыта только интеграционным тестом, который не входит в этот прогон — не единственный подобный случай в проекте, см. README); `go build/vet -tags=integration ./...` — компилируется; `make verify` — чисто (docs-check: 1148 ссылок, 0 ошибок).
5. **Обновлённая документация:** README `internal/infrastructure` (полностью переписан).
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-047 — Postgres-адаптеры `ProjectStore`/`TaskStore` на этом раннере и пуле; первая реальная миграция с таблицами `projects`/`tasks`.
