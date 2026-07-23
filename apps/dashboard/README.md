# apps/dashboard — веб-интерфейс

## Назначение

Веб-интерфейс на Next.js 15 ([EPIC-009](../../docs/roadmap/EPIC-009-dashboard.md), v0.8): наблюдение за проектами и задачами. Работает исключительно через `apps/api` ([module-boundaries.md](../../docs/architecture/module-boundaries.md)) — без прямого обращения к `internal/`, БД или шине событий. Первая версия — read-only, без форм действий, без аутентификации (ADR-012, Вариант 1), без канала реального времени (см. EPIC-009 «Контекст»).

## Содержание

### Толчейн (ADR-009, зафиксировано при открытии EPIC-009)

Node.js 22 LTS, TypeScript, pnpm, ESLint (`next/core-web-vitals` + `next/typescript` + `prettier`) + Prettier, Vitest + React Testing Library (jsdom).

### Структура

| Путь             | Содержимое                                                                                         |
| ---------------- | -------------------------------------------------------------------------------------------------- |
| `src/app/`       | Страницы (App Router) и корневой layout (навигация)                                                |
| `src/lib/api.ts` | Типизированный клиент `apps/api` — типы вручную зеркалят `docs/api/*.md`, без генерации из OpenAPI |

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

В работе (EPIC-009)

## Последнее обновление

2026-07-23
