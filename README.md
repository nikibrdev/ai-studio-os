# AI Studio OS

## Назначение

**AI Studio OS** — открытая платформа для управления разработкой программного обеспечения, в которой роли команды могут выполняться как людьми, так и AI-агентами.

Платформа организует полный цикл разработки — от постановки задачи до проверки качества — и позволяет гибко распределять роли между людьми и агентами без изменения процессов.

Инженерное видение продукта на горизонтах 1/2/5 лет — [VISION.md](VISION.md).

## Содержание

### Цели проекта

- Создать единый процесс разработки, одинаково пригодный для людей и AI-агентов.
- Сделать AI-агентов полноценными участниками команды с понятными ролями и зонами ответственности.
- Обеспечить независимость ядра платформы от конкретных AI-моделей и провайдеров.
- Предоставить прозрачный, документируемый и воспроизводимый процесс разработки.

### Возможности (MVP)

Четыре роли исполняются агентами, две точки процесса остаются за человеком:

| Роль | Что делает | Кто решает |
| --- | --- | --- |
| Project Manager | Уточняет цель и критерии приёмки задачи | **Человек** принимает Definition of Ready |
| Developer | Пишет код в изолированной песочнице, открывает Pull Request | — |
| Reviewer | Изучает изменения, выносит вердикт ревью | Агент ([ADR-008](docs/adr/ADR-008-git-policies.md) это допускает) |
| QA Engineer | Проверяет изменения, готовит отчёт | **Человек** принимает приёмочное решение |

Все роли исполняет один и тот же адаптер и один Docker-образ — различие только в промпте ([ADR-007](docs/adr/ADR-007-pm-qa-executors.md)). Агенты **не имеют доступа к API платформы**: результат передаётся файлом, применяет его сама платформа — поэтому контрольные точки человека недоступны агенту структурно, а не по договорённости.

По умолчанию используется Claude Code, но архитектура независима от конкретной модели: OpenAI Codex, OpenHands, Gemini и другие агенты подключаются через [контракт Executor](docs/adr/ADR-005-executor-contract.md) без изменения ядра.

### Текущий статус

**Платформа доводит задачу до Pull Request'а сама; за человеком остаются две контрольные точки.**

Реализованы все слои — домен, сценарии использования, инфраструктура (PostgreSQL, GitHub, Qdrant), исполнение агентов в изолированных Docker-контейнерах, память, REST API и веб-интерфейс. Оркестратор реагирует на события жизненного цикла задачи и запускает исполнителей сам.

**Что человек решает лично:** приём Definition of Ready и финальное приёмочное решение (оно сливает Pull Request). Это не незавершённость, а [принятое решение](docs/adr/ADR-007-pm-qa-executors.md): агенты **готовят** — человек **принимает**. Нормативное описание — [workflow.md](docs/architecture/workflow.md#контрольные-точки-человека).

**Честные ограничения:** аутентификации нет (доверенная однопользовательская установка, [ADR-012](docs/adr/ADR-012-identity-and-auth.md)); слой инструментов (`tools/`) не реализован; содержательное качество работы агентов не проверялось вживую — нет ключа AI-провайдера, проверена вся механика и все отказные пути.

Актуальное состояние одним взглядом: **[PROJECT_MANIFEST.md](PROJECT_MANIFEST.md)** (паспорт), **[PROJECT_HEALTH.md](PROJECT_HEALTH.md)** (прогресс), **[индекс решений](docs/adr/DECISIONS_INDEX.md)**.

### Архитектурные принципы

- Clean Architecture
- SOLID
- KISS
- DRY
- Interface First
- Documentation First
- Event-Driven Architecture
- Modular Monolith (на первом этапе)
- Расширяемость без изменения ядра
- Минимум магии
- Максимальная читаемость

Подробнее: [docs/architecture/overview.md](docs/architecture/overview.md) и [CONSTITUTION.md](CONSTITUTION.md).

### Запуск

Нужны Go 1.24, Node.js 22 LTS, pnpm и Docker ([CONTRIBUTING.md](CONTRIBUTING.md) — настройка окружения разработки, включая Dev Container).

1. **Хранилища.** PostgreSQL и Qdrant поднимаются готовым compose-файлом ([docker-compose.yml](docker-compose.yml), только для разработки — учётные данные намеренно простые):

   ```bash
   docker compose up -d
   ```

2. **Образ песочницы исполнения** — нужен только для запуска агентов (шаг 5); без него API и Dashboard работают:

   ```bash
   docker build -t ai-studio-os-execution -f docker/execution/Dockerfile .
   ```

3. **API** — применяет миграции при старте, отвечает на `http://localhost:8080` ([apps/api/README.md](apps/api/README.md)):

   ```bash
   export DATABASE_URL="postgres://ai_studio_os:ai_studio_os@localhost:5432/ai_studio_os?sslmode=disable"
   export QDRANT_URL="http://localhost:6333"   # опционально: без него память отключена
   go run ./apps/api
   ```

4. **Dashboard** — `http://localhost:3000`, требует запущенного API ([apps/dashboard/README.md](apps/dashboard/README.md)):

   ```bash
   cd apps/dashboard && pnpm install && pnpm dev
   ```

5. **Orchestrator** — запускает исполнителей; **`GITHUB_TOKEN` обязателен**, без него процесс завершается с ошибкой ([apps/orchestrator/README.md](apps/orchestrator/README.md)):

   ```bash
   export DATABASE_URL="postgres://ai_studio_os:ai_studio_os@localhost:5432/ai_studio_os?sslmode=disable"
   export GITHUB_TOKEN="<токен с доступом к управляемому репозиторию>"
   export ANTHROPIC_API_KEY="<ключ>"   # пустое значение допустимо: песочница запустится, вызова провайдера не будет
   go run ./apps/orchestrator
   ```

Порядок операций для первого проекта (создать → подключить репозиторий → активировать → поставить задачу) описан в [docs/api/projects.md](docs/api/projects.md) и [docs/api/tasks.md](docs/api/tasks.md); что ждёт решения человека — `GET /decisions` и раздел «Ждут решения» в Dashboard.

### Технологический стек

| Слой | Технология |
| --- | --- |
| Backend | Go |
| Frontend | Next.js |
| База данных | PostgreSQL (источник истины, журнал событий) |
| Векторный поиск | Qdrant (память агентов) |
| Кэш | Redis — **не используется**; предусмотрен как возможная шина событий ([ADR-002](docs/adr/ADR-002-event-delivery.md)) |
| Тестирование | Go testing, Vitest, Playwright |
| Git-хостинг | GitHub |
| Контейнеры | Docker |
| AI Developer | Claude Code |

### Структура проекта

```
ai-studio-os/
├── .github/          # Шаблоны PR/Issue, workflows CI
├── .claude/          # Инструкции, роли, команды и контекст для AI-агентов
├── apps/
│   ├── api/          # Backend API (Go)
│   ├── dashboard/    # Веб-интерфейс (Next.js)
│   └── orchestrator/ # Оркестрация ролей и процессов
├── internal/         # Внутренние пакеты ядра
├── pkg/              # Переиспользуемые публичные пакеты
├── agents/           # Определения и адаптеры AI-агентов
├── tools/            # Слой инструментов (зарезервирован, не реализован)
├── memory/           # Память агентов и знания проекта
├── projects/         # Зарезервирован; метаданные проектов — в PostgreSQL (ADR-013)
├── tasks/            # Файловая система задач (жизненный цикл)
├── docs/             # Документация (архитектура, ADR, процессы)
├── docker/           # Docker-конфигурация (образ песочницы исполнения)
├── scripts/          # Служебные скрипты
└── examples/         # Примеры использования
```

Подробнее: [docs/architecture/project-structure.md](docs/architecture/project-structure.md).

### Roadmap

Источник истины — **[ROADMAP.md](ROADMAP.md)**; здесь только ориентир.

| Версия | Название | Состояние |
| --- | --- | --- |
| v0.1 | Foundation | Завершено |
| v0.2 | Architecture & Engineering Platform | Завершено |
| v0.3 | Domain Layer | Завершено |
| v0.4 | Application Layer | Завершено |
| v0.5 | Infrastructure Layer | Завершено |
| v0.6 | AI Agent Runtime | Завершено |
| v0.7 | Memory System | Завершено |
| v0.8 | Dashboard | Завершено |
| v0.9 | API Layer | Завершено |
| v1.0 | First Public MVP | В работе — стабилизация и релиз |

Нумерация версий пересмотрена 2026-07-20: прежняя описывала стек ролей, нынешняя — готовность продукта по слоям ([решение](engineering/decisions/2026-07-20-release-milestones.md)).

### Лицензия

Проект распространяется под лицензией **Apache License 2.0** — см. [LICENSE](LICENSE) и [ADR-001](docs/adr/ADR-001-license.md). Контрибуции принимаются на условиях той же лицензии.

## Статус

Актуален

## Последнее обновление

2026-07-27
