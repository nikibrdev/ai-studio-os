# EPIC-002: Foundation — контракты ядра

## Цель

Заложить кодовый фундамент платформы: Go-модуль, инструменты качества и публичные контракты ядра (`internal/core`) плюс интерфейсные пакеты модулей (`internal/events`, `internal/workflow`, `internal/tasks`, `internal/project`). **Только интерфейсы — никакой логики, реализации, API, БД.** Соответствует началу v0.2 ([ROADMAP.md](../../ROADMAP.md)).

## Контекст

Architecture Freeze выполнен: ADR-002, 003, 004, 009, 014 приняты. Контракты пишутся строго по [interfaces.md](../architecture/interfaces.md), [core.md](../architecture/core.md), [module-boundaries.md](../architecture/module-boundaries.md), [state-machine.md](../architecture/state-machine.md), [events.md](../architecture/events.md).

Процесс: один Epic = малые задачи; одна задача — не более одного модуля; после каждой задачи проект собирается и проходит проверки (`go build`, `go vet`, форматирование).

## Scope

### Входит

- Инициализация Go-модуля (единый `go.mod`, Go 1.24), `.golangci.yml`, рабочие цели Makefile.
- Интерфейсы в `internal/core`: EventBus, Agent, Tool, Workflow, Repository (Provider), Memory (Provider) + минимальные общие типы, необходимые для сигнатур.
- Интерфейсные пакеты: `internal/events` (типы событий), `internal/workflow`, `internal/tasks`, `internal/project`.

### Не входит

- Любая реализация и бизнес-логика (Domain Layer — следующий эпик).
- API, Dashboard, БД, миграции, Docker, CI/CD, тесты.
- Контракты, зависящие от нерешённых ADR (детали Agent-обмена — ADR-005/006).

## Критерии завершения

- [ ] Все задачи эпика в Done; каждая прошла сборку и проверки.
- [ ] `go build ./...` и `go vet ./...` — без ошибок; файлы отформатированы gofumpt-совместимо.
- [ ] В пакетах нет исполняемой логики — только интерфейсы, типы и константы.
- [ ] Контракты соответствуют [interfaces.md](../architecture/interfaces.md) и не противоречат замороженной архитектуре.

## Декомпозиция

| Задача | Модуль | Статус |
| --- | --- | --- |
| TASK-001 Инициализация Go-модуля и инструментов качества | — (корень) | ready → … |
| TASK-002 Интерфейс EventBus | internal/core | ready → … |
| TASK-003 Интерфейс Agent | internal/core | ready → … |
| TASK-004 Интерфейс Tool | internal/core | ready → … |
| TASK-005 Интерфейс Workflow | internal/core | ready → … |
| TASK-006 Интерфейс Repository Provider | internal/core | ready → … |
| TASK-007 Интерфейс Memory Provider | internal/core | ready → … |
| TASK-008 Типы событий | internal/events | ready → … |
| TASK-009 Контракты определений процесса | internal/workflow | ready → … |
| TASK-010 Контракты Task Engine | internal/tasks | ready → … |
| TASK-011 Контракты реестра проектов | internal/project | ready → … |

## Риски и зависимости

- Репозиторий ещё не является git-репозиторием — ветки/PR по [git-workflow.md](../development/git-workflow.md) невозможны до `git init` + настройки GitHub (Open Question процесса, не архитектуры).
- Детали контракта Agent ограничены минимумом до ADR-005/006.
- Идентификаторы — `string` до ADR-011.

## Статус

В работе

## Последнее обновление

2026-07-19
