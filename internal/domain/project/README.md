# Модуль: internal/domain/project

## Назначение

Сущность Project и контракт реестра проектов, которыми управляет платформа: граница инициативы, в которой существуют Epic, Task и Artifact ([полная спецификация](../../../docs/specifications/domain/project.md), статус Утверждена; [domain-model.md](../../../docs/architecture/domain-model.md), «Project»).

## Содержание

### Ответственность

- Тип `Project` — сущность с инвариантами в коде (TASK-038): Lifecycle `Created → Active → Archived`; переход Created → Active — явная команда `Activate` с guard-условием «подключён хотя бы один Repository» (решение финального ревью спецификации), а не побочный эффект `ConnectRepository`.
- Набор подключённых Repository только растёт — отключение не предусмотрено v1 (сознательное ограничение, Decision Log спецификации); повторное подключение того же Repository — идемпотентный no-op.
- Предикат `AcceptsNewContent` — новые Epic/Task/Artifact создаются только в Active (Behavioral Invariant 4); само создание — в модулях-владельцах.
- `Registry` — контракт жизненного цикла: Create → ConnectRepository → **Activate** (расширение этапа 2) → Archive; архив неизменяем.
- Владение данными: Project, назначения исполнителей ролей в проекте (контракт назначений — анонсированная future work, не специфицирован).

### Зависимости

- Разрешено: stdlib.
- Запрещено: другие доменные модули, application, infrastructure, драйверы.

### События

`Created`, `RepositoryConnected` (на каждое подключение, не только первое), `Activated`, `Archived` — возвращаются командами сущности как значения; публикация через Event Bus — вне модуля.

### Ограничения

Формат подключения репозиториев и содержимое `projects/` — Decision Required ([ADR-013](../../../docs/adr/ADR-013-managed-projects.md)); Repository — строка-ссылка до его принятия.

## Статус

Актуален

## Последнее обновление

2026-07-21
