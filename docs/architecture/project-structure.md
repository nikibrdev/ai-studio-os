# Структура проекта

## Назначение

Полное описание структуры репозитория: назначение каждого каталога и правила его использования. Изменение структуры допускается только через ADR.

## Содержание

### Дерево репозитория

```
ai-studio-os/
├── .github/
│   ├── ISSUE_TEMPLATE/          # Шаблоны issue (наполняются позже)
│   ├── workflows/               # CI/CD (запрещено на этапе Foundation)
│   └── PULL_REQUEST_TEMPLATE.md # Шаблон Pull Request
├── .claude/
│   ├── agents/                  # Описания ролей агентов
│   ├── commands/                # Шаблоны команд для агентов
│   ├── templates/               # Шаблоны документов (ADR, Task, Epic и др.)
│   ├── checklists/              # Чек-листы (Architecture, PR, QA, Release)
│   └── context/                 # Контекст проекта для агентов
├── apps/
│   ├── api/                     # Backend API (Go)
│   ├── dashboard/               # Веб-интерфейс (Next.js)
│   └── orchestrator/            # Оркестрация ролей и процессов
├── internal/                    # Ядро, закрыто от внешнего импорта (ADR-015)
│   ├── domain/                  # Предметная область
│   │   ├── shared/              # Язык домена: Role, TaskState, ...
│   │   └── ...                  # Модули: task, project, event, workflow, ...
│   ├── application/             # Application Layer (сценарии, проекции)
│   ├── platform/                # Абстракции платформы: EventBus, Agent, Tool, ...
│   └── infrastructure/          # Адаптеры (PostgreSQL, шина, GitHub, ...)
├── engineering/
│   ├── reviews/                 # Записи code review
│   ├── retrospective/           # Ретроспективы
│   ├── decisions/               # Процессные решения
│   └── metrics/                 # Метрики процесса
├── pkg/                         # Публичные переиспользуемые пакеты
├── agents/                      # Определения и адаптеры AI-агентов
├── tools/                       # Слой инструментов для агентов
├── memory/                      # Память агентов и знания проекта
├── projects/                    # Проекты, управляемые платформой
├── tasks/
│   ├── backlog/                 # Идеи и несформулированные задачи
│   ├── ready/                   # Задачи, готовые к работе (DoR выполнен)
│   ├── in-progress/             # Задачи в работе
│   ├── review/                  # Задачи на ревью
│   ├── blocked/                 # Заблокированные задачи
│   ├── done/                    # Выполненные задачи
│   └── archive/                 # Архив
├── docs/
│   ├── adr/                     # Architecture Decision Records
│   ├── architecture/            # Архитектурная документация
│   ├── api/                     # Документация API (placeholder)
│   ├── development/             # Процессы разработки
│   ├── operations/              # Эксплуатация (placeholder)
│   ├── roadmap/                 # Детализация roadmap, эпики
│   └── specifications/          # Полные спецификации модулей (ТЗ)
│       ├── domain/
│       ├── application/
│       ├── platform/
│       └── infrastructure/
├── docker/                      # Docker-конфигурация (наполняется позже)
├── scripts/                     # Служебные скрипты
├── examples/                    # Примеры использования
├── README.md
├── CLAUDE.md                    # Инструкция для Claude Code
├── CONSTITUTION.md              # Конституция проекта
├── ROADMAP.md
├── CHANGELOG.md
├── LICENSE                      # Apache License 2.0 (ADR-001)
├── Makefile
├── .gitattributes
└── .gitignore
```

### Правила

1. Новые каталоги верхнего уровня не создаются без ADR.
2. Пустые каталоги содержат `.gitkeep` до появления содержимого.
3. Код в `internal/` недоступен для импорта извне (соглашение Go).
4. `pkg/` не зависит от `internal/` и не содержит доменной логики.
5. Документация живёт в `docs/`; документы для агентов — в `.claude/`; правило разграничения: `docs/` описывает систему, `.claude/` управляет поведением агентов.

## Статус

Актуален

## Последнее обновление

2026-07-19
