# PROJECT_MANIFEST — паспорт проекта

## Назначение

Единый документ, отвечающий на вопрос «что сейчас представляет собой проект»: версия, статус, текущий эпик, состояние слоёв и качества. Обновляется в каждом изменении, меняющем состояние проекта (в том же PR).

## Содержание

### Паспорт

| Поле | Значение |
| --- | --- |
| **Project** | AI Studio OS |
| **Version** | v0.8 Dashboard — завершён (2026-07-23); v0.9 API — завершён ([порядок реализации](engineering/decisions/2026-07-22-api-before-dashboard-build-order.md)); v1.0 First Public MVP — **в работе** (открыт 2026-07-23, декомпозирован на 4 эпика) ([ROADMAP.md](ROADMAP.md)) |
| **Status** | **Architecture Frozen** (2026-07-19) |
| **Current Epic** | [EPIC-010 Orchestrator](docs/roadmap/EPIC-010-orchestrator.md) — первый из 4 эпиков декомпозиции v1.0: автоматический запуск Developer-исполнителя на `TaskPlanned` через контракт Executor, TASK-079…085 |
| **Current Sprint** | — (спринты не введены; итерации ведутся эпиками из 5–15 задач) |
| **Current Branch** | main |
| **Repository** | [github.com/nikibrdev/ai-studio-os](https://github.com/nikibrdev/ai-studio-os) (public) |
| **License** | Apache License 2.0 ([ADR-001](docs/adr/ADR-001-license.md)) |

### Состояние слоёв и компонентов

| Компонент | Состояние |
| --- | --- |
| Architecture | **Frozen** (ADR-002, 003, 004, 009, 014, 015 приняты) |
| Domain | **Implemented** — 5 сущностей (artifact/execution/executor/task/project) + `workflow.Machine`, инварианты покрыты тестами (81.8–100%), сквозной сценарий слоя зелёный ([goldenpath_test.go](internal/domain/goldenpath_test.go)) |
| Application | **Implemented** — 5 use-case-сервисов (Project/TaskPlanning/Work/Result/Completion) + проекция чтения (ключ — (ProjectID, ID), BUGFIX-003; несёт title/type/scope/acceptanceCriteria — TASK-076), порты хранения на in-memory фейках ([README](internal/application/README.md)) |
| Platform (контракты) | Contracts ready |
| Infrastructure | **Implemented** — PostgreSQL (пять Store, `pgx/v5`, самописные миграции, составной ключ Task — BUGFIX-003), производственный EventBus с журналом, GitHub Repository Provider, Memory Provider (файлы + Qdrant, ADR-018) ([README](internal/infrastructure/README.md)) |
| API | **Implemented** (EPIC-008/009, `apps/api` — REST над `internal/application`, ADR-003, весь golden path + списковые операции, 17 операций; без auth — ADR-012 Вариант 1) ([README](apps/api/README.md)) |
| Dashboard | **Implemented** (EPIC-009, v0.8 — `apps/dashboard` Next.js, read-only: список проектов, задачи проекта, детали задачи) ([README](apps/dashboard/README.md)) |
| Developer Engine | **Implemented** — первый реальный адаптер Executor ([agents/claude-code](agents/claude-code/README.md)): Docker-контейнер на Execution, сетевой allowlist, короткоживущие секреты (ADR-005/006); реальный AI-вызов не проверен — нет ключа в этой сессии (честный предел, TASK-056) |
| Workflow | **Machine реализована** — каноническая state machine (20 переходов, 100% покрытия); Definition/Step — контракты до появления потребителя (v0.4) |
| Memory | **Implemented** — `platform.MemoryProvider` целиком (Record/Search/Reindex), файлы — источник истины, Qdrant — производный индекс; эмбеддинг наивный (feature hashing, ADR-018), не семантический по сути — честно задокументированное ограничение MVP |

### Контрольные точки

| Поле | Значение |
| --- | --- |
| **Last ADR** | [ADR-007](docs/adr/ADR-007-pm-qa-executors.md)/[ADR-010](docs/adr/ADR-010-documentation-language.md)/[ADR-013](docs/adr/ADR-013-managed-projects.md) приняты при открытии v1.0 (2026-07-23) — все 18 ADR проекта теперь Accepted |
| **Last Review** | 2026-07-23 — EPIC-009 (TASK-072…078, PR #90–#95) закрыт целиком; предыдущее ревью — EPIC-008 (TASK-064…071 + BUGFIX-003, PR #80–#87) |
| **Quality** | All checks passed; CI: GitHub Actions `verify` — green, required status check (теперь включает `apps/dashboard`: pnpm lint/format/test/build, TASK-077); `main` защищена; toolchain честно закреплён — Go 1.24 без маскировки ([BUGFIX-001](tasks/done/BUGFIX-001-pin-gofumpt.md), [BUGFIX-002](tasks/done/BUGFIX-002-pin-golangci-lint-and-toolchain.md)), pnpm — через `packageManager` в `package.json` (TASK-077); локальная среда воспроизводима и практически проверена — git-хуки (реальные негативные тесты) и Dev Container (реальная сборка, `0 issues.`) |
| **Открытые решения** | 0 ADR в статусе Decision Required — [индекс](docs/adr/DECISIONS_INDEX.md) (все 18 приняты, 2026-07-23) |
| **Прогресс** | [PROJECT_HEALTH.md](PROJECT_HEALTH.md) |

### Правило обновления

Манифест — часть Definition of Done для изменений, затрагивающих: версию, эпик, состояние слоя, последний ADR/ревью. Устаревший манифест — блокирующее замечание ревью.

## Статус

Актуален

## Последнее обновление

2026-07-23
