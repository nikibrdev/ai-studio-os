# Решение: Execution — не Bounded Context, Artifact — первичная сущность, порядок Domain Layer

## Назначение

Фиксирует четыре решения архитектора проекта от 2026-07-20 (тот же разговор, что и [ADR-005](../../docs/adr/ADR-005-executor-contract.md)/TASK-026), довершающие доменные основания перед стартом Domain Layer (EPIC-003).

## Содержание

### Контекст

После принятия ADR-005 (Executor Contract) и переименования `platform.Agent` → `platform.Executor` (TASK-026) архитектор развил решение дальше: уточнил границы Bounded Contexts, статус сущности Artifact и порядок проектирования Domain Layer — чтобы будущая реализация (EPIC-003) не начинала с Task «по умолчанию», а отражала настоящую цель системы.

### Решение

1. **Execution — не Bounded Context.** Исполнение — сквозная техническая возможность, координируемая Application Layer, а не отдельная предметная область: все контексты нуждаются в исполнении, поэтому выделение его в контекст означало бы владение тем, чем пользуются все. Оставшиеся Bounded Contexts: **Planning → Development → Review → Knowledge** (переименован из Memory — только название контекста, сущность/модуль `memory` не переименованы). Задокументировано в [bounded-contexts.md](../../docs/domain/bounded-contexts.md).
2. **Agent и Executor — разные понятия** (закреплено ранее ADR-005, доведено до домен-модели TASK-027): домен-модуль `agent` переименован в `executor` во всех концептуальных списках модулей ([core.md](../../docs/architecture/core.md), [components.md](../../docs/architecture/components.md), [domain-model.md](../../docs/architecture/domain-model.md)); понятие «Agent» (роль-исполнитель, «Developer Agent») отдельным модулем не становится — выражается связкой Role + Executor.
3. **Artifact — первичная сущность результата работы** — не Result, не Output, не Response. Сущность `Result` убрана из доменной модели: Execution напрямую производит Artifact и несёт `ExecutionStatus`, без промежуточной обёртки — согласовано с уже принятым контрактом ADR-005 (`Accept`/`Artifacts`/`Status`/`Finish`). Задокументировано в [domain-model.md](../../docs/architecture/domain-model.md).
4. **Порядок проектирования Domain Layer (EPIC-003): Artifact → Execution → Executor → Task → Project**, не Task в первую очередь. Обоснование: конечная цель системы — не хранение задач, а производство артефактов; Task — способ организовать работу, Artifact — ценность, которую эта работа производит. Задокументировано в [domain-model.md](../../docs/architecture/domain-model.md) и усилено дословной цитатой в [VISION.md](../../VISION.md):

   > AI Studio OS — это операционная система исполнения инженерной работы. LLM, человек, Claude Code, Codex, OpenHands — это просто разные исполнители. Задачи — это способ организовать работу. А артефакты — это настоящая ценность, которую производит система.

### Согласованные открытые вопросы

При планировании (TASK-027) было явно согласовано три уточнения:

- Task-состояние `Testing` (QA) не привязывается ни к одному из четырёх Bounded Context — показывается как переход через сквозную возможность Execution, чтобы не смешивать роли Reviewer и QA Engineer.
- Переименование Memory → Knowledge касается только названия Bounded Context, не сущности/модуля `memory`.
- Домен-модуль `agent` переименован в `executor`; отдельного модуля/сущности «Agent» не вводится.

### Последствия

- `docs/domain/bounded-contexts.md`, `docs/domain/ubiquitous-language.md`, `docs/architecture/domain-model.md`, `docs/architecture/core.md`, `docs/architecture/components.md`, `VISION.md` — синхронизированы (TASK-027).
- Открытый вопрос о точном месте типа `Artifact` (отдельный модуль `artifact` vs владение `execution`) остаётся нерешённым — зафиксирован в самом [ADR-005](../../docs/adr/ADR-005-executor-contract.md), требует архитектора при проектировании модуля Artifact (EPIC-003).
- EPIC-003 при старте должен явно учитывать порядок Artifact → Execution → Executor → Task → Project при написании спецификаций модулей ([docs/specifications/domain/](../../docs/specifications/domain/)).

## Статус

Актуален

## Последнее обновление

2026-07-20
