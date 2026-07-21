# PROJECT_MANIFEST — паспорт проекта

## Назначение

Единый документ, отвечающий на вопрос «что сейчас представляет собой проект»: версия, статус, текущий эпик, состояние слоёв и качества. Обновляется в каждом изменении, меняющем состояние проекта (в том же PR).

## Содержание

### Паспорт

| Поле | Значение |
| --- | --- |
| **Project** | AI Studio OS |
| **Version** | v0.5 Infrastructure Layer — **в работе** (открыт 2026-07-21); v0.4 Application Layer — завершён ([ROADMAP.md](ROADMAP.md)) |
| **Status** | **Architecture Frozen** (2026-07-19) |
| **Current Epic** | [EPIC-005 Infrastructure Layer](docs/roadmap/EPIC-005-infrastructure-layer.md) — **в работе** (открыт 2026-07-21): реализация портов EPIC-004 на PostgreSQL (`pgx/v5`, [ADR-017](docs/adr/ADR-017-postgresql-driver.md)), производственный EventBus с журналом, GitHub Repository Provider (TASK-046…051) |
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
| Infrastructure | **В работе** (v0.5, EPIC-005: каркас БД, Postgres-адаптеры, EventBus, GitHub Provider) |
| API | Not Started (REST, после Infrastructure) |
| Dashboard | Not Started (v0.6) |
| Developer Engine | Planning (ADR-005 и ADR-006 приняты — контракт и среда исполнения определены; реализация — v0.6) |
| Workflow | **Machine реализована** — каноническая state machine (20 переходов, 100% покрытия); Definition/Step — контракты до появления потребителя (v0.4) |

### Контрольные точки

| Поле | Значение |
| --- | --- |
| **Last ADR** | [ADR-017](docs/adr/ADR-017-postgresql-driver.md) (драйвер PostgreSQL — `pgx/v5`; миграции — самописный раннер) |
| **Last Review** | 2026-07-21 — Code Review EPIC-004 (TASK-040…045, PR #53–#57); эпик закрыт целиком |
| **Quality** | All checks passed; CI: GitHub Actions `verify` — green, required status check; `main` защищена; toolchain честно закреплён на Go 1.24 без маскировки ([BUGFIX-001](tasks/done/BUGFIX-001-pin-gofumpt.md), [BUGFIX-002](tasks/done/BUGFIX-002-pin-golangci-lint-and-toolchain.md)); локальная среда воспроизводима и практически проверена — git-хуки (реальные негативные тесты) и Dev Container (реальная сборка, `0 issues.`) |
| **Открытые решения** | 4 ADR в статусе Decision Required — [индекс](docs/adr/DECISIONS_INDEX.md) |
| **Прогресс** | [PROJECT_HEALTH.md](PROJECT_HEALTH.md) |

### Правило обновления

Манифест — часть Definition of Done для изменений, затрагивающих: версию, эпик, состояние слоя, последний ADR/ревью. Устаревший манифест — блокирующее замечание ревью.

## Статус

Актуален

## Последнее обновление

2026-07-21
