# TASK-082: Диспетчеризация Developer на TaskPlanned

## Тип

feature

## Эпик

[EPIC-010 Orchestrator](../../docs/roadmap/EPIC-010-orchestrator.md)

## Цель

Заставить `apps/orchestrator` реагировать на `TaskPlanned`: выбрать зарегистрированного Developer-исполнителя, перевести задачу в работу (`WorkService.StartTask`), создать ветку и запустить реального `agents/claude-code` (`Executor.Accept`). Это первая половина автоматизации Developer-шага golden path; вторая половина (слежение за исполнением до Review) — TASK-083.

## Контекст

`platform.ExecutorTask.Repository`/`Branch` — git-координаты, которые обязан подготовить вызывающий код (`agents/claude-code/README.md`: адаптер клонирует уже существующую ветку, не создаёт её). Имя ветки — по конвенции `docs/development/git-workflow.md` (`feature/<task-id>-<short-name>`). Репозиторий — берётся из `Project.Repositories()` (первый/единственный, для v1.0 самой платформы). `WorkService.StartTask` (EPIC-004) не меняется — уже принимает `ExecutorID` извне.

## Scope

### Входит

- `apps/orchestrator` — обработчик события `TaskPlanned` в цикле опроса (TASK-081): получить `Project`/`Task` по `ProjectID`/`SubjectID` события, выбрать Active Developer-исполнителя через `ExecutorStore.List`, вызвать `WorkService.StartTask`.
- `RepositoryProvider.CreateBranch` — имя ветки формируется из `TaskID` и заголовка задачи (slug), базовая ветка — `main`.
- Построение `platform.ExecutorTask{TaskID, ProjectID, Role: "developer", Title, Type, Scope, AcceptanceCriteria, Repository, Branch}` из `TaskView` (поля уже есть, TASK-076) и репозитория проекта.
- Конструирование реального `agents/claude-code.Executor` (`New(image, gitToken, providerAPIKey)`, значения — из конфигурации TASK-081) и вызов `Accept`.
- Обработка ошибок: если подходящего исполнителя нет или `Accept` возвращает ошибку — логировать и не терять событие молча (задача остаётся в `In Progress`, ручной путь через `apps/api` не заблокирован).
- Юнит-тесты на фейках (`ExecutorStore`, `RepositoryProvider`, фейковая реализация `platform.Executor`) — сценарий целиком, без реального Docker.

### Не входит

- Опрос `Status`/сбор `Artifacts`/открытие PR/`RequestReview`/`Finish` — TASK-083.
- Диспетчеризация других ролей — вне эпика.

## Критерии приёмки

- [ ] На `TaskPlanned` Orchestrator создаёт ветку и вызывает `Accept` реального исполнителя с корректно заполненным `ExecutorTask`.
- [ ] Задача при этом действительно переходит Ready → In Progress (через `WorkService.StartTask`, не в обход).
- [ ] Отсутствие подходящего исполнителя не приводит к панике/потере события — логируется, задача остаётся в Ready.
- [ ] Юнит-тесты на фейках покрывают успешный путь и отсутствие исполнителя.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/orchestrator/dispatch.go` (или аналог), `apps/orchestrator/dispatch_test.go`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-081

## План реализации

<Заполняется исполнителем до начала работы; реализация начинается только после утверждения плана.>

## История

2026-07-23 — Architect — EPIC-010 открыт; задача поставлена в очередь, зависит от TASK-081.

## Отчёт о выполнении

<Заполняется исполнителем после завершения.>
