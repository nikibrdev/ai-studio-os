# apps/dashboard — веб-интерфейс

## Назначение

Веб-интерфейс на Next.js 15 ([EPIC-009](../../docs/roadmap/EPIC-009-dashboard.md), v0.8): наблюдение за проектами и задачами. Работает исключительно через `apps/api` ([module-boundaries.md](../../docs/architecture/module-boundaries.md)) — без прямого обращения к `internal/`, БД или шине событий. Первая версия — read-only, без форм действий, без аутентификации (ADR-012, Вариант 1), без канала реального времени (см. EPIC-009 «Контекст»).

## Содержание

### Толчейн (ADR-009, зафиксировано при открытии EPIC-009)

Node.js 22 LTS, TypeScript, pnpm, ESLint (`next/core-web-vitals` + `next/typescript` + `prettier`) + Prettier, Vitest + React Testing Library (jsdom).

### Структура

- `src/app/layout.tsx` — корневой layout, шапка с навигацией.
- `src/app/page.tsx` — `/`, список проектов (TASK-075).
- `src/app/loading.tsx`/`error.tsx` — общее состояние загрузки (Suspense) и ошибки (Error Boundary с кнопкой «Попробовать снова») для всех страниц.
- `src/app/projects/[id]/page.tsx` — список задач проекта (TASK-076).
- `src/app/projects/[id]/tasks/[taskId]/page.tsx` — детали задачи: заголовок, тип, scope, критерии приёмки, состояние, время обновления (TASK-076).
- `src/lib/api.ts` — типизированный клиент `apps/api`, типы вручную зеркалят `docs/api/*.md`, без генерации из OpenAPI.

Все страницы — `export const dynamic = "force-dynamic"`: данные всегда актуальны только на момент запроса, статическая генерация `next build` неприменима (иначе сборка зависает на `fetch` к недоступному во время сборки `apps/api` — обнаружено в TASK-075).

### Проверено вживую

Полный путь навигации (список проектов → задачи проекта → детали задачи) проверен против реального `apps/api` + PostgreSQL (Playwright, headless Chromium): данные отображаются корректно, включая кириллицу; недоступный API и несуществующие проект/задача дают читаемую ошибку (`error.tsx`), а не падение страницы.

### CI

`.github/workflows/verify.yml` — шаги `pnpm install --frozen-lockfile`/`lint`/`format:check`/`test`/`build` внутри job `verify` (TASK-077) — тот же единственный обязательный статус-чек ветки, что и для Go-кода.

### Запуск локально

```bash
pnpm install
export NEXT_PUBLIC_API_URL="http://localhost:8080"  # apps/api, опционально — по умолчанию localhost:8080
pnpm dev
```

Требует запущенный `apps/api` (см. [apps/api/README.md](../api/README.md)) для реальных данных.

### Проверки

```bash
pnpm lint
pnpm format:check
pnpm test
pnpm build
```

## Статус

Завершён (EPIC-009)

## Последнее обновление

2026-07-23
