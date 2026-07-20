# TASK-016: Защита ветки main

## Тип

chore

## Эпик

EPIC-002.5 ([docs/roadmap/EPIC-002.5-engineering-platform.md](../../docs/roadmap/EPIC-002.5-engineering-platform.md))

## Цель

Прямой push в `main` невозможен (в том числе для владельца репозитория); обязателен зелёный статус-чек `verify`; запрещены force-push и удаление ветки.

## Контекст

Разблокировано TASK-013 (чек `verify` существует и стабилен). Настройка — состояние репозитория на GitHub, не файл в дереве; применена через Branch Protection API, задокументирована и проверена здесь.

## Scope

### Входит

- Классическая защита ветки `main`: required status check `verify` (strict — ветка должна быть актуальна), `enforce_admins: true`, `allow_force_pushes: false`, `allow_deletions: false`.
- Проверка блокировки прямого push.

### Не входит

- Обязательное число формальных GitHub-approve — см. Open Questions.

## Критерии приёмки

- [x] Прямой push в `main` отклонён GitHub (`GH006: Protected branch update failed`).
- [x] Чек `verify` — required, strict.
- [x] force-push и удаление ветки запрещены.

## План реализации

Применить настройки через `PUT /repos/{owner}/{repo}/branches/main/protection` (Branch Protection API): required_status_checks (strict, contexts: ["verify"]), enforce_admins: true, allow_force_pushes: false, allow_deletions: false. required_pull_request_reviews — включить структурно (объект должен присутствовать, чтобы активировать «require PR before merge»), но required_approving_review_count выставить в 0 — обоснование в Open Questions. Проверить блокировку прямого push тестовым `git push` (пустой коммит, затем `git reset --hard` локально — на удалённую историю коммит не попал, так как push был отклонён).

## Затрагиваемые модули и документы

- Настройки репозитория GitHub (branch protection на `main`); файл задачи.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-20 — Architect — сигнал на продолжение EPIC-002.5 (план эпика утверждён).
2026-07-20 — Claude Code (Developer) — применена защита через API, проверена тестовым push (отклонён), задача оформлена и переведена в review (PR #5).

## Отчёт о выполнении

1. **Задача:** TASK-016 — защита ветки `main`.
2. **Что сделано:** Branch Protection применена через API: required status check `verify` (strict=true), `enforce_admins: true`, `allow_force_pushes: false`, `allow_deletions: false`, `required_pull_request_reviews` присутствует с `required_approving_review_count: 0`.
3. **Проверка:** прямой push пустого коммита в `main` отклонён GitHub: `GH006: Protected branch update failed... Changes must be made through a pull request... Required status check "verify" is expected.` Локальный коммит и HEAD откачены (`git reset --hard`) — в удалённую историю ничего не попало.
4. **Изменённые файлы:** только этот файл задачи (настройка — состояние репозитория, не файл в дереве).
5. **Обновлённая документация:** не требуется отдельно; ссылка на задачу — достаточный источник истины для настройки.
6. **Open Questions:**
   - **required_approving_review_count = 0, а не 1.** GitHub не засчитывает самоодобрение (self-approval) как approving review; я аутентифицируюсь как единственный аккаунт (`nikibrdev`) — при count ≥ 1 я не смог бы смержить ни один будущий PR, включая этот. Api-запрос с count=0 был принят (HTTP 200) и структурно всё равно требует прохождения через PR (объект `required_pull_request_reviews` активирует «require PR before merge»); формального GitHub-«Approved» сейчас нет — вместо него ревью оформляется комментарием к PR (см. PR #1–4). Это прямое следствие открытого вопроса **[ADR-008](../../docs/adr/ADR-008-git-policies.md)** (число обязательных ревьюеров, допустимость агента как единственного ревьюера) — решение не принималось мной, требуется ваше.
   - **enforce_admins: true** — правило действует и на владельца (соответствует «даже если работаешь один»); это означает, что при появлении второго участника с правами admin то же ограничение (0 approvals) будет действовать и на него — тоже вопрос к ADR-008.
7. **Рекомендации:** после принятия ADR-008 — если решение введёт обязательное количество approve ≥ 1, реализация потребует либо второго формального ревьюера (человека или отдельного GitHub-аккаунта/App для агента), либо оставить текущую схему (комментарий + merge) как осознанно принятую практику для соло-этапа проекта.
