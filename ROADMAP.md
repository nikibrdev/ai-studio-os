# Roadmap AI Studio OS

## Назначение

Определяет релизные вехи проекта от фундамента до первого публичного MVP. Каждая веха отражает готовность продукта, а не только объём выполненных задач; содержание уточняется эпиками (`docs/roadmap/EPIC-*.md`) и задачами (`tasks/`). Вехи служат горизонту 1 года из [VISION.md](VISION.md); горизонты 2 и 5 лет там же выходят за рамки этого документа.

## Содержание

### История ревизии

Версия 2026-07-19 (первоначальная) была написана до проектирования архитектуры и описывала стек ролей (Project Management, Developer Engine, QA Engine, Workflow) как последовательные версии. После Architecture Freeze (ADR-002…015, [overview.md](docs/architecture/overview.md)) и завершения EPIC-002.5 стала известна реальная слоистая структура (`domain → application → platform → infrastructure`, [ADR-015](docs/adr/ADR-015-internal-layering.md)). Roadmap пересмотрен 2026-07-20 архитектором проекта, чтобы версии отражали готовность продукта по слоям, а не предположения этапа Foundation. Решение зафиксировано: [engineering/decisions/2026-07-20-release-milestones.md](engineering/decisions/2026-07-20-release-milestones.md).

Прежние роли-версии не исчезли — они распределены по новым вехам: Project Management → часть Domain/Application Layer; Developer Engine и Multi Agent → AI Agent Runtime; Workflow → Domain Layer (`internal/domain/workflow`) и Application Layer (оркестрация).

### Внутренние архитектурные вехи (не версии продукта)

Отдельно от пользовательских версий v0.X — рубежи, отмечающие качественные изменения в самом процессе разработки, а не в готовности продукта.

- **First Reference Domain Specification** (2026-07-20) — [Artifact](docs/specifications/domain/artifact.md) стал первой доменной спецификацией проекта со статусом **Reference**: не просто утверждённой, а официально признанной образцом для всех последующих спецификаций Domain Layer ([решение](engineering/decisions/2026-07-20-domain-specification-reference-status.md)). Момент, когда у проекта впервые появился не только код и документация, но и проверенный, воспроизводимый метод доменного моделирования, на который опираются все следующие сущности.
- **Domain Specifications First — этап 1 закрыт** (2026-07-21) — все пять спецификаций Domain Layer ([Artifact](docs/specifications/domain/artifact.md), [Execution](docs/specifications/domain/execution.md), [Executor](docs/specifications/domain/executor.md), [Task](docs/specifications/domain/task.md), [Project](docs/specifications/domain/project.md)) утверждены архитектором. Каждая закрыта в тот же день, когда взята в работу — метод, откалиброванный на Artifact, воспроизвёлся четыре раза подряд без потери в качестве ревью (реальные раунды Changes Requested/Approve на каждой). Два реальных расхождения со старым кодом EPIC-002 (`internal/domain/task`, `internal/domain/project`) не устранялись правкой кода, а получили явное архитектурное решение — контракты целенаправленно расширяются вслед за спецификацией на этапе 2.

### v0.1 Foundation — фундамент — **Завершено**

Структура репозитория, базовая документация, система задач, роли агентов на уровне обязанностей.

### v0.2 Architecture & Engineering Platform — **Завершено** (2026-07-20)

- Architecture Freeze: ADR-002, 003, 004, 009, 014, 015 приняты ([DECISIONS_INDEX](docs/adr/DECISIONS_INDEX.md)).
- Контракты ядра без реализации: `internal/{domain,platform}` (EPIC-002).
- Инженерная платформа: GitHub-репозиторий, CI (`verify`), защита `main`, CODEOWNERS, шаблоны, Dependabot, релизный процесс (EPIC-002.5).
- Процесс «план → утверждение → код → ревью → merge»; правило необратимости решений; двухуровневая документация модулей (README + Specification).

**Результат:** архитектура заморожена, инженерная платформа исключает попадание непроверенного изменения в `main`.

### v0.3 Domain Layer — предметная область — **Завершено** (2026-07-21)

- [EPIC-003](docs/roadmap/EPIC-003-domain-layer.md): реализация доменных модулей, начиная не с `task`, а с порядка **Artifact → Execution → Executor → Task → Project** ([domain-model.md](docs/architecture/domain-model.md), [ADR-016](docs/adr/ADR-016-artifact-aggregate-root.md)).
- Этап 1 — Domain Specifications First: полные спецификации всех пяти модулей утверждены архитектором до единой строки Go (закрыт 2026-07-21).
- Этап 2 — реализация: пять сущностей + каноническая state machine (`workflow.Machine`), TASK-034…039 (закрыт 2026-07-21).

**Результат достигнут:** доменная логика (в т.ч. state machine задачи) работает и покрыта тестами, без внешних зависимостей — подтверждено сквозным сценарием слоя ([internal/domain/goldenpath_test.go](internal/domain/goldenpath_test.go)): Task проходит все девять канонических состояний, порождая Execution и опубликованный Artifact.

### v0.4 Application Layer — сценарии использования — **Завершено** (2026-07-21)

- [EPIC-004](docs/roadmap/EPIC-004-application-layer.md): сценарии команд поверх доменных модулей ([ADR-014](docs/adr/ADR-014-module-interaction.md)) — постановка задачи, запуск работы, производство результата, завершение задачи (merge после Testing, ADR-008); проекция для чтения из событий.

**Результат достигнут:** платформа исполняет use-case'ы, не завязанные на конкретную инфраструктуру — подтверждено сквозным тестом ([internal/application/e2e_test.go](internal/application/e2e_test.go)): вся золотая дорожка на in-memory адаптерах, включая ветки «changes requested» и «tests failed», состояние читается только через проекцию.

### v0.5 Infrastructure Layer — инфраструктура — **Завершено** (2026-07-21)

- [EPIC-005](docs/roadmap/EPIC-005-infrastructure-layer.md): адаптеры портов EPIC-004 к реальным технологиям — PostgreSQL (источник истины задач, [ADR-004](docs/adr/ADR-004-task-storage.md), драйвер `pgx/v5` — [ADR-017](docs/adr/ADR-017-postgresql-driver.md)), производственный In-Memory Event Bus с журналом в PostgreSQL ([ADR-002](docs/adr/ADR-002-event-delivery.md)), GitHub (Repository Provider).

**Результат достигнут:** платформа работает end-to-end на реальных хранилищах и интеграциях — подтверждено интеграционным тестом ([internal/infrastructure/wiring/golden_path_integration_test.go](internal/infrastructure/wiring/golden_path_integration_test.go)): та же золотая дорожка, что и в v0.4, на реальных PostgreSQL-адаптерах и производственном EventBus (с журналом, восстановимым отдельным select'ом), без единой строки изменений в `internal/application`/`internal/domain`. Единственное исключение — `RepositoryProvider`: покрыт unit-тестами против GitHub REST API, но не проверен вживую на реальном репозитории (нет тестового репозитория и токена в этой сессии) — принятый открытый риск, не блокирующий результат.

### v0.6 AI Agent Runtime — исполнение агентов — **В работе** (открыт 2026-07-21)

- [EPIC-006](docs/roadmap/EPIC-006-ai-agent-runtime.md): целевая модель исполнения из [ADR-006](docs/adr/ADR-006-agent-execution-environment.md) (принят) — одно Execution = один Docker-контейнер, сетевой allowlist, короткоживущие секреты; первый реальный адаптер `Executor` ([ADR-005](docs/adr/ADR-005-executor-contract.md), принят) — Claude Code.
- Роль Developer исполняется агентом по процессу из [CLAUDE.md](CLAUDE.md); оформление изменений через GitHub.

**Результат:** агент-разработчик выполняет подготовленные задачи в исполняемой платформе.

### v0.7 Memory System — память

- Память агентов и знания проекта (`memory/`); векторный поиск на Qdrant.

**Результат:** агенты используют накопленный контекст проекта.

### v0.8 Dashboard — веб-интерфейс

- Next.js: проекты, задачи, статусы, отчёты агентов; наблюдаемость процесса для человека.

**Результат:** состояние платформы видно и управляемо через веб-интерфейс.

### v0.9 API — публичный интерфейс

- REST API ([ADR-003](docs/adr/ADR-003-api-protocol.md)) для внешних потребителей сверх Dashboard.

**Результат:** платформа доступна внешним клиентам, не только собственному UI.

### v1.0 First Public MVP — первый публичный релиз

- Стабилизация контрактов; полная документация для пользователей и контрибьюторов.
- Публичный релиз ([релизный процесс](docs/development/git-workflow.md)).

**Результат:** платформа готова к использованию внешними командами.

## Статус

Актуален

## Последнее обновление

2026-07-21
