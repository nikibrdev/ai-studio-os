# TASK-012: Инициализация Git и GitHub-репозитория

## Тип

chore

## Эпик

EPIC-002.5 ([docs/roadmap/EPIC-002.5-engineering-platform.md](../../docs/roadmap/EPIC-002.5-engineering-platform.md))

## Цель

Проект в git с историей на GitHub: `git init`, первый коммит, публичный репозиторий `ai-studio-os` на личном аккаунте мейнтейнера, push `main`.

## Контекст

Решения зафиксированы: [engineering/decisions/2026-07-19-github-repository.md](../../engineering/decisions/2026-07-19-github-repository.md). Порядок команд задан архитектором дословно.

## Scope

### Входит

- `git init`, ветка `main`, первый коммит `chore: bootstrap AI Studio OS`.
- `git remote add origin <URL>`, `git push -u origin main` — после создания репозитория мейнтейнером.

### Не входит

- Защита ветки, Actions, CODEOWNERS (TASK-013…016); переименование Go-модуля (отдельная задача после push).

## Критерии приёмки

- [ ] Локальный репозиторий: ветка `main`, один коммит со всем содержимым проекта.
- [ ] Репозиторий на GitHub, push выполнен, `main` отслеживает origin.

## План реализации

1. `git init` → `git branch -M main`.
2. Проверить git-конфигурацию (user.name/user.email).
3. `git add .` → `git commit -m "chore: bootstrap AI Studio OS"`.
4. После получения URL от мейнтейнера: `git remote add origin <URL>` → `git push -u origin main`.

План соответствует командам, заданным архитектором (утверждено постановкой).

## Затрагиваемые модули и документы

- Корень репозитория (git-метаданные).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-19 — Architect — задача поставлена (EPIC-002.5), порядок команд задан, план утверждён постановкой.
2026-07-19 — Claude Code (Developer) — выполнена локальная часть: git init, main, первый коммит.
2026-07-19 — Claude Code (Developer) — переведена в blocked: ожидается создание пустого публичного репозитория `ai-studio-os` мейнтейнером в веб-интерфейсе GitHub и URL для remote/push.

## Отчёт о выполнении

(будет дополнен после push)
