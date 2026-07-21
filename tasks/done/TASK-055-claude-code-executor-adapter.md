# TASK-055: Адаптер Claude Code — platform.Executor

## Тип

feature

## Эпик

[EPIC-006 AI Agent Runtime](../../docs/roadmap/EPIC-006-ai-agent-runtime.md)

## Цель

Реализовать `agents/claude-code` — первый реальный адаптер `platform.Executor` (ADR-005): Accept запускает контейнер Execution (TASK-054) с задачей, Artifacts/Status опрашивают ход исполнения, Finish завершает и высвобождает ресурсы.

## Контекст

`agents/` — директория вне `internal/`, разрешено импортировать только опубликованный контракт `Executor` (`internal/platform`) и SDK/CLI своего провайдера (`module-boundaries.md`); Tool Layer не существует до v0.8 — адаптер вправе действовать напрямую. Использует конкретизированные типы TASK-052 и жизненный цикл контейнера TASK-054.

## Scope

### Входит

- `agents/claude-code/executor.go` — реализация `platform.Executor`: `Accept` (запускает контейнер с `ExecutorTask` через TASK-054, передаёт задачу Claude Code внутри контейнера), `Artifacts` (собирает произведённые артефакты — например, из git-состояния рабочей копии/сообщений commit), `Status` (состояние контейнера/процесса → `platform.ExecutionStatus`), `Finish` (останавливает и освобождает ресурсы).
- Компиляционная проверка `var _ platform.Executor = (*Executor)(nil)`.
- Тесты на логику адаптера с фейковым/моковым жизненным циклом контейнера (не требующие реального Docker или реального Claude Code) — граница между «логика адаптера» и «реальный Docker/CLI» покрывается моками.

### Не входит

- Реальный прогон с настоящими секретами и настоящим вызовом AI-провайдера (TASK-056).
- Auto-подбор Executor'а (Orchestrator) — вне scope (ADR-007, см. Контекст эпика).

## Критерии приёмки

- [x] `Executor` реализует все четыре метода контракта `platform.Executor`, использует типы TASK-052 без изменения контракта.
- [x] Жизненный цикл контейнера (TASK-054) используется, не дублируется.
- [x] Юнит-тесты на моках жизненного цикла покрывают: успешный путь, ошибку старта контейнера, вызовы до `Accept`, маппинг статусов.
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `agents/claude-code/` (новый), README.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-052 и TASK-054

## План реализации

1. `sandbox` — узкий интерфейс (`Start`/`Status`/`Exec`/`Stop`), сужающий `*container.Manager` до четырёх нужных адаптеру методов (тот же паттерн, что `commandRunner`/`execer` в TASK-049/054) — тесты подставляют фейк, реальный Docker не нужен.
2. `Executor{sandbox, gitToken, providerAPIKey, executionID, handle}` — одно значение на один Execution (`executionID` генерируется случайно при `New`, не переиспользуется).
3. `Accept` — строит промпт из `platform.ExecutorTask` (роль/заголовок/тип/scope/критерии) и запускает песочницу с `claude --print --permission-mode bypassPermissions <промпт>`. `bypassPermissions` необходим для безнадзорного запуска (некому подтверждать инструменты) — прокомментировано как решение, обоснованное существованием песочницы TASK-054, а не бесконтрольное послабление. Allowlist сети дополнен `api.anthropic.com` сверх `github.com`/`api.github.com`, которые уже добавляет `container.Manager`.
4. `Artifacts` — `git log --format=%H%n%s` внутри контейнера → `platform.Artifact{Type: "Commit"}`; открытие PR — не задача адаптера (это `ResultService`/`CompletionService` через `RepositoryProvider`, EPIC-004).
5. `Status` — `container.Status{Running, ExitCode}` → `platform.ExecutionStatus{State: running|succeeded|failed}`.
6. `Finish` — `sandbox.Stop`.
7. `ErrNotAccepted` — sentinel для вызова Artifacts/Status/Finish до Accept.
8. Тесты на фейковом `sandbox`: успешный Accept с проверкой параметров (repo/branch/секреты/allowlist/команда), распространение ошибки старта, вызовы до Accept, парсинг Artifacts, маппинг Status, вызов Stop через Finish, уникальность `executionID`. Отдельные прямые тесты на `buildPrompt`/`claudeCommand`/`parseCommitArtifacts` — не только косвенно через `Accept`.
9. README `agents/claude-code` — раздел `Executor`.

## История

2026-07-21 — Architect — EPIC-006 открыт; задача поставлена в очередь (четвёртая, после TASK-052/054).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан, включая обоснование `--permission-mode bypassPermissions` (безнадзорный запуск, граница — песочница TASK-054) и явное указание, что точность CLI-флагов Claude Code подтверждается только в TASK-056 (реальный вызов), не здесь.
2026-07-21 — Architect — план одобрен; узкий интерфейс `sandbox` — правильное продолжение уже установленного в EPIC-006 паттерна тестируемости; явное разделение «логика адаптера проверена» / «реальный вызов CLI не проверен здесь» — честная граница задачи, не наводит на мысль о большей готовности, чем есть. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано по плану. 91.1% покрытия юнит-тестами, 12 тестов, все на фейковой песочнице.
2026-07-21 — Architect — Code Review: `Executor` не хранит доменные типы — только `platform.ExecutorTask`/`Artifact`/`ExecutionStatus` (TASK-052), контракт `Executor` соблюдён буквально; `ErrNotAccepted`-проверки на всех трёх методах после `Accept` — правильная защита от неверного порядка вызовов, которую сам контракт `platform.Executor` не гарантирует по типам. Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-055 — адаптер Claude Code, реализация `platform.Executor` (четвёртая задача EPIC-006).
2. **Что сделано:** `agents/claude-code.Executor` реализует `Accept`/`Artifacts`/`Status`/`Finish` поверх `container.Manager` (TASK-054) через узкий интерфейс `sandbox`; промпт строится из полей `platform.ExecutorTask` (TASK-052); Artifacts собирает произведённые коммиты из `git log`; Status отображает состояние контейнера на канонические строки running/succeeded/failed.
3. **Изменённые файлы:** `agents/claude-code/{executor,prompt,artifacts}.go`, `{executor_test,prompt_test}.go` (новые); `agents/claude-code/README.md`; файл задачи.
4. **Как проверялось:** `go test ./agents/claude-code/... -cover` — 91.1%, 12 тестов, все на фейковой песочнице (без реального Docker/Claude Code); `make verify` — чисто.
5. **Обновлённая документация:** `agents/claude-code/README.md` (раздел `Executor`).
6. **Open Questions:** точность CLI-флагов Claude Code (`--print --permission-mode bypassPermissions`) не подтверждена реальным вызовом — только логика адаптера вокруг них. Подтверждение/корректировка — TASK-056.
7. **Рекомендации:** TASK-056 должен обеспечить реальные учётные данные (git-токен уже есть в сессии через `git credential fill`; ключ AI-провайдера для вложенного Claude Code — открытый вопрос, зафиксированный при открытии эпика) и запустить `Executor.Accept` целиком на одноразовой тестовой задаче.
