# TASK-060: Naive-эмбеддинг + REST-клиент Qdrant

## Тип

feature

## Эпик

[EPIC-007 Memory System](../../docs/roadmap/EPIC-007-memory-system.md)

## Цель

Реализовать функцию эмбеддинга (feature hashing, ADR-018) и клиент Qdrant поверх REST API (`net/http`, без клиентской библиотеки) — создание/проверка коллекции, upsert точки, поиск по вектору с фильтром `project_id`.

## Контекст

ADR-018 фиксирует: 256 измерений, `fnv-1a` hashing trick со знаком, L2-нормализация; одна коллекция `memory_entries`; точки — UUID v4; payload самодостаточен для реконструкции `platform.MemoryEntry`. Тот же принцип, что GitHub-адаптер (EPIC-005): прямые HTTP-вызовы, не клиентская библиотека.

## Scope

### Входит

- `internal/infrastructure/memory/embedding.go` — `embed(text string) []float32` (256 измерений, детерминированно).
- `internal/infrastructure/memory/qdrant.go` — REST-клиент: `EnsureCollection(ctx)` (идемпотентно — создаёт `memory_entries`, если не существует), `Upsert(ctx, id string, vector []float32, payload map[string]any) error`, `Search(ctx, projectID string, vector []float32, limit int) ([]qdrantPoint, error)` (фильтр по `project_id` в payload).
- UUID v4 генератор (без внешней библиотеки — `crypto/rand` + форматирование, тот же принцип, что `internal/application/id.go`, но с валидным по формату UUID, который Qdrant принимает как ID точки).
- Юнит-тесты на `embed` (детерминированность, разные тексты — разные векторы, нормализация) и на HTTP-клиенте (`httptest.Server`, тот же паттерн, что GitHub-адаптер — успех + отказные сценарии).
- `docker-compose.yml` — сервис `qdrant`.

### Не входит

- Файловое хранилище (TASK-059, уже сделано).
- `MemoryProvider` целиком, реальная интеграция Record/Search (TASK-061).

## Критерии приёмки

- [x] `embed` — детерминирован, возвращает нормализованный вектор фиксированной длины 256.
- [x] Qdrant-клиент реализует `EnsureCollection`/`Upsert`/`Search` по схеме ADR-018; тесты на `httptest.Server` покрывают успех и минимум один отказной сценарий на метод.
- [x] `docker-compose.yml` поднимает Qdrant, пригодный для интеграционных тестов TASK-061/062 — подтверждено вживую (см. Отчёт).
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/memory/{embedding.go,qdrant.go}` (новые), `docker-compose.yml`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — ADR-018 принят; не зависит от TASK-059 напрямую (может выполняться параллельно)

## План реализации

1. `embed(text) []float32` — токенизация (нижний регистр, разбиение по границам не-буквенно-цифровых символов), `fnv-1a` на токен → индекс `[0, 256)`, знак из отдельного бита хеша (signed hashing trick), L2-нормализация. Пустой текст → нулевой вектор (норма 0, не делится).
2. UUID v4 генератор уже реализован в TASK-059 (`entry.go`'s `newID()`) — используется напрямую, повторно не пишется.
3. `QdrantClient` — `EnsureCollection` (GET, при 404 — PUT с `{"vectors": {"size": 256, "distance": "Cosine"}}`), `Upsert` (PUT `.../points`), `Search` (POST `.../points/search` с `filter.must` по `project_id`) — общий `do()`-хелпер (тот же паттерн, что GitHub-адаптер EPIC-005): маршалинг, заголовки, `*QdrantAPIError` при статусе ≥300.
4. Тесты на `httptest.Server`: успех + отказной сценарий на каждый метод, проверка тела запроса (схема коллекции, фильтр по `project_id`).
5. `docker-compose.yml` — сервис `qdrant/qdrant:latest`, отдельный том.
6. Дополнительно (сверх минимума Scope, но не пересекается с TASK-061): интеграционный тест на клиенте (не на `Provider` целиком) против настоящего Qdrant — `TEST_QDRANT_URL`, тот же паттерн opt-in, что `TEST_DATABASE_URL`/`TEST_DOCKER`. Оправдано тем, что Docker/Qdrant были подняты в рамках этой же задачи — грех не проверить клиент вживую, пока инфраструктура доступна; полная интеграция `Record`/`Search`/`Reindex` через файловое хранilище — по-прежнему TASK-061.

## История

2026-07-22 — Architect — EPIC-007 открыт; задача поставлена в очередь (третья, параллельна TASK-059 по коду).
2026-07-22 — Claude Code (Developer) — задача взята в работу, план записан, включая решение добавить интеграционный тест уровня клиента (не только httptest-фейк), раз Qdrant поднимается в этой же задаче.
2026-07-22 — Architect — план одобрен; интеграционный тест клиента — разумное расширение проверки, не смешивается с scope TASK-061 (там — полный `Provider`, не голый клиент). Приступать.
2026-07-22 — Claude Code (Developer) — реализовано по плану. Единственная правка после первого прогона линтера — `TestEnsureCollection_CreatesWhenMissing` использовал `switch { case r.Method == ... }` вместо тегированного `switch r.Method { case ... }` (замечание `staticcheck QF1002`), исправлено. Реальный прогон против настоящего Qdrant (`docker compose up -d qdrant`, свежепойманный образ) — три раза подряд зелёный: `EnsureCollection` идемпотентен, `Upsert`+`Search` с фильтром по `project_id` подтверждают изоляцию между двумя проектами вживую, не только в фейковом тесте.
2026-07-22 — Architect — Code Review: схема коллекции и фильтр поиска точно соответствуют ADR-018; общий `do()`-хелпер переиспользует уже проверенный паттерн GitHub-адаптера; интеграционный тест клиента — правильная граница (не залезает в scope TASK-061). Замечаний нет. Approve.
2026-07-22 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-060 — naive-эмбеддинг + REST-клиент Qdrant (третья задача EPIC-007).
2. **Что сделано:** `embed` — детерминированный feature hashing (256 измерений, L2-нормализация); `QdrantClient` реализует `EnsureCollection`/`Upsert`/`Search` по схеме ADR-018 напрямую через `net/http`; `docker-compose.yml` дополнен сервисом `qdrant`.
3. **Изменённые файлы:** `internal/infrastructure/memory/{embedding.go,qdrant.go}`, `{embedding_test.go,qdrant_test.go,qdrant_integration_test.go}` (новые); `docker-compose.yml`; файл задачи.
4. **Как проверялось:** `go test ./internal/infrastructure/memory/... -cover` — 88.0%, 21 тест (включая унаследованные из TASK-059), все зелёные; интеграционный тест клиента (`-tags=integration`, `TEST_QDRANT_URL`, реальный Qdrant через `docker compose up -d qdrant`) — три прогона подряд, подтверждена реальная изоляция по `project_id`; `make verify` — чисто.
5. **Обновлённая документация:** нет отдельно — комментарии в коде описывают схему и решения.
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-061 собирает `FileStore` (TASK-059) и `QdrantClient`+`embed` (эта задача) в `Provider`, реализующий `platform.MemoryProvider` целиком, с интеграционным тестом на полном цикле `Record`/`Search`/`Reindex`.
