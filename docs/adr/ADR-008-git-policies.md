# ADR-008: Git-политики

## Статус

**Decision Required** — политики слияния и защиты веток закрепляются в настройках GitHub и влияют на процесс каждой задачи; порядок слияния относительно Testing меняет state machine процесса. Требуется решение архитектора/мейнтейнера репозитория.

## Дата

2026-07-19 (заготовка)

## Проблема

Не зафиксированы: способ слияния PR, правила защиты `main`, число обязательных ревьюеров, допустимость AI-агента как единственного ревьюера, момент слияния относительно стадии Testing, подпись коммитов, шаблоны Issue ([git-workflow.md](../development/git-workflow.md), [review-process.md](../development/review-process.md), [state-machine.md](../architecture/state-machine.md)).

## Контекст

- GitHub Flow и Conventional Commits уже приняты как процесс ([git-workflow.md](../development/git-workflow.md)).
- Каноническая state machine вводит Testing после Review; где в этой цепочке слияние — не определено.
- Большинство PR будут созданы AI-агентами; ревью человеком — узкое место, ревью агентом — вопрос доверия.

## Возможные варианты (по подвопросам)

### Способ слияния

- **Squash:** линейная история, один коммит на задачу (плюс: чистота; минус: теряется структура коммитов агента).
- **Merge commit:** полная история (плюс: прослеживаемость; минус: шумная история).
- **Rebase:** линейность без squash (минус: переписывание истории).

### Момент слияния относительно Testing

- **Слияние после Testing:** QA проверяет ветку; в `main` попадает только проверенное (плюс: чистый `main`; минус: тестируется не итоговое состояние `main`).
- **Слияние после Review, Testing на `main`:** проверяется реальный результат слияния (плюс: тестируется итог; минус: непроверенный код временно в `main`).

### Ревью

- 1 обязательный ревьюер / 2 ревьюера / человек обязателен для всех PR / агент-ревьюер допустим с эскалацией к человеку.

### Прочее

- Защита `main` (запрет прямых пушей — предполагается, закрепить); обязательность прохождения проверок перед merge; подпись коммитов; шаблоны Issue (`.github/ISSUE_TEMPLATE/`): bug / feature / task.

## Влияние на систему

Настройки GitHub-репозиториев (платформы и управляемых проектов); условия переходов Testing → Done ([state-machine.md](../architecture/state-machine.md)); порядок событий MergeCompleted / TaskCompleted ([events.md](../architecture/events.md)); [git-workflow.md](../development/git-workflow.md) и [review-process.md](../development/review-process.md); поведение Repository Provider.

## Связанные материалы

[git-workflow.md](../development/git-workflow.md) · [review-process.md](../development/review-process.md) · [state-machine.md](../architecture/state-machine.md) · [events.md](../architecture/events.md)
