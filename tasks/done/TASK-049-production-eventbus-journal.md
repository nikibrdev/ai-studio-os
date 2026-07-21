# TASK-049: Производственный EventBus + журнал событий в PostgreSQL

## Тип

feature

## Эпик

[EPIC-005 Infrastructure Layer](../../docs/roadmap/EPIC-005-infrastructure-layer.md)

## Цель

Реализовать производственную версию `platform.EventBus` — синхронную внутрипроцессную шину (ADR-002 — интерфейс не меняется при будущей замене на Redis Streams/NATS), потокобезопасную (в отличие от тестового фейка EPIC-004, не обязанного быть потокобезопасным), плюс подписчик-журнал, сохраняющий каждое опубликованное событие в PostgreSQL (обязательное требование ADR-002/event-model.md).

## Контекст

Тестовый `inmemory.EventBus` (EPIC-004) остаётся в `internal/application/inmemory` для юнит-тестов use-case'ов — не удаляется и не переиспользуется напрямую как продакшен-реализация (он не потокобезопасен и не пишет журнал). Эта задача создаёт отдельную, производственную реализацию в `internal/infrastructure`.

## Scope

### Входит

- `internal/infrastructure/eventbus/bus.go` — реализация `platform.EventBus` (Publish/Subscribe), потокобезопасная (мьютекс или эквивалент), сохраняющая семантику доставки in-memory фейка (синхронно, всем текущим подписчикам типа, отменяемые подписки).
- Миграция: таблица `event_journal` (append-only: id, type, schema_version, occurred_at, source, actor, project_id, subject_id, payload).
- Подписчик-журнал: реализация, подписывающаяся на все типы событий (или общий враппер `Publish`, сохраняющий перед доставкой — выбор фиксируется в плане) и пишущая строку в `event_journal`.
- Компиляционная проверка `var _ platform.EventBus = (*Bus)(nil)`.
- Интеграционный тест (`//go:build integration`): публикация события → видно в журнале PostgreSQL и доставлено подписчику.

### Не входит

- Restart-recovery/replay из журнала как отдельный use-case (журнал существует для аудита и будущего перестроения проекций — сам механизм replay не проектируется в этой задаче, если явно не потребуется для теста).
- Redis/NATS — не входит ни в этот эпик, ни в этот MVP (ADR-002).

## Критерии приёмки

- [x] `Bus` реализует `platform.EventBus`, потокобезопасен, семантика доставки (синхронно, всем подписчикам типа на момент публикации) сохранена.
- [x] Каждый `Publish` фиксируется строкой в `event_journal` до возврата (решение: журналируем ДО доставки — см. План/Отчёт).
- [x] Миграция применяется существующим раннером (TASK-046) без его правок.
- [x] Интеграционный тест зелёный при поднятом PostgreSQL; без него — пропускается.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/eventbus/` (новый), `internal/infrastructure/postgres/migrations/` (новая `.sql`), README `internal/infrastructure`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-046

## План реализации

1. Миграция `0004_event_journal.sql` — append-only таблица `event_journal` (id, type, schema_version, occurred_at, source, actor, project_id, subject_id, `data JSONB`).
2. `internal/infrastructure/eventbus/bus.go` — `Bus{pool execer, subscribers}`; `pool` — сужен до интерфейса `execer` (только `Exec`), а не конкретный `*pgxpool.Pool`, чтобы `Publish`/`Subscribe`/`Cancel`-логику можно было протестировать фейком без реальной БД.
3. Решение по порядку журнал/доставка: **журналируем первым**. Обоснование: ADR-002 называет журнал условием переживания события падения процесса; если бы доставка шла первой, сбой между доставкой и журналированием означал бы, что подписчики уже отреагировали на событие, которого нет в журнале для восстановления/аудита — хуже, чем отказ `Publish` целиком при сбое журнала (когда вообще ничего не произошло).
4. Данные, специфичные для типа события (`Envelope.Data()`, EPIC-004), — не часть контракта `platform.Event`; чтобы журналировать их не импортируя `internal/application` (пакет `eventbus` не должен знать о конкретном типе конверта), используется локальный структурный интерфейс `dataCarrier { Data() map[string]string }` — `Envelope` ему соответствует автоматически.
5. Тесты (`bus_test.go`, фейковый `execer`): доставка в порядке регистрации, фильтрация по типу события, `Cancel` останавливает доставку, ошибка журнала блокирует доставку, ошибка обработчика распространяется, `dataCarrier`-payload корректно попадает в `data`.
6. Интеграционный тест (`bus_integration_test.go`, тег `integration`): реальная публикация → строка видна в `event_journal` с ожидаемыми полями и `data`.
7. **Обнаружено интеграционными тестами (Docker был поднят в TASK-048 и оставался запущен):** `go test ./internal/infrastructure/...` реально гоняет пакеты `eventbus` и `postgres` как отдельные процессы, оба вызывающие `Migrate` против одной и той же локальной БД — конкурентный вызов ловил гонку `CREATE TABLE`/`pg_type_typname_nsp_index` (классическая гонка Postgres на создании одноимённой таблицы). Исправлено точечно в `internal/infrastructure/postgres/migrate.go` (файл TASK-046): `Migrate` берёт выделенное соединение (`pool.Acquire`) и держит на нём `pg_advisory_lock` на всё время миграции — не входит в исходный scope этой задачи, но обнаружено её же тестами и относится к общему раннеру, используемому всеми Store.
8. README `internal/infrastructure`.
9. `make verify`, повторные прогоны интеграционных тестов против реального PostgreSQL (Docker поднят) для проверки отсутствия гонки, PR, CI, merge.

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (четвёртая, независима от TASK-047/048 по коду, но использует общий раннер миграций TASK-046).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан, включая решение «журнал прежде доставки» и обнаруженную гонку в раннере миграций.
2026-07-21 — Architect — план одобрен; порядок «журнал → доставка» и структурный интерфейс `dataCarrier` (без импорта `internal/application`) — оба решения корректны и минимальны; фикс гонки в `migrate.go` — принять как точечное исправление TASK-046 в рамках этой задачи (тот же паттерн, что и предыдущие «обнаружено по ходу» правки). Приступать.
2026-07-21 — Claude Code (Developer) — реализовано по плану. `Bus` зелёный на 93.3% юнит-покрытия без БД (фейковый `execer`); интеграционный тест подтверждает реальную запись в `event_journal`. Гонка в `Migrate` воспроизведена (`ERROR: duplicate key value violates unique constraint "pg_type_typname_nsp_index"`) и исправлена advisory lock'ом; после фикса три подряд прогона `go test -tags=integration ./internal/infrastructure/...` против реального PostgreSQL (Docker Desktop запущен) — все зелёные. Также обнаружен и исправлен второй мелкий баг в собственном интеграционном тесте (`bus_integration_test.go`): захардкоженный `id` конфликтовал с PRIMARY KEY `event_journal` при повторном прогоне на не пересозданной БД — заменён на уникальный ID с временной меткой.
2026-07-21 — Architect — Code Review: сужение пула до `execer`/`dbConn` интерфейсов — правильный компромисс тестируемости без моков всей библиотеки pgx; порядок журнал-затем-доставка обоснован корректно; advisory lock на выделенном соединении (а не на пуле) — единственный верный способ, поскольку сессионные advisory locks привязаны к соединению; `dataCarrier` как структурный интерфейс — не создаёт зависимости `eventbus` → `application`, при этом не теряет данные `Envelope`. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-049 — производственный `EventBus` + журнал событий в PostgreSQL (четвёртая задача EPIC-005).
2. **Что сделано:** `internal/infrastructure/eventbus.Bus` реализует `platform.EventBus` (та же семантика доставки, что и тестовый фейк EPIC-004) плюс журналирование в `event_journal` (миграция `0004_event_journal.sql`) — журнал пишется до доставки, событие-специфичные данные (`Envelope.WithData`) улавливаются структурным интерфейсом `dataCarrier` без импорта `internal/application`. Побочно обнаружена и исправлена гонка в раннере миграций (`internal/infrastructure/postgres/migrate.go`, TASK-046): `Migrate` теперь держит PostgreSQL advisory lock на выделенном соединении на всё время миграции.
3. **Изменённые файлы:** `internal/infrastructure/eventbus/{doc.go,bus.go,bus_test.go,bus_integration_test.go}` (новые); `internal/infrastructure/postgres/migrations/0004_event_journal.sql` (новая); `internal/infrastructure/postgres/migrate.go` (изменён — advisory lock, точечное исправление гонки); `internal/infrastructure/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/infrastructure/eventbus/... -cover` — 93.3%, без реальной БД (фейковый `execer`); Docker Desktop запущен (оставался с TASK-048) — `go test -tags=integration ./internal/infrastructure/...` три раза подряд против настоящего PostgreSQL, все зелёные, включая проверку реальной строки в `event_journal` и отсутствие гонки после фикса `migrate.go`; `make verify` — чисто.
5. **Обновлённая документация:** README `internal/infrastructure` (разделы про `eventbus` и про advisory lock раннера).
6. **Open Questions:** нет.
7. **Рекомендации:** TASK-050 (GitHub-адаптер) не зависит от PostgreSQL — может выполняться без Docker; TASK-051 должен зафиксировать CI-job с сервис-контейнером PostgreSQL, чтобы гонки такого рода (и любые будущие) ловились автоматически, а не только при случайном параллельном локальном прогоне, как в этой задаче.
