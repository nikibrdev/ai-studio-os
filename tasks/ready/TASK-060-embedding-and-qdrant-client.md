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

- [ ] `embed` — детерминирован, возвращает нормализованный вектор фиксированной длины 256.
- [ ] Qdrant-клиент реализует `EnsureCollection`/`Upsert`/`Search` по схеме ADR-018; тесты на `httptest.Server` покрывают успех и минимум один отказной сценарий на метод.
- [ ] `docker-compose.yml` поднимает Qdrant, пригодный для интеграционных тестов TASK-061/062.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/memory/{embedding.go,qdrant.go}` (новые), `docker-compose.yml`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — ADR-018 принят; не зависит от TASK-059 напрямую (может выполняться параллельно)

## План реализации

## История

2026-07-22 — Architect — EPIC-007 открыт; задача поставлена в очередь (третья, параллельна TASK-059 по коду).

## Отчёт о выполнении
