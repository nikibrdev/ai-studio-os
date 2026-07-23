# Слой: internal/infrastructure

## Назначение

Infrastructure Layer — адаптеры, реализующие контракты `internal/platform` и порты `internal/application` для конкретных технологий: PostgreSQL (персистентность агрегатов, журнал событий), производственный In-Memory Event Bus ([ADR-002](../../docs/adr/ADR-002-event-delivery.md)), GitHub (RepositoryProvider). Реализовано эпиком [EPIC-005](../../docs/roadmap/EPIC-005-infrastructure-layer.md), v0.5. Дополнено `platform.MemoryProvider` (файлы + Qdrant) эпиком [EPIC-007](../../docs/roadmap/EPIC-007-memory-system.md), v0.7.

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
| `wiring/` | Composition root: собирает `System` из всех адаптеров выше, применяет миграции | TASK-051 |
| `memory/` | `platform.MemoryProvider` — файловое хранилище (источник истины) + Qdrant (производный индекс) | TASK-059…062 |

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
- `Migrate` держит сессионный PostgreSQL advisory lock (`pg_advisory_lock`) на выделенном соединении (`pool.Acquire`) на всё время миграции — без этого два процесса (или два пакета тестов, которые `go test` по умолчанию запускает параллельно), мигрирующие свежую базу одновременно, могут оба решить, что миграция не применена, и оба попытаться `CREATE TABLE`: один упадёт с конфликтом в системном каталоге вместо чистой ошибки «уже существует». Обнаружено интеграционными тестами TASK-049 (`eventbus` и `postgres` конкурировали за одну и ту же локальную БД).

Новая миграция — новый файл с очередным номером; редактировать уже применённые файлы нельзя (история миграций только растёт).

### `postgres` — Store-адаптеры

`ProjectStore` и `TaskStore` (TASK-047) реализуют `application.ProjectStore`/`application.TaskStore` — те же контракты, что и in-memory фейки EPIC-004, без изменений. `Get` на отсутствующей строке возвращает `application.ErrNotFound` (тот же sentinel, что и у фейков — use-case'ы не отличают технологию хранения). `TaskStore.Save` — upsert по `(project_id, id)`, не по голому `id`: публичный `TASK-NNN` уникален только в рамках Project (ADR-011), а не глобально — до BUGFIX-003 `tasks` был с `PRIMARY KEY (id)`, и два разных проекта с одинаковым `TASK-001` молча портили друг друга через `ON CONFLICT (id) DO UPDATE`; миграция `0006` меняет ключ на составной, `executions.task_id` стал неформальной ссылкой без FK (тот же принцип, что уже был у `artifacts.produced_by`). `ProjectStore.List` (TASK-072, EPIC-009) — `SELECT ... ORDER BY id`, без пагинации (не нужна при текущем объёме).

`ExecutorStore`, `ExecutionStore` и `ArtifactStore` (TASK-048) реализуют оставшиеся три порта тем же образом.

`TaskStore.NextID` (TASK-065, EPIC-008) реализует `application.TaskIDGenerator` (ADR-011): атомарно выдаёт следующий `TASK-NNN` на проект одним `INSERT ... ON CONFLICT DO UPDATE ... RETURNING` — собственная блокировка строки PostgreSQL сериализует конкурентных вызывающих без блокировок на уровне приложения; таблица `task_sequences` (миграция `0005`) хранит счётчик на проект. Проверено вживую: 50 конкурентных вызовов для одного проекта дают ровно числа 1…50 без повторов и пропусков (три прогона подряд).

Доменные агрегаты (`project.Project`, `task.Task`, `executor.Executor`, `execution.Execution`, `artifact.Artifact`) хранят поля неэкспортированными и не давали способа собрать их из уже сохранённых данных — только через бизнес-команды (`New`, `Activate`, ...). Во все пять пакетов добавлена `Restore(...)` — чистая реконструкция из уже провалидированных при сохранении данных, без бизнес-правил и без события; вызывать её вне Store-адаптера не следует.

### Локальный PostgreSQL

`docker-compose.yml` в корне репозитория поднимает PostgreSQL для разработки и интеграционных тестов:

```bash
docker compose up -d
# DSN: postgres://ai_studio_os:ai_studio_os@localhost:5432/ai_studio_os?sslmode=disable
```

### `github` — RepositoryProvider

`Provider` (TASK-050) реализует `platform.RepositoryProvider` напрямую через GitHub REST API (`net/http`) — без клиентской библиотеки (`stack.md` её не перечисляет; шесть операций контракта не оправдывают новую зависимость, тот же принцип, что и решение ADR-017 не вводить библиотеку миграций). Токен — переменная окружения `GITHUB_TOKEN` (`github.TokenEnv`).

Два места, где контракт не полностью однозначен, потребовали решения при реализации:

- **`OpenPullRequest` не принимает целевую ветку** — контракт (`internal/platform/repository.go`) фиксирует её в тексте doc-комментария: «opens a pull request from the branch into the main branch». Реализовано буквально — целевая ветка всегда `main`.
- **`RequestReview` не принимает личность ревьюера** — GitHub API «request reviewers» требует конкретных логинов/команд, которых в контракте нет. [ADR-008](../../docs/adr/ADR-008-git-policies.md) явно говорит: «обязательность ревью обеспечивается стадией Review канонической state machine, а не настройкой GitHub» — то есть платформа не полагается на нативный GitHub-механизм review-required. Реализовано как публикация видимого комментария к PR («Запрошено ревью.») — видимый сигнал без придумывания несуществующего в контракте параметра.

Юнит-тесты — на `httptest.Server` (без обращения к реальному GitHub): успешный путь и минимум один отказной сценарий на метод, 89.6% покрытия. Интеграционного прогона против настоящего GitHub нет — нужен тестовый репозиторий и секрет в CI, решение об этом не входит в scope EPIC-005 (см. риски эпика).

### `memory` — Memory Provider

`Provider` (TASK-061, ADR-018) реализует `platform.MemoryProvider` над двумя компонентами, каждый — durable-источник истины или его производный индекс (тот же принцип, что `event_journal`/`ReadJournal`, EPIC-005):

- `FileStore` (TASK-059) — `memory/<projectID>/<id>.md` (frontmatter + Markdown), человекочитаемое, аудируемое хранилище; формат и политика записи зафиксированы отдельным decision-документом ([2026-07-22-memory-file-format.md](../../engineering/decisions/2026-07-22-memory-file-format.md)), не ADR (это формат файлов и организационная политика, не выбор технологии).
- `QdrantClient` (TASK-060) — REST-клиент Qdrant напрямую через `net/http`, без клиентской библиотеки (тот же принцип, что и `github`-адаптер); коллекция `memory_entries` общая на все проекты, `project_id` — поле payload, а не отдельная коллекция.

`embed(text string) []float32` (TASK-060) — наивный детерминированный эмбеддинг (feature hashing, 256 измерений, `hash/fnv` + `math`, ADR-018): без нейросети, без внешних вызовов и секретов; осознанно не семантический (совпадение хешированных токенов, не смысла) — честно задокументированное ограничение MVP, локализованное в одной функции для будущей замены.

`Provider.Record` пишет файл, затем индексирует запись в Qdrant — в этом порядке, чтобы сбой Qdrant не терял данные (файл уже сохранён). `Provider.Search` встраивает запрос и реконструирует записи из самодостаточного payload Qdrant (`project_id`/`kind`/`content`/`source`/`recorded_at`) без обращения к файлам. `Provider.Reindex(ctx, projectID)` перестраивает индекс проекта из файлов с нуля — восстановление после расхождения файла и Qdrant (двухшаговая запись `Record` не атомарна между шагами, известное ограничение).

### Интеграционные тесты

Тесты, требующие реального PostgreSQL или Qdrant, — за build-тегом `integration` и пропускаются (`t.Skip`), если не задана соответствующая переменная (`TEST_DATABASE_URL`/`TEST_QDRANT_URL`) — обычный `go test ./...` (и, следовательно, `make verify`) их не запускает и не видит Docker как зависимость.

```bash
docker compose up -d
export TEST_DATABASE_URL="postgres://ai_studio_os:ai_studio_os@localhost:5432/ai_studio_os?sslmode=disable"
export TEST_QDRANT_URL="http://localhost:6333"
go test -tags=integration ./...
```

CI-job `integration` (`.github/workflows/verify.yml`, TASK-051/062) поднимает сервис-контейнеры PostgreSQL и Qdrant и запускает эти тесты автоматически на каждый PR — отдельно от обязательного статус-чека `verify`, его падение не блокирует merge (см. риски EPIC-005: решения о хранении секретов вроде GitHub PAT в этот job не входят).

### `eventbus` — производственная шина

`Bus` (TASK-049) реализует `platform.EventBus`: та же семантика, что и тестовый `internal/application/inmemory.EventBus` из EPIC-004 (синхронная доставка всем текущим подписчикам типа, в порядке регистрации, отменяемые подписки) — плюс журнал в PostgreSQL (`event_journal`, обязателен по ADR-002).

`Publish` сначала журналирует событие, затем доставляет подписчикам — если запись в журнал не удалась, `Publish` возвращает ошибку и ни один обработчик не вызывается: журнал никогда не отстаёт от того, что видели подписчики.

Событие несёт только фиксированные поля контракта `platform.Event` (ID/Type/SchemaVersion/OccurredAt/Source/Actor/ProjectID/SubjectID) — специфичные для типа события данные (например, `ReviewCompleted`'s `to`) добавляются через `Envelope.WithData`/`Data()` (EPIC-004) поверх контракта. `eventbus` не импортирует `internal/application`: он ловит такие данные через локальный структурный интерфейс `dataCarrier` (`Data() map[string]string`) — любой конверт с таким методом (включая `Envelope`) будет прожурналирован полностью, без явной зависимости пакета.

Тесты `Bus.Publish`/`Subscribe`/`Cancel` не требуют реальной БД: пул сужен до интерфейса `execer` (только нужный `Exec`), тесты подставляют фейк — 93.3% покрытия без Docker. Отдельный интеграционный тест подтверждает реальную запись в `event_journal`.

Сами адаптеры доменных событий не порождают — события создаются use-case'ами `internal/application`.

`eventbus.ReadJournal(ctx, pool)` (TASK-051) читает `event_journal` целиком и восстанавливает значения `platform.Event` (включая `dataCarrier`-данные) — назначение журнала из ADR-002/event-model.md («перестроение проекций») доведено до работающего кода: `TaskProjection.Rebuild` может строиться из реальной БД, не только из живой шины.

### `wiring` — composition root

`wiring.New(ctx, dsn)` (TASK-051) подключается к PostgreSQL, применяет миграции и собирает `System`: пять Store, производственный `EventBus`, `RepositoryProvider` (nil, если `GITHUB_TOKEN` не задан — GitHub-адаптер не зависит от PostgreSQL и его отсутствие не должно ломать остальное). Не поднимает HTTP-сервер и ничего не доставляет наружу — это задача v0.9 (API); `System` существует, чтобы одни и те же сервисы `internal/application` могли работать на реальной инфраструктуре в тестах уже сейчас и за будущим API-слоем позже.

Интеграционный тест `TestGoldenPath_Infrastructure` — тот же сценарий, что и `TestGoldenPath_Application` (EPIC-004), на реальных адаптерах: ни `internal/application`, ни `internal/domain` не изменены ни на строку. Единственное исключение — `RepositoryProvider`: используется тот же in-memory фейк EPIC-004, поскольку реального GitHub-токена в этой среде нет (см. Open Question TASK-050).

## Статус

Завершён (EPIC-005)

## Последнее обновление

2026-07-22
