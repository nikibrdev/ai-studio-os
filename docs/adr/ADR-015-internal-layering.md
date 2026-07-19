# ADR-015: Слои internal/ — domain, application, platform, infrastructure

## Статус

**Принято** (решение архитектора проекта, 2026-07-19)

## Дата

2026-07-19

## Контекст

По итогам ревью EPIC-002 кросс-доменные контракты (EventBus, Agent, Tool, MemoryProvider, RepositoryProvider) и словари (Role, TaskState) были временно размещены в `internal/core`. Требовалось решить окончательное размещение: слово «core» быстро становится свалкой — в крупных проектах через год в core лежат десятки несвязанных вещей.

## Рассмотренные варианты

1. **Оставить `internal/core`** — просто, но без разделения по ответственности; риск «свалки».
2. **Разделение по ответственности: platform + domain/shared** — платформенные абстракции отдельно от языка домена.

## Решение

Принят вариант 2. Структура `internal/`:

```
internal/
├── domain/          # Предметная область
│   ├── shared/      # Язык домена: Role, TaskState, ID, ошибки, value objects
│   ├── task/
│   ├── project/
│   ├── workflow/
│   └── event/
├── application/     # Сценарии использования, проекции
├── platform/        # Абстракции платформы: EventBus, Agent, Tool,
│                    # MemoryProvider, RepositoryProvider
└── infrastructure/  # Адаптеры: PostgreSQL, In-Memory Bus, GitHub, память
```

Разделение по ответственности:

- **`platform/`** — инфраструктурные абстракции платформы, а не предметной области: EventBus, Agent, Tool, MemoryProvider, RepositoryProvider. Слой домен-агностичен (не импортирует domain).
- **`domain/shared/`** — язык предметной области: Role, TaskState; по мере принятия решений — идентификаторы (ADR-011), доменные ошибки, value objects.
- Контракт применения state machine («Workflow» из interfaces.md) — доменное правило: тип `Rules` в `internal/domain/workflow`.

Правила зависимостей:

| Слой | Может импортировать |
|---|---|
| `domain/shared` | только stdlib |
| `domain/<module>` | stdlib, `domain/shared`, публичные контракты соседних доменных модулей |
| `application` | domain, platform, `pkg/` |
| `platform` | только stdlib (и `pkg/`) |
| `infrastructure` | platform (реализует), порты domain, драйверы своей системы |

`internal/core` упразднён.

## Последствия

### Положительные

- Ясное разделение: язык домена отделён от механики платформы; «свалка core» невозможна по построению.
- Платформенные контракты заменяемы независимо от домена (agent-agnostic ядро — architектурный принцип).
- Слои совпадают с планом эпиков: Domain Layer → Application Layer → Infrastructure.

### Отрицательные

- Ранняя фиксация слоёв: перенос типа между слоями после появления зависимостей потребует рефакторинга (смягчено тем, что решение принято до какой-либо реализации).

### Влияние на существующие документы и код

Код перенесён (`internal/platform`, `internal/domain/shared`, `workflow.Rules`); обновлены [module-boundaries.md](../architecture/module-boundaries.md), [core.md](../architecture/core.md), [components.md](../architecture/components.md), [project-structure.md](../architecture/project-structure.md), [interfaces.md](../architecture/interfaces.md), README слоёв и модулей, контекст агентов.

## Связанные материалы

[Ревью EPIC-002](../../engineering/reviews/2026-07-19-epic-002-code-review.md) · [module-boundaries.md](../architecture/module-boundaries.md) · [ADR-014](ADR-014-module-interaction.md)
