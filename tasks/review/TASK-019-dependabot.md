# TASK-019: Dependabot

## Тип

chore

## Эпик

EPIC-002.5 ([docs/roadmap/EPIC-002.5-engineering-platform.md](../../docs/roadmap/EPIC-002.5-engineering-platform.md))

## Цель

Автоматические PR на обновление зависимостей: `gomod` и `github-actions`; `npm`/`pnpm` — отдельной задачей при появлении `apps/dashboard`.

## Контекст

По плану эпика (TASK-019). `go.mod` пока не содержит `require`-зависимостей (только контракты; линтеры/форматтеры ставятся в CI через `go install`, не импортируются кодом) — конфигурация готова заранее, к моменту появления зависимостей в Domain Layer.

## Scope

### Входит

- `.github/dependabot.yml`: `gomod` (директория `/`, weekly, commit-message `chore` + scope), `github-actions` (директория `/`, weekly, commit-message `ci` + scope — используя тип, добавленный в TASK-015).

### Не входит

- `npm`/`pnpm` ecosystem (нет `apps/dashboard` с `package.json`).

## Критерии приёмки

- [x] YAML валиден (проверено `js-yaml`).
- [ ] Прогон verify в PR — зелёный.
- [ ] Dependabot принимает конфигурацию (проверяется вкладкой Insights → Dependency graph → Dependabot после merge).

## План реализации

`.github/dependabot.yml`, `version: 2`, два `updates`-блока (gomod, github-actions), `schedule.interval: weekly`, `commit-message.include: scope` для соответствия Conventional Commits (TASK-018 проверит эти PR так же, как любые другие).

## Затрагиваемые модули и документы

- `.github/dependabot.yml` (новый).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-20 — Architect — сигнал на продолжение EPIC-002.5.
2026-07-20 — Claude Code (Developer) — ready → in-progress → выполнена, переведена в review (PR #8).

## Отчёт о выполнении

1. **Задача:** TASK-019 — Dependabot.
2. **Что сделано:** `.github/dependabot.yml` с двумя экосистемами (gomod, github-actions), еженедельно, commit-message с префиксами `chore`/`ci` и scope — совместимо с проверкой TASK-018.
3. **Изменённые файлы:** `.github/dependabot.yml` (новый), файл задачи.
4. **Как проверялось:** YAML провалидирован `npx js-yaml` (корректный парсинг, ожидаемая структура); прогон verify — в PR #8.
5. **Обновлённая документация:** не требуется отдельно.
6. **Open Questions:** нет.
7. **Рекомендации:** добавить `npm`/`pnpm` ecosystem в этот же файл отдельной задачей при создании `apps/dashboard/package.json` (v0.6).
