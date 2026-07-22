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

- [ ] `pnpm install && pnpm build` проходит локально.
- [ ] `pnpm lint`, `pnpm test` (пустой/минимальный набор тестов — хотя бы один smoke-тест) проходят.
- [ ] `pnpm dev` запускает приложение локально, базовый layout отображается в браузере.

## Затрагиваемые модули и документы

- `apps/dashboard/*` (новый пакет).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — независима от TASK-072/073, может выполняться параллельно

## План реализации

## История

2026-07-22 — Architect — EPIC-009 открыт; задача поставлена в очередь.

## Отчёт о выполнении
