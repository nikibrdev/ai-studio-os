# ADR-005: Executor Contract

## Статус

**Принято** (решение архитектора проекта, 2026-07-20)

## Дата

2026-07-20 (заготовка — 2026-07-19; принято — 2026-07-20)

## Контекст

Контракт `platform.Agent.Execute(ctx, Request) (Response, error)` был намеренно абстрактным с момента ревью EPIC-002 — `Request`/`Response` без полей, чтобы не зашивать в ядро предположения о формате обмена с конкретным AI-провайдером до принятия этого ADR. Без решения Domain Layer (v0.3) рисковал начать строиться вокруг незафиксированной абстракции.

Отдельно от формы контракта возникла терминологическая проблема: слово «Agent» смешивало два разных понятия — логическую роль-исполнителя («Developer») и конкретный технический бэкенд, который эту роль реально исполняет (Claude Code, Codex, человек, OpenHands). Смешение мешало держать ядро независимым от провайдера на практике, а не только в декларациях.

## Рассмотренные варианты

Форма контракта:

1. **Один блокирующий вызов** (`Execute(ctx, Request) (Response, error)`) — просто, но смешивает «начать работу», «отчитаться о ходе», «отдать результат» и «завершить» в одну непрозрачную операцию.
2. **Явные отдельные возможности** — каждая из четырёх функций Executor'а как отдельный метод контракта.

Терминология:

1. Оставить «Agent» единственным термином и для роли, и для исполнителя.
2. Разделить: **Agent** — логическая роль-исполнитель (например, Developer Agent); **Executor** — конкретный технический бэкенд (Claude Code, Codex, Human, OpenHands), связанный с Agent через назначение.

## Решение

Принят вариант 2 по обоим вопросам.

**Терминология.** Agent и Executor — разные понятия ([docs/domain/ubiquitous-language.md](../domain/ubiquitous-language.md)). Цепочка: Role (Developer) → Agent (Developer Agent — логический исполнитель роли) → Executor (Claude Code — реальный технический бэкенд). В коде платформы (`internal/platform`) используется **Executor** — платформа запускает исполнителей, не агентов. В документации о ролях термин «Agent»/«агент» остаётся допустимым (`.claude/agents/` по-прежнему описывает именно этот, Agent-уровень).

**Форма контракта.** Executor обязан уметь ровно четыре вещи — не больше:

```
Accept Task → Produce Artifact → Report Status → Finish Execution
```

```go
type Executor interface {
    Accept(ctx context.Context, task ExecutorTask) error
    Artifacts(ctx context.Context) ([]Artifact, error)
    Status(ctx context.Context) (ExecutionStatus, error)
    Finish(ctx context.Context) error
}
```

`ExecutorTask`, `Artifact`, `ExecutionStatus` — намеренно абстрактны (`any`) и этим ADR не конкретизируются: документ фиксирует **имена и число возможностей контракта**, а не форму передаваемых данных. Форма — задача Domain Layer (v0.3, EPIC-003), когда появятся реальные Task/Artifact/Execution.

## Последствия

### Положительные

- Ядро окончательно не знает, кто исполняет работу — только то, что Executor принимает задание и производит артефакты, статус и завершение. Соответствует видению платформы как ОС для любых исполнителей ([VISION.md](../../VISION.md)).
- Явное разделение возможностей контракта (а не один непрозрачный `Execute`) читаемо и не создаёт архитектурного долга при будущей детализации.
- Терминологическая точность (Role → Agent → Executor) убирает двусмысленность, которая иначе протекла бы в Domain Layer.

### Отрицательные

- Три места неопределённости вместо одного: `ExecutorTask`, `Artifact`, `ExecutionStatus` по отдельности остаются абстрактными до Domain Layer — решение не закрывает вопрос формы данных, только форму контракта.
- Переименование `Agent` → `Executor` в коде — потенциально breaking change; на 2026-07-20 внешних потребителей контракта нет (адаптеры не реализованы), цена перехода нулевая.

### Влияние на существующие документы и код

`internal/platform/agent.go` → `executor.go` (интерфейс `Agent` → `Executor`, `Request`/`Response` → `ExecutorTask`/`Artifact`/`ExecutionStatus`); [docs/architecture/agents.md](../architecture/agents.md), [interfaces.md](../architecture/interfaces.md), [module-boundaries.md](../architecture/module-boundaries.md), [components.md](../architecture/components.md), [overview.md](../architecture/overview.md), [DECISIONS_INDEX.md](DECISIONS_INDEX.md) — обновлены (TASK-026).

### Открытый вопрос

Где размещается тип `Artifact`, когда он перестанет быть абстрактным placeholder'ом: `platform` домен-агностичен по [ADR-015](ADR-015-internal-layering.md) и не может зависеть от `domain`, но Artifact — по философии проекта («артефакты — настоящая ценность», [VISION.md](../../VISION.md)) — это язык предметной области, а не механика платформы. Вероятное решение: `platform.Artifact` остаётся минимальной ссылкой/маркером (то, чем оперирует контракт исполнения), а полноценная сущность Artifact с метаданными и связями определяется в `domain` и оборачивает/ссылается на platform-примитив — не решено, требует архитектора при проектировании модуля Artifact (Domain Layer, EPIC-003).

## Связанные материалы

[VISION.md](../../VISION.md) · [docs/domain/ubiquitous-language.md](../domain/ubiquitous-language.md) · [interfaces.md](../architecture/interfaces.md) · [ADR-006](ADR-006-agent-execution-environment.md) (среда выполнения — Decision Required, следующий вопрос)
