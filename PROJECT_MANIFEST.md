# PROJECT_MANIFEST — паспорт проекта

## Назначение

Единый документ, отвечающий на вопрос «что сейчас представляет собой проект»: версия, статус, текущий эпик, состояние слоёв и качества. Обновляется в каждом изменении, меняющем состояние проекта (в том же PR).

## Содержание

### Паспорт

| Поле | Значение |
| --- | --- |
| **Project** | AI Studio OS |
| **Version** | v0.3 Domain Layer — **завершён** (2026-07-21); следующий — v0.4 Application Layer ([ROADMAP.md](ROADMAP.md)) |
| **Status** | **Architecture Frozen** (2026-07-19) |
| **Current Epic** | [EPIC-003 Domain Layer](docs/roadmap/EPIC-003-domain-layer.md) — **закрыт** (2026-07-21): этап 1 — пять спецификаций утверждены; этап 2 — пять сущностей + `workflow.Machine` реализованы (TASK-034…039), сквозной сценарий слоя зелёный. EPIC-004 (Application Layer) — не открыт |
| **Current Sprint** | — (спринты не введены; итерации ведутся эпиками из 5–15 задач) |
| **Current Branch** | main |
| **Repository** | [github.com/nikibrdev/ai-studio-os](https://github.com/nikibrdev/ai-studio-os) (public) |
| **License** | Apache License 2.0 ([ADR-001](docs/adr/ADR-001-license.md)) |

### Состояние слоёв и компонентов

| Компонент | Состояние |
| --- | --- |
| Architecture | **Frozen** (ADR-002, 003, 004, 009, 014, 015 приняты) |
| Domain | **Implemented** — 5 сущностей (artifact/execution/executor/task/project) + `workflow.Machine`, инварианты покрыты тестами (81.8–100%), сквозной сценарий слоя зелёный ([goldenpath_test.go](internal/domain/goldenpath_test.go)) |
| Application | Not Started (v0.4, следующий эпик) |
| Platform (контракты) | Contracts ready |
| Infrastructure | Not Started |
| API | Not Started (REST, после Application) |
| Dashboard | Not Started (v0.6) |
| Developer Engine | Planning (ADR-005 принят — Executor Contract; блокер — ADR-006) |
| Workflow | **Machine реализована** — каноническая state machine (20 переходов, 100% покрытия); Definition/Step — контракты до появления потребителя (v0.4) |

### Контрольные точки

| Поле | Значение |
| --- | --- |
| **Last ADR** | [ADR-011](docs/adr/ADR-011-task-identifiers.md) (идентификаторы `TASK-NNN`/`EPIC-NNN` — последовательные в рамках Project; суррогатный ключ в БД) |
| **Last Review** | 2026-07-21 — Code Review этапа 2 EPIC-003 (TASK-034…039, PR #42–#47); эпик закрыт целиком |
| **Quality** | All checks passed; CI: GitHub Actions `verify` — green, required status check; `main` защищена; toolchain честно закреплён на Go 1.24 без маскировки ([BUGFIX-001](tasks/done/BUGFIX-001-pin-gofumpt.md), [BUGFIX-002](tasks/done/BUGFIX-002-pin-golangci-lint-and-toolchain.md)); локальная среда воспроизводима и практически проверена — git-хуки (реальные негативные тесты) и Dev Container (реальная сборка, `0 issues.`) |
| **Открытые решения** | 6 ADR в статусе Decision Required — [индекс](docs/adr/DECISIONS_INDEX.md) |
| **Прогресс** | [PROJECT_HEALTH.md](PROJECT_HEALTH.md) |

### Правило обновления

Манифест — часть Definition of Done для изменений, затрагивающих: версию, эпик, состояние слоя, последний ADR/ревью. Устаревший манифест — блокирующее замечание ревью.

## Статус

Актуален

## Последнее обновление

2026-07-21
