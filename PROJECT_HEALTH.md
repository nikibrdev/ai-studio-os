# PROJECT_HEALTH — прогресс проекта

## Назначение

Состояние проекта одним взглядом: экспертная оценка готовности по направлениям. Обновляется вместе с [PROJECT_MANIFEST.md](PROJECT_MANIFEST.md); со временем показатели будут выводиться из [engineering/metrics/](engineering/metrics/).

## Содержание

### Готовность по направлениям

| Направление | Готовность | Комментарий |
| --- | --- | --- |
| Architecture | **98%** | Заморожена; 2% — 3 ADR в Decision Required, ни один не блокирует ближайшие вехи (v1.0) ([индекс](docs/adr/DECISIONS_INDEX.md)) |
| Documentation | **100%** | Для текущего этапа: 17 архитектурных документов, процессы, спецификационная структура |
| Workflow | **60%** | Каноническая state machine реализована (`workflow.Machine`, 100% покрытия) и используется всеми use-case'ами Application Layer; Definition/Step — по решению архитектора, не раньше v1.0 |
| Implementation | **90%** | Domain (EPIC-003), Application (EPIC-004), Infrastructure (EPIC-005), AI Agent Runtime (EPIC-006), Memory System (EPIC-007), API Layer (EPIC-008) и Dashboard (EPIC-009) завершены: `apps/dashboard` (Next.js 15) показывает список проектов → задачи → детали задачи через реальные HTTP-запросы к `apps/api`, подтверждено вживую (Playwright, реальный `apps/api`+PostgreSQL); по ходу закрыты два реальных пробела API (списковые операции — TASK-072; описательные поля задачи в `TaskProjection` — TASK-076); read-only (формы действий отложены); без auth (ADR-012, доверенная установка); сам AI-вызов Executor'а не проверен (нет ключа в этой сессии, честный предел, TASK-056) |
| Testing | **55%** | 298 unit-тестов Go (Domain 91.1–100%, Application 83.7%/88.4%, Infrastructure — юнит-часть 3.5–90% + интеграционные на реальном PostgreSQL/Qdrant, Agent Runtime — юнит-часть 84–91% + интеграционный тест на реальном Docker, API — `httpapi` 85.1% + сквозной HTTP-тест на реальном PostgreSQL) + 8 unit-тестов `apps/dashboard` (Vitest); CI-job `integration` (Postgres + Qdrant сервис-контейнеры), CI-job `verify` включает `apps/dashboard` (TASK-077); e2e/QA Engine — по решению архитектора, не раньше v1.0 |
| API | **100%** | REST реализован (ADR-003, EPIC-008/009): весь golden path + списковые операции через `apps/api`, без auth (ADR-012, Вариант 1) |
| Dashboard | **100%** | Read-only реализован (EPIC-009): список проектов, задачи проекта, детали задачи; формы действий/auth/realtime — сознательно отложены |

### Методика

Оценки — экспертные, относительно scope соответствующего направления в MVP (v1.0). Правила округления не формализованы; спорные изменения оценок обсуждаются на ревью. Автоматизируемые показатели (задачи, пакеты, покрытие документацией) — в снимках [engineering/metrics/](engineering/metrics/).

## Статус

Актуален

## Последнее обновление

2026-07-23
