# PROJECT_MANIFEST — паспорт проекта

## Назначение

Единый документ, отвечающий на вопрос «что сейчас представляет собой проект»: версия, статус, текущий эпик, состояние слоёв и качества. Обновляется в каждом изменении, меняющем состояние проекта (в том же PR).

## Содержание

### Паспорт

| Поле | Значение |
| --- | --- |
| **Project** | AI Studio OS |
| **Version** | v0.7 Memory System — **завершён** (2026-07-22) ([ROADMAP.md](ROADMAP.md)) |
| **Status** | **Architecture Frozen** (2026-07-19) |
| **Current Epic** | Нет открытого эпика — [EPIC-007 Memory System](docs/roadmap/EPIC-007-memory-system.md) закрыт (2026-07-22): `platform.MemoryProvider` реализован целиком — файловое хранилище + поиск на Qdrant (наивный локальный эмбеддинг, ADR-018), TASK-058…063 |
| **Current Sprint** | — (спринты не введены; итерации ведутся эпиками из 5–15 задач) |
| **Current Branch** | main |
| **Repository** | [github.com/nikibrdev/ai-studio-os](https://github.com/nikibrdev/ai-studio-os) (public) |
| **License** | Apache License 2.0 ([ADR-001](docs/adr/ADR-001-license.md)) |

### Состояние слоёв и компонентов

| Компонент | Состояние |
| --- | --- |
| Architecture | **Frozen** (ADR-002, 003, 004, 009, 014, 015 приняты) |
| Domain | **Implemented** — 5 сущностей (artifact/execution/executor/task/project) + `workflow.Machine`, инварианты покрыты тестами (81.8–100%), сквозной сценарий слоя зелёный ([goldenpath_test.go](internal/domain/goldenpath_test.go)) |
| Application | **Implemented** — 4 use-case-сервиса (TaskPlanning/Work/Result/Completion) + проекция чтения, порты хранения на in-memory фейках, покрытие 83.1% ([README](internal/application/README.md)) |
| Platform (контракты) | Contracts ready |
| Infrastructure | **Implemented** — PostgreSQL (пять Store, `pgx/v5`, самописные миграции), производственный EventBus с журналом, GitHub Repository Provider, Memory Provider (файлы + Qdrant, ADR-018) ([README](internal/infrastructure/README.md)) |
| API | Not Started (REST, после Infrastructure) |
| Dashboard | Not Started (v0.6) |
| Developer Engine | **Implemented** — первый реальный адаптер Executor ([agents/claude-code](agents/claude-code/README.md)): Docker-контейнер на Execution, сетевой allowlist, короткоживущие секреты (ADR-005/006); реальный AI-вызов не проверен — нет ключа в этой сессии (честный предел, TASK-056) |
| Workflow | **Machine реализована** — каноническая state machine (20 переходов, 100% покрытия); Definition/Step — контракты до появления потребителя (v0.4) |
| Memory | **Implemented** — `platform.MemoryProvider` целиком (Record/Search/Reindex), файлы — источник истины, Qdrant — производный индекс; эмбеддинг наивный (feature hashing, ADR-018), не семантический по сути — честно задокументированное ограничение MVP |

### Контрольные точки

| Поле | Значение |
| --- | --- |
| **Last ADR** | [ADR-018](docs/adr/ADR-018-memory-embeddings-and-qdrant-schema.md) (эмбеддинги Memory — наивный локальный feature hashing; схема Qdrant — коллекция `memory_entries`) |
| **Last Review** | 2026-07-22 — EPIC-007 (TASK-058…063, PR #75–#78) закрыт целиком; предыдущее ревью — EPIC-006 (TASK-052…057, PR #67–#72) |
| **Quality** | All checks passed; CI: GitHub Actions `verify` — green, required status check; `main` защищена; toolchain честно закреплён на Go 1.24 без маскировки ([BUGFIX-001](tasks/done/BUGFIX-001-pin-gofumpt.md), [BUGFIX-002](tasks/done/BUGFIX-002-pin-golangci-lint-and-toolchain.md)); локальная среда воспроизводима и практически проверена — git-хуки (реальные негативные тесты) и Dev Container (реальная сборка, `0 issues.`) |
| **Открытые решения** | 4 ADR в статусе Decision Required — [индекс](docs/adr/DECISIONS_INDEX.md) |
| **Прогресс** | [PROJECT_HEALTH.md](PROJECT_HEALTH.md) |

### Правило обновления

Манифест — часть Definition of Done для изменений, затрагивающих: версию, эпик, состояние слоя, последний ADR/ревью. Устаревший манифест — блокирующее замечание ревью.

## Статус

Актуален

## Последнее обновление

2026-07-22
