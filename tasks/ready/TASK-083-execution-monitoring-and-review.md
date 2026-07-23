# TASK-083: Слежение за исполнением, Pull Request, RequestReview

## Тип

feature

## Эпик

[EPIC-010 Orchestrator](../../docs/roadmap/EPIC-010-orchestrator.md)

## Цель

Довести автоматизацию Developer-шага golden path до конца: после `Accept` (TASK-082) следить за исполнением до терминального состояния, зафиксировать произведённые артефакты, открыть Pull Request и перевести задачу в Review — либо, при неудаче, корректно завершить Execution как Failed.

## Контекст

`ResultService`/`CompletionService` (EPIC-004) уже реализуют нужные use-case'ы (`RecordDraftArtifact`, `PublishArtifact`, `SucceedExecution`, `FailExecution`, `RequestReview`) — эта задача только вызывает их в правильном порядке из `apps/orchestrator`, ничего не меняя в Application Layer. Открытие Pull Request'а — обязанность вызывающего кода (`agents/claude-code/README.md`), не адаптера: `Executor.Artifacts()` возвращает только `Commit`.

## Scope

### Входит

- Опрос `Executor.Status(ctx)` после `Accept` до терминального состояния (`succeeded`/`failed`), с таймаутом (константа, например 30 минут — предохранитель от зависшего контейнера).
- При успехе: `Executor.Artifacts(ctx)` → `ResultService.RecordDraftArtifact` для каждого коммита → `PublishArtifact` → `RepositoryProvider.OpenPullRequest` (заголовок/описание — из полей задачи) → `ResultService.SucceedExecution` → `CompletionService.RequestReview`.
- При неудаче/таймауте: `ResultService.FailExecution`.
- `Executor.Finish(ctx)` — вызывается всегда (успех, неудача, таймаут, паника обработчика — `defer`), чтобы контейнер не оставался запущенным.
- Юнит-тесты на фейковом `platform.Executor`/`RepositoryProvider`: успешный путь целиком, путь с ошибкой исполнения, таймаут.

### Не входит

- Ретраи при сбое — сознательно не входит в эпик (см. «Не входит» EPIC-010).
- Диспетчеризация Reviewer/QA после `ReviewRequested` — следующий эпик декомпозиции v1.0.

## Критерии приёмки

- [ ] Успешное исполнение приводит задачу в состояние Review с опубликованным Artifact и открытым Pull Request'ом.
- [ ] Неудачное исполнение (ошибка `Status`/таймаут) переводит Execution в Failed, не оставляя задачу в неопределённом состоянии.
- [ ] `Executor.Finish` вызывается во всех путях (успех/неудача/таймаут).
- [ ] Юнит-тесты покрывают все три пути на фейках.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/orchestrator/monitor.go` (или аналог), `apps/orchestrator/monitor_test.go`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-082

## План реализации

<Заполняется исполнителем до начала работы; реализация начинается только после утверждения плана.>

## История

2026-07-23 — Architect — EPIC-010 открыт; задача поставлена в очередь, зависит от TASK-082.

## Отчёт о выполнении

<Заполняется исполнителем после завершения.>
