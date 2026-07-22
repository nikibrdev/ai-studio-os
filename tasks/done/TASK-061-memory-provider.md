# TASK-061: MemoryProvider целиком (Record/Search/Reindex)

## Тип

feature

## Эпик

[EPIC-007 Memory System](../../docs/roadmap/EPIC-007-memory-system.md)

## Цель

Собрать файловое хранилище (TASK-059) и Qdrant-клиент (TASK-060) в реализацию `platform.MemoryProvider` целиком: `Record` пишет файл и индексирует в Qdrant; `Search` ищет по вектору с фильтром `project_id`; `Reindex` перестраивает коллекцию Qdrant из файлов с нуля (тот же принцип, что `eventbus.ReadJournal`/`TaskProjection.Rebuild`, EPIC-005).

## Контекст

Файлы — источник истины (durable, аудируемый), Qdrant — производный индекс. Известное ограничение: запись файла и upsert в Qdrant — две отдельные операции, не атомарны между собой (тот же класс ограничений, что межагрегатные транзакции в EPIC-004/005) — `Reindex` существует в том числе как средство восстановления после расхождения.

## Scope

### Входит

- `internal/infrastructure/memory/provider.go` — `Provider{filestore, qdrant}` реализует `platform.MemoryProvider`: `Record` (файл → Qdrant, в этом порядке — при сбое Qdrant файл уже сохранён и переиндексируем); `Search` (эмбеддинг запроса → Qdrant.Search с фильтром `project_id` → реконструкция `[]platform.MemoryEntry` из payload).
- `Reindex(ctx, projectID string) error` — читает все файлы проекта, для каждого — `embed` + `Upsert`.
- Компиляционная проверка `var _ platform.MemoryProvider = (*Provider)(nil)`.
- Юнит-тесты на фейковых файловом хранилище/Qdrant-клиенте (интерфейсы, тот же паттерн, что `sandbox` в `agents/claude-code`).
- Интеграционный тест на реальном Qdrant (тег `integration`, за переменной окружения, тот же паттерн, что `TEST_DATABASE_URL`/`TEST_DOCKER`): запись в двух проектах, поиск в одном не возвращает записи другого; `Reindex` после ручной очистки коллекции восстанавливает результаты поиска.

### Не входит

- Композиция/wiring, CI-job (TASK-062).

## Критерии приёмки

- [x] `Provider` реализует `platform.MemoryProvider` без изменения контракта.
- [x] `Record` → `Search` находит только что записанную запись (базовая проверка релевантности — по совпадению токенов, честно ожидаемо для naive-эмбеддинга).
- [x] Изоляция по `project_id` подтверждена реальным тестом (не только фильтром в коде — вживую на Qdrant).
- [x] `Reindex` восстанавливает индекс из файлов с нуля — подтверждено вживую (очистить коллекцию, вызвать `Reindex`, убедиться, что поиск снова работает).
- [x] `make verify` — чисто; интеграционный тест — реальный Qdrant, пропускается без него.

## Затрагиваемые модули и документы

- `internal/infrastructure/memory/{provider.go,provider_test.go,provider_integration_test.go}`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-059 и TASK-060

## План реализации

1. `internal/infrastructure/memory/provider.go`: узкие интерфейсы `fileStore`/`vectorIndex` (подмножества `*FileStore`/`*QdrantClient`, тот же паттерн, что `sandbox` в `agents/claude-code`) — позволяют юнит-тестам подставлять фейки без файловой системы/Docker; `Provider{files, index}` реализует `platform.MemoryProvider`.
2. `Record`: конвертировать входной `platform.MemoryEntry` (интерфейс — может быть не только `memory.Entry`) в конкретный `Entry` через `restoreEntry`; записать файл, затем проиндексировать в Qdrant (`embed(content)` + payload); при ошибке индексации файл уже сохранён (известное ограничение неатомарности, EPIC-007 «Риски»).
3. `Search`: эмбеддинг запроса, `index.Search` с фильтром `project_id`, реконструкция `[]platform.MemoryEntry` из самодостаточного payload (без обращения к файлам, ADR-018).
4. `Reindex(ctx, projectID)`: `files.List` + `index.Upsert` для каждой записи — восстановление индекса из файлов с нуля.
5. Компиляционная проверка `var _ platform.MemoryProvider = (*Provider)(nil)`.
6. Юнит-тесты (`provider_test.go`) на фейковых `fileStore`/`vectorIndex`: порядок операций Record (файл → индекс, индекс не вызывается при ошибке записи файла), распространение ошибок, корректность payload/эмбеддинга, реконструкция Search, обход всех файлов в Reindex.
7. Интеграционный тест (`provider_integration_test.go`, тег `integration`, `TEST_QDRANT_URL`): запись в двух проектах, изоляция поиска, симуляция потери индекса (прямое удаление точек проекта через REST — не добавляя лишний метод `Delete` в `QdrantClient`, у которого нет для этого производственного применения) и восстановление через `Reindex` — вживую на реальном Qdrant.
8. `make verify`, затем живой прогон интеграционного теста (`-count=1`, трижды) против поднятого Docker/Qdrant.

## История

2026-07-22 — Architect — EPIC-007 открыт; задача поставлена в очередь (четвёртая, после TASK-059/060).

2026-07-22 — Developer — задача взята в работу, реализована и проверена вживую (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-061 — MemoryProvider целиком (Record/Search/Reindex).

### Что сделано

- `internal/infrastructure/memory/provider.go`: `Provider` реализует `platform.MemoryProvider` над узкими интерфейсами `fileStore`/`vectorIndex` (тот же паттерн тестируемости, что `sandbox` в `agents/claude-code/executor.go`).
- `Record` пишет файл (durable источник истины), затем индексирует в Qdrant (`embed` + самодостаточный payload: `project_id`/`kind`/`content`/`source`/`recorded_at`, ADR-018) — в этом порядке, чтобы сбой Qdrant не терял данные.
- `Search` встраивает запрос, ищет в Qdrant с фильтром `project_id`, реконструирует записи из payload без обращения к файлам.
- `Reindex(ctx, projectID)` перестраивает присутствие проекта в Qdrant из файлов с нуля — та же схема, что `eventbus.ReadJournal`/`TaskProjection.Rebuild` (EPIC-005).
- Юнит-тесты на фейках (`provider_test.go`): порядок записи (файл → индекс, индекс не трогается при ошибке файла), распространение ошибок на каждом шаге, содержимое payload и вектора, реконструкция `Search`, полный обход `Reindex`.
- Интеграционный тест (`provider_integration_test.go`, тег `integration`, `TEST_QDRANT_URL`) — сверх минимума задачи: помимо изоляции по `project_id` и восстановления `Reindex`, добавлена симуляция реальной потери индекса (прямое удаление точек проекта через Qdrant REST в самом тесте, без добавления лишнего метода в `QdrantClient`) — доказывает, что `Reindex` действительно восстанавливает результаты поиска, а не просто идемпотентно переиндексирует уже согласованное состояние.

### Изменённые файлы

- `internal/infrastructure/memory/provider.go` — реализация `Provider`.
- `internal/infrastructure/memory/provider_test.go` — юнит-тесты на фейковых `fileStore`/`vectorIndex`.
- `internal/infrastructure/memory/provider_integration_test.go` — интеграционный тест на реальном Qdrant.

### Как проверялось

- `go test ./internal/infrastructure/memory/... -cover` — 89.6% покрытия пакета, все тесты зелёные.
- `make verify` — чисто (fmt, lint, vet, тесты, markdownlint, docs-check).
- Живая проверка: поднят `docker compose up -d qdrant` (контейнер уже работал с TASK-060), `curl http://localhost:6333/healthz` — здоров; `TEST_QDRANT_URL=http://localhost:6333 go test -tags=integration -count=1 ./internal/infrastructure/memory/... -run TestProvider_RecordSearchIsolationAndReindex -v` прогнан трижды подряд с `-count=1` (без кеша) — все три прогона зелёные, подтверждена реальная изоляция по `project_id` и реальное восстановление после симулированной потери индекса.

### Обновлённая документация

- `docs/roadmap/EPIC-007-memory-system.md` — строка TASK-061 в декомпозиции помечена `done`.

### Open Questions

Нет.

### Рекомендации

TASK-062 (композиция/wiring, CI-job, README) может использовать `NewProvider(files, index)` напрямую, вызвав `EnsureCollection` один раз при старте (тот же принцип, что `postgres.Migrate` в `wiring.New`), а не при каждом обращении к `Provider`.
