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

- [x] Guard Testing → Done в use-case требует успешного вызова MergePullRequest — без него переход не происходит.
- [x] Порядок событий воспроизводимо TestsPassed → MergeCompleted → TaskCompleted (тест проверяет последовательность, не только факт публикации).
- [x] Покрытие 80.7% (та же природа остатка, что в TASK-042/043); `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — ADR-008 принят; зависит от TASK-040/041

## План реализации

1. `internal/application/completion.go` — `CompletionService{Tasks, Repositories platform.RepositoryProvider, Events, Rules}`:
   - `RequestReview` — In Progress → Review, публикует `ReviewRequested` (source=`task`, по каталогу events.md).
   - `CompleteReview(taskID, approved bool, actor)` — Review → Testing (approved) или Review → In Progress (changes requested), публикует `ReviewCompleted` (source=`git`, по каталогу).
   - `CompleteTesting(params{TaskID, Passed, Repository, PullRequestID, Actor})` — при провале: `TestsFailed` (source=`execution`) + Testing → In Progress. При успехе — точно порядок ADR-008: публикация `TestsPassed` (source=`execution`) → `Repositories.MergePullRequest` → публикация `MergeCompleted` (source=`git`) → **только затем** `Task.Transition(Done)` → публикация `TaskCompleted` (source=`task`). Если merge вернёт ошибку — Task не переходит в Done (guard в коде, не только в ADR-тексте).
2. Тесты: RequestReview/CompleteReview обе ветки; CompleteTesting провал; CompleteTesting успех — явная проверка порядка 3 событий (TestsPassed, MergeCompleted, TaskCompleted) через `bus.Published()`; отказ `MergePullRequest` — Task остаётся в Testing, `TaskCompleted` не публикуется (это и есть ADR-008 в виде теста, не только доккомментария); Task не найден.
3. In-memory фейк `RepositoryProvider` в `internal/application/inmemory` — ещё один пробел TASK-040, закрывается здесь по уже сложившейся практике (TASK-041 закрыла EventBus, TASK-042 — каталог событий).
4. README, `make verify`, PR, CI, merge.

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (пятая); реализует ADR-008 в коде.
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен; источники событий по каталогу events.md (ReviewCompleted/MergeCompleted = git, TestsPassed/TestsFailed = execution) — не по модулю, физически выполняющему переход, а по модулю-источнику факта. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: `inmemory/repository.go` (фейк RepositoryProvider с инъекцией ошибки MergeErr и журналом MergeCalls — специально для проверки гейта ADR-008); `completion.go` (RequestReview/CompleteReview/CompleteTesting); `completion_test.go` (12 тестов, включая `TestCompleteTesting_Passed_EventOrderMatchesADR008` — явная проверка порядка трёх событий по срезу, и `TestCompleteTesting_MergeFailure_BlocksDone` — Task остаётся в Testing, TaskCompleted не публикуется, если MergePullRequest вернул ошибку). README дополнен разделом «ADR-008 в коде». `make verify` — чисто.
2026-07-21 — Architect — Code Review: гейт реализован верно — переход в Done физически недостижим без успешного вызова Save после Transition(Done), который сам недостижим без успешного merge (короткое замыкание на ошибке); источники событий (git/execution/task) соответствуют каталогу, не интуиции о том, «кто нажал кнопку»; тест порядка событий через срез `bus.Published()[before:]` — правильный уровень строгости для этого решения. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-044 — use-case «Завершение задачи» (EPIC-004, пятая задача, реализует ADR-008).
2. **Что сделано:** `CompletionService` — Review (обе ветки), Testing (провал/успех); при успехе — точный порядок ADR-008 (TestsPassed → merge → MergeCompleted → Done → TaskCompleted) с merge как код-гейтом, не просто задокументированным ожиданием.
3. **Изменённые файлы:** `internal/application/{completion,completion_test}.go`, `internal/application/inmemory/{repository,repository_test}.go`, `internal/application/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/application/... -cover` — 80.7%/86.8%; `make verify` — чисто.
5. **Обновлённая документация:** README `internal/application` («ADR-008 в коде»).
6. **Open Questions:** нет; Task пока не хранит собственную ссылку на Repository/PullRequest (домен `git` вне scope EPIC-003/004) — параметры передаются вызывающей стороной, как и Executor в TASK-042.
7. **Рекомендации:** TASK-045 — последняя задача эпика: проекция чтения и сквозной тест приложения поверх уже готовых четырёх сервисов; закрывает EPIC-004.
