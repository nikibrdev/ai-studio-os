# Рабочий процесс задач

## Назначение

Описывает организацию процесса работы над задачами: участие ролей, правила и файловую реализацию в `tasks/`. Канонический жизненный цикл (состояния и переходы) — [state-machine.md](state-machine.md); этот документ его не дублирует.

## Содержание

### Канонический жизненный цикл

Задача проходит состояния: Backlog → Ready → In Progress → Review → Testing → Done → Archived, с особыми состояниями Blocked и Cancelled. Диаграмма, полная таблица переходов, guard-условия и инварианты — [state-machine.md](state-machine.md).

### Участие ролей по стадиям

| Стадия | Ответственная роль | Что делает |
| --- | --- | --- |
| Backlog | Project Manager | Фиксирует и готовит задачи, приоритизирует |
| Ready | Project Manager | Подтверждает Definition of Ready |
| In Progress | Developer | Реализует задачу, оформляет PR |
| Review | Reviewer | Проверяет PR ([review-process.md](../development/review-process.md)) |
| Testing | QA Engineer | Проверяет поведение и качество ([QA-чеклист](../../.claude/checklists/QA.md)) |
| Done | QA Engineer | Подтверждает Definition of Done |
| Blocked | Project Manager | Организует снятие блокировки |
| Cancelled / Archived | Project Manager | Отменяет с причиной, архивирует |

Обязанности ролей: [.claude/agents/](../../.claude/agents/).

### Правила процесса

1. Задача — markdown-файл по шаблону [Task.md](../../.claude/templates/Task.md); история переходов, замечания и отчёты накапливаются в самом файле.
2. В работу берутся только задачи из `ready/` (DoR проверен, см. [CONSTITUTION.md](../../CONSTITUTION.md)).
3. Одна задача — один исполнитель роли Developer и один PR.
4. Возврат с ревью или тестирования: задача возвращается в In Progress с замечаниями/дефектами, зафиксированными в файле задачи.
5. Блокировка и отмена всегда сопровождаются записью причины; для блокировки — также требуемого решения.
6. Допустимость перехода определяется правилами state machine; нарушение — ошибка процесса.

### Файловая система `tasks/`

**Принято ([ADR-004](../adr/ADR-004-task-storage.md)):** целевая модель — PostgreSQL как источник истины, Task Engine как единственная точка переходов, `tasks/` — markdown-экспорт.

**Переходный период** (до ввода Task Engine): стадиям соответствуют каталоги `backlog/`, `ready/`, `in-progress/`, `review/`, `blocked/`, `done/`, `archive/`; перемещение файла = переход состояния; Testing — в `review/` (запись в «Истории»), Cancelled — в `archive/` (причина в «Истории»).

### Статус решений

- [ADR-004](../adr/ADR-004-task-storage.md) — **принято** (см. выше).
- [ADR-011](../adr/ADR-011-task-identifiers.md) — **Decision Required**: формат идентификаторов задач.
- [ADR-007](../adr/ADR-007-pm-qa-executors.md) — **Decision Required**: исполнители ролей PM и QA в MVP.

## Статус

Актуален

## Последнее обновление

2026-07-19
