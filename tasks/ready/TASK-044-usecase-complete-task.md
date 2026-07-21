# TASK-044: Use-case «Завершение задачи»

## Тип

feature

## Эпик

[EPIC-004 Application Layer](../../docs/roadmap/EPIC-004-application-layer.md)

## Цель

Use-case'ы завершающей части golden path: RequestReview (In Progress → Review), CompleteReview (Review → In Progress при «changes requested» или Review → Testing при «approved»), CompleteTesting (Testing → In Progress при провале или Testing → Done при успехе — с merge PR **до** перехода в Done, по ADR-008).

## Контекст

Golden path, шаги «Reviewer проверяет → QA подтверждает → задача закрывается». ADR-008: слияние — после Testing (TestsPassed), порядок событий TestsPassed → MergeCompleted → TaskCompleted; guard перехода Testing → Done включает факт слияния.

## Scope

### Входит

- Три сервиса выше; CompleteTesting вызывает `RepositoryProvider.MergePullRequest` (порт, in-memory фейк в тестах) после TestsPassed и до перевода Task в Done; событие `MergeCompleted` публикуется между `TestsPassed` и `TaskCompleted`.
- Тесты: обе ветки Review (changes requested/approved); обе ветки Testing (failed/passed); порядок публикации событий проверяется явно (это и есть ADR-008 в коде).

### Не входит

- Реальный GitHub-адаптер RepositoryProvider (EPIC-005); интерфейс роли Reviewer/QA как исполнителей (ADR-007).

## Критерии приёмки

- [ ] Guard Testing → Done в use-case требует успешного вызова MergePullRequest — без него переход не происходит.
- [ ] Порядок событий воспроизводимо TestsPassed → MergeCompleted → TaskCompleted (тест проверяет последовательность, не только факт публикации).
- [ ] Покрытие ≥ 85%; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — ADR-008 принят; зависит от TASK-040/041

## План реализации

<Заполняется при взятии задачи в работу.>

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (пятая); реализует ADR-008 в коде.
