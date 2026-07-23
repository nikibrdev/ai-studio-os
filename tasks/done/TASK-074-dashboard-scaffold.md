# TASK-074: Каркас apps/dashboard (Next.js, толчейн, API-клиент)

## Тип

feature

## Эпик

[EPIC-009 Dashboard](../../docs/roadmap/EPIC-009-dashboard.md)

## Цель

Создать несущую конструкцию `apps/dashboard`: приложение Next.js 15, инструменты качества, типизированный клиент REST API — без бизнес-логики, только каркас, на который лягут страницы (TASK-075/076).

## Контекст

Толчейн зафиксирован при открытии эпика — дополнение [ADR-009](../../docs/adr/ADR-009-toolchain.md): Node.js 22 LTS, TypeScript, ESLint + Prettier, Vitest + React Testing Library, pnpm. `apps/dashboard` может зависеть только от `apps/api` ([module-boundaries.md](../../docs/architecture/module-boundaries.md)) — никакого прямого обращения к `internal/`, БД или шине событий.

## Scope

### Входит

- `apps/dashboard/` — Next.js 15 (App Router, TypeScript), `package.json`/`pnpm-lock.yaml`, `tsconfig.json`, ESLint (`next/core-web-vitals` + TypeScript) + Prettier конфигурация, Vitest конфигурация (+ React Testing Library, `jsdom`).
- Типизированный клиент REST API (`lib/api.ts` или аналог) — типы ответов зеркалят `docs/api/*.md` (ручное соответствие — без генерации из OpenAPI, вне scope EPIC-008/009); базовый URL — через переменную окружения (`NEXT_PUBLIC_API_URL` или аналог).
- Базовый layout (навигация, обёртка страниц) — без конкретных страниц с данными (TASK-075/076).
- `apps/dashboard/README.md` (черновик, дополняется TASK-078).

### Не входит

- Конкретные страницы со списками/деталями (TASK-075/076).
- CI-job (TASK-077).

## Критерии приёмки

- [x] `pnpm install && pnpm build` проходит локально.
- [x] `pnpm lint`, `pnpm test` (пустой/минимальный набор тестов — хотя бы один smoke-тест) проходят.
- [x] `pnpm dev` запускает приложение локально, базовый layout отображается в браузере.

## Затрагиваемые модули и документы

- `apps/dashboard/*` (новый пакет).
- `Makefile`, `scripts/verify-docs.sh` — точечный фикс (см. Отчёт).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — независима от TASK-072/073, может выполняться параллельно

## План реализации

1. `npx create-next-app@15 apps/dashboard --typescript --eslint --app --src-dir --import-alias "@/*" --use-pnpm --no-tailwind --turbopack` — базовый каркас; удалить плейсхолдерный `.gitkeep`, не тронутый до этого.
2. pnpm 11+ по умолчанию блокирует postinstall/build-скрипты неизвестных пакетов (supply-chain защита) — `sharp` (оптимизация изображений Next.js) и `unrs-resolver` (резолвер ESLint) заблокированы при первой установке; одобрены явно в `pnpm-workspace.yaml` (`allowBuilds: {sharp: true, unrs-resolver: true}`) — оба легитимные, широко используемые зависимости самого тулчейна Next.js/ESLint, не сторонний код с неясным происхождением.
3. Prettier + `eslint-config-prettier` (отключает стилистические правила ESLint, конфликтующие с Prettier — стандартная пара); Vitest + `@vitejs/plugin-react-swc` + jsdom + React Testing Library, по решению ADR-009 (дополнено при открытии эпика).
4. `src/lib/api.ts` — типизированный клиент (`listProjects`, `listProjectTasks`, `getTask`), типы вручную зеркалят `docs/api/projects.md`/`docs/api/tasks.md`; `ApiError` несёт HTTP-статус.
5. Базовый layout (`src/app/layout.tsx`) — шапка с навигацией; `page.tsx` — временная заглушка до TASK-075.
6. Тесты: `src/lib/api.test.ts` (мок `fetch`, успех/ошибка), `src/app/page.test.tsx` (smoke-рендер).
7. `package.json`: `engines.node` (`>=22 <23`, ADR-009), скрипты `format`/`format:check`/`test`; `.nvmrc` (`22`) — для будущего CI (TASK-077) и локального `nvm use`.
8. `apps/dashboard/README.md` (черновик) — по стандарту README модуля проекта, не generic-заготовка `create-next-app`.
9. **Обнаружено и исправлено при верификации**: `make verify` зависал на десятках тысяч файлов внутри свежесозданного `apps/dashboard/node_modules` — оба скрипта (`Makefile`'s `md-lint`, `scripts/verify-docs.sh`) исключали только КОРНЕВОЙ `node_modules` (`#node_modules`, `-not -path "./node_modules/*"`, `grep -rl` без `--exclude-dir`), не вложенный. Это первый пакет в репозитории со своим `node_modules`, поэтому пробел не проявлялся раньше. Исправлено на рекурсивные исключения (`#**/node_modules/**` для markdownlint-cli2; `-not -path "*/node_modules/*"` для `find`; `--exclude-dir=node_modules` для `grep -r`), с тем же заодно исключением `.next`.
10. Живая проверка в браузере (Playwright, `chromium-cli` недоступен в этом окружении — использован fallback на прямой `playwright`-скрипт из временного проекта в scratchpad): `pnpm dev`, реальный HTTP-запрос и скриншот headless Chromium — заголовок, навигация и текст-заглушка отображаются, `console --errors` пуст.
11. `make verify` (корень репозитория) — чисто, включая проверку, что исправление не сломало проверку остальных 238 существующих markdown-файлов.

## История

2026-07-22 — Architect — EPIC-009 открыт; задача поставлена в очередь.

2026-07-23 — Developer — задача взята в работу, реализована и проверена (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-074 — каркас `apps/dashboard` (Next.js, толчейн, API-клиент).

### Что сделано

- `apps/dashboard` — Next.js 15 (App Router, TypeScript), pnpm; ESLint (`next/core-web-vitals` + `next/typescript` + `prettier`) + Prettier; Vitest + React Testing Library (jsdom) + `@vitejs/plugin-react-swc`.
- `src/lib/api.ts` — типизированный клиент `apps/api` (`listProjects`, `listProjectTasks`, `getTask`), типы вручную зеркалят `docs/api/*.md`; `ApiError` с HTTP-статусом.
- Базовый layout с навигацией (`src/app/layout.tsx`), временная страница-заглушка (`src/app/page.tsx`) до TASK-075.
- `package.json.engines.node` (`>=22 <23`) + `.nvmrc` (`22`) — ADR-009.
- 3 теста (2 файла): клиент API на моке `fetch` (успех/ошибка), smoke-рендер главной страницы.
- `apps/dashboard/README.md` — черновик по стандарту README модуля проекта.
- **Найден и исправлен реальный баг инфраструктуры репозитория**: `make verify` зависал (сканировал десятки тысяч файлов в `apps/dashboard/node_modules`) — оба скрипта проверки Markdown исключали только корневой `node_modules`; это первый пакет в репозитории с собственным `node_modules`, пробел не проявлялся раньше. Исправлено рекурсивными исключениями в `Makefile` и `scripts/verify-docs.sh` (заодно добавлено исключение `.next`).

### Изменённые файлы

- `apps/dashboard/*` — весь новый пакет (см. структуру в README).
- `Makefile`, `scripts/verify-docs.sh` — исправление исключения вложенного `node_modules`/`.next`.

### Как проверялось

- `pnpm install && pnpm build` — чисто (Turbopack, статическая генерация, 0 ошибок).
- `pnpm lint`, `pnpm format:check`, `pnpm test` — все чисто (3/3 теста зелёных).
- Живая проверка в браузере: `pnpm dev` на `localhost:3000`, headless Chromium (Playwright, временный проект в scratchpad — `chromium-cli` недоступен в этом окружении) — реальный скриншот подтверждает рендер шапки/навигации/текста-заглушки; `console --errors` пуст; заголовок вкладки — «AI Studio OS — Dashboard».
- `make verify` (корень репозитория) — чисто после фикса; 239 markdown-файлов, 1319 ссылок проверено, 0 ошибок (было 238/238 до добавления README дашборда).

### Обновлённая документация

- `apps/dashboard/README.md` (новый, черновик — уточняется TASK-078).

### Open Questions

Нет.

### Рекомендации

- TASK-077 (CI для `apps/dashboard`) может использовать уже добавленный `.nvmrc`/`package.json.engines.node` для настройки `actions/setup-node@v4` (`node-version-file: apps/dashboard/.nvmrc`).
- Предупреждение `[vite:react-swc] We recommend switching to @vitejs/plugin-react...` при `pnpm test` — информационное (Vitest предлагает более быстрый плагин без SWC-специфичных возможностей, которые здесь не используются); оставлено как есть, не блокирует и не портит результат — `@vitejs/plugin-react-swc` согласуется с тем, что сам Next.js уже использует SWC.
