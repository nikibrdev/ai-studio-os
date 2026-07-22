# PROJECT_HEALTH — прогресс проекта

## Назначение

Состояние проекта одним взглядом: экспертная оценка готовности по направлениям. Обновляется вместе с [PROJECT_MANIFEST.md](PROJECT_MANIFEST.md); со временем показатели будут выводиться из [engineering/metrics/](engineering/metrics/).

## Содержание

### Готовность по направлениям

| Направление | Готовность | Комментарий |
| --- | --- | --- |
| Architecture | **97%** | Заморожена; 3% — 4 ADR в Decision Required, ни один не блокирует ближайшие вехи (v0.7+) ([индекс](docs/adr/DECISIONS_INDEX.md)) |
| Documentation | **100%** | Для текущего этапа: 17 архитектурных документов, процессы, спецификационная структура |
| Workflow | **60%** | Каноническая state machine реализована (`workflow.Machine`, 100% покрытия) и используется всеми use-case'ами Application Layer; Definition/Step — v0.7+ |
| Implementation | **70%** | Domain (EPIC-003), Application (EPIC-004), Infrastructure (EPIC-005), AI Agent Runtime (EPIC-006) и Memory System (EPIC-007) завершены: `platform.MemoryProvider` реализован целиком (файлы + Qdrant) и собран в `wiring.System`, подтверждено реальным прогоном (запись/поиск/изоляция/восстановление индекса); эмбеддинг наивный, не семантический по сути — честно задокументированное ограничение (ADR-018); сам AI-вызов Executor'а не проверен (нет ключа в этой сессии, честный предел, TASK-056) |
| Testing | **48%** | 253 unit-теста (Domain 91.1–100%, Application 83.1%/86.8%, Infrastructure — юнит-часть 50–90% + интеграционные на реальном PostgreSQL/Qdrant, Agent Runtime — юнит-часть 84–91% + интеграционный тест на реальном Docker), CI-job `integration` (Postgres + Qdrant сервис-контейнеры); e2e/QA Engine — v0.8+ |
| API | **0%** | REST принят (ADR-003); Infrastructure Layer готова, реализация API — v0.9 |
| Dashboard | **0%** | v0.8 |

### Методика

Оценки — экспертные, относительно scope соответствующего направления в MVP (v1.0). Правила округления не формализованы; спорные изменения оценок обсуждаются на ревью. Автоматизируемые показатели (задачи, пакеты, покрытие документацией) — в снимках [engineering/metrics/](engineering/metrics/).

## Статус

Актуален

## Последнее обновление

2026-07-22
