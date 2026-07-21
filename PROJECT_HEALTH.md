# PROJECT_HEALTH — прогресс проекта

## Назначение

Состояние проекта одним взглядом: экспертная оценка готовности по направлениям. Обновляется вместе с [PROJECT_MANIFEST.md](PROJECT_MANIFEST.md); со временем показатели будут выводиться из [engineering/metrics/](engineering/metrics/).

## Содержание

### Готовность по направлениям

| Направление | Готовность | Комментарий |
| --- | --- | --- |
| Architecture | **97%** | Заморожена; 3% — 4 ADR в Decision Required, ни один не блокирует v0.5–v0.6 ([индекс](docs/adr/DECISIONS_INDEX.md)) |
| Documentation | **100%** | Для текущего этапа: 17 архитектурных документов, процессы, спецификационная структура |
| Workflow | **60%** | Каноническая state machine реализована (`workflow.Machine`, 100% покрытия) и используется всеми use-case'ами Application Layer; Definition/Step — v0.6+ |
| Implementation | **55%** | Domain (EPIC-003), Application (EPIC-004) и Infrastructure Layer (EPIC-005) завершены: 5 сущностей + Machine + 4 use-case-сервиса + проекция чтения — теперь работают на реальном PostgreSQL, производственном EventBus (журнал в БД) и GitHub-адаптере, не только на in-memory фейках |
| Testing | **40%** | 170 unit-теста (Domain 91.1–100%, Application 83.1%/86.8%, Infrastructure — юнит-часть 50–93%, DB-логика покрыта интеграционными тестами) + интеграционные тесты на реальном PostgreSQL (golden path, все Store, EventBus/журнал), CI-job `integration`; e2e/QA Engine — v0.6+ |
| API | **0%** | REST принят (ADR-003); Infrastructure Layer готова, реализация API — v0.9 |
| Dashboard | **0%** | v0.6 |

### Методика

Оценки — экспертные, относительно scope соответствующего направления в MVP (v1.0). Правила округления не формализованы; спорные изменения оценок обсуждаются на ревью. Автоматизируемые показатели (задачи, пакеты, покрытие документацией) — в снимках [engineering/metrics/](engineering/metrics/).

## Статус

Актуален

## Последнее обновление

2026-07-21
