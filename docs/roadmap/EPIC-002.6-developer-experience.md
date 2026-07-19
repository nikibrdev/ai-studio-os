# EPIC-002.6: Developer Experience (DX)

## Цель

Подготовить среду разработки так, чтобы новый участник — человек или AI-агент — получал рабочее окружение по принципу «клонируй → одна команда → проект готов», а качество контролировалось ещё до коммита.

## Контекст

Предложен архитектором проекта 2026-07-19 как обязательный шаг перед Domain Layer. Опирается на готовые `make verify`, `.golangci.yml`, `scripts/verify-docs.sh`. Выполняется после EPIC-002.5 (нужен репозиторий на GitHub).

## Scope

### Входит

1. Pre-commit хуки: автоматический запуск `make verify` перед коммитом.
2. `.editorconfig` — единые редакторские настройки.
3. `.vscode/` — рекомендуемые расширения (Go, EditorConfig, markdownlint, Mermaid) и задачи (verify, build).
4. Devcontainer — единое окружение (вопрос ниже).
5. `docs/development/onboarding.md` — инструкция нового разработчика + установочный скрипт (`scripts/setup`).

### Не входит

- Domain Layer и прикладной код; CI (это EPIC-002.5); деплой.

## Декомпозиция

| Задача | Содержание |
|---|---|
| TASK-021 | Git-хуки по уровням ([решение](../../engineering/decisions/2026-07-19-quality-gates.md)): pre-commit — gofumpt → golangci-lint → go vet; pre-push — полный `make verify`; установка хуков через setup-скрипт |
| TASK-022 | `.editorconfig` |
| TASK-023 | `.vscode/extensions.json` + `.vscode/tasks.json` |
| TASK-024 | Devcontainer — минимальный: Go 1.24, Node LTS, pnpm, golangci-lint, gofumpt, markdownlint, Git, Make; без PostgreSQL/Redis/Qdrant (инфраструктура — отдельно через Docker Compose) |
| TASK-025 | `scripts/setup` + `docs/development/onboarding.md` («клонируй → одна команда → готов») |

## Решения по вопросам (приняты 2026-07-19)

1. Devcontainer — **делаем сразу**, минимальный (состав в TASK-024); отвечает только за среду разработки.
2. Pre-commit — **быстрый набор** (fmt + lint + vet, секунды); полный `make verify` — на pre-push и обязательно в GitHub Actions, без исключений.

## Критерии завершения

- [ ] Новый разработчик получает рабочее окружение одной командой после клонирования.
- [ ] Коммит с падающими проверками невозможен локально (хук) и в CI (verify).
- [ ] Все конфиги описаны в onboarding-инструкции.

## Статус

Ожидает утверждения плана (после EPIC-002.5)

## Последнее обновление

2026-07-19
