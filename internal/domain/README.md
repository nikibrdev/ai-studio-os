# Слой: internal/domain

## Назначение

Доменный слой — центр системы. Каждый подкаталог — один доменный модуль с единственной ответственностью и владением своими сущностями ([docs/architecture/domain-model.md](../../docs/architecture/domain-model.md)). Всё строится вокруг домена: application и infrastructure зависят от domain, но не наоборот.

## Содержание

### Состав

| Пакет | Ответственность | README |
| --- | --- | --- |
| `shared/` | Язык домена: Role, TaskState (позже ID, ошибки, value objects) | [shared/README.md](shared/README.md) |
| `artifact/` | Сущность Artifact — инварианты, Lifecycle и команды, реализованные в коде (не только контракты) | [artifact/README.md](artifact/README.md) |
| `execution/` | Сущность Execution — единичный запуск Executor'а: Lifecycle, команды, гонка Fail/Abort порядком выполнения | [execution/README.md](execution/README.md) |
| `task/` | Контракты записи/чтения/экспорта задач | [task/README.md](task/README.md) |
| `project/` | Реестр управляемых проектов | [project/README.md](project/README.md) |
| `event/` | Словарь типов событий | [event/README.md](event/README.md) |
| `workflow/` | Правила state machine (Rules) и определения процессов (Definition, Step) | [workflow/README.md](workflow/README.md) |

Все пять спецификаций EPIC-003 (Artifact, Execution, Executor, Task, Project) утверждены — [docs/roadmap/EPIC-003-domain-layer.md](../../docs/roadmap/EPIC-003-domain-layer.md). `artifact/` (TASK-034) и `execution/` (TASK-035) — пакеты с реализованной логикой этапа 2; `executor` и расширения `task`/`project` появятся отдельными задачами того же этапа, в том же порядке проектирования. Остальные модули доменной модели (`tool`, `memory`, `git`, `identity`) — вне scope EPIC-003, по решению архитектора отдельно.

### Зависимости

- Разрешено: стандартная библиотека, `internal/domain/shared` (язык домена), `pkg/`.
- Запрещено: `internal/platform` (домен не знает о платформенных абстракциях), `internal/application`, `internal/infrastructure`, `apps/`, `agents/`, `tools/`, инфраструктурные библиотеки, внутренние пакеты соседних доменных модулей.

### События

Междоменное взаимодействие — только через события ([ADR-014](../../docs/adr/ADR-014-module-interaction.md)); схемы payload определяются модулями-источниками (Domain Layer, следующий эпик).

### Правила

Каждый модуль обязан иметь README (назначение, зависимости, события, ответственность) — [docs/development/documentation.md](../../docs/development/documentation.md).

## Статус

Актуален

## Последнее обновление

2026-07-21
