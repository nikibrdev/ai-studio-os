# Решение: видение продукта и доменные основания перед Domain Layer

## Назначение

Фиксирует стратегический сдвиг, заданный архитектором проекта 2026-07-20: от построения платформы к построению продукта, и обязательные доменные основания перед началом Domain Layer (v0.3, EPIC-003).

## Содержание

### Контекст

После завершения EPIC-002.5 (Engineering Platform) и EPIC-002.6 (Developer Experience) проект имеет зрелый инженерный фундамент: архитектуру, ADR-процесс, инженерные правила, спецификации, CI, ревью, quality gates, DX, документацию, релизные вехи. Архитектор отметил: это больше похоже на фундамент зрелого open-source проекта, чем на эксперимент с агентами — и предложил сменить фокус со строительства платформы на строительство продукта.

### Решение

1. **[VISION.md](../../VISION.md)** — инженерное (не маркетинговое) видение на горизонтах 1/2/5 лет. Архитектура обслуживает видение, а не наоборот.
2. **Ubiquitous Language до моделей.** Перед проектированием доменных Go-структур — единый язык предметной области: [docs/domain/ubiquitous-language.md](../../docs/domain/ubiquitous-language.md) (10 понятий: Project, Epic, Task, Workflow, Review, Artifact, Execution, Agent/Executor, Context, Memory).
3. **Bounded Contexts.** [docs/domain/bounded-contexts.md](../../docs/domain/bounded-contexts.md) — пять контекстов (Planning, Development, Review, Execution, Memory), карта их связей; один открытый вопрос зафиксирован честно (граница контекста Execution — двоякое прочтение исходной формулировки).
4. **Commands / Events / Queries.** Явное разделение записи, факта и чтения — новый принцип в [engineering-principles.md](../../docs/architecture/engineering-principles.md); формализует уже применённый в EPIC-002 паттерн (`internal/domain/task`).
5. **Golden Path.** [docs/architecture/golden-path.md](../../docs/architecture/golden-path.md) — эталонный сквозной сценарий; критерий приоритизации: каждый новый модуль оценивается по тому, приближает ли он систему к этому сценарию.
6. **Агенты — плагины.** Ядро — операционная система для любых Исполнителей (Executor); AI-агент — один из видов Executor'а, не центральное понятие. Переформулировано в [agents.md](../../docs/architecture/agents.md); переименование `platform.Agent` в код не внесено — зафиксировано как терминологический Decision Required к моменту [ADR-005](../../docs/adr/ADR-005-agent-adapter-contract.md).

### Последствия

- Порядок работы над Domain Layer: язык (`docs/domain/`) → границы контекста → спецификация модуля (`docs/specifications/domain/`) → реализация.
- README.md, ROADMAP.md, docs/architecture/overview.md, docs/development/documentation.md — синхронизированы с новыми документами.
- Открытый вопрос по границе контекста Execution требует подтверждения архитектора при проектировании модуля `execution`.

## Статус

Актуален

## Последнее обновление

2026-07-20
