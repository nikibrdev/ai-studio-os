# TASK-045: Проекция чтения задач + сквозной golden-path тест приложения

## Тип

feature

## Эпик

[EPIC-004 Application Layer](../../docs/roadmap/EPIC-004-application-layer.md)

## Цель

Проекция чтения статусов задач, построенная исключительно из событий через `EventBus.Subscribe` (ADR-014 — никаких синхронных чтений чужих модулей), и сквозной тест уровня приложения: весь golden path (TASK-041→044) через in-memory адаптеры, включая ветки «changes requested» и «tests failed». Последняя задача эпика — после неё EPIC-004 закрывается.

## Контекст

Первая проекция для чтения на платформе (основа будущих Dashboard/API, v0.6+); намеренно read-only и полностью перестраиваемая — «not the source of truth» (ADR-004/014 в применении к чтению).

## Scope

### Входит

- `internal/application/…`: `TaskProjection` — подписывается на TaskCreated/TaskPlanned/TaskStarted/ReviewRequested/ReviewCompleted/TestsFailed/TestsPassed/TaskCompleted, строит текущее состояние по задаче; метод пересборки с нуля из журнала (in-memory для теста).
- Сквозной тест: создать → запланировать → начать → произвести и опубликовать Artifact → отправить на ревью (обе ветки) → протестировать (обе ветки) → завершить; проверка состояния через проекцию, а не напрямую через хранилище.
- Закрытие эпика: EPIC-004 критерии отмечены, PROJECT_MANIFEST/HEALTH/ROADMAP/CHANGELOG синхронизированы (v0.4 — Завершено).

### Не входит

- Персистентная проекция (БД) — v0.5+; HTTP-доступ к проекции — v0.9 (API).

## Критерии приёмки

- [x] Проекция не читает ничего, кроме потока событий; пересборка с нуля даёт то же состояние, что инкрементальное обновление.
- [x] Сквозной тест проходит обе ветки Review и обе ветки Testing; порядок событий ADR-008 подтверждён на уровне всего сценария, не только TASK-044.
- [x] EPIC-004 закрыт: чек-лист критериев завершения отмечен, манифест/health/roadmap/changelog обновлены.
- [x] Покрытие 83.1% (та же природа остатка, что в TASK-042/043/044); `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/`; `docs/roadmap/EPIC-004-application-layer.md`; PROJECT_MANIFEST.md; PROJECT_HEALTH.md; ROADMAP.md; CHANGELOG.md.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-040…044

## План реализации

1. Обнаружен и закрыт пробел: `Envelope` (TASK-040) несёт только общие поля `platform.Event`, без места для данных, специфичных для конкретного события — `ReviewCompleted` не может передать «Testing или обратно в In Progress» одним типом события с общими полями. Решение: `Envelope.WithData(map[string]string) Envelope` + `Data() map[string]string` — метод сверх интерфейса `platform.Event` (сам интерфейс не меняется), доступен через type assertion на конкретный тип `Envelope`, знакомый только `internal/application`. `CompletionService.CompleteReview` публикует `ReviewCompleted` с `Data{"to": <целевое состояние>}`.
2. `internal/application/projection.go` — `TaskProjection` (map id→`TaskView{ID, ProjectID, State, UpdatedAt}`), `Subscribe(bus)` — восемь подписок по именам из Scope задачи, `Handle(ctx, event)` — экспортирован отдельно от Subscribe, чтобы одну и ту же логику можно было прогнать как «вживую» через шину, так и вручную по журналу `bus.Published()` для пересборки с нуля.
3. `internal/application/projection_test.go` — инкрементальное обновление; пересборка с нуля из журнала даёт тот же результат.
4. `internal/application/e2e_test.go` — golden path целиком через все четыре сервиса на in-memory адаптерах: создать → запланировать → начать → произвести и опубликовать Artifact → успешно завершить исполнение → ревью с одной итерацией «changes requested» → повторное ревью «approved» → тесты сначала провалены → повторно пройдены → done; проверка состояния только через проекцию.
5. Закрытие EPIC-004: чек-лист критериев эпика, PROJECT_MANIFEST/PROJECT_HEALTH/ROADMAP/CHANGELOG (v0.4 — Завершено).
6. README, `make verify`, PR, CI, merge.

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (последняя, закрывает эпик).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан; обнаружен пробел Envelope (нет места для данных, специфичных для события) — включён в план.
2026-07-21 — Architect — план одобрен; `WithData` как метод сверх `platform.Event`, а не изменение самого интерфейса, — верное решение: не трогает контракт, принятый в EPIC-002/TASK-040. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: `event.go` расширен (`WithData`/`Data`); `CompletionService.CompleteReview` прикрепляет целевое состояние; `projection.go` (`TaskProjection`, восемь подписок, `Handle` отдельно от `Subscribe` для пересборки, `targetState` явно документирует, что TestsPassed не двигает состояние — Done только вместе с TaskCompleted); `projection_test.go` (4 теста, включая явную проверку неоднозначности ReviewCompleted и пересборку с нуля); `e2e_test.go` (`TestGoldenPath_Application` — полный сценарий через все четыре сервиса: create → plan → start → produce+publish artifact → succeed execution → review changes-requested → review approved → testing failed → review снова → testing passed с merge → Done; состояние проверяется только через проекцию, включая пересобранную с нуля в конце теста). README дополнен. `make verify` — чисто.
2026-07-21 — Architect — Code Review: `targetState` для ReviewCompleted корректно требует явных данных, не гадает по умолчанию; отсутствие ветки для TestsPassed в `targetState` — осознанное решение, задокументированное комментарием, а не забытый case; сквозной тест — именно то, что нужно для подтверждения результата v0.4. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`. **EPIC-004 закрыт: все шесть задач этапа выполнены.**

## Отчёт о выполнении

1. **Задача:** TASK-045 — проекция чтения + сквозной тест приложения (EPIC-004, шестая и последняя задача, закрывает эпик).
2. **Что сделано:** `TaskProjection` — read-модель Task, построенная только из восьми событий golden path, с доказанной пересобираемостью с нуля (`Rebuild`); попутно закрыт пробел `Envelope` (не было места для данных, специфичных для события) через `WithData`, не меняющий контракт `platform.Event`. Сквозной тест проводит одну задачу через весь golden path целиком на in-memory адаптерах, включая обе неблагополучные ветки (changes requested, tests failed), состояние проверяется исключительно через проекцию.
3. **Изменённые файлы:** `internal/application/{projection,projection_test,e2e_test,event,completion}.go`, `internal/application/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/application/... -cover` — 83.1%/86.8%; `go test ./internal/...` — весь Domain+Application Layer зелёный; `make verify` — чисто.
5. **Обновлённая документация:** README `internal/application`; закрытие EPIC-004 (roadmap/manifest/health/changelog) — в этом же PR, следующим шагом.
6. **Open Questions:** нет новых; известное ограничение (нет межагрегатной транзакции, TASK-042/043) остаётся в силе, распространяется и на общий golden-path сценарий, не решается здесь.
7. **Рекомендации:** EPIC-005 (Infrastructure Layer, v0.5) — первая задача может напрямую переиспользовать интерфейсы портов из TASK-040 без изменений сигнатур; известное ограничение отсутствия транзакции — решить явным ADR-подобным решением архитектора при проектировании PostgreSQL-адаптера, не по умолчанию.
