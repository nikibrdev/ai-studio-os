# TASK-065: Генерация последовательного TASK-NNN на проект (ADR-011)

## Тип

feature

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Реализовать генерацию последовательного публичного идентификатора `TASK-NNN` на проект — [ADR-011](../../docs/adr/ADR-011-task-identifiers.md) поручал это модулю `task` при реализации PostgreSQL-адаптера (v0.5, EPIC-005), но выдача номера не была реализована: сейчас `CreateTaskParams.ID` — поле, которое передаёт вызывающий код. Внешний клиент `apps/api` не может безопасно вычислить следующий номер сам (гонка параллельных запросов).

## Контекст

ADR-011: «Выдача номера — модуль `task` через единый путь записи: PostgreSQL-последовательность на проект исключает коллизии при параллельном создании людьми и агентами». Не нативная PostgreSQL `SEQUENCE` (её нельзя завести динамически на проект без DDL при создании каждого нового проекта) — счётчик в отдельной таблице с атомарным `UPDATE ... RETURNING` (или `INSERT ... ON CONFLICT DO UPDATE ... RETURNING`), тот же уровень строгости, что advisory lock в раннере миграций (EPIC-005) для конкурентного доступа.

## Scope

### Входит

- Новая миграция `internal/infrastructure/postgres/migrations/0005_task_sequences.sql` — таблица счётчика на проект.
- `internal/infrastructure/postgres.TaskStore.NextID` — атомарно возвращает следующий номер `TASK-NNN` для данного `projectID` (одна операция, без гонки — конкурентный вызов проверен тестом).
- **Добавлено при реализации, не было в исходном описании задачи**: `internal/application` — новый порт `TaskIDGenerator` (`ports.go`) и опциональное поле `IDs TaskIDGenerator` на `TaskPlanningService`. Без этого генератор был бы недостижим из `apps/api` — `module-boundaries.md` запрещает `apps/api` импортировать `internal/infrastructure` напрямую, только `internal/application`. `CreateTask` использует `IDs.NextID`, только если `CreateTaskParams.ID` пуст — полностью обратно совместимо с уже принятым EPIC-004 (все существующие вызовы передают ID явно).
- Интеграционный тест на реальном PostgreSQL: N конкурентных вызовов для одного проекта не дают повторов и не теряют номер.

### Не входит

- Аналогичная последовательность для `EPIC-NNN` — эпики создаются архитектором вручную (текущая практика), не через API; не входит в эту задачу.
- Изменение формата ID (`TASK-NNN`, три цифры и более, растёт без ограничения) — уже зафиксировано ADR-011.

## Критерии приёмки

- [x] Публичный ID `TASK-NNN` для нового проекта начинается с 001 и растёт без пропусков при последовательных вызовах.
- [x] N конкурентных вызовов (реальный PostgreSQL, `-tags=integration`) для одного проекта дают N уникальных номеров без повторов и без пропусков.
- [x] `make verify` — чисто; интеграционный тест — реальный PostgreSQL, пропускается без него (`TEST_DATABASE_URL`).

## Затрагиваемые модули и документы

- `internal/infrastructure/postgres/migrations/0005_task_sequences.sql`, `internal/infrastructure/postgres/task_store.go`, `internal/infrastructure/postgres/task_id_generator_integration_test.go`, `internal/application/ports.go`, `internal/application/task_planning.go`, `internal/application/task_planning_test.go`, `internal/application/README.md`, `internal/infrastructure/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — независима от TASK-064, может выполняться параллельно

## План реализации

1. Миграция `0005_task_sequences.sql` — таблица `task_sequences(project_id PK, next_number)`, без FK на `projects` (счётчик — самостоятельная утилита, не часть агрегата Project).
2. `TaskStore.NextID(ctx, projectID)` — один `INSERT ... ON CONFLICT DO UPDATE ... RETURNING`, атомарность — на уровне блокировки строки PostgreSQL, без advisory lock (в отличие от раннера миграций — там гонка на уровне DDL, здесь — на уровне DML одной строки, разные механизмы решения).
3. Обнаружено при реализации: `apps/api` (потребитель этой задачи) не может вызвать `TaskStore.NextID` напрямую — `module-boundaries.md` разрешает `apps/api` только `internal/application`. Добавлен порт `TaskIDGenerator` в `internal/application/ports.go` и опциональное поле `IDs` на `TaskPlanningService`; `CreateTask` использует его, только если `CreateTaskParams.ID` пуст — сохраняет полную обратную совместимость с EPIC-004.
4. Тесты: юнит (`internal/application`) — фейковый `TaskIDGenerator`, путь генерации/явного ID/ошибки/отсутствия генератора; интеграционные (`internal/infrastructure/postgres`, тег `integration`) — последовательность с 1, конкурентный прогон (50 горутин) на уникальность и отсутствие пропусков.
5. `internal/application/README.md`, `internal/infrastructure/README.md` — задокументировать оба дополнения.
6. `make verify`, затем живой прогон интеграционных тестов на реальном PostgreSQL (`-count=1`, трижды).

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь.

2026-07-22 — Developer — задача взята в работу, реализована и проверена вживую (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-065 — генерация последовательного TASK-NNN на проект (ADR-011).

### Что сделано

- `internal/infrastructure/postgres/migrations/0005_task_sequences.sql` — таблица-счётчик `task_sequences(project_id, next_number)`.
- `TaskStore.NextID` — атомарная выдача следующего `TASK-NNN` одним `INSERT ... ON CONFLICT DO UPDATE ... RETURNING`; блокировка строки PostgreSQL сериализует конкурентных вызывающих без блокировок на уровне приложения.
- **Сверх исходного описания задачи**: порт `application.TaskIDGenerator` и опциональное поле `IDs` на `TaskPlanningService` — без них `apps/api` не смог бы достичь генератора (`module-boundaries.md`: `apps/api` не импортирует `internal/infrastructure`). `CreateTask` вызывает `IDs.NextID`, только если `CreateTaskParams.ID` пуст, — все существующие вызовы EPIC-004 (передающие ID явно) не затронуты.
- 5 юнит-тестов на фейковом `TaskIDGenerator` (генерация при пустом ID, явный ID игнорирует генератор, ошибка генератора пробрасывается без публикации событий, пустой ID без генератора — прежняя ошибка домена) + 2 интеграционных теста на реальном PostgreSQL (последовательность с 1; 50 конкурентных вызовов — ровно числа 1…50 без повторов и пропусков).

### Изменённые файлы

- `internal/infrastructure/postgres/migrations/0005_task_sequences.sql` — новая миграция.
- `internal/infrastructure/postgres/task_store.go` — метод `NextID`.
- `internal/infrastructure/postgres/task_id_generator_integration_test.go` — интеграционные тесты.
- `internal/application/ports.go` — порт `TaskIDGenerator`.
- `internal/application/task_planning.go` — поле `IDs`, генерация ID в `CreateTask`.
- `internal/application/task_planning_test.go` — юнит-тесты на фейке.
- `internal/application/README.md`, `internal/infrastructure/README.md` — документация.

### Как проверялось

- `go test ./internal/application/... ./internal/infrastructure/postgres/... -cover` — все тесты зелёные (application 83.0%, postgres 3.8% — юнит-часть не касается Store-методов, как и раньше).
- `make verify` — чисто.
- Живая проверка: `docker compose up -d postgres`, `TEST_DATABASE_URL` выставлен, `go test -tags=integration -count=1 ./internal/infrastructure/postgres/... -run TestTaskStore_NextID -v` прогнан трижды подряд без кеша — все прогоны зелёные, включая 50 конкурентных вызовов на один проект (числа 1…50, без повторов и пропусков, каждый раз). `docker compose down` — чисто.

### Обновлённая документация

- `internal/application/README.md`, `internal/infrastructure/README.md`.

### Open Questions

Нет.

### Рекомендации

TASK-068 (хендлеры Projects/Tasks) должен собирать `TaskPlanningService.IDs` из `postgres.NewTaskStore(pool)` в `wiring`/`main.go` — тот же `*TaskStore` уже используется как `Tasks TaskStore`, никакой отдельной сборки не требуется (один и тот же объект реализует оба порта).
