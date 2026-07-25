# TASK-080: Курсор журнала событий (монотонная последовательность) и подключение EventJournal в wiring.System

## Тип

feature

## Эпик

[EPIC-010 Orchestrator](../../docs/roadmap/EPIC-010-orchestrator.md)

## Цель

Реализовать курсорный запрос к журналу событий (`ReadJournalSince`) и подключить его как реализацию порта `EventJournal` (TASK-079) в `internal/infrastructure/wiring.System` — единственный способ, которым `apps/orchestrator`, будучи отдельным процессом, может узнавать о новых событиях (см. «Контекст» EPIC-010: продуктивная `eventbus.Bus` — только внутрипроцессная).

## Контекст

`internal/infrastructure/eventbus/journal.go` уже содержит `ReadJournal` (весь журнал целиком, для восстановления проекций) и таблицу `event_journal` (миграция 0004). Нужно добавить курсорную выборку, не заменяя существующую функцию (`ReadJournal` продолжает использоваться для полного восстановления).

**Изменение проектного решения при реализации (решение архитектора, подтверждено владельцем проекта 2026-07-25).** TASK-079 объявил порт как `Since(ctx, after time.Time)`. При разборе схемы перед реализацией обнаружено, что курсор по `occurred_at` **молча теряет события** — и это не теоретический риск:

- `occurred_at` заполняется доменным `time.Now()` **до** записи строки в БД. При двух параллельных запросах событие A может получить штамп `T1`, событие B — `T2 > T1`, но commit B произойти раньше. Опросчик увидит B, сдвинет курсор на `T2` — и строка A с `T1 < T2` не будет прочитана никогда.
- `WHERE occurred_at > $1` пропускает «братьев» с тем же штампом (`WorkService.StartTask` публикует три события подряд; `TIMESTAMPTZ` — микросекундная точность).

В отличие от уже принятого риска эпика («курсор в памяти, теряется при перезапуске» — оператор видит, что перезапустил процесс), этот отказ **бесшумный**: задача навсегда останется в Ready, и ни одной ошибки нигде не появится. Строить на таком курсоре цикл событий Orchestrator'а — заложить в фундамент класс дефектов, который трудно диагностировать позже.

**Решение:** курсор — по монотонной последовательности вставки (`seq BIGSERIAL`), а не по времени. Это стандартный приём для опроса журнала/outbox: последовательность выдаётся в момент `INSERT`, поэтому порядок курсора и есть порядок записи — рассинхронизация со временем домена перестаёт что-либо значить. Цена — одна миграция и изменение сигнатуры порта, объявленного в TASK-079 (та же задача, тот же эпик; обнаружение при реализации — нормальный ход, а не переоткрытие решения).

## Scope

### Входит

- `internal/infrastructure/postgres/migrations/0007_event_journal_seq.sql` — `ALTER TABLE event_journal ADD COLUMN seq BIGSERIAL` + индекс по `seq`; комментарий-обоснование в стиле 0006.
- `internal/application/ports.go` — порт `EventJournal` меняет форму: `JournalEntry{Seq int64, Event platform.Event}` и `Since(ctx context.Context, afterSeq int64) ([]JournalEntry, error)`. Курсор возвращается вызывающему явно (иначе ему нечем сдвинуть позицию), а не скрыт внутри реализации.
- `internal/infrastructure/eventbus/journal.go` — `Journal` (тип с пулом) и его метод `Since`; общий хелпер сканирования строк, чтобы `ReadJournal` и `Since` не дублировали разбор.
- `internal/infrastructure/wiring/wiring.go` — `System.EventJournal application.EventJournal` (поле типа интерфейса — присваивание и есть проверка соответствия на этапе компиляции, как у `Events`/`Repository`/`Memory`; пакет `eventbus` при этом не начинает импортировать `internal/application`).
- Юнит-тесты на фейке (по образцу `bus_test.go`), интеграционный тест на реальном PostgreSQL (`//go:build integration`): проверяет монотонность `seq`, выборку строго после курсора и — главное — что событие с **более ранним** `occurred_at`, записанное **позже**, курсором не теряется (тест ровно на тот дефект, из-за которого поменялось решение).

### Не входит

- Само использование в `apps/orchestrator` (цикл опроса, хранение курсора) — TASK-081.
- Устойчивое хранение курсора между перезапусками — сознательно не входит в эпик (см. «Риски» EPIC-010). Замечание: с `seq` такое хранение становится тривиальным (одно число), если потребность появится.
- Обратное заполнение `seq` для строк, записанных до миграции — `BIGSERIAL` в `ALTER TABLE` заполняет существующие строки автоматически, порядок для них произволен (журнал до этой миграции никем по курсору не читался, поэтому это ни на что не влияет).

## Критерии приёмки

- [x] Миграция 0007 добавляет `seq` и индекс; применяется на существующей БД с данными без ошибок.
- [x] `Journal.Since(ctx, afterSeq)` возвращает только записи со `seq` строго больше курсора, в порядке `seq`.
- [x] Событие с более ранним `occurred_at`, но записанное позже, **не теряется** — подтверждено тестом.
- [x] `wiring.System.EventJournal` реализует порт `application.EventJournal`.
- [x] Интеграционный тест на реальном PostgreSQL подтверждает курсорную выборку.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/postgres/migrations/0007_event_journal_seq.sql` (новый), `internal/application/ports.go`, `internal/infrastructure/eventbus/journal.go`, `internal/infrastructure/eventbus/journal_test.go` (новый), `internal/infrastructure/eventbus/journal_integration_test.go` (новый), `internal/infrastructure/wiring/wiring.go`, `internal/application/README.md`, `internal/infrastructure/README.md`, `docs/roadmap/EPIC-010-orchestrator.md` (фиксация изменённого решения).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-079 (порт `EventJournal` должен быть объявлен)

## План реализации

1. Миграция `0007_event_journal_seq.sql`: `ALTER TABLE event_journal ADD COLUMN seq BIGSERIAL;` + `CREATE INDEX event_journal_seq_idx ON event_journal (seq);`. Комментарий сверху — почему курсор не по `occurred_at` (в стиле обоснования 0006).
2. `internal/application/ports.go`: `JournalEntry{Seq int64, Event platform.Event}`; сигнатура `EventJournal.Since(ctx, afterSeq int64) ([]JournalEntry, error)`; doc-комментарий фиксирует, почему курсор по последовательности, а не по времени. Импорт `time` убрать, если больше не нужен.
3. `internal/infrastructure/eventbus/journal.go`:
   - Выделить `scanJournalRows(rows) ([]platform.Event, error)` из `ReadJournal` (разбор строки в `journalEvent` — общий для обеих выборок).
   - `Journal{pool *pgxpool.Pool}` + `NewJournal(pool)`; метод `Since(ctx, afterSeq int64) ([]application.JournalEntry, error)` — `WHERE seq > $1 ORDER BY seq`, возвращает записи со `seq`.
   - Проверить, придётся ли `eventbus` импортировать `internal/application` (тип `JournalEntry` в возвращаемом значении — да, придётся). Это допустимо: `internal/infrastructure/postgres` уже импортирует `internal/application` (`ErrNotFound`, проверки соответствия портам); прежний комментарий в `bus.go` про отказ от импорта касался конкретно `dataCarrier`, который сопоставляется структурно, и остаётся в силе.
4. `internal/infrastructure/wiring/wiring.go`: поле `EventJournal application.EventJournal`, заполняется `eventbus.NewJournal(pool)`.
5. Тесты:
   - `journal_test.go` — юнит-тесты на фейке пула (по образцу `bus_test.go`): формирование запроса с курсором, разбор `seq`.
   - `journal_integration_test.go` (`//go:build integration`) — на реальном PostgreSQL: (а) `seq` монотонна по порядку записи; (б) `Since` отдаёт только записи после курсора; (в) **регрессионный тест на исходный дефект** — записать событие с `occurred_at` в прошлом ПОСЛЕ события с более поздним штампом и убедиться, что `Since` его отдаёт (курсор по времени его бы потерял).
6. Документация: `internal/application/README.md` (строка `ports.go` — уточнить форму порта), `internal/infrastructure/README.md` (журнал/курсор), `docs/roadmap/EPIC-010-orchestrator.md` («Контекст» и «Риски» — зафиксировать изменённое решение и то, что риск потери событий по времени снят).
7. `make verify`; интеграционные тесты на живом PostgreSQL (`docker compose up -d postgres`).

## История

2026-07-23 — Architect — EPIC-010 открыт; задача поставлена в очередь, зависит от TASK-079.

2026-07-25 — Developer — при разборе схемы перед реализацией обнаружен дефект проектного решения TASK-079 (курсор по `occurred_at` бесшумно теряет события); вынесено владельцу проекта как развилка.

2026-07-25 — Владелец проекта — выбран вариант «курсор по монотонной последовательности (`seq BIGSERIAL`)»; план пересмотрен под это решение, реализация начата.

2026-07-25 — Developer — задача реализована и проверена на реальном PostgreSQL (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-080 — курсорное чтение журнала событий (`eventbus.Journal`), миграция 0007, подключение порта `EventJournal` в `wiring.System`.

### Что сделано

- **Миграция `0007_event_journal_seq.sql`** — `seq BIGSERIAL` + индекс `event_journal_seq_idx`. Комментарий в файле подробно фиксирует, почему курсор не по `occurred_at`: это не стилистический выбор, а исправление дефекта, и следующий читатель должен видеть причину, не восстанавливая её из истории задач.
- **Порт `EventJournal` переработан** (`internal/application/ports.go`): `JournalEntry{Seq int64, Event platform.Event}`, `Since(ctx, afterSeq int64) ([]JournalEntry, error)`. `Seq` возвращается вызывающему явно — курсором владеет опросчик, а не журнал; спрятать позицию внутрь реализации значило бы сделать её непроверяемой и невосстановимой.
- **`eventbus.Journal`** — реализация порта. Разбор строки вынесен в общие `journalEventColumns`/`scanDest`/`decodeData`, поэтому список колонок и порядок сканирования у `ReadJournal` и `Since` физически не могут разойтись по отдельности.
- **`ReadJournal` переведён с `ORDER BY occurred_at` на `ORDER BY seq`** — не было в плане, но следует из той же причины: пересборка проекции (`TaskProjection.Rebuild`) должна воспроизводить события в причинном порядке (порядок записи), а доменный штамп времени его не задаёт. Оставить здесь сортировку по времени значило бы починить курсор и сохранить тот же дефект в восстановлении проекций.
- **`wiring.System.EventJournal`** — поле типа порта (`application.EventJournal`), заполняется `eventbus.NewJournal(pool)`. Присваивание интерфейсного поля и есть проверка соответствия на этапе компиляции — так же типизированы `Events`/`Repository`/`Memory`.
- **`eventbus` теперь импортирует `internal/application`** (тип `JournalEntry` в возвращаемом значении). Это не нарушение слоёв: `internal/infrastructure/postgres` импортирует `internal/application` с EPIC-005 (`ErrNotFound`, проверки соответствия портам). Прежний комментарий в `bus.go` об отказе от этого импорта касался конкретно `dataCarrier`, который по-прежнему сопоставляется структурно, и остаётся в силе.
- **4 интеграционных теста** на реальном PostgreSQL, включая **регрессионный тест на исходный дефект** (`TestJournal_Since_DoesNotSkipRowWrittenWithEarlierOccurredAt`): событие со штампом на час в прошлом записывается ПОСЛЕ события с текущим штампом; курсор по `seq` его отдаёт, курсор по времени потерял бы. Тест проверяет и то, что условие всё ещё воспроизводится (`entry[1].OccurredAt < entry[0].OccurredAt`) — иначе он тихо перестал бы что-либо доказывать.

### Изменённые файлы

- `internal/infrastructure/postgres/migrations/0007_event_journal_seq.sql` (новый) — `seq` + индекс, обоснование.
- `internal/application/ports.go` — `JournalEntry`, новая сигнатура `EventJournal.Since`.
- `internal/infrastructure/eventbus/journal.go` — `Journal`/`NewJournal`/`Since`, общий разбор строки, `ReadJournal` на `ORDER BY seq`.
- `internal/infrastructure/eventbus/journal_integration_test.go` (новый) — 4 теста.
- `internal/infrastructure/wiring/wiring.go` — поле `EventJournal`, импорт `internal/application`.
- `internal/application/README.md`, `internal/infrastructure/README.md` — состав, курсор и его обоснование.
- `docs/roadmap/EPIC-010-orchestrator.md` — уточнение решения в «Контексте», Scope, критерии, «Риски», декомпозиция.

### Как проверялось

- `make verify` — чисто (fmt, golangci-lint `0 issues.`, vet, тесты, markdownlint `0 issues`, docs-check 1334 ссылки).
- **Миграция на живой БД с данными**: применена к базе, где уже было 6 строк журнала от прогонов TASK-079 — все получили непустой `seq` (1…6), индекс `event_journal_seq_idx` создан (проверено `psql`: `\d event_journal`, `SELECT count(seq)`). Это не гипотетическая проверка «на чистой базе», а именно тот случай, который бывает у реальной установки.
- **Интеграционные тесты на реальном PostgreSQL** (`postgres:16-alpine`): 4/4 PASS — курсорная выборка, монотонность `seq`, сохранность полей и `Data()`, регрессия на потерю событий.
- **Весь интеграционный набор проекта** (`go test -tags integration -count=1 ./...`) — зелёный, включая `TestGoldenPath_Infrastructure` и `apps/api` (подтверждает, что смена сортировки `ReadJournal` ничего не сломала).

### Обновлённая документация

- `internal/application/README.md`, `internal/infrastructure/README.md`, `docs/roadmap/EPIC-010-orchestrator.md`.

### Open Questions

Нет.

### Рекомендации

- **Отклонение от плана (шаг 5, первый пункт), сознательное:** юнит-тесты на фейке пула не написаны. `Journal` читает через `pool.Query`, а подделка `pgx.Rows` — это девять методов-заглушек, из которых используются четыре; такой тест проверял бы в основном саму подделку, тогда как реальная ценность здесь — поведение `BIGSERIAL`/`JSONB`/`TIMESTAMPTZ` и курсора, доказуемое только против настоящей БД. Вся проверка перенесена в интеграционные тесты; сужение пула до интерфейса `querier` (как `execer` для `Bus`) осталось несделанным осознанно — вводить абстракцию ради теста, который ничего не доказывает, хуже, чем не иметь его.
- **Реализованный риск, найден и исправлен по ходу:** первая версия интеграционных тестов утверждала на всём окне журнала после курсора и **падала при прогоне полного набора** — `go test` запускает пакеты параллельно, и в общий журнал одновременно писали тесты `postgres`/`wiring`/`httpapi`. Исправлено фильтрацией по собственному уникальному суффиксу событий — тот же приём, который уже был закодирован в `TestProjectStore_List_ReturnsCreatedProjectsInOrder` и который стоило применить сразу. Полезный факт для будущих интеграционных тестов в этом репозитории: тестовая БД общая, изоляции между пакетами нет.
- Устойчивое хранение курсора между перезапусками (снятие последнего принятого риска эпика) теперь стоит одно целое число в таблице — если TASK-085 покажет, что перезапуски на практике мешают, это уже небольшая отдельная задача, а не переработка.
