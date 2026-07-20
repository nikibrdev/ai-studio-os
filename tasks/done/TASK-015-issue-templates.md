# TASK-015: Шаблоны Issue и актуализация шаблона PR

## Тип

chore

## Эпик

EPIC-002.5 ([docs/roadmap/EPIC-002.5-engineering-platform.md](../../docs/roadmap/EPIC-002.5-engineering-platform.md))

## Цель

Три структурированных шаблона Issue (bug / feature / task) в формате GitHub Issue Forms; актуализирован шаблон PR — добавлен тип `ci`, закрывающий Open Question из TASK-013.

## Контекст

По плану эпика (п. 3): `.github/ISSUE_TEMPLATE/` + сверка PR-шаблона. При ревью TASK-013 зафиксирован открытый вопрос: типа `ci` не было ни в шаблоне PR, ни в git-workflow.md — использовался `chore(ci)`. Решение здесь: добавить `ci` как отдельный тип (CI/CD — не совсем «служебное» в смысле chore, отдельный тип точнее для будущей автогенерации release notes по типам, EPIC-002.5 TASK-020).

## Scope

### Входит

- `.github/ISSUE_TEMPLATE/bug_report.yml`, `feature_request.yml`, `task.yml`, `config.yml` (`blank_issues_enabled: false` — Issue только по шаблону).
- `.github/PULL_REQUEST_TEMPLATE.md` — добавлен пункт `ci`.
- `docs/development/git-workflow.md` — `ci` в таблице Conventional Commits.

### Не входит

- Автоматизация release notes по типам (TASK-020).

## Критерии приёмки

- [ ] Три Issue-шаблона валидны (YAML Issue Forms); `config.yml` отключает произвольный Issue.
- [ ] PR-шаблон и git-workflow.md согласованы (тип `ci` в обоих).
- [ ] Прогон verify — зелёный.

## План реализации

По утверждённому плану эпика (п. 3), с уточнением по Open Question TASK-013: добавить тип `ci`.

## Затрагиваемые модули и документы

- `.github/ISSUE_TEMPLATE/*`, `.github/PULL_REQUEST_TEMPLATE.md`, `docs/development/git-workflow.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-20 — Architect — сигнал на продолжение EPIC-002.5.
2026-07-20 — Claude Code (Developer) — ready → in-progress → выполнена, переведена в review (PR #4).

## Отчёт о выполнении

1. **Задача:** TASK-015 — шаблоны Issue + актуализация PR-шаблона.
2. **Что сделано:** 3 Issue Forms (bug_report, feature_request, task) + config.yml; PR-шаблон и git-workflow.md дополнены типом `ci` (закрыт Open Question TASK-013).
3. **Изменённые файлы:** `.github/ISSUE_TEMPLATE/{bug_report,feature_request,task,config}.yml`, `.github/PULL_REQUEST_TEMPLATE.md`, `docs/development/git-workflow.md`, файл задачи.
4. **Как проверялось:** прогон verify в PR #4 (merge только после зелёного).
5. **Обновлённая документация:** git-workflow.md.
6. **Open Questions:** нет.
7. **Рекомендации:** метки `type: bug` / `type: feature` / `type: task` из шаблонов Issue стоит создать в репозитории (GitHub создаёт их автоматически при первом использовании формы, но можно завести заранее в отдельной задаче).
