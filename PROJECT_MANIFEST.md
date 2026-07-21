# PROJECT_MANIFEST — паспорт проекта

## Назначение

Единый документ, отвечающий на вопрос «что сейчас представляет собой проект»: версия, статус, текущий эпик, состояние слоёв и качества. Обновляется в каждом изменении, меняющем состояние проекта (в том же PR).

## Содержание

### Паспорт

| Поле | Значение |
| --- | --- |
| **Project** | AI Studio OS |
| **Version** | v0.3 Domain Layer — этап 1 (Domain Specifications First) закрыт; этап 2 (Реализация) не начат ([ROADMAP.md](ROADMAP.md)) |
| **Status** | **Architecture Frozen** (2026-07-19) |
| **Current Epic** | [EPIC-003 Domain Layer](docs/roadmap/EPIC-003-domain-layer.md) — этап 1 закрыт: все пять спецификаций утверждены ([Artifact](docs/specifications/domain/artifact.md) — Reference; [Execution](docs/specifications/domain/execution.md), [Executor](docs/specifications/domain/executor.md), [Task](docs/specifications/domain/task.md), [Project](docs/specifications/domain/project.md) — Утверждена); этап 2 (реализация на Go) — следующий |
| **Current Sprint** | — (спринты не введены; итерации ведутся эпиками из 5–15 задач) |
| **Current Branch** | main |
| **Repository** | [github.com/nikibrdev/ai-studio-os](https://github.com/nikibrdev/ai-studio-os) (public) |
| **License** | Apache License 2.0 ([ADR-001](docs/adr/ADR-001-license.md)) |

### Состояние слоёв и компонентов

| Компонент | Состояние |
| --- | --- |
| Architecture | **Frozen** (ADR-002, 003, 004, 009, 014, 015 приняты) |
| Domain | Contracts ready; все 5 спецификаций утверждены ([Artifact](docs/specifications/domain/artifact.md) — Reference; Execution/Executor/Task/Project — Утверждена); логика — Not Started (этап 2) |
| Application | Not Started |
| Platform (контракты) | Contracts ready |
| Infrastructure | Not Started |
| API | Not Started (REST, после Application) |
| Dashboard | Not Started (v0.6) |
| Developer Engine | Planning (v0.3; ADR-005 принят — Executor Contract; блокер — ADR-006) |
| Workflow | State machine спроектирована; контракты готовы; реализация Not Started |

### Контрольные точки

| Поле | Значение |
| --- | --- |
| **Last ADR** | [ADR-016](docs/adr/ADR-016-artifact-aggregate-root.md) (Artifact — самостоятельный Aggregate Root, не часть Execution/Task/Project) |
| **Last Review** | 2026-07-21 — Final Architecture Review, спецификации Execution/Executor/Task/Project (TASK-030…033); этап 1 EPIC-003 закрыт |
| **Quality** | All checks passed; CI: GitHub Actions `verify` — green, required status check; `main` защищена; toolchain честно закреплён на Go 1.24 без маскировки ([BUGFIX-001](tasks/done/BUGFIX-001-pin-gofumpt.md), [BUGFIX-002](tasks/done/BUGFIX-002-pin-golangci-lint-and-toolchain.md)); локальная среда воспроизводима и практически проверена — git-хуки (реальные негативные тесты) и Dev Container (реальная сборка, `0 issues.`) |
| **Открытые решения** | 7 ADR в статусе Decision Required — [индекс](docs/adr/DECISIONS_INDEX.md) |
| **Прогресс** | [PROJECT_HEALTH.md](PROJECT_HEALTH.md) |

### Правило обновления

Манифест — часть Definition of Done для изменений, затрагивающих: версию, эпик, состояние слоя, последний ADR/ревью. Устаревший манифест — блокирующее замечание ревью.

## Статус

Актуален

## Последнее обновление

2026-07-21
