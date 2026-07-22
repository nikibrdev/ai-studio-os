# PROJECT_HEALTH — прогресс проекта

## Назначение

Состояние проекта одним взглядом: экспертная оценка готовности по направлениям. Обновляется вместе с [PROJECT_MANIFEST.md](PROJECT_MANIFEST.md); со временем показатели будут выводиться из [engineering/metrics/](engineering/metrics/).

## Содержание

### Готовность по направлениям

| Направление | Готовность | Комментарий |
| --- | --- | --- |
| Architecture | **98%** | Заморожена; 2% — 3 ADR в Decision Required, ни один не блокирует ближайшие вехи (v0.8+) ([индекс](docs/adr/DECISIONS_INDEX.md)) |
| Documentation | **100%** | Для текущего этапа: 17 архитектурных документов, процессы, спецификационная структура |
| Workflow | **60%** | Каноническая state machine реализована (`workflow.Machine`, 100% покрытия) и используется всеми use-case'ами Application Layer; Definition/Step — v0.7+ |
| Implementation | **80%** | Domain (EPIC-003), Application (EPIC-004), Infrastructure (EPIC-005), AI Agent Runtime (EPIC-006), Memory System (EPIC-007) и API Layer (EPIC-008) завершены: `apps/api` реализует весь golden path через REST (15 операций, Documentation First), подтверждено сквозным HTTP-тестом на реальном PostgreSQL; по ходу найден и исправлен реальный баг межпроектной коллизии `TASK-NNN` (BUGFIX-003); без auth (ADR-012, доверенная установка); сам AI-вызов Executor'а не проверен (нет ключа в этой сессии, честный предел, TASK-056) |
| Testing | **52%** | 285 unit-тестов (Domain 91.1–100%, Application 83.1%/87.2%, Infrastructure — юнит-часть 3.8–90% + интеграционные на реальном PostgreSQL/Qdrant, Agent Runtime — юнит-часть 84–91% + интеграционный тест на реальном Docker, API — `httpapi` 84.8% + сквозной HTTP-тест на реальном PostgreSQL), CI-job `integration` (Postgres + Qdrant сервис-контейнеры); e2e/QA Engine — v0.8+ |
| API | **100%** | REST реализован (ADR-003, EPIC-008): весь golden path через `apps/api`, без auth (ADR-012, Вариант 1) |
| Dashboard | **0%** | v0.8, следующий эпик; строится после API — [decision](engineering/decisions/2026-07-22-api-before-dashboard-build-order.md) |

### Методика

Оценки — экспертные, относительно scope соответствующего направления в MVP (v1.0). Правила округления не формализованы; спорные изменения оценок обсуждаются на ревью. Автоматизируемые показатели (задачи, пакеты, покрытие документацией) — в снимках [engineering/metrics/](engineering/metrics/).

## Статус

Актуален

## Последнее обновление

2026-07-22
