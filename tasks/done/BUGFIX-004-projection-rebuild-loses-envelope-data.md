# BUGFIX-004: пересборка проекции из журнала теряет данные Envelope.WithData

## Тип

fix

## Эпик

Вне эпика — исправление; обнаружено при разборе TASK-082 (EPIC-010), затрагивает уже закрытые EPIC-004 (`TaskProjection`) и EPIC-005 (`ReadJournal`).

## Цель

`TaskProjection.Rebuild`, применённый к событиям из настоящего журнала (`eventbus.ReadJournal`), молча теряет все данные, привязанные через `Envelope.WithData`: описательные поля задачи (`Title`/`Type`/`Scope`/`AcceptanceCriteria`) остаются пустыми, а `ReviewCompleted` не двигает состояние. Задача — сделать так, чтобы проекция, пересобранная из журнала, была идентична построенной на живой шине, как и обещает документация `Rebuild`.

## Контекст

Обнаружено при подготовке TASK-082: Orchestrator — отдельный процесс, он не подписан на шину `apps/api` и обязан наполнять свою `TaskProjection` из журнала (`ADR-014`: проекция — единственный путь чтения Task; `module-boundaries.md`: прямой доступ к хранилищам запрещён). Для сборки `platform.ExecutorTask` ему нужны `Title`/`Type`/`Scope`/`AcceptanceCriteria` — то есть ровно те поля, которые теряются. Без исправления агент получил бы пустой промпт.

**Причина.** `applyCreatedData` и `targetState` (`internal/application/projection.go`) извлекают данные через приведение к **конкретному** типу:

```go
env, ok := e.(Envelope)
```

События, восстановленные из журнала, имеют тип `eventbus.journalEvent`, а не `application.Envelope` — приведение не срабатывает, функция молча возвращается (`applyCreatedData`) или сообщает «состояние неизвестно» (`targetState` для `ReviewCompleted`). Пакет `eventbus` при этом на своей стороне той же границы уже делает это правильно: он читает данные через структурный интерфейс `dataCarrier` (`Data() map[string]string`) именно для того, чтобы не зависеть от конкретного типа.

**Почему не заметили раньше.** Единственный тест, пересобиравший проекцию из настоящего журнала (`TestGoldenPath_Infrastructure`, TASK-051), проверял только финальное состояние — `Done`, которое достигается по `TaskCompleted`, а этот переход в `targetState` разобран обычным `case` и в `Envelope` не нуждается. Описательные поля после пересборки не проверялись ни разу, а неоднозначность `ReviewCompleted` маскировалась тем, что дальше в сценарии всё равно приходил `TaskCompleted`. Остальные тесты `Rebuild` (`projection_test.go`, `e2e_test.go`) подают `bus.Published()` — а это настоящие значения `Envelope`, на которых приведение работает.

**Подтверждено эмпирически** до исправления (разовый пробный тест против реального PostgreSQL, журнал после прогона `TestGoldenPath_Infrastructure`):

```text
REBUILT FROM REAL JOURNAL: state="done" title="" type="" scope="" ac=[]
BUG CONFIRMED: Title is empty after rebuilding from the real journal
```

## Scope

### Входит

- `internal/application/projection.go` — `applyCreatedData` и `targetState` читают данные через структурный интерфейс (`interface{ Data() map[string]string }`) вместо приведения к `Envelope`; тот же приём, что `eventbus.dataCarrier` на другой стороне границы. Именованный приватный тип для интерфейса, чтобы обе функции ссылались на одно объявление.
- Тест, доказывающий исправление на уровне пакета: пересборка из событий, реализующих `Data()`, но **не** являющихся `Envelope` (имитация `journalEvent`) — сохраняет описательные поля и разрешает `ReviewCompleted`.
- `internal/infrastructure/wiring/golden_path_integration_test.go` — усилить существующую проверку пересборки: сравнивать не только состояние, но и описательные поля (то есть закрыть пробел, из-за которого дефект жил незамеченным).
- `internal/application/README.md` — уточнить, что данные `WithData` читаются структурно, поэтому пересборка из журнала равносильна живой шине.

### Не входит

- Изменение контракта `platform.Event` (ADR-002) — `Data()` остаётся расширением сверх контракта, а не его частью; исправление меняет только способ чтения.
- Перенос `Data()` в `platform.Event` — это было бы изменением принятого контракта, требующим ADR; структурного чтения достаточно.
- `internal/infrastructure/eventbus` — не меняется, его сторона границы уже корректна.

## Критерии приёмки

- [x] Проекция, пересобранная из настоящего журнала PostgreSQL, содержит `Title`/`Type`/`Scope`/`AcceptanceCriteria` — подтверждено интеграционным тестом на реальной БД.
- [x] `ReviewCompleted` из журнала корректно разрешается в целевое состояние.
- [x] Юнит-тест доказывает работу на событии, реализующем `Data()`, но не являющемся `Envelope`.
- [x] `make verify` — чисто; весь набор `-tags integration` — зелёный.

## Затрагиваемые модули и документы

- `internal/application/projection.go`, `internal/application/projection_test.go`, `internal/infrastructure/wiring/golden_path_integration_test.go`, `internal/application/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — блокирует TASK-082; дефект подтверждён эмпирически до начала работы

## План реализации

1. `internal/application/projection.go` — объявить приватный `dataCarrier interface { Data() map[string]string }`; в `applyCreatedData` и `targetState` заменить `e.(Envelope)` на `e.(dataCarrier)`. `Envelope` продолжает удовлетворять ему как есть, поэтому существующие тесты на `bus.Published()` остаются валидны без правок.
2. `internal/application/projection_test.go` — тест с событием-фейком, реализующим `Data()`, но не `Envelope`: описательные поля восстанавливаются, `ReviewCompleted` разрешается.
3. `internal/infrastructure/wiring/golden_path_integration_test.go` — после `Rebuild` из реального журнала сверять описательные поля с теми, что были переданы в `CreateTask` (пробел, из-за которого дефект не был виден).
4. `internal/application/README.md` — одна уточняющая фраза о структурном чтении.
5. `make verify`; полный набор `-tags integration` на живом PostgreSQL.

## История

2026-07-25 — Developer — дефект обнаружен при разборе TASK-082, подтверждён эмпирически против реального журнала; задача открыта как отдельное исправление (TASK-082 от неё зависит), по образцу BUGFIX-003.

2026-07-25 — Developer — исправлено и проверено; оба теста подтверждены как ловящие дефект (см. Отчёт).

## Отчёт о выполнении

### Задача

BUGFIX-004 — пересборка `TaskProjection` из журнала теряла данные `Envelope.WithData`.

### Что сделано

- **Исправление** (`internal/application/projection.go`): объявлен приватный `dataCarrier interface { Data() map[string]string }`; `applyCreatedData` и `targetState` приводят событие к нему вместо конкретного `Envelope`. Ровно тот же приём, что `eventbus.dataCarrier` уже применял на своей стороне этой же границы — асимметрия и была дефектом. `Envelope` удовлетворяет интерфейсу как есть, поэтому ни один существующий тест не потребовал правок.
- **Юнит-тест** (`projection_test.go`): `journalLikeEvent` — намеренно **чужой** тип, несущий данные через `Data()`, но не являющийся `Envelope`. Это единственное, что различает дефект и исправление: `bus.Published()` возвращает настоящие `Envelope`, на которых старое приведение работало, поэтому все прежние тесты дефекта не видели.
- **Закрыт пробел в интеграционном тесте** (`wiring/golden_path_integration_test.go`): после пересборки из реального журнала теперь сверяются и описательные поля, а не только состояние. Проверка только состояния и была причиной, по которой дефект жил незамеченным: `Done` достигается по `TaskCompleted`, чей переход разобран обычным `case` и в данных не нуждается.
- `internal/application/README.md` — раздел про `WithData` объясняет, почему чтение структурное, со ссылкой на этот BUGFIX.

### Изменённые файлы

- `internal/application/projection.go` — `dataCarrier`, два приведения.
- `internal/application/projection_test.go` — `journalLikeEvent` + `TestTaskProjection_RebuildFromNonEnvelopeEventsKeepsAttachedData`.
- `internal/infrastructure/wiring/golden_path_integration_test.go` — проверка описательных полей после пересборки.
- `internal/application/README.md` — обоснование структурного чтения.

### Как проверялось

- **Дефект подтверждён до исправления** (разовый пробный тест против реального PostgreSQL): `state="done" title="" type="" scope="" ac=[]` — состояние выжило, все данные `WithData` потеряны.
- **Оба новых теста подтверждены как ловящие дефект** — это ключевая проверка, без неё тест мог бы просто «проходить всегда». Приведение временно возвращено к `e.(Envelope)`:
  - юнит-тест: `FAIL` — пустые `Title`/`Type`/`Scope`, `AcceptanceCriteria = []`, и `State = review` вместо `testing`;
  - интеграционный: `FAIL` — `rebuilt view = {Title:"" Type:""}`, `Scope = ""`, `AcceptanceCriteria = []`.

  После возврата исправления оба — `PASS`.
- `make verify` — чисто (fmt, golangci-lint `0 issues.`, vet, тесты, markdownlint `0 issues`, docs-check 1342 ссылки).
- **Весь набор `-tags integration` на живом PostgreSQL** — зелёный (18 пакетов, включая `wiring`, `apps/api`, `apps/orchestrator`).

### Обновлённая документация

- `internal/application/README.md`.

### Open Questions

Нет.

### Рекомендации

- **Урок для будущих проекций:** данные сверх контракта `platform.Event` следует читать только структурно. Приведение к конкретному типу конверта работает на живой шине и незаметно ломается на любом другом источнике тех же событий (журнал, будущая внешняя шина по ADR-002). Сейчас в коде таких приведений не осталось — проверено поиском по `e.(Envelope)`.
- **Урок для тестов пересборки:** проверять не только конечное состояние, но и поля, которые несёт только `WithData`. Состояние — самый устойчивый к этому дефекту признак (большинство переходов не нуждаются в данных), поэтому именно оно создавало ложную уверенность.
- Формально это дефект EPIC-004 (`projection.go`, TASK-045), проявившийся с появлением второго источника событий в EPIC-005 (`ReadJournal`, TASK-051). Обещание «пересобираемости из журнала» было записано в документации раньше, чем реально проверено на журнале, — расхождение между заявленным и проверенным, а не ошибка в рассуждении.
