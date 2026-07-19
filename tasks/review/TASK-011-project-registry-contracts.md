# TASK-011: Контракты реестра проектов (internal/project)

## Тип

feature

## Эпик

EPIC-002 ([docs/roadmap/EPIC-002-foundation.md](../../docs/roadmap/EPIC-002-foundation.md))

## Цель

Пакет `internal/project` с контрактом реестра проектов: Registry (создание, архивирование, подключение репозитория) — по [domain-model.md](../../docs/architecture/domain-model.md) (Project).

## Контекст

Способ подключения репозиториев и формат `projects/` — Decision Required ([ADR-013](../../docs/adr/ADR-013-managed-projects.md)): контракт фиксирует только операции, не зависящие от этого решения; идентификаторы — строки (до [ADR-011](../../docs/adr/ADR-011-task-identifiers.md)).

## Scope

### Входит

- `internal/project/registry.go`: интерфейс Registry с doc-комментариями (жизненный цикл Project: Created → Active → Archived).

### Не входит

- Реализация; формат манифестов проектов; назначения исполнителей ролей (Domain Layer).

## Критерии приёмки

- [ ] `go build ./...`, `go vet ./...` — без ошибок; форматирование gofumpt.
- [ ] Только интерфейсы; зависимостей за пределами stdlib нет.

## Затрагиваемые модули и документы

- `internal/project/`

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-19 — Architect — задача определена в составе EPIC-002.
2026-07-19 — Claude Code (Developer) — задача оформлена в ready.
2026-07-19 — Claude Code (Developer) — выполнена, переведена в review.
2026-07-19 — Architect — code review, замечание № 3 (структура): пакет перенесён `internal/project` → `internal/domain/project` ([запись ревью](../../engineering/reviews/2026-07-19-epic-002-code-review.md)).
2026-07-19 — Claude Code (Developer) — перенос выполнен; проверки verify пройдены.

## Отчёт о выполнении

1. **Задача:** TASK-011 — контракты реестра проектов (internal/project).
2. **Что сделано:** создан пакет `internal/project` (`registry.go`): интерфейс Registry (Create, ConnectRepository, Archive) с жизненным циклом Project (Created → Active → Archived, архив неизменяем); зависимости — только stdlib.
3. **Изменённые файлы:** `internal/project/registry.go` (новый).
4. **Как проверялось:** `go build ./...` — OK; `go vet ./...` — OK; `gofumpt -l .` — чисто; `go list ./...` — 5 пакетов (core, events, project, tasks, workflow).
5. **Обновлённая документация:** не требуется.
6. **Open Questions:** нет (формат подключения — ADR-013).
7. **Рекомендации:** назначения исполнителей ролей в проекте (владение — модуль project по domain-model.md) добавить контрактом в Domain Layer.
