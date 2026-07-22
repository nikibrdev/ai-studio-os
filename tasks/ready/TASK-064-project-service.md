# TASK-064: ProjectService (CreateProject/Activate)

## Тип

feature

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Добавить в `internal/application` минимальный use-case-сервис для создания и активации Project — сейчас Application Layer (EPIC-004) не содержит такого сценария (все четыре существующих сервиса действуют уже внутри Active-проекта); без него `apps/api` не сможет создавать проекты.

## Контекст

По образцу уже существующих сервисов EPIC-004 (`TaskPlanningService`, `internal/application/task_planning.go`): узкий порт хранения (`ProjectStore`, уже объявлен в `internal/application/ports.go`), события — через `platform.EventBus` и `Envelope`. Доменная модель уже готова: `internal/domain/project` (`project.New`, `project.Activate`, guard «≥1 Repository»). Решение открыть эту задачу — раздел «Контекст» [EPIC-008](../../docs/roadmap/EPIC-008-api-layer.md).

## Scope

### Входит

- `internal/application/project.go` — `ProjectService{Projects ProjectStore, Events platform.EventBus}`.
- `CreateProjectParams` + `CreateProject(ctx, params) (*project.Project, error)` — оборачивает `project.New`, сохраняет, публикует событие создания.
- `Activate(ctx, projectID, actor string) error` — оборачивает `project.Activate` (guard «≥1 Repository» уже в домене), сохраняет, публикует событие активации.
- Юнит-тесты на `inmemory`-фейке `ProjectStore`/`EventBus` (тот же паттерн, что тесты остальных сервисов EPIC-004).

### Не входит

- HTTP-хендлеры (TASK-068).
- Изменение доменного пакета `internal/domain/project` — используется как есть.

## Критерии приёмки

- [ ] `ProjectService.CreateProject`/`Activate` реализованы, стиль идентичен остальным сервисам `internal/application` (узкий порт, `Envelope`, sentinel-ошибки).
- [ ] Guard «≥1 Repository» перед активацией — проверен тестом (ошибка домена пробрасывается, а не дублируется в Application Layer).
- [ ] Юнит-тесты покрывают успешный путь и минимум по одному отказному сценарию на метод.
- [ ] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/project.go`, `internal/application/project_test.go`, `internal/application/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — использует уже принятый `internal/domain/project` и `internal/application/ports.go`

## План реализации

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь (первая — остальные хендлеры зависят от неё).

## Отчёт о выполнении
