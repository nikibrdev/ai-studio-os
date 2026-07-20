# Решение: уровни проверок качества и Dev Container

## Назначение

Фиксирует решения архитектора проекта (2026-07-19) по разделению проверок качества и составу Dev Container (уточнение scope EPIC-002.6).

## Содержание

### Уровни проверок

| Уровень | Что выполняется | Требование |
| --- | --- | --- |
| **pre-commit** | gofumpt → golangci-lint → go vet | Секунды; не раздражает |
| **pre-push** | `make verify` полностью (+ go test, markdownlint, Mermaid, ссылки) | Тяжёлые проверки перед отправкой |
| **GitHub Actions** | `make verify` полностью | **Обязательно, без исключений** |

Полный `make verify` на каждый коммит отклонён осознанно: медленные локальные хуки быстро начинают обходить.

### Dev Container — делаем сразу, минимальный

Внутри только среда разработки: **Go 1.24, Node LTS, pnpm, golangci-lint, gofumpt, markdownlint, Git, Make**.

Без PostgreSQL, Redis и Qdrant: инфраструктура запускается отдельно через Docker Compose. Dev Container отвечает только за среду разработки.

### Последствия

- EPIC-002.6: TASK-021 — pre-commit и pre-push хуки по таблице выше; TASK-024 — devcontainer в утверждённом составе.
- CI-workflow (EPIC-002.5, TASK-013) выполняет полный `make verify` на каждый PR и push в main.

## Статус

Актуален

## Последнее обновление

2026-07-19
