# TASK-026: Принять ADR-005 (Executor Contract) и переименовать Agent → Executor в platform/

## Тип

refactor

## Эпик

Вне эпика — архитектурное решение, принятое архитектором проекта 2026-07-20 (до начала EPIC-003), закрывающее последний блокирующий вопрос перед Domain Layer.

## Цель

`internal/platform` использует `Executor`, а не `Agent`; контракт зафиксирован минимальным — ровно четыре возможности (Accept Task, Produce Artifact, Report Status, Finish Execution); ADR-005 принят.

## Контекст

Решение архитектора: Agent (логическая роль, например Developer) и Executor (реальный исполнитель — Claude Code/Codex/Human/OpenHands) — разные понятия ([docs/domain/ubiquitous-language.md](../../docs/domain/ubiquitous-language.md)). В коде платформы («она запускает исполнителей, не агентов») должен использоваться `Executor`; в документации о ролях `Agent` остаётся допустимым.

Контракт `platform.Agent.Execute(ctx, Request) (Response, error)` был намеренно абстрактным до принятия этого ADR (ревью EPIC-002); абстрактность была правильным решением — теперь она конкретизируется в форме, названной архитектором, а не додуманной агентом-разработчиком.

## Scope

### Входит

- `docs/adr/ADR-005-agent-adapter-contract.md` → переименован в `ADR-005-executor-contract.md`, статус **Принято**, содержание — по решению архитектора.
- `internal/platform/agent.go` → `internal/platform/executor.go`: интерфейс `Agent` → `Executor` с четырьмя методами (`Accept`, `Artifacts`, `Status`, `Finish`); типы `Request`/`Response` заменены на `ExecutorTask`, `Artifact`, `ExecutionStatus` — по-прежнему абстрактны (`any`), т.к. Domain Layer ещё не спроектирован; конкретизирует только ИМЕНА возможностей контракта, не форму данных.
- Обновлены все документы, ссылавшиеся на `ADR-005-agent-adapter-contract.md` или `platform.Agent` (список — из `grep`, см. план).
- `DECISIONS_INDEX.md`: ADR-005 Decision Required → Accepted.

### Не входит

- Реализация адаптеров (`agents/`) — вне scope, будущий эпик.
- Полная форма `Artifact`/`ExecutorTask` — намеренно не проектируется (Domain Layer).
- Остальные решения этого разговора (bounded contexts, ubiquitous language, domain-model.md, VISION.md) — отдельная задача (TASK-027), чтобы не смешивать код-переименование с широкой доке-перестройкой в одном PR.

## Критерии приёмки

- [x] `go build ./...`, `go vet ./...`, `golangci-lint run ./...` — чисто.
- [x] Ни одной оставшейся ссылки на `platform.Agent` или файл `ADR-005-agent-adapter-contract.md` в репозитории (проверено `grep`).
- [x] ADR-005 — статус Принято, ровно четыре метода контракта.
- [x] `docs-check` — 0 битых ссылок после переименования файла ADR.

## План реализации

1. `git mv docs/adr/ADR-005-agent-adapter-contract.md docs/adr/ADR-005-executor-contract.md`; переписать содержание — Принято, четыре метода, зафиксировать открытый вопрос о размещении типа `Artifact` (platform vs domain, см. отчёт).
2. `git mv internal/platform/agent.go internal/platform/executor.go`; переписать: `Executor` вместо `Agent`, `ExecutorTask`/`Artifact`/`ExecutionStatus` вместо `Request`/`Response`, четыре метода.
3. Обновить `DECISIONS_INDEX.md` (статус, тема, «Блокирует»).
4. Обновить каждый файл из списка `grep` (10 файлов) — заменить ссылки на новый путь ADR и, где уместно, терминологию.
5. Полная локальная верификация (`build`, `vet`, `lint`, `docs-check`).

## Затрагиваемые модули и документы

- `internal/platform/` (rename + переписывание контракта).
- `docs/adr/ADR-005-*`, `DECISIONS_INDEX.md`, и файлы, ссылавшиеся на старый путь: ADR-006, components.md, interfaces.md, module-boundaries.md, overview.md, agents.md, ubiquitous-language.md, VISION.md, engineering/decisions/2026-07-20-vision-and-domain-foundations.md, tasks/done/TASK-003 (историческая ссылка — не менять текст задачи, только если ссылка ломается; задача в `done/` — история, а не живой документ, проверить точечно).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-20 — Architect — шесть архитектурных решений; ADR-005 сформулирован дословно (Accept Task → Produce Artifact → Report Status → Finish Execution).
2026-07-20 — Claude Code (Developer) — план записан, задача взята в работу.
2026-07-20 — Claude Code (Developer) — реализовано, локальная верификация пройдена, переведена в review.

## Отчёт о выполнении

1. **Задача:** TASK-026 — принять ADR-005 (Executor Contract) и переименовать `Agent` → `Executor` в `internal/platform`.
2. **Что сделано:**
   - `docs/adr/ADR-005-agent-adapter-contract.md` переименован (`git mv`) в `ADR-005-executor-contract.md`; статус изменён на **Принято**; содержание переписано по решению архитектора — контракт из ровно четырёх возможностей (`Accept`, `Artifacts`, `Status`, `Finish`), терминологическая цепочка Role → Agent → Executor, зафиксирован открытый вопрос о месте размещения полноценной сущности `Artifact` (platform-примитив vs domain-сущность) для решения при проектировании Domain Layer.
   - `internal/platform/agent.go` переименован (`git mv`) в `executor.go`; интерфейс `Agent` заменён на `Executor` с методами `Accept`, `Artifacts`, `Status`, `Finish`; типы `Request`/`Response` заменены на `ExecutorTask`, `Artifact`, `ExecutionStatus` — по-прежнему абстрактны (`any`) до Domain Layer.
   - `internal/platform/README.md` обновлён под новую терминологию и контракт.
   - `docs/adr/DECISIONS_INDEX.md`: ADR-005 переведён Decision Required → **Accepted**, тема переформулирована, счётчики сводки обновлены (Принято: 8, Decision Required: 7).
   - Все 9 документов, ссылавшихся на старый путь ADR-005 и/или на `platform.Agent`, обновлены: `docs/adr/ADR-006-agent-execution-environment.md`, `docs/architecture/components.md`, `docs/architecture/interfaces.md`, `docs/architecture/module-boundaries.md`, `docs/architecture/overview.md`, `docs/architecture/agents.md`, `docs/domain/ubiquitous-language.md` (раздел Agent/Executor переписан под цепочку Role → Agent → Executor, открытый вопрос о переименовании закрыт со ссылкой на ADR-005), `VISION.md`, `engineering/decisions/2026-07-20-vision-and-domain-foundations.md`.
   - `tasks/done/TASK-003-agent-interface.md` (историческая задача): точечно поправлена только ссылка на новый путь файла ADR-005 (иначе ломался `docs-check`); исторический текст задачи не переписывался.
   - В `module-boundaries.md` и `overview.md`, где терминология цитирует замороженный ADR-014 (использующий дословно «Agent Runtime»), сам ADR-014 не редактировался (принцип неизменности принятого ADR); добавлена сноска, что отображаемая терминология обновлена на «Executor Runtime» после ADR-005, а суть решения (три запрещённых перехода, четырёхэтапный пайплайн) не менялась.
3. **Изменённые файлы:**
   - `docs/adr/ADR-005-agent-adapter-contract.md` → `docs/adr/ADR-005-executor-contract.md` (rename + переписан)
   - `internal/platform/agent.go` → `internal/platform/executor.go` (rename + переписан)
   - `internal/platform/README.md`
   - `docs/adr/ADR-006-agent-execution-environment.md`
   - `docs/adr/DECISIONS_INDEX.md`
   - `docs/architecture/agents.md`
   - `docs/architecture/components.md`
   - `docs/architecture/interfaces.md`
   - `docs/architecture/module-boundaries.md`
   - `docs/architecture/overview.md`
   - `docs/domain/ubiquitous-language.md`
   - `VISION.md`
   - `engineering/decisions/2026-07-20-vision-and-domain-foundations.md`
   - `tasks/done/TASK-003-agent-interface.md` (точечная правка ссылки)
   - `PROJECT_MANIFEST.md` (Last ADR, Developer Engine, счётчик открытых решений, дата)
   - `CHANGELOG.md` (запись Unreleased/Added про ADR-005)
   - `tasks/review/TASK-026-executor-contract.md` (эта задача)
4. **Как проверялось:**
   - `go build ./...` — OK.
   - `go vet ./...` — OK.
   - `gofumpt -l .` — чисто (пустой вывод).
   - `golangci-lint run ./...` — `0 issues.`
   - `go test ./...` — OK (пакеты без тестов помечены `[no test files]`, тестов, требующих проверки, не затронуто).
   - `bash scripts/verify-docs.sh` — 721 ссылка проверена, 11 mermaid-блоков, 0 ошибок.
   - `npx markdownlint-cli2` — 149 файлов, 0 issues.
   - `grep -rn "ADR-005-agent-adapter-contract"` по всему репозиторию — совпадений нет вне текста самой задачи TASK-026 (описывающего сам факт переименования).
5. **Обновлённая документация:** список файлов — см. «Изменённые файлы» выше; все относятся к документации, кроме `internal/platform/*.go`.
6. **Open Questions:** зафиксирован один — в тексте ADR-005 (раздел «Открытый вопрос»): где размещается полноценная сущность `Artifact` (в `platform` как минимальный примитив контракта vs в `domain` как сущность языка предметной области) — требует решения архитектора при проектировании модуля Artifact в рамках Domain Layer (EPIC-003), намеренно не решён этой задачей.
7. **Рекомендации:** приступить к TASK-027 (широкая доке-перестройка: bounded-contexts.md — убрать Execution как контекст, переименовать Memory → Knowledge; domain-model.md — поднять статус Artifact, зафиксировать порядок Artifact → Execution → Executor → Task → Project; усилить закрывающий раздел VISION.md цитатой архитектора; новая запись в `engineering/decisions/`) — как и было согласовано, вне scope этой задачи.
