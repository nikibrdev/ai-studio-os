# TASK-062: Композиция Memory Provider, CI-интеграция, README

## Тип

feature

## Эпик

[EPIC-007 Memory System](../../docs/roadmap/EPIC-007-memory-system.md)

## Цель

Подключить `internal/infrastructure/memory.Provider` к composition root (`internal/infrastructure/wiring`), расширить CI-job `integration` сервис-контейнером Qdrant, задокументировать слой.

## Контекст

`internal/infrastructure/wiring.System` (EPIC-005) уже собирает Postgres-адаптеры, EventBus и best-effort `RepositoryProvider`. `platform.MemoryProvider` — ещё один порт того же уровня; добавляется тем же образом.

## Scope

### Входит

- `wiring.System.Memory platform.MemoryProvider` — собирается в `wiring.New`, использует тот же `DATABASE_URL`-независимый путь (Qdrant — отдельный DSN/адрес, переменная окружения по аналогии с `DatabaseURLEnv`).
- `.github/workflows/verify.yml` — job `integration` дополняется сервис-контейнером Qdrant.
- README `internal/infrastructure` — раздел про Memory Provider (аналогично разделам про postgres/eventbus/github).
- `agents/README.md`/`agents/claude-code/README.md` — если уместно (Memory используется агентами) — точечное упоминание, не переработка.

### Не входит

- Автоматический вызов `Record` из `agents/claude-code` или application-сервисов — вне scope эпика (см. EPIC-007 Scope «Не входит»).

## Критерии приёмки

- [x] `wiring.System.Memory` собирается и доступен наравне с остальными адаптерами.
- [x] CI-job `integration` включает Qdrant, интеграционные тесты TASK-061 проходят в CI.
- [x] README обновлены.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/wiring/wiring.go`, `.github/workflows/verify.yml`, `internal/infrastructure/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-061

## План реализации

1. `wiring.System`: добавить поле `Memory platform.MemoryProvider`; `New(ctx, dsn, qdrantURL string)` — новый параметр `qdrantURL`, по аналогии с best-effort `Repository` (пусто → `Memory` остаётся `nil`, а не ошибка): если `qdrantURL != ""` — создать `memory.NewQdrantClient`, вызвать `EnsureCollection` один раз при старте (тот же принцип, что `postgres.Migrate`), собрать `memory.NewProvider(memory.NewFileStore(memoryRootDir), qdrant)`. `memoryRootDir = "memory"` — фиксированный путь (каталог знаний — часть репозитория, TASK-058), в отличие от `dsn`/`qdrantURL`, которые об окружении.
2. Обновить существующий вызов в `golden_path_integration_test.go` под новую сигнатуру (`qdrantURL` необязателен для этого сценария — он не касается Memory).
3. Новый `memory_integration_test.go` (тег `integration`, `TEST_DATABASE_URL`+`TEST_QDRANT_URL`) — доказывает, что `System.Memory` реально работает (`Record`→`Search` через собранный `wiring.New`), не только что поле компилируется.
4. `.github/workflows/verify.yml`: сервис-контейнер `qdrant` в job `integration`, `TEST_QDRANT_URL` в `env`; т.к. образ Qdrant не имеет утилиты для `--health-cmd`, добавлен отдельный шаг поллинга `/healthz` перед прогоном тестов.
5. `internal/infrastructure/README.md`: раздел про `memory` (по аналогии с разделами `postgres`/`eventbus`/`github`), обновление таблицы модулей и раздела «Интеграционные тесты».
6. `agents/README.md`/`agents/claude-code/README.md` — рассмотрено и осознанно не тронуто: агенты не импортируют `internal/infrastructure` (`module-boundaries.md`), автоматический вызов `Record` — вне scope эпика; точечное упоминание было бы неподкреплённым реальным кодом.
7. `make verify`, затем живой прогон интеграционных тестов (`docker compose up -d`, трижды `-count=1`) — `TestNew_WiresMemoryWhenQdrantURLProvided` и (для регрессии) `TestGoldenPath_Infrastructure`.

## История

2026-07-22 — Architect — EPIC-007 открыт; задача поставлена в очередь (пятая, после TASK-061).

2026-07-22 — Developer — задача взята в работу, реализована и проверена вживую (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-062 — композиция Memory Provider в `wiring.System`, CI-интеграция, README.

### Что сделано

- `wiring.System.Memory platform.MemoryProvider` — собирается в `wiring.New(ctx, dsn, qdrantURL)`; пустой `qdrantURL` оставляет `Memory` равным `nil` (та же терпимость к отсутствующей внешней зависимости, что best-effort `Repository` при отсутствии `GITHUB_TOKEN`, TASK-050). При непустом `qdrantURL` `wiring.New` вызывает `EnsureCollection` один раз при старте — коллекция Qdrant не создаётся заново при каждом обращении к `Provider`.
- `.github/workflows/verify.yml`: job `integration` дополнен сервис-контейнером `qdrant/qdrant:latest` и шагом ожидания `/healthz` (образ не поставляет утилиту для декларативного `--health-cmd`, в отличие от `pg_isready` у Postgres) — интеграционные тесты TASK-061/062 теперь реально прогоняются в CI, а не только локально.
- `internal/infrastructure/wiring/golden_path_integration_test.go` — обновлён вызов `New` под новую сигнатуру; `TEST_QDRANT_URL` для этого теста необязателен (сценарий не касается Memory).
- `internal/infrastructure/wiring/memory_integration_test.go` (новый, тег `integration`) — доказывает, что `System.Memory`, собранный через `wiring.New`, реально работает (`Record` → `Search`), а не просто компилируется.
- `internal/infrastructure/README.md` — раздел про `memory`, запись в таблице модулей, обновлён раздел «Интеграционные тесты» (упоминание `TEST_QDRANT_URL`).
- `agents/README.md`/`agents/claude-code/README.md` — рассмотрены, оставлены без изменений: агенты не импортируют `internal/infrastructure`, автоматическая запись в Memory — вне scope эпика; добавлять упоминание было бы опережением несуществующей интеграции.

### Изменённые файлы

- `internal/infrastructure/wiring/wiring.go` — поле `Memory`, новый параметр `qdrantURL`, сборка `Provider`.
- `internal/infrastructure/wiring/golden_path_integration_test.go` — вызов `New` под новую сигнатуру.
- `internal/infrastructure/wiring/memory_integration_test.go` — новый интеграционный тест.
- `.github/workflows/verify.yml` — сервис-контейнер Qdrant, шаг ожидания готовности, `TEST_QDRANT_URL`.
- `internal/infrastructure/README.md` — раздел про Memory Provider.
- `docs/roadmap/EPIC-007-memory-system.md` — строка TASK-062 в декомпозиции помечена `done`.

### Как проверялось

- `make verify` — чисто (fmt, lint, vet, тесты, markdownlint, docs-check).
- Живая проверка: `docker compose up -d` (Postgres + Qdrant), оба здоровы; `TEST_DATABASE_URL`+`TEST_QDRANT_URL` выставлены; `go test -tags=integration -count=1 ./internal/infrastructure/wiring/... -run TestNew_WiresMemoryWhenQdrantURLProvided -v` прогнан трижды подряд без кеша — все три прогона зелёные, `System.Memory` реально пишет и находит запись через настоящие Postgres+Qdrant. Регрессия: `TestGoldenPath_Infrastructure` (сценарий EPIC-005, Memory не используется) по-прежнему проходит с новой сигнатурой `New`. `docker compose down` после проверки — не осталось висящих контейнеров/файлов в `memory/` (тест сам подчищает через `t.Cleanup`).

### Обновлённая документация

- `internal/infrastructure/README.md`.
- `docs/roadmap/EPIC-007-memory-system.md`.

### Open Questions

Нет.

### Рекомендации

TASK-063 (закрытие эпика) может ссылаться на CI-подтверждение (`integration`-job теперь включает оба сервис-контейнера) как на доказательство критерия «интеграционные тесты — реальный прогон, а не только на фейках» при закрытии EPIC-007.
