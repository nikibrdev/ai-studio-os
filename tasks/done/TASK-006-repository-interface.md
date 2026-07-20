# TASK-006: Интерфейс Repository Provider

## Тип

feature

## Эпик

EPIC-002 ([docs/roadmap/EPIC-002-foundation.md](../../docs/roadmap/EPIC-002-foundation.md))

## Цель

Контракт git-хостинга в `internal/core`: RepositoryProvider (ветки, PR, ревью, слияние, статус) и тип PullRequestState — по [interfaces.md](../../docs/architecture/interfaces.md) и [domain-model.md](../../docs/architecture/domain-model.md).

## Контекст

Весь доступ платформы к GitHub — только через этот контракт. Политики слияния ([ADR-008](../../docs/adr/ADR-008-git-policies.md)) и способ подключения репозиториев ([ADR-013](../../docs/adr/ADR-013-managed-projects.md)) — Decision Required: контракт не фиксирует эти детали (идентификаторы — строки).

## Scope

### Входит

- `internal/core/repository.go`: интерфейс RepositoryProvider; тип PullRequestState с константами (Open, Merged, Closed).

### Не входит

- GitHub-адаптер (Infrastructure); политики слияния; вебхуки/наблюдение.

## Критерии приёмки

- [ ] `go build ./...`, `go vet ./...` — без ошибок; форматирование gofumpt.
- [ ] Состояния PR соответствуют domain-model.md.
- [ ] Только интерфейсы/типы/константы.

## Затрагиваемые модули и документы

- `internal/core/`

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-19 — Architect — задача определена в составе EPIC-002.
2026-07-19 — Claude Code (Developer) — задача оформлена в ready.
2026-07-19 — Claude Code (Developer) — выполнена, переведена в review.

## Отчёт о выполнении

1. **Задача:** TASK-006 — интерфейс Repository Provider.
2. **Что сделано:** создан `internal/core/repository.go`: интерфейс RepositoryProvider (CreateBranch, OpenPullRequest, RequestReview, MergePullRequest, ClosePullRequest, PullRequestState) и тип PullRequestState (open/merged/closed по domain-model.md). Зафиксировано: провайдер исполняет операции, но не принимает процессных решений; политики слияния — за ADR-008; идентификаторы — строки до ADR-013.
3. **Изменённые файлы:** `internal/core/repository.go` (новый).
4. **Как проверялось:** `go build ./...` — OK; `go vet ./...` — OK; `gofumpt -l .` — чисто.
5. **Обновлённая документация:** не требуется.
6. **Open Questions:** нет (ADR-008/013 уже отслеживаются).
7. **Рекомендации:** события git-процесса (PR открыт, ревью завершено, слито) добавить в контракт наблюдения при проектировании GitHub-адаптера.
