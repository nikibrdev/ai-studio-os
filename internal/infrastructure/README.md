# Слой: internal/infrastructure

## Назначение

Infrastructure Layer — адаптеры, реализующие контракты `internal/platform` и порты `internal/application` для конкретных технологий: PostgreSQL (персистентность агрегатов, журнал событий), производственный In-Memory Event Bus ([ADR-002](../../docs/adr/ADR-002-event-delivery.md)), GitHub (RepositoryProvider). Реализуется эпиком [EPIC-005](../../docs/roadmap/EPIC-005-infrastructure-layer.md), v0.5.

## Содержание

### Ответственность

- Реализации контрактов `internal/platform` и портов `internal/application`.
- Никаких доменных решений: адаптер исполняет, домен и application решают.

### Зависимости

- Разрешено: контракты `internal/platform`, порты `internal/application`, доменные типы (для сериализации), драйвер своей системы.
- Запрещено: доменная логика; зависимость адаптеров друг от друга; реализация «чужих» портов заодно.

### Структура

| Пакет | Содержимое | Задача |
| --- | --- | --- |
| `postgres/` | Подключение (`pgxpool.Pool`), раннер миграций, Store-адаптеры пяти агрегатов | TASK-046 (каркас), TASK-047 (Project+Task), TASK-048 (Executor+Execution+Artifact) |
| `eventbus/` | Производственный `platform.EventBus` + журнал событий в PostgreSQL | TASK-049 |
| `github/` | `platform.RepositoryProvider` — GitHub REST API | TASK-050 |

### `postgres` — подключение

Драйвер — `pgx/v5` через `pgxpool.Pool` ([ADR-017](../../docs/adr/ADR-017-postgresql-driver.md)). DSN — переменная окружения `DATABASE_URL` (`postgres.DatabaseURLEnv`):

```go
pool, err := postgres.NewPool(ctx) // читает DATABASE_URL
// или
pool, err := postgres.NewPoolFromDSN(ctx, dsn) // явный DSN
```

`NewPool`/`NewPoolFromDSN` пингуют соединение перед возвратом — ошибка подключения обнаруживается сразу, не при первом реальном запросе.

### `postgres` — миграции

Самописный раннер, не внешняя библиотека (ADR-017: единственная используемая возможность таких библиотек — применить пронумерованные `.sql`-файлы один раз при старте, вводить зависимость ради этого нецелесообразно).

- `.sql`-файлы лежат в `postgres/migrations/`, встроены через `embed.FS`.
- Именование — числовой префикс с ведущими нулями (`0001_init.sql`, `0002_...`) — раннер применяет их в лексикографическом порядке имени файла.
- Применённые версии фиксируются в таблице `schema_migrations` (создаётся раннером автоматически, `CREATE TABLE IF NOT EXISTS`).
- `postgres.Migrate(ctx, pool)` — применяет все ещё не применённые миграции; повторный вызов на уже мигрированной базе — no-op.
- Каждая миграция применяется в отдельной транзакции (файл + запись в `schema_migrations` — атомарно).

Новая миграция — новый файл с очередным номером; редактировать уже применённые файлы нельзя (история миграций только растёт).

### `postgres` — Store-адаптеры

`ProjectStore` и `TaskStore` (TASK-047) реализуют `application.ProjectStore`/`application.TaskStore` — те же контракты, что и in-memory фейки EPIC-004, без изменений. `Save` — upsert по `id`; `Get` на отсутствующей строке возвращает `application.ErrNotFound` (тот же sentinel, что и у фейков — use-case'ы не отличают технологию хранения).

`ExecutorStore`, `ExecutionStore` и `ArtifactStore` (TASK-048) реализуют оставшиеся три порта тем же образом.

Доменные агрегаты (`project.Project`, `task.Task`, `executor.Executor`, `execution.Execution`, `artifact.Artifact`) хранят поля неэкспортированными и не давали способа собрать их из уже сохранённых данных — только через бизнес-команды (`New`, `Activate`, ...). Во все пять пакетов добавлена `Restore(...)` — чистая реконструкция из уже провалидированных при сохранении данных, без бизнес-правил и без события; вызывать её вне Store-адаптера не следует.

### Локальный PostgreSQL

`docker-compose.yml` в корне репозитория поднимает PostgreSQL для разработки и интеграционных тестов:

```bash
docker compose up -d
# DSN: postgres://ai_studio_os:ai_studio_os@localhost:5432/ai_studio_os?sslmode=disable
```

### Интеграционные тесты

Тесты, требующие реального PostgreSQL, — за build-тегом `integration` и пропускаются (`t.Skip`), если не задана переменная `TEST_DATABASE_URL` — обычный `go test ./...` (и, следовательно, `make verify`) их не запускает и не видит Docker как зависимость.

```bash
docker compose up -d
export TEST_DATABASE_URL="postgres://ai_studio_os:ai_studio_os@localhost:5432/ai_studio_os?sslmode=disable"
go test -tags=integration ./...
```

CI-job, поднимающий сервис-контейнер PostgreSQL и запускающий эти тесты автоматически, — TASK-051 (закрытие эпика); отдельно от обязательного статус-чека `verify`.

### События

Производственный EventBus (TASK-049) доставляет события; он же (или подписчик рядом) сохраняет их в PostgreSQL для журнала. Сами адаптеры доменных событий не порождают.

## Статус

В работе (EPIC-005)

## Последнее обновление

2026-07-21
