# PROJECT_MANIFEST — паспорт проекта

## Назначение

Единый документ, отвечающий на вопрос «что сейчас представляет собой проект»: версия, статус, текущий эпик, состояние слоёв и качества. Обновляется в каждом изменении, меняющем состояние проекта (в том же PR).

## Содержание

### Паспорт

| Поле | Значение |
| --- | --- |
| **Project** | AI Studio OS |
| **Version** | v0.2 Architecture & Engineering Platform — Завершено; следующая — v0.3 Domain Layer ([ROADMAP.md](ROADMAP.md)) |
| **Status** | **Architecture Frozen** (2026-07-19) |
| **Current Epic** | [EPIC-002.6 Developer Experience](docs/roadmap/EPIC-002.6-developer-experience.md) — Утверждён, выполняется |
| **Current Sprint** | — (спринты не введены; итерации ведутся эпиками из 5–15 задач) |
| **Current Branch** | main |
| **Repository** | [github.com/nikibrdev/ai-studio-os](https://github.com/nikibrdev/ai-studio-os) (public) |
| **License** | Apache License 2.0 ([ADR-001](docs/adr/ADR-001-license.md)) |

### Состояние слоёв и компонентов

| Компонент | Состояние |
| --- | --- |
| Architecture | **Frozen** (ADR-002, 003, 004, 009, 014, 015 приняты) |
| Domain | Contracts ready; логика Not Started (EPIC-003) |
| Application | Not Started |
| Platform (контракты) | Contracts ready |
| Infrastructure | Not Started |
| API | Not Started (REST, после Application) |
| Dashboard | Not Started (v0.6) |
| Developer Engine | Planning (v0.3; блокеры — ADR-005, ADR-006) |
| Workflow | State machine спроектирована; контракты готовы; реализация Not Started |

### Контрольные точки

| Поле | Значение |
| --- | --- |
| **Last ADR** | [ADR-015](docs/adr/ADR-015-internal-layering.md) (internal layering) |
| **Last Review** | 2026-07-19 — [EPIC-002 code review](engineering/reviews/2026-07-19-epic-002-code-review.md) |
| **Quality** | All checks passed; CI: GitHub Actions `verify` — green, required status check; `main` защищена (прямой push отклонён, force-push/удаление запрещены); toolchain честно закреплён на Go 1.24 без маскировки ([BUGFIX-001](tasks/done/BUGFIX-001-pin-gofumpt.md), [BUGFIX-002](tasks/done/BUGFIX-002-pin-golangci-lint-and-toolchain.md)) |
| **Открытые решения** | 8 ADR в статусе Decision Required — [индекс](docs/adr/DECISIONS_INDEX.md) |
| **Прогресс** | [PROJECT_HEALTH.md](PROJECT_HEALTH.md) |

### Правило обновления

Манифест — часть Definition of Done для изменений, затрагивающих: версию, эпик, состояние слоя, последний ADR/ревью. Устаревший манифест — блокирующее замечание ревью.

## Статус

Актуален

## Последнее обновление

2026-07-19
