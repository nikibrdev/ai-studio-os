# ADR-016: Artifact — самостоятельный Aggregate Root

## Статус

**Принято** (решение архитектора проекта, 2026-07-20)

## Дата

2026-07-20

## Контекст

[ADR-005](ADR-005-executor-contract.md) ввело `Artifact` как абстрактный тип, возвращаемый Executor'ом (`Accept`/`Artifacts`/`Status`/`Finish`), но намеренно оставило открытым вопрос о его размещении в доменной модели: владеет ли им модуль `execution` (как предполагалось изначально в `domain-model.md`) или Artifact — самостоятельная сущность.

Последнее обсуждение перед стартом Domain Layer (EPIC-003) показало, что вопрос владения — не техническая деталь, а решение, определяющее ценность всей платформы: AI Studio OS производит артефакты, а не хранит задачи ([VISION.md](../../VISION.md)). Artifact может существовать до задачи (импортированный документ), во время исполнения, после завершения задачи и независимо от конкретного исполнителя — жизненный цикл Artifact не совпадает с жизненным циклом Execution, Task или Project, которые могли бы его «владеть».

## Рассмотренные варианты

### Вариант 1: Artifact принадлежит Execution

Artifact существует только в рамках произведшего его Execution (текущее до этого ADR предположение в `domain-model.md`, унаследованное от ранней модели с сущностью Result).

- **Плюсы:** просто; не требует отдельного модуля.
- **Минусы:** не может выразить Artifact, существующий до задачи или переживающий Execution/Task/Project; смешивает «что было произведено» с «как это было произведено».

### Вариант 2: Artifact — самостоятельный Aggregate Root

Artifact — отдельная сущность верхнего уровня, на которую Execution лишь ссылается («я произвёл вот этот Artifact»), не владея им.

- **Плюсы:** соответствует реальному жизненному циклу Artifact (может существовать независимо); делает Artifact первичной ценностью системы явно, а не побочным продуктом Execution; согласуется с [ADR-005](ADR-005-executor-contract.md) (`Artifacts()` уже возвращает срез, а не вложенный в Result объект).
- **Минусы:** на один доменный модуль больше (11 вместо 10); требует дисциплины — Execution не должно превращаться в скрытого владельца через код.

## Решение

Принят вариант 2.

**Artifact — самостоятельный Aggregate Root**, не часть Execution, не часть Task, не часть Project. Модель:

```
Project
    ├── Task
    └── Artifact

Task → создаёт → Execution
Execution → использует → Executor
Execution → производит → Artifact
```

- Task инициирует работу.
- Execution описывает процесс выполнения.
- Executor выполняет работу.
- Artifact остаётся после выполнения.

**Определение.** Artifact — любое долговременное инженерное произведение, созданное или изменённое системой исполнения. Примеры: commit, Pull Request, source file, Markdown, ADR, спецификация, test report, build report, diagram, screenshot, Figma-файл, release note. В перспективе: video, audio, dataset, prompt, knowledge entry.

**Границы.** НЕ является Artifact: временный лог, прогресс выполнения (progress), heartbeat, токен LLM, внутреннее сообщение агента — это принадлежит Execution (процессу), но не является результатом, который система обязана сохранить.

**Внутреннее устройство.** Artifact состоит из двух частей:

- **Metadata** — то, что знает платформа: ID, Type, Author, CreatedAt, ProducedByExecution, Version.
- **Payload** — сами данные, содержимое которых платформа не интерпретирует и тип которых зависит от Type (Markdown, Git Commit, PDF, JSON, Binary и т.д.).

**Правило владения.** Execution никогда не владеет Artifact. Execution знает только факт: «я произвёл вот этот Artifact» (ссылка по идентификатору), а не хранит его как часть своего состояния.

## Последствия

### Положительные

- Модель выражает реальный жизненный цикл Artifact (независимый от Execution/Task/Project), а не подгоняет его под жизненный цикл процесса, который его произвёл.
- Artifact становится буквально тем, чем его называет видение проекта — первичной сущностью ценности, а не деталью реализации Execution.
- Устраняет источник будущей путаницы: без этого решения было бы легко реализовать Artifact как поле/срез внутри Execution, что противоречило бы VISION.md и самому ADR-005.

### Отрицательные

- Доменных модулей — 11 вместо 10 (`artifact` — новый); больше поверхности для проектирования на этапе EPIC-003.
- Metadata/Payload разделение вводит дополнительную концептуальную сложность (два типа данных вместо одного) — цена, оправданная тем, что платформа не должна знать о содержимом конкретных типов артефактов.

### Влияние на существующие документы и код

`docs/architecture/domain-model.md` (Artifact — Aggregate Root, отдельный владелец данных, связи без владения через Execution), `docs/adr/ADR-005-executor-contract.md` (открытый вопрос закрыт этой ссылкой), `docs/domain/ubiquitous-language.md`, `docs/domain/bounded-contexts.md`, `docs/architecture/core.md`, `docs/architecture/components.md`, `docs/architecture/module-boundaries.md` (список доменных модулей — 11) — обновлены (TASK-028). Реализация модуля `artifact` — EPIC-003, после утверждённой спецификации ([engineering/decisions/2026-07-20-domain-layer-specification-requirement.md](../../engineering/decisions/2026-07-20-domain-layer-specification-requirement.md)).

## Связанные материалы

[VISION.md](../../VISION.md) · [ADR-005](ADR-005-executor-contract.md) · [domain-model.md](../architecture/domain-model.md) · [ubiquitous-language.md](../domain/ubiquitous-language.md)
