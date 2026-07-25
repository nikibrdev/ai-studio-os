# EPIC-010: Orchestrator — автоматический запуск исполнителей

## Цель

Реализовать `apps/orchestrator` (первый из четырёх эпиков декомпозиции v1.0, [ROADMAP.md](../../ROADMAP.md)): координатор, который реагирует на события жизненного цикла задачи и сам запускает исполнителя через контракт Executor — без участия человека в рутинном шаге «взять задачу в работу». Результат: golden path от `TaskPlanned` (Ready) до `ReviewRequested` (Review) проходит автоматически силами реального `agents/claude-code`-исполнителя, не только вручную через `apps/api`.

## Контекст

При открытии эпика (ADR-007, принят при открытии v1.0) Orchestrator назван первым шагом декомпозиции, потому что от него зависят оба следующих эпика («Роли PM/QA» — расширяет тот же механизм на другие роли; «Механизм подтверждения человеком» — добавляет паузу поверх уже работающего автозапуска).

Разбор существующего кода при открытии эпика вскрыл блокер, не позволяющий `apps/orchestrator` быть отдельным процессом в том виде, как предполагает диаграмма [components.md](../architecture/components.md):

- **Продуктивная шина событий — только внутрипроцессная.** `internal/infrastructure/eventbus.Bus` (ADR-002) — «synchronous, in-process bus... plus a durable journal»: `Subscribe` регистрирует обработчик в памяти конкретного процесса; `apps/api` и будущий `apps/orchestrator` — разные процессы (разные `wiring.System`), поэтому подписка Orchestrator'а на шину `apps/api` физически невозможна без общей памяти. Журнал (`event_journal`, PostgreSQL) уже существует, но его текущее назначение — восстановление проекций (`ReadJournal` вычитывает всё целиком), не потоковая доставка новым процессам.

  Решение архитектора: `apps/orchestrator` **опрашивает журнал по курсору** (`Journal.Since` в `internal/infrastructure/eventbus`), а не подписывается на `Bus.Subscribe`. Это не новая архитектура, а прямое продолжение уже принятого ADR-002 («интерфейс стабилен, позже — Redis Streams/NATS»): опрос журнала — временный, объяснимый шаг между «только in-process» и настоящей шиной сообщений, не требующий новой ADR и не вводящий внешней зависимости. Курсор хранится в памяти процесса Orchestrator'а (без сохранения между перезапусками) — осознанное ограничение v1.0 (см. «Риски»), тот же принцип прагматизма, что уже применялся в ADR-012/018.

  **Уточнение решения при реализации (TASK-080, 2026-07-25, подтверждено владельцем проекта).** Изначально курсор предполагался по времени (`after time.Time`). При разборе схемы перед реализацией обнаружено, что такой курсор **бесшумно теряет события**: `occurred_at` заполняется доменным `time.Now()` до записи строки, поэтому при двух параллельных запросах строка может стать видимой со штампом более ранним, чем уже прочитанная, — курсор по времени перешагнёт её навсегда, не сообщив ни об одной ошибке (задача останется в Ready). Курсор переведён на монотонную последовательность вставки (`seq BIGSERIAL`, миграция 0007): `seq` выдаётся в момент `INSERT`, поэтому порядок курсора и есть порядок записи. Порт — `Since(ctx, afterSeq int64) ([]JournalEntry, error)`. Регрессионный тест на исходный дефект — `TestJournal_Since_DoesNotSkipRowWrittenWithEarlierOccurredAt`. Побочный эффект: `ReadJournal` (пересборка проекций) тоже переведён на `ORDER BY seq` — воспроизведение истории в причинном порядке, а не в порядке доменных штампов.

  Чтобы не открывать `apps/orchestrator` прямой доступ к `internal/infrastructure` (`module-boundaries.md`: «Запрещено: прямой доступ к хранилищам»), опрос курсора оборачивается новым узким портом `internal/application` (`EventJournal.Since`), реализация которого подключается через `wiring.System` — тот же паттерн, что все остальные порты Application Layer.

- **Нет способа зарегистрировать Executor.** `internal/domain/executor` полностью реализован (EPIC-003), но `internal/application` не содержит ни одного use-case для Register/Activate — тесты создают `Executor` напрямую через пакет домена, `apps/api` не выставляет ни одного эндпоинта для исполнителей. Без этого Orchestrator'у неоткуда взять исполнителя для назначения. Решение архитектора: минимальный `ExecutorService` (Register/Activate) в этом эпике — то же обоснование, что `ProjectService` в EPIC-008 (узкий порт хранения, события через `platform.EventBus`), плюс `ExecutorStore.List` (запрос, без доменной команды) — тот же прецедент, что `ProjectStore.List`/`TaskProjection.ListByProject` (EPIC-009).

- **`ResultService`/`CompletionService` уже готовы принимать вызовы от реального исполнителя.** `RecordDraftArtifact`/`PublishArtifact`/`SucceedExecution`/`FailExecution`/`RequestReview` (EPIC-004) не меняются — их сигнатуры уже сегодня рассчитаны на «Application-adjacent orchestration, later a real Executor adapter» (`internal/application/result.go`, `CompleteTestingParams`). Открытие Pull Request'а — не часть контракта Executor (`agents/claude-code/README.md`: «открытие Pull Request'а — задача вызывающего application-сервиса... через `platform.RepositoryProvider`»), поэтому Orchestrator, а не `agents/claude-code`, вызывает `RepositoryProvider.CreateBranch` (до `Accept`) и `OpenPullRequest` (после успешного `Artifacts`).

Роль Orchestrator'а в этом эпике — только Developer: `TaskPlanned` → `WorkService.StartTask` → реальный `agents/claude-code` → `ReviewRequested`. Диспетчеризация Reviewer/QA/PM — сознательно не входит (см. «Не входит»): она требует role-aware промптов, которые — предмет следующего эпика декомпозиции.

## Scope

### Входит

- `internal/application`: `ExecutorService` (Register/Activate, по образцу `ProjectService`), `ExecutorStore.List` (запрос), новый порт `EventJournal.Since(ctx, afterSeq int64) ([]JournalEntry, error)` (курсор по последовательности — см. уточнение в «Контексте»).
- `internal/infrastructure`: миграция `0007_event_journal_seq.sql` (`seq BIGSERIAL` + индекс); `eventbus.Journal` — курсорный запрос к `event_journal`; подключение в `wiring.System` как реализация `EventJournal`.
- `apps/orchestrator`: каркас (`main.go`, сборка через `wiring.System`, конфигурация — образ контейнера, GitHub-токен, ключ AI-провайдера, репозиторий проекта — переменные окружения, по образцу `apps/api`), идемпотентный бутстрап одного Developer-исполнителя (`agents/claude-code`) через `ExecutorService`, цикл опроса `EventJournal.Since` с курсором в памяти.
- `apps/orchestrator`: диспетчеризация Developer — на `TaskPlanned`: выбор Active Developer-исполнителя (`ExecutorStore.List`), `WorkService.StartTask`, `RepositoryProvider.CreateBranch`, построение `platform.ExecutorTask` из полей Task (Title/Type/Scope/AcceptanceCriteria — уже в `TaskView`, TASK-076), `Executor.Accept` (реальный `agents/claude-code.New`).
- `apps/orchestrator`: слежение за исполнением — опрос `Executor.Status` до терминального состояния; при успехе — `Executor.Artifacts` → `ResultService.RecordDraftArtifact`/`PublishArtifact` для каждого, `RepositoryProvider.OpenPullRequest`, `ResultService.SucceedExecution`, `CompletionService.RequestReview`; при неудаче — `FailExecution`; `Executor.Finish` — всегда, независимо от исхода.
- `apps/orchestrator/README.md`; `docs/architecture/orchestrator.md` (новый) — механизм диспетчеризации, курсорный опрос журнала, ограничения; `module-boundaries.md` — уточнение терминологии «Core» → `internal/application` в разделе `apps/orchestrator` (тот же комментарий, что уже сделан для `apps/api`), фиксация порта `EventJournal`.
- Живая проверка на реальной инфраструктуре (PostgreSQL + Docker, тот же принцип, что TASK-056/070): переходы состояния, запуск и уничтожение контейнера, открытие PR — с честно принятым ограничением по отсутствию реального `ANTHROPIC_API_KEY` в этой сессии (проверяется механика вызовов, не качество ответа AI-провайдера — тот же принятый разрыв, что EPIC-006).
- Закрытие эпика: критерии, ROADMAP (v1.0 — прогресс), PROJECT_MANIFEST, PROJECT_HEALTH, CHANGELOG.

### Не входит

- Диспетчеризация ролей Reviewer/PM/QA и их промпты — отдельный эпик декомпозиции v1.0 («Роли PM/QA»), расширяющий этот же механизм.
- Механизм подтверждения человеком (принятие DoR, финальное решение Done) — отдельный эпик декомпозиции v1.0.
- Настоящая межпроцессная шина сообщений (Redis Streams/NATS) — остаётся будущим шагом ADR-002; курсорный опрос журнала в этом эпике — временное, явно объявленное решение, не такая замена.
- Устойчивость курсора между перезапусками Orchestrator'а (потеря позиции = пропуск событий, случившихся во время простоя) — принятый риск v1.0 (см. «Риски»).
- Повторные попытки (retry) при сбое исполнителя — при `Executor.Accept`/`Status` с ошибкой задача остаётся в `In Progress` без исполнения; ручной перезапуск через `apps/api` — существующий путь, ничего нового не требует.
- API/Dashboard для управления исполнителями (список, ручная регистрация через HTTP) — `ExecutorService` в этом эпике используется только самим Orchestrator'ом для бутстрапа; выставлять наружу — решение по реальной потребности, не здесь.

## Критерии завершения

- [x] `ExecutorService` (Register/Activate) и `ExecutorStore.List` реализованы и покрыты тестами — TASK-079.
- [x] Порт `EventJournal.Since` и его реализация (`eventbus.Journal`) покрыты тестами; интеграционный тест подтверждает курсорную выборку на реальном PostgreSQL, включая регрессию на потерю событий при курсоре по времени — TASK-080.
- [x] `apps/orchestrator` при старте идемпотентно регистрирует и активирует один Developer-исполнитель — TASK-081, подтверждено вживую в БД (повторный запуск не добавил ни записи, ни событий).
- [ ] `apps/orchestrator` на событие `TaskPlanned` автоматически проводит задачу Ready → In Progress → Review через реальный `agents/claude-code` (создание ветки, запуск контейнера, сбор коммитов как Artifact, открытие Pull Request), без вызова `apps/api` человеком.
- [ ] Сценарий подтверждён вживую на реальной инфраструктуре (PostgreSQL + Docker); ограничение по отсутствию реального `ANTHROPIC_API_KEY` явно задокументировано, если применимо на момент проверки.
- [x] `docs/architecture/orchestrator.md` (новый), `module-boundaries.md`, `components.md`, `golden-path.md` синхронизированы с реализацией — TASK-082 (границы) и TASK-084 (остальное).
- [ ] PROJECT_MANIFEST/PROJECT_HEALTH/ROADMAP/CHANGELOG синхронизированы при закрытии.

## Декомпозиция

| Задача | Содержание | Статус |
| --- | --- | --- |
| TASK-079 | `ExecutorService` (Register/Activate), `ExecutorStore.List`, порт `EventJournal.Since` в `internal/application` | done |
| TASK-080 | Курсор журнала по `seq` (миграция 0007), `eventbus.Journal`, подключение в `wiring.System` | done |
| TASK-081 | Каркас `apps/orchestrator`: `main.go`, бутстрап Developer-исполнителя, цикл опроса журнала с курсором | done |
| TASK-082 | Диспетчеризация Developer: `TaskPlanned` → выбор исполнителя → ветка → `StartTask` → `Executor.Accept` (порядок исправлен по итогам живой проверки) | done |
| TASK-083 | Слежение за исполнением: опрос `Status`, сбор `Artifacts`, Pull Request, `SucceedExecution`/`FailExecution`, `RequestReview`, `Finish` | done |
| TASK-084 | `apps/orchestrator/README.md`, `docs/architecture/orchestrator.md`, синхронизация `module-boundaries.md`/`components.md`/`golden-path.md` | done |
| TASK-085 | Живая проверка на реальной инфраструктуре, закрытие эпика | ready |

## Риски и зависимости

- **Курсор опроса журнала — в памяти, не сохраняется между перезапусками Orchestrator'а.** Перезапуск во время простоя пропускает события, случившиеся за это время (задача останется в Ready, пока её не запустят вручную через `apps/api`). Принятый риск v1.0 — самостоятельная эксплуатация с одним долгоживущим процессом делает это редким; устойчивый курсор — решение по реальной потребности, не раньше. Важно: этот риск **операционно виден** (оператор знает, что перезапустил процесс), в отличие от бесшумной потери событий, которую снял переход на курсор по `seq` (TASK-080). После этого перехода устойчивое хранение курсора становится тривиальным — одно целое число.
- **Единственный Developer-исполнитель, бутстрап жёстко задан.** Подходит для доверенной однопользовательской установки (тот же принцип, что ADR-012); масштабирование на несколько исполнителей одной роли — не в этом эпике.
- **Наследуется ограничение EPIC-006**: реальный вызов AI-провайдера требует `ANTHROPIC_API_KEY`, которого может не быть в среде проверки — тот же принятый разрыв между «механика подтверждена» и «качество ответа проверено».
- **Опрос вместо подписки — временное решение**, явно привязанное к будущему шагу ADR-002 (настоящая шина сообщений); переход на неё, если/когда потребуется несколько подписчиков или доставка в реальном времени, — самостоятельная задача, не расширение этого эпика.
- Зависит от: EPIC-004 (Application Layer, ResultService/CompletionService/WorkService), EPIC-005 (Infrastructure Layer, EventBus/PostgreSQL), EPIC-006 (agents/claude-code), ADR-007 (принят при открытии v1.0).
- **Реализованные риски — два дефекта в уже закрытых эпиках, найденные этим эпиком и исправленные до его закрытия.** Оба жили в швах между слоями и были невидимы, потому что каждая сторона тестировалась своими данными:
  - [BUGFIX-004](../../tasks/done/BUGFIX-004-projection-rebuild-loses-envelope-data.md) — пересборка `TaskProjection` из журнала теряла все данные `Envelope.WithData` (приведение к конкретному типу вместо структурного). Блокировал эпик: Orchestrator наполняет проекцию из журнала и собрал бы агенту пустой промпт.
  - [BUGFIX-005](../../tasks/done/BUGFIX-005-github-repo-identifier-format.md) — GitHub-адаптер строил неверный путь API из принятого в проекте формата идентификатора репозитория. Заодно закрыт формат, который [ADR-013](../adr/ADR-013-managed-projects.md) должна была зафиксировать и не зафиксировала.

  Это ожидаемо, а не отклонение: TASK-082 — первый код проекта, который проходит весь стек до внешних систем не в тесте. Урок для будущих адаптеров: тест должен подавать в адаптер **то же значение, которое реально хранит платформа**.
- **Порядок шагов диспетчеризации исправлен по итогам живой проверки** (TASK-082): ветка создаётся до `StartTask`, иначе отказ `CreateBranch` оставлял задачу In Progress с живым Execution и без ветки, без возможности повтора. Обнаружено только вживую — на фейках `CreateBranch` не падает.

## Статус

В работе

## Последнее обновление

2026-07-25
