# EPIC-005: Infrastructure Layer — инфраструктура

## Цель

Реализовать Infrastructure Layer AI Studio OS (v0.5, [ROADMAP.md](../../ROADMAP.md)): адаптеры портов `internal/platform`/`internal/application` к реальным технологиям — PostgreSQL (источник истины задач, [ADR-004](../adr/ADR-004-task-storage.md)), Event Bus с журналом событий в PostgreSQL ([ADR-002](../adr/ADR-002-event-delivery.md)), GitHub Repository Provider. Результат по ROADMAP: «платформа работает end-to-end на реальных хранилищах и интеграциях» — те же use-case'ы EPIC-004, без единой строки изменений в `internal/application`/`internal/domain`, впервые исполняются не на in-memory фейках.

## Контекст

Перед этим эпиком закрыт EPIC-004 (четыре use-case-сервиса золотого пути + проекция чтения, 136 unit-тестов, сквозной тест приложения на in-memory адаптерах). Порты (`application.ProjectStore/TaskStore/ExecutorStore/ExecutionStore/ArtifactStore`, `platform.EventBus`, `platform.RepositoryProvider`) уже объявлены и обкатаны тестовыми фейками — этот эпик их не меняет, только реализует.

Ключевые архитектурные рамки:

- **Драйвер PostgreSQL** — `pgx/v5`, миграции — самописный раннер по встроенным `.sql` ([ADR-017](../adr/ADR-017-postgresql-driver.md), принят при открытии этого эпика).
- **Хранение агрегатов не транзакционно между собой** — известное ограничение, зафиксированное при закрытии EPIC-004 ([README](../../internal/application/README.md), «Известное ограничение»): `Store.Get/Save` — по одному агрегату, без сквозной транзакции между несколькими `Save` в одном use-case. Это ограничение эпик **не устраняет** — адаптеры реализуют контракт `Store[T]` как есть (одна операция — одна транзакция БД на уровне одной таблицы); вопрос единой транзакции/saga через несколько агрегатов — решение архитектора при появлении реальной необходимости, не блокирует этот эпик.
- **Event Bus MVP — по-прежнему in-process** ([ADR-002](../adr/ADR-002-event-delivery.md)): реализация этого эпика не Redis/NATS, а производственная (не тестовая) версия синхронной внутрипроцессной шины плюс подписчик-журнал, сохраняющий события в PostgreSQL. Интерфейс `platform.EventBus` не меняется.
- **Repository Provider — только GitHub**, идентификатор репозитория — строка `owner/name` (уже зафиксировано текущей сигнатурой `platform.RepositoryProvider`, ADR-013 о подключении управляемых проектов остаётся Decision Required и не блокирует этот узкий контракт — эпик не решает вопрос многорепозиторных/managed-проектов).
- **Интеграционные тесты — с реальными зависимостями в контейнерах** ([testing.md](../development/testing.md), правило 5): Docker Compose с PostgreSQL для локального запуска; в CI — сервис-контейнер PostgreSQL в отдельном job. Тесты за build-тегом `integration`, пропускаются при отсутствии `TEST_DATABASE_URL` — `make verify`/`go test ./...` остаются independent от Docker.

## Scope

### Входит

- `internal/infrastructure` — новый слой: подключение к PostgreSQL (`pgxpool`), раннер миграций, схема БД для пяти агрегатов и журнала событий.
- Postgres-адаптеры пяти портов хранения (`ProjectStore`, `TaskStore`, `ExecutorStore`, `ExecutionStore`, `ArtifactStore`), реализующие контракты `internal/application/ports.go` без их изменения.
- Производственная реализация `platform.EventBus` (in-process) + подписчик-журнал, сохраняющий каждое опубликованное событие в PostgreSQL.
- `platform.RepositoryProvider` — адаптер к GitHub REST API (`net/http`, без новых зависимостей, аутентификация — токен из окружения).
- Composition root: точка сборки всех адаптеров в рабочую систему (для интеграционных тестов и будущего API-слоя v0.9).
- Docker Compose для локального PostgreSQL; интеграционные тесты золотого пути на реальной БД.

### Не входит

- Redis (кэш, не используется для доставки событий в MVP согласно ADR-002 — вводится по отдельной задаче/эпику при реальной потребности).
- HTTP/REST-доставка (v0.9, API) — composition root не поднимает HTTP-сервер, только собирает зависимости для тестов и будущего использования.
- Docker-среда исполнения агентов (ADR-006) — это про запуск Executor'ов, не про инфраструктуру хранения; v0.6.
- Managed Projects / многорепозиторное подключение (ADR-013) — GitHub-адаптер работает с одним репозиторием по фиксированному идентификатору `owner/name`.
- Down-миграции, blue-green, репликация — вне зоны MVP.

## Критерии завершения

- [ ] Пять портов хранения реализованы на PostgreSQL, покрывают Get/Save/ErrNotFound так же, как in-memory фейки (те же контрактные тесты, где применимо).
- [ ] Событие, опубликованное через производственный `EventBus`, доставляется подписчикам синхронно (как и раньше) и сохраняется в журнал PostgreSQL — журнал восстановим отдельным select.
- [ ] `RepositoryProvider` выполняет реальные операции с GitHub (CreateBranch/OpenPullRequest/RequestReview/MergePullRequest/ClosePullRequest/PullRequestState) на тестовом репозитории.
- [ ] Golden path (`TestGoldenPath_Application` эквивалент) проходит на реальных адаптерах в интеграционном тесте с PostgreSQL в контейнере — `internal/application` не изменён ни на строку.
- [ ] `make verify` — чисто; интеграционные тесты — отдельный CI job с сервис-контейнером PostgreSQL, не блокирующий обычный `make verify` при отсутствии БД.
- [ ] PROJECT_MANIFEST/PROJECT_HEALTH/ROADMAP/CHANGELOG синхронизированы при закрытии.

## Декомпозиция

| Задача | Содержание | Статус |
| --- | --- | --- |
| TASK-046 | Каркас `internal/infrastructure`: конфигурация подключения, `pgxpool`, раннер миграций, Docker Compose, README | done |
| TASK-047 | Postgres-адаптеры `ProjectStore` + `TaskStore` (источник истины задач по ADR-004), схема таблиц | done |
| TASK-048 | Postgres-адаптеры `ExecutorStore` + `ExecutionStore` + `ArtifactStore`, схема таблиц | done |
| TASK-049 | Производственный `EventBus` (in-process) + подписчик-журнал в PostgreSQL | ready |
| TASK-050 | `RepositoryProvider` — адаптер GitHub REST API | ready |
| TASK-051 | Composition root, интеграционный golden-path тест на реальной БД, CI job, закрытие эпика | ready |

## Риски и зависимости

- ~~Docker недоступен на машине разработчика в моменте~~ — снято в TASK-048: Docker Desktop запущен, `docker compose up -d` + все интеграционные тесты (раннер миграций, пять Store) прогнаны вживую против настоящего PostgreSQL и зелёные. Автоматизация в CI (сервис-контейнер, независимый от локальной машины) — по-прежнему TASK-051.
- **GitHub-адаптер тестируется против реального API** — нужен тестовый репозиторий и токен с правами на него; вноситься в CI как секрет не будет (не входит в scope без отдельного решения человека о хранении секрета) — TASK-050 покрывается unit-тестами на HTTP-уровне (httptest) плюс ручная проверка автором задачи; интеграционный прогон против настоящего GitHub — вне автоматического CI.
- Зависимость на ADR-017 (принят при открытии эпика) — снята.
- Наследуется от EPIC-004: отсутствие сквозной транзакции между агрегатами — не решается здесь (см. «Контекст»).

## Статус

В работе

## Последнее обновление

2026-07-21
