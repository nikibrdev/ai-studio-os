# Спецификации модулей

## Назначение

Полные технические спецификации модулей — техническое задание для реализации. Именно отсюда Claude Code (и любой исполнитель) берёт требования перед реализацией модуля ([решение](../../engineering/decisions/2026-07-19-architecture-spec-implementation.md)).

## Содержание

### Два уровня документации модуля

| Документ | Где | Содержание | Объём |
| --- | --- | --- | --- |
| **README** | в каталоге модуля | Назначение, зависимости, экспортируемые сущности, ограничения, ссылки | 1–2 страницы |
| **Specification** | здесь, `docs/specifications/<слой>/<модуль>.md` | Требования, сценарии использования, инварианты, события, ограничения, будущие расширения | полное ТЗ |

README — краткая документация для разработчика; Specification — полноценное техническое задание.

### Структура каталога

```
docs/specifications/
├── domain/          # Спецификации доменных модулей (task, project, ...)
├── application/     # Спецификации сценариев использования
├── platform/        # Спецификации платформенных абстракций
└── infrastructure/  # Спецификации адаптеров
```

Спецификации оформляются по шаблону [.claude/templates/Specification.md](../../.claude/templates/Specification.md).

### Правило нового пакета

Ни один новый пакет не может появиться без:

1. **README** (в каталоге пакета);
2. **Specification** (здесь);
3. **TASK** (в системе задач);
4. **Acceptance Criteria** (в задаче).

Задача на реализацию без спецификации не соответствует Definition of Ready.

### Усиленное требование для Domain Layer

Начиная с v0.3 (Domain Layer, EPIC-003) спецификация доменного пакета оформляется по отдельному, более строгому шаблону — [Specification-Domain.md](../../.claude/templates/Specification-Domain.md): девятнадцать обязательных разделов, начиная с One Sentence и Identity — модель определяется до текста о назначении (Model First) — и заканчивая Open Questions; три независимых прохода проверки перед утверждением ([Domain Specification Review — чек-лист](../../.claude/checklists/DomainSpecificationReview.md): Internal Consistency, Cross-domain Consistency, Future-proof Review); пишется тремя отдельными PR (фундамент → поведение → завершение). Спецификация должна быть **утверждена** до начала реализации. Цикл: Architecture → Specification → Implementation → Review → Merge. Решения: [2026-07-20-domain-layer-specification-requirement.md](../../engineering/decisions/2026-07-20-domain-layer-specification-requirement.md) (минимальная база), [2026-07-20-domain-specification-review.md](../../engineering/decisions/2026-07-20-domain-specification-review.md) (12 разделов), [2026-07-20-domain-specification-three-pass-review.md](../../engineering/decisions/2026-07-20-domain-specification-three-pass-review.md) (+4 раздела, три прохода проверки, сознательный темп работы), [2026-07-20-domain-specification-model-first.md](../../engineering/decisions/2026-07-20-domain-specification-model-first.md) (+3 раздела, структура PR 1, главный принцип).

## Статус

Актуален

## Последнее обновление

2026-07-20
