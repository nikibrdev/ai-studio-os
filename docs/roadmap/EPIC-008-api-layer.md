# EPIC-008: API Layer — REST-слой платформы

## Цель

Реализовать `apps/api` (v0.9, [ROADMAP.md](../../ROADMAP.md)): тонкий REST-слой доставки ([ADR-003](../adr/ADR-003-api-protocol.md)) над уже готовым Application Layer (EPIC-004) — без собственной бизнес-логики, без auth в первой версии ([ADR-012](../adr/ADR-012-identity-and-auth.md), Вариант 1). Результат по ROADMAP: «платформа доступна внешним клиентам, не только собственному UI; Dashboard (v0.8) становится возможным».

## Контекст

Строится **до** v0.8 Dashboard, вопреки номеру версии — `apps/dashboard` может обращаться к платформе только через `apps/api` ([module-boundaries.md](../architecture/module-boundaries.md)); порядок и причина зафиксированы отдельно ([decision](../../engineering/decisions/2026-07-22-api-before-dashboard-build-order.md)).

При открытии эпика найдены и решены два блокера, не позволявшие построить полезный, реально работающий API поверх уже существующего Application Layer:

- **Не было способа создать Project** — `internal/application` не содержит use-case для создания/активации Project (только четыре сервиса золотого пути, начинающегося УЖЕ внутри Active-проекта); тесты создают Project напрямую через `internal/domain/project`, минуя Application Layer. Решение архитектора: добавить минимальный `ProjectService` (`CreateProject`/`Activate`) в этом эпике — то же обоснование, что и остальные use-case-сервисы EPIC-004 (узкий порт хранения, события через `platform.EventBus`).
- **Не было генерации последовательного `TASK-NNN` на проект** — [ADR-011](../adr/ADR-011-task-identifiers.md) требовал этого в модели данных `task` при реализации PostgreSQL-адаптера (v0.5, EPIC-005), но реализация выдачи номера не была сделана: `CreateTaskParams.ID` — поле, которое передаёт вызывающий код. Внутри проекта (тесты, golden path) это работало, потому что вызывающий код и так знает следующий номер. Внешний клиент API вычислить следующий номер безопасно не может (гонка параллельных запросов — именно то, что ADR-011 поручал решить единому пути записи). Решение архитектора: реализовать генерацию последовательности в этом эпике, как и предписывал ADR-011.

Оба решения — точечные, ограниченные дополнения к уже закрытым слоям (Application Layer, PostgreSQL-адаптер), обоснованные тем, что без них API не может выполнить свою заявленную роль (ROADMAP: «платформа доступна внешним клиентам»), а не расширение их scope «заодно».

**Аутентификация — ADR-012 (принят при открытии эпика), Вариант 1**: доверенная однопользовательская установка, без auth в запросах API. Пересмотр — при появлении первого реального внешнего потребителя или многопользовательского сценария.

## Scope

### Входит

- `internal/application/project.go` — `ProjectService.CreateProject`/`Activate` (по образцу уже существующих сервисов EPIC-004: узкий порт `ProjectStore`, события через `platform.EventBus`).
- Генерация последовательного `TASK-NNN` на проект (ADR-011) — точечное дополнение `internal/infrastructure/postgres` (счётчик на проект, атомарный `UPDATE ... RETURNING` — не нативная PostgreSQL `SEQUENCE`, так как её нельзя завести динамически на проект без DDL при каждом новом проекте).
- `docs/api/*.md` — спецификации по ресурсам (Projects, Tasks, Artifacts, Executions) по шаблону [API.md](../../.claude/templates/API.md), **Documentation First** — до реализации хендлеров (тот же принцип, что Domain Specifications First, EPIC-003, приведён к масштабу этого слоя).
- `apps/api` — HTTP-сервер на стандартном `net/http` (`ServeMux` с маршрутизацией по методу и пути, Go 1.24 — без сторонней библиотеки-роутера, тот же принцип «не добавлять зависимость без необходимости», что и REST-клиенты GitHub/Qdrant): хендлеры вызывают use-case-сервисы `internal/application`, зависимости собираются через `internal/infrastructure/wiring.System`; JSON-кодирование запросов/ответов; единообразное отображение ошибок домена/приложения в HTTP-коды.
- `apps/api/README.md`.
- Интеграционный тест: сквозной сценарий golden path через реальные HTTP-вызовы `apps/api` на настоящем PostgreSQL (тот же принцип, что `TestGoldenPath_Infrastructure`, EPIC-005, но через HTTP, а не прямые вызовы сервисов).
- Закрытие эпика: критерии, ROADMAP (v0.9 — Завершено), PROJECT_MANIFEST, PROJECT_HEALTH, CHANGELOG.

### Не входит

- Аутентификация/авторизация запросов — отложены (ADR-012, Вариант 1).
- Дашборд (`apps/dashboard`) — отдельный эпик (v0.8), после этого.
- Канал доставки событий в реальном времени (SSE/WebSocket) — по [ADR-003](../adr/ADR-003-api-protocol.md) проектируется вместе с Dashboard, не здесь.
- Генерация OpenAPI-документа и клиентских типов из него — `docs/api/*.md` остаются исполняемой Markdown-спецификацией (тот же формат, что API.md уже определяет); инструментарий генерации OpenAPI не входит в `stack.md` и не вводится без отдельного решения.
- Дополнительные проекции чтения сверх уже существующей `TaskProjection` (например, список задач/проектов) — `TaskProjection.Get` сейчас отдаёт только одну запись по ID; список — решение с реальной потребностью Dashboard, не раньше.
- Роли PM/QA как исполнители (ADR-007, отдельный Decision Required, не блокирует этот эпик).

## Критерии завершения

- [x] `ProjectService` (CreateProject/Activate) реализован и покрыт тестами по образцу остальных сервисов Application Layer — TASK-064.
- [x] Последовательный `TASK-NNN` на проект генерируется платформой (не вызывающим кодом), проверено тестом на конкурентный вызов — TASK-065, [postgres/task_id_generator_integration_test.go](../../internal/infrastructure/postgres/task_id_generator_integration_test.go).
- [x] `docs/api/*.md` — спецификации всех операций по шаблону API.md, написаны до реализации соответствующих хендлеров — TASK-066.
- [x] `apps/api` реализует весь golden path через HTTP: создание проекта → подключение репозитория → активация → создание и планирование задачи → запуск работы → черновик и публикация артефакта → результат исполнения → ревью → тестирование → завершение — TASK-067…069.
- [x] Сквозной HTTP-сценарий подтверждён вживую на реальном PostgreSQL (не только на фейках) — TASK-070, [golden_path_integration_test.go](../../apps/api/httpapi/golden_path_integration_test.go).
- [x] `docs/architecture/*.md`, затронутые эпиком (module-boundaries.md, system-design.md при необходимости), синхронизированы — TASK-071.
- [x] PROJECT_MANIFEST/PROJECT_HEALTH/ROADMAP/CHANGELOG синхронизированы при закрытии — TASK-071.

## Декомпозиция

| Задача | Содержание | Статус |
| --- | --- | --- |
| TASK-064 | `ProjectService` (CreateProject/Activate) в `internal/application` | done |
| TASK-065 | Генерация последовательного `TASK-NNN` на проект (ADR-011) в `internal/infrastructure/postgres` | done |
| TASK-066 | `docs/api/*.md` — спецификации по ресурсам (Documentation First) | done |
| TASK-067 | Каркас `apps/api`: `main.go`, сборка через `wiring.System`, маршрутизация, JSON-хелперы, отображение ошибок в HTTP-коды | done |
| TASK-068 | Хендлеры Projects/Tasks (CreateProject/Activate, CreateTask/PlanTask, чтение через TaskProjection) | done |
| TASK-069 | Хендлеры Work/Result/Completion (StartTask, черновик/публикация Artifact, Succeed/FailExecution, Review/Testing) | done |
| TASK-070 | Сквозной интеграционный тест golden path через реальные HTTP-вызовы на настоящем PostgreSQL | done |
| TASK-071 | README `apps/api`, синхронизация документации, закрытие эпика | ready |

## Риски и зависимости

- **Два точечных дополнения к закрытым эпикам** (ProjectService в Application Layer, ID-последовательность в Infrastructure Layer) — решение архитектора при открытии этого эпика, не самовольное расширение scope; обоснование — раздел «Контекст» выше.
- **Без auth** (ADR-012, Вариант 1) — API пригоден для доверенной установки, не для публичного доступа без дополнительного периметра (например, обратный прокси с собственной аутентификацией) — явно принятое ограничение, не скрытое.
- **Известное ограничение EPIC-004/005 наследуется**: отсутствие сквозной транзакции между агрегатами (`WorkService.StartTask`, `ResultService.RecordDraftArtifact`) теперь также видно снаружи через HTTP — не решается в этом эпике.
- **`TaskProjection.Get` — только одна запись по (project, id)**, списковых операций (все задачи проекта и т.п.) нет — самый заметный практический предел этой версии API для будущего Dashboard; расширение проекций — решение по реальной потребности, не здесь.
- **Реализованный риск, материализовался и исправлен**: живая проверка TASK-069 вскрыла, что `TASK-NNN` неуникален глобально между проектами (`tasks.id` был единственным `PRIMARY KEY`, без `project_id`) — исправлено по всему стеку в [BUGFIX-003](../../tasks/done/BUGFIX-003-task-project-scoped-key.md) в этом же эпике, до его закрытия.

## Статус

Закрыт

## Последнее обновление

2026-07-22
