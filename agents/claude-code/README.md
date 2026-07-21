# agents/claude-code — адаптер Claude Code

## Назначение

Первый реальный адаптер `platform.Executor` ([ADR-005](../../docs/adr/ADR-005-executor-contract.md)): исполняет роль Developer через Claude Code, запущенный внутри изолированного Docker-контейнера ([ADR-006](../../docs/adr/ADR-006-agent-execution-environment.md)).

## Содержание

### Структура

| Пакет | Содержимое | Задача |
| --- | --- | --- |
| `container/` | Жизненный цикл контейнера Execution: клон рабочей копии, сетевой allowlist, инъекция секретов | TASK-054 |
| (корень) | Реализация `platform.Executor` (Accept/Artifacts/Status/Finish) | TASK-055 |

### `container` — жизненный цикл Execution

`Manager` (TASK-054) управляет Docker-контейнером Execution через прямые вызовы `docker` CLI (`os/exec`) — без клиентской библиотеки Docker: `agents/` не может импортировать `internal/infrastructure`, а горстка команд (`run`, `network create`, `exec`, `rm`) не оправдывает новую зависимость (тот же принцип, что ADR-017 и GitHub-адаптер).

**Сетевой allowlist** — по ADR-006: две Docker-сети на Execution — внутренняя (`--internal`, без маршрута наружу) и обычный `bridge` (с интернетом). Контейнер исполнения подключён только к внутренней сети; прокси-контейнер (`ubuntu/squid` — существующий публичный образ, не собственная сборка, тот же принцип, что `postgres:16-alpine` для БД) подключён к обеим и разрешает `CONNECT`/HTTP только к доменам allowlist'а (по умолчанию — `github.com`/`api.github.com`, плюс явно переданные), остальное — `deny all`. У контейнера исполнения физически нет иного пути наружу, кроме как через прокси.

**Секреты** — git-токен и ключ AI-провайдера передаются только переменными окружения контейнера (`GIT_TOKEN`, `ANTHROPIC_API_KEY`). Токен не встраивается в URL клонирования и не передаётся флагом (оба способа оставляют секрет в списке процессов контейнера, видном через `docker inspect`/`ps`) — вместо этого генерируется `GIT_ASKPASS`-скрипт, читающий токен из переменной окружения `git`'ом самостоятельно.

**Рабочая копия** — клонируется командой `git clone --branch <branch>` внутри контейнера при старте; уничтожается вместе с контейнером при `Stop` (ADR-006: эфемерна на всё время Execution, не дольше).

### Тестирование

Юнит-тесты (`lifecycle_test.go`, `proxy_test.go`) не требуют реального Docker: пул команд сужен до интерфейса `commandRunner`, тесты подставляют фейк — проверяют построение команд, порядок операций, откат при ошибке, что секреты не попадают в аргументы команд.

Интеграционный тест (`lifecycle_integration_test.go`, тег `integration`) требует реальный Docker и собранный образ `docker/execution` (TASK-053); пропускается, если не задана `TEST_DOCKER` (тот же паттерн, что `TEST_DATABASE_URL` в `internal/infrastructure` — тест создаёт настоящие Docker-ресурсы, включение по явному согласию):

```bash
docker build -t ai-studio-os-execution -f docker/execution/Dockerfile .
export TEST_DOCKER=1
go test -tags=integration ./agents/claude-code/container/...
```

Прогнан вживую (TASK-054): реальное клонирование `nikibrdev/ai-studio-os`, обращение к `api.github.com` (в allowlist'е — успех) и к `example.com` (не в allowlist'е — заблокировано), три прогона подряд стабильны.

## Статус

В работе (EPIC-006)

## Последнее обновление

2026-07-21
