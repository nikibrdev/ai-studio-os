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

- [ ] `Provider` реализует `platform.MemoryProvider` без изменения контракта.
- [ ] `Record` → `Search` находит только что записанную запись (базовая проверка релевантности — по совпадению токенов, честно ожидаемо для naive-эмбеддинга).
- [ ] Изоляция по `project_id` подтверждена реальным тестом (не только фильтром в коде — вживую на Qdrant).
- [ ] `Reindex` восстанавливает индекс из файлов с нуля — подтверждено вживую (очистить коллекцию, вызвать `Reindex`, убедиться, что поиск снова работает).
- [ ] `make verify` — чисто; интеграционный тест — реальный Qdrant, пропускается без него.

## Затрагиваемые модули и документы

- `internal/infrastructure/memory/{provider.go,provider_test.go,provider_integration_test.go}`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-059 и TASK-060

## План реализации

## История

2026-07-22 — Architect — EPIC-007 открыт; задача поставлена в очередь (четвёртая, после TASK-059/060).

## Отчёт о выполнении
