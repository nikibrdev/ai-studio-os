# TASK-018: Проверка Conventional Commits в CI

## Тип

ci

## Эпик

EPIC-002.5 ([docs/roadmap/EPIC-002.5-engineering-platform.md](../../docs/roadmap/EPIC-002.5-engineering-platform.md))

## Цель

Каждый PR проверяется на соответствие коммитов формату Conventional Commits ([git-workflow.md](../../docs/development/git-workflow.md)); нарушение — красный статус.

## Контекст

Разблокировано TASK-013 (workflow verify существует). Проверяется диапазон коммитов PR (base…head); мерж-коммиты (например, после слияния main для разрешения конфликтов) исключены из проверки. На push в `main` шаг не запускается — туда попадают уже проверенные merge-коммиты PR.

## Scope

### Входит

- `scripts/check-commits.sh <base> <head>` — regex по типам из таблицы git-workflow.md (`feat|fix|docs|refactor|test|chore|ci`), опциональный `(scope)`.
- Шаг в `.github/workflows/verify.yml`: `if: github.event_name == 'pull_request'`, диапазон `origin/<base-ref>..HEAD`.
- `actions/checkout@v4` — `fetch-depth: 0` (нужна полная история для сравнения диапазонов; репозиторий пока мал — не будет ощутимой платы).

### Не входит

- Отдельная make-цель (не предусмотрена планом эпика — скрипт вызывается напрямую в CI и вручную).

## Критерии приёмки

- [x] Скрипт локально: на реальной истории — 0 нарушений; на подставном плохом сообщении коммита — детектирует и возвращает ненулевой код.
- [ ] Шаг в CI выполняется на этом PR и проходит (коммиты этой задачи — конвенциональные).
- [ ] На push в main шаг не запускается (проверяется по логам следующего push после merge).

## План реализации

`scripts/check-commits.sh` — `git log --format='%H<TAB>%s' base..head`, фильтр `Merge *`, regex-проверка типа; код возврата 1 при нарушениях. Шаг CI — после `make verify`, только для pull_request, с `origin/<base>..HEAD`.

## Затрагиваемые модули и документы

- `scripts/check-commits.sh` (новый), `.github/workflows/verify.yml`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-20 — Architect — сигнал на продолжение EPIC-002.5.
2026-07-20 — Claude Code (Developer) — ready → in-progress; скрипт протестирован локально (позитивный и негативный кейс); выполнена, переведена в review (PR #7).

## Отчёт о выполнении

1. **Задача:** TASK-018 — проверка Conventional Commits в CI.
2. **Что сделано:** `scripts/check-commits.sh` (regex по 7 типам + опциональный scope, пропуск Merge-коммитов); шаг `Check Conventional Commits` в `verify.yml` для событий `pull_request`; `fetch-depth: 0` в checkout.
3. **Изменённые файлы:** `scripts/check-commits.sh` (новый), `.github/workflows/verify.yml`, файл задачи.
4. **Как проверялось:** локально — позитивный тест (19 реальных коммитов репозитория, 0 нарушений) и негативный тест (подставной коммит `this is not a conventional commit`, скрипт вернул код 1 и указал точную строку; тестовый коммит затем отменён `git reset --hard`, в удалённую историю не попадал). В CI — этот PR: коммиты задачи сами являются позитивным тестом шага.
5. **Обновлённая документация:** не требуется отдельно (правила уже описаны в git-workflow.md).
6. **Open Questions:** нет.
7. **Рекомендации:** при появлении внешних контрибьюторов проверить поведение на PR из форков (`github.event.pull_request.base.ref` должен работать одинаково — источник ветки не влияет на доступность base ref после fetch-depth: 0).
