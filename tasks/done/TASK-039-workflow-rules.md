# TASK-039: Реализация workflow.Rules — каноническая state machine

## Тип

feature

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 2 (Реализация)

## Цель

Реализация контракта [`workflow.Rules`](../../internal/domain/workflow/rules.go) — таблица переходов и ролей строго по каноническим [state-machine.md](../../docs/architecture/state-machine.md) (20 разрешённых переходов, 9 состояний) и [workflow.md](../../docs/architecture/workflow.md) (таблица «Участие ролей по стадиям»). Последняя задача кода этапа 2: после неё все контракты Domain Layer, заявленные к реализации, реализованы.

## Контекст

Шестая задача этапа 2 EPIC-003 (PROJECT_HEALTH: «реализация правил — EPIC-003 этап 2»). Спецификации workflow среди пяти спецификаций этапа 1 сознательно не было ([EPIC-003](../../docs/roadmap/EPIC-003-domain-layer.md), «Не входит»); реализация ведётся напрямую по каноническому state-machine.md — решение оформлено отдельным decision-документом ([2026-07-21-workflow-rules-canonical-source.md](../../engineering/decisions/2026-07-21-workflow-rules-canonical-source.md)), чтобы не создавать прецедент «тихого» пропуска Domain Specifications First. Сущность Task (TASK-037) уже принимает Rules в Transition — эта задача даёт первую реальную реализацию вместо тестовых стабов.

## Scope

### Входит

- `internal/domain/workflow/machine.go` — тип `Machine`, реализующий `Rules`: `CanTransition` по полной таблице state-machine.md, `NextRole` по таблице workflow.md; без состояния и I/O (контрактные ограничения `rules.go`).
- Decision-документ о реализации без отдельной 20-раздельной спецификации.
- Исчерпывающие unit-тесты: все 81 пары состояний (9×9) сверяются с таблицей; NextRole для всех девяти состояний; неизвестные состояния — ошибка.
- README модуля workflow обновляется.

### Не входит

- Полная 20-раздельная спецификация workflow — отдельное решение архитектора после этапа 1 (см. decision-документ).
- Definition/Step (структуры версий процесса) — контракты остаются контрактами до появления потребителя (Application Layer, v0.4).
- Проверка причины Blocked/Cancelled — уже в сущности Task (инвариант 3 — данные, которых Rules не видит).

## Критерии приёмки

- [x] `Machine` реализует `Rules`; ровно 20 разрешённых переходов state-machine.md, все прочие пары — ошибка с указанием нарушенного правила.
- [x] `NextRole` возвращает роль по таблице workflow.md для всех девяти состояний; неизвестное состояние — ошибка.
- [x] Детерминированность и отсутствие mutable state/I/O — по построению (value-тип без полей, таблицы — неизменяемые package-level данные).
- [x] Тест покрывает все 81 пары состояний исчерпывающим перебором, не выборочно.
- [x] `make verify` — чисто; decision-документ создан; README обновлён.

## Затрагиваемые модули и документы

- `internal/domain/workflow/` (новый файл machine.go + тесты); `engineering/decisions/2026-07-21-workflow-rules-canonical-source.md` (новый); README модуля.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — канонические источники: state-machine.md (переходы), workflow.md (роли); контракт Rules принят (EPIC-002)

## План реализации

1. Decision-документ: почему реализация по каноническому state-machine.md без отдельной доменной спецификации — не прецедент пропуска процесса.
2. `machine.go` — `Machine{}` (value-тип), package-level таблицы `allowedTransitions` (map from→set of to) и `stateRoles` (map state→Role); sentinel-ошибки ErrTransitionNotAllowed/ErrUnknownState, обе оборачиваются с деталями from/to.
3. `machine_test.go` — исчерпывающий перебор 9×9 против явно записанного ожидаемого множества из 20 переходов; NextRole для 9 состояний; неизвестные состояния.
4. README модуля, `make verify`, PR, CI, merge.

## История

2026-07-21 — Architect — этап 2 EPIC-003: задача поставлена в очередь (последняя задача кода этапа); решение о каноническом источнике вместо отдельной спецификации фиксируется decision-документом в этом же PR.
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: decision-документ `2026-07-21-workflow-rules-canonical-source.md` (границы решения: только Machine, не Definition/Step; не прецедент пропуска Domain Specifications First; изменение таблиц — только через ADR); `machine.go` (Machine — value-тип без полей, таблицы allowedTransitions/stateRoles как package-level данные с комментариями-именами событий на каждой строке, compile-time проверка `var _ Rules = Machine{}`); `machine_test.go` (6 тестов, 100% покрытия — исчерпывающий перебор всех 81 пары с независимо записанной ожидаемой таблицей, проверка счётчика «ровно 20», терминальность Archived, таблица ролей целиком, золотой путь Backlog→…→Archived). README модуля переписан (устаревшие ссылки на `internal/core` заменены на `internal/domain/shared`, ADR-015), `internal/domain/README.md` синхронизирован. `make verify` — чисто.
2026-07-21 — Architect — Code Review: ожидаемая таблица в тесте записана независимо от таблицы реализации — тест сверяет код с документом, а не сам с собой (ключевое качество для канона); счётчик «ровно 20» защищает от молчаливого расширения таблицы; decision-документ корректно ограничивает своё действие. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`. **Весь код этапа 2 EPIC-003 завершён** — остаётся сквозной тест фазы и закрытие эпика.

## Отчёт о выполнении

1. **Задача:** TASK-039 — реализация workflow.Rules (EPIC-003, этап 2, последняя задача кода).
2. **Что сделано:** `Machine` — реализация контракта Rules по каноническим state-machine.md (20 переходов, 9 состояний) и workflow.md (таблица ролей); решение о реализации без отдельной 20-раздельной спецификации оформлено decision-документом с явными границами.
3. **Изменённые файлы:** `internal/domain/workflow/{machine,machine_test}.go` (новые), `internal/domain/workflow/README.md` (переписан, устранены устаревшие ссылки на internal/core), `engineering/decisions/2026-07-21-workflow-rules-canonical-source.md` (новый), `internal/domain/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/domain/workflow/... -cover` — 6 тестов, 100% покрытия, исчерпывающий перебор 9×9; `make verify` — чисто.
5. **Обновлённая документация:** README модуля workflow, README слоя domain, decision-документ.
6. **Open Questions:** нет.
7. **Рекомендации:** сквозной unit-сценарий фазы A (Task проходит все состояния через реальную Machine, порождая Execution и Artifact) — следующий шаг, затем закрытие этапа 2 EPIC-003.
