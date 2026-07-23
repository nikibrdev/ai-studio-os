# TASK-077: CI для apps/dashboard

## Тип

chore

## Эпик

[EPIC-009 Dashboard](../../docs/roadmap/EPIC-009-dashboard.md)

## Цель

`apps/dashboard` проверяется в CI наравне с Go-кодом — без этого frontend остаётся непроверяемым по факту, тот же принцип строгости, что применялся к каждому Go-пакету этой сессии.

## Контекст

`.github/workflows/verify.yml` сейчас знает только про Go/Markdown. Требуется job для pnpm install/lint/test/build — по аналогии с тем, как `integration`-job был добавлен для PostgreSQL/Qdrant (EPIC-005/007), но как часть обязательного `verify` (frontend lint/test/build — не интеграционный тест, требующий внешней инфраструктуры, а обычная проверка кода).

## Scope

### Входит

- `.github/workflows/verify.yml` — job (или шаги в существующем job) для `apps/dashboard`: `pnpm install --frozen-lockfile`, `pnpm lint`, `pnpm test`, `pnpm build`; кэширование pnpm store.
- Обязательный статус-чек — падение этого job'а блокирует merge (в отличие от `integration`, который не требует внешней инфраструктуры и должен быть надёжным).

### Не входит

- Деплой/публикация собранного приложения — не входит ни в этот эпик, ни в MVP.

## Критерии приёмки

- [x] CI зелёный на PR, вносящем изменения в `apps/dashboard`.
- [x] Сознательно сломанный lint/test в `apps/dashboard` (временная проверка) даёт красный CI — job действительно проверяет, а не пропускает молча.

## Затрагиваемые модули и документы

- `.github/workflows/verify.yml`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-074 (нужен реальный `package.json`/скрипты)

## План реализации

1. `.github/workflows/verify.yml` — шаги для `apps/dashboard` внутри уже существующего job `verify` (не отдельный job): `verify` уже единственный обязательный статус-чек ветки; отдельный job потребовал бы отдельной регистрации в branch protection settings GitHub, не выражаемой в самом workflow-файле.
2. `pnpm/action-setup@v4` (устанавливает pnpm) → `actions/setup-node@v7` с `cache: 'pnpm'` (нужен pnpm на PATH раньше, чтобы вычислить ключ кеша) → `pnpm install --frozen-lockfile` → `pnpm lint` → `pnpm format:check` → `pnpm test` → `pnpm build`, все с `working-directory: apps/dashboard`.
3. Живая проверка на реальном PR (не только локально) — обнаружены и исправлены две реальные ошибки конфигурации, ни одна не воспроизводилась локально (там `pnpm install` уже видел версию pnpm иначе):
   - `pnpm/action-setup@v4` требует явную версию pnpm (через `version` или `packageManager` в `package.json`) — добавлено `"packageManager": "pnpm@11.16.0"` в `apps/dashboard/package.json` (тот же принцип фиксации версий, что gofumpt/golangci-lint, BUGFIX-002).
   - Действие по умолчанию ищет `package.json` в корне репозитория — там его нет (Go-монорепозиторий); добавлен `package_json_file: apps/dashboard/package.json`.
4. Доказательство отрицательного случая (обязательный критерий приёмки): временно испорчено одно тестовое ожидание в `api.test.ts`, запушено, дождался реального красного CI именно на шаге `Test dashboard` (не на посторонней ошибке конфигурации), затем откачено обратно.
5. `make verify` (локально, Go-часть) — чисто на каждом шаге.

## История

2026-07-22 — Architect — EPIC-009 открыт; задача поставлена в очередь.

2026-07-23 — Developer — задача взята в работу; два реальных прогона CI вскрыли и позволили исправить два дефекта конфигурации (версия pnpm не указана, package.json не в корне); негативный сценарий подтверждён реальным красным прогоном; реализовано и проверено (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-077 — CI для `apps/dashboard`.

### Что сделано

- `.github/workflows/verify.yml`: шаги для `apps/dashboard` (`pnpm install --frozen-lockfile`/`lint`/`format:check`/`test`/`build`) добавлены в существующий job `verify` — единственный обязательный статус-чек ветки, без необходимости менять branch protection settings.
- `apps/dashboard/package.json`: `"packageManager": "pnpm@11.16.0"` — требование `pnpm/action-setup@v4`, найденное первым реальным прогоном CI (без этого поля или явного `version` действие падает с «No pnpm version is specified»).
- `.github/workflows/verify.yml`: `package_json_file: apps/dashboard/package.json` в шаге `Set up pnpm` — действие по умолчанию ищет `package.json` в корне репозитория, которого там нет (найдено вторым реальным прогоном CI).
- Негативный сценарий доказан вживую: временно испорченное тестовое ожидание дало реальный красный CI именно на шаге `Test dashboard` (не раньше, не по посторонней причине) — job действительно проверяет код, а не пропускает его молча.

### Изменённые файлы

- `.github/workflows/verify.yml` — шаги для `apps/dashboard`, `pnpm/action-setup` с `package_json_file`.
- `apps/dashboard/package.json` — `packageManager`.

### Как проверялось

- Три реальных прогона CI на PR: (1) красный на «No pnpm version is specified» → исправлено; (2) красный на «package.json не найден» → исправлено; (3) зелёный, с подтверждением по логу, что шаги `Lint dashboard`/`Check dashboard formatting`/`Test dashboard` (4 файла, 8 тестов)/`Build dashboard` реально выполнились (не пропущены), `verify` стал заметно дольше (2m02s против обычных ~1m30s Go-only).
- Четвёртый прогон: намеренно испорченное тестовое ожидание — реальный красный CI именно на шаге `Test dashboard`, подтверждено логом; откачено, пятый прогон — снова зелёный.
- `make verify` (Go-часть, локально) — чисто на каждом шаге.

### Обновлённая документация

Нет отдельных изменений документации сверх кода и `package.json`.

### Open Questions

Нет.

### Рекомендации

Нет.
