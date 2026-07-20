# Roadmap AI Studio OS

## Назначение

Определяет релизные вехи проекта от фундамента до первого публичного MVP. Каждая веха отражает готовность продукта, а не только объём выполненных задач; содержание уточняется эпиками (`docs/roadmap/EPIC-*.md`) и задачами (`tasks/`). Вехи служат горизонту 1 года из [VISION.md](VISION.md); горизонты 2 и 5 лет там же выходят за рамки этого документа.

## Содержание

### История ревизии

Версия 2026-07-19 (первоначальная) была написана до проектирования архитектуры и описывала стек ролей (Project Management, Developer Engine, QA Engine, Workflow) как последовательные версии. После Architecture Freeze (ADR-002…015, [overview.md](docs/architecture/overview.md)) и завершения EPIC-002.5 стала известна реальная слоистая структура (`domain → application → platform → infrastructure`, [ADR-015](docs/adr/ADR-015-internal-layering.md)). Roadmap пересмотрен 2026-07-20 архитектором проекта, чтобы версии отражали готовность продукта по слоям, а не предположения этапа Foundation. Решение зафиксировано: [engineering/decisions/2026-07-20-release-milestones.md](engineering/decisions/2026-07-20-release-milestones.md).

Прежние роли-версии не исчезли — они распределены по новым вехам: Project Management → часть Domain/Application Layer; Developer Engine и Multi Agent → AI Agent Runtime; Workflow → Domain Layer (`internal/domain/workflow`) и Application Layer (оркестрация).

### v0.1 Foundation — фундамент — **Завершено**

Структура репозитория, базовая документация, система задач, роли агентов на уровне обязанностей.

### v0.2 Architecture & Engineering Platform — **Завершено** (2026-07-20)

- Architecture Freeze: ADR-002, 003, 004, 009, 014, 015 приняты ([DECISIONS_INDEX](docs/adr/DECISIONS_INDEX.md)).
- Контракты ядра без реализации: `internal/{domain,platform}` (EPIC-002).
- Инженерная платформа: GitHub-репозиторий, CI (`verify`), защита `main`, CODEOWNERS, шаблоны, Dependabot, релизный процесс (EPIC-002.5).
- Процесс «план → утверждение → код → ревью → merge»; правило необратимости решений; двухуровневая документация модулей (README + Specification).

**Результат:** архитектура заморожена, инженерная платформа исключает попадание непроверенного изменения в `main`.

### v0.3 Domain Layer — предметная область

- [EPIC-003](docs/roadmap/EPIC-003-domain-layer.md): реализация доменных модулей, начиная не с `task`, а с порядка **Artifact → Execution → Executor → Task → Project** ([domain-model.md](docs/architecture/domain-model.md), [ADR-016](docs/adr/ADR-016-artifact-aggregate-root.md)).
- Этап 1 — Domain Specifications First: полные спецификации всех пяти модулей утверждаются архитектором до единой строки Go ([Specification.md](.claude/templates/Specification.md); требование Domain Layer — [engineering/decisions/2026-07-20-domain-layer-specification-requirement.md](engineering/decisions/2026-07-20-domain-layer-specification-requirement.md)).
- Этап 2 — реализация пяти модулей в том же порядке, только после утверждения этапа 1.

**Результат:** доменная логика (в т.ч. state machine задачи) работает и покрыта тестами, без внешних зависимостей.

### v0.4 Application Layer — сценарии использования

- Сценарии команд поверх доменных модулей; проекции для чтения из событий ([ADR-014](docs/adr/ADR-014-module-interaction.md)).

**Результат:** платформа исполняет use-case'ы, не завязанные на конкретную инфраструктуру.

### v0.5 Infrastructure Layer — инфраструктура

- Адаптеры: PostgreSQL (источник истины задач, [ADR-004](docs/adr/ADR-004-task-storage.md)), In-Memory Event Bus ([ADR-002](docs/adr/ADR-002-event-delivery.md)), GitHub (Repository Provider).

**Результат:** платформа работает end-to-end на реальных хранилищах и интеграциях.

### v0.6 AI Agent Runtime — исполнение агентов

- Контракт адаптера агента (ADR-005, ADR-006 — Decision Required); первый адаптер — Claude Code.
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

2026-07-20
