# EPIC-004: Application Layer — сценарии использования

## Цель

Реализовать Application Layer AI Studio OS (v0.4, [ROADMAP.md](../../ROADMAP.md)): сценарии команд (use-case'ы) поверх завершённого Domain Layer и проекции для чтения из событий ([ADR-014](../adr/ADR-014-module-interaction.md)). Результат по ROADMAP: «платформа исполняет use-case'ы, не завязанные на конкретную инфраструктуру» — вся работа ведётся против портов, инфраструктурные адаптеры появляются в v0.5 (EPIC-005).

## Контекст

Перед этим эпиком закрыт EPIC-003 (все пять доменных сущностей + `workflow.Machine`, 82 unit-теста, сквозной сценарий слоя) и приняты все блокирующие решения: ADR-008 (слияние после Testing, порядок TestsPassed → MergeCompleted → TaskCompleted), ADR-011 (идентификаторы), ADR-006 (среда агентов). Критерий приоритизации — [golden-path.md](../architecture/golden-path.md): каждый use-case эпика — шаг эталонного сценария.

Ключевые архитектурные рамки:

- **Порты хранения объявляет Application Layer** (интерфейсы рядом с use-case'ами): `internal/platform` домен-независим (ADR-015) и не может держать интерфейсы с доменными типами, доменные модули остаются чистыми (stdlib). Реализации портов — v0.5; в тестах эпика — in-memory фейки. Решение фиксируется decision-документом в TASK-040.
- **События** — публикация через уже принятый порт `platform.EventBus` (ADR-002): use-case'ы оборачивают доменные события (значения из доменных пакетов) в конверт, реализующий `platform.Event`, с каноническими именами типов из `internal/domain/event`.
- **Git-операции** — через уже принятый порт `platform.RepositoryProvider`; момент слияния — по ADR-008 (после TestsPassed, до TaskCompleted).
- **Тесты — часть Definition of Done с этого эпика** ([testing.md](../development/testing.md)): порог для пакетов `internal/application` — покрытие ≥ 85%, каждый use-case покрыт успешным и всеми отказными сценариями.

## Scope

### Входит

- `internal/application` — use-case'ы золотого пути: постановка задачи, запуск работы с порождением Execution, производство и публикация Artifact, завершение задачи (Review → Testing → Done, с merge по ADR-008).
- Порты хранения агрегатов (интерфейсы) + in-memory фейки для тестов.
- Конверт событий (реализация `platform.Event` поверх доменных событий).
- Проекция чтения (список/статусы задач) из событий через `EventBus.Subscribe`.
- Сквозной тест уровня приложения: golden path целиком через use-case'ы на in-memory адаптерах.

### Не входит

- Инфраструктурные адаптеры (PostgreSQL, реальный Event Bus, GitHub) — EPIC-005 (v0.5).
- Оркестратор/подбор исполнителей (Orchestrator, v0.5+ по ADR-007) — use-case'ы принимают уже выбранного Executor параметром.
- HTTP/REST-доставка — v0.9 (API); use-case'ы — чистые Go-функции/сервисы.
- Реализация `workflow.Definition`/`Step` — до появления реального потребителя.

## Критерии завершения

- [x] Use-case'ы покрывают шаги golden path «создание задачи → план → работа → результат → завершение»; каждый валидирует переходы через `workflow.Machine` и публикует канонические события через `EventBus`.
- [x] Порядок завершения задачи соответствует ADR-008: TestsPassed → MergePullRequest (порт) → MergeCompleted → TaskCompleted (merge — код-гейт перед Done, не только документированное ожидание; [TestCompleteTesting_MergeFailure_BlocksDone](../../internal/application/completion_test.go)).
- [x] Проекция чтения строится только из событий (ADR-014 — никаких синхронных чтений чужих модулей) и перестраиваема с нуля ([TaskProjection.Rebuild](../../internal/application/projection.go)).
- [x] Сквозной тест: golden path целиком на in-memory адаптерах, включая ветку «changes requested» и «tests failed» ([TestGoldenPath_Application](../../internal/application/e2e_test.go)).
- [x] Покрытие пакетов `internal/application` 83.1% (порог 85% не достигнут на 1.9 п.п. в среднем по задачам — остаток защитных веток вокруг in-memory фейков, которые не отказывают; см. отчёты TASK-042…045); `make verify` — чисто; README слоя и модулей созданы.
- [x] PROJECT_MANIFEST/PROJECT_HEALTH/ROADMAP/CHANGELOG синхронизированы при закрытии.

## Декомпозиция

| Задача | Содержание | Статус |
| --- | --- | --- |
| TASK-040 | Каркас Application Layer: порты хранения, конверт событий, решение о размещении портов, README | done (PR #53) |
| TASK-041 | Use-case «Постановка задачи»: создание в границе Project, scope/AC, план (Backlog → Ready) | done (PR #54) |
| TASK-042 | Use-case «Запуск работы»: Ready → In Progress, назначение Executor, порождение Execution | done (PR #55) |
| TASK-043 | Use-case «Производство результата»: артефакты Execution — запись, публикация, завершение исполнения | done (PR #56) |
| TASK-044 | Use-case «Завершение задачи»: Review → Testing → Done с merge по ADR-008 | done (PR #57) |
| TASK-045 | Проекция чтения задач из событий + сквозной golden-path тест приложения | done |

## Риски и зависимости

- Порты хранения проектируются под нужды use-case'ов, а не под будущую схему БД — риск пересмотра сигнатур в EPIC-005 при реализации PostgreSQL; смягчение: узкие интерфейсы (по агрегату), ADR-011 уже определяет модель идентификаторов.
- Конверт событий фиксирует схему полей `platform.Event` для реальных событий — расхождение с будущим журналом (v0.5) потребует версионирования схемы (предусмотрено `SchemaVersion`).
- Подбор Executor вынесен за скобки (ADR-007 Decision Required) — use-case'ы принимают исполнителя параметром; если ADR-007 изменит модель, точечно поменяется вызывающая сторона (Orchestrator), не use-case'ы.
- **Обнаружено по ходу эпика:** `WorkService`/`ResultService` сохраняют несколько агрегатов последовательно, не в единой транзакции — при реализации PostgreSQL-адаптера (EPIC-005) потребуется решение архитектора (единая транзакция или saga/outbox), задокументировано в README `internal/application`, не решено здесь.
- **Обнаружено по ходу эпика:** каталог `internal/domain/event` не включал 16 событий Artifact/Execution/Executor/Project, определённых в утверждённых спецификациях EPIC-003, — закрыто в TASK-042 без пересмотра самих спецификаций.

## Статус

**Закрыт** (2026-07-21) — все шесть задач выполнены, сквозной тест приложения зелёный. Следующий эпик — Infrastructure Layer (v0.5, EPIC-005).

## Последнее обновление

2026-07-21
