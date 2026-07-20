# Git Workflow

## Назначение

Правила работы с git и GitHub в AI Studio OS: ветвление, коммиты, Pull Request'ы. Обязательны для людей и AI-агентов.

## Содержание

### Модель ветвления

Используется **GitHub Flow**:

1. Основная ветка — `main`; она всегда в рабочем состоянии.
2. Любое изменение выполняется в отдельной ветке от `main`.
3. Изменение попадает в `main` только через Pull Request после ревью.
4. Прямые коммиты в `main` запрещены **без исключений, даже при работе в одиночку**; ревью может выполнять Claude ([решение](../../engineering/decisions/2026-07-19-no-direct-main-commits.md)).

### Именование веток

```
feature/<task-id>-<short-name>   # новая функциональность
bugfix/<task-id>-<short-name>    # исправление ошибки
docs/<task-id>-<short-name>      # документация
refactor/<task-id>-<short-name>  # рефакторинг
```

Пример: `feature/TASK-012-task-lifecycle`.

### Коммиты — Conventional Commits

Формат: `<type>(<scope>): <описание>`

| Тип | Назначение |
| --- | --- |
| `feat` | Новая функциональность |
| `fix` | Исправление ошибки |
| `docs` | Документация |
| `refactor` | Рефакторинг без изменения поведения |
| `test` | Тесты |
| `chore` | Служебные изменения (конфигурация, зависимости) |
| `ci` | CI/CD (workflows, конфигурация проверок) |

Правила:

1. Описание — в повелительном наклонении, без точки в конце: `feat(tasks): add task lifecycle`.
2. Один коммит — одно логическое изменение.
3. `scope` — необязателен, но рекомендуется (имя модуля или области).

### Pull Request

1. Один PR — одна задача из `tasks/`; scope PR не расширяется.
2. PR оформляется по шаблону [.github/PULL_REQUEST_TEMPLATE.md](../../.github/PULL_REQUEST_TEMPLATE.md).
3. PR ссылается на файл задачи и содержит описание изменений и проверок.
4. PR помечается GitHub-меткой, соответствующей основному типу Conventional Commits (`feat`/`fix`/`docs`/`refactor`/`test`/`chore`/`ci`) — метка используется категоризацией release notes ([release.yml](../../.github/release.yml)).
5. Перед запросом ревью автор проходит чек-лист [.claude/checklists/PR.md](../../.claude/checklists/PR.md).
6. Обязательный статус-чек `verify` и проверка Conventional Commits — в CI ([verify.yml](../../.github/workflows/verify.yml)); `main` защищена (прямой push, force-push и удаление ветки запрещены — [TASK-016](../../tasks/done/TASK-016-branch-protection.md)).
7. Слияние — после ревью и зелёного `verify`; текущая практика — merge commit (единообразная история PR); окончательный способ слияния и обязательное число формальных approve — [ADR-008](../adr/ADR-008-git-policies.md), Decision Required.

### Релизный процесс

1. **Версионирование** — [Semantic Versioning](https://semver.org/lang/ru/): `vMAJOR.MINOR.PATCH` как git-тег на `main` после слияния соответствующих PR.
2. **CHANGELOG.md** — куратируемый источник истины (формат Keep a Changelog): при релизе раздел `[Unreleased]` переименовывается в версию с датой, начинается новый `[Unreleased]`.
3. **GitHub Release notes** — вспомогательный, полуавтоматический артефакт: кнопка «Generate release notes» на странице релиза группирует PR по меткам согласно [.github/release.yml](../../.github/release.yml). Не заменяет CHANGELOG.md — при необходимости содержимое сверяется/переносится в него.
4. Категоризация release notes работает только если у PR проставлена метка типа (см. правило Pull Request, п. 4); PR без метки попадают в категорию «Other Changes».

### Decision Required

Способ слияния PR (см. текущую практику выше), правила защиты `main` (базово настроены — [TASK-016](../../tasks/done/TASK-016-branch-protection.md); обязательное число approve не зафиксировано), момент слияния относительно стадии Testing и политика подписи коммитов — [ADR-008](../adr/ADR-008-git-policies.md).

## Статус

Актуален

## Последнее обновление

2026-07-20
