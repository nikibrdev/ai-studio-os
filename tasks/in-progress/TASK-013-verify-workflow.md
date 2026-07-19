# TASK-013: GitHub Actions — workflow verify

## Тип

chore

## Эпик

EPIC-002.5 ([docs/roadmap/EPIC-002.5-engineering-platform.md](../../docs/roadmap/EPIC-002.5-engineering-platform.md))

## Цель

Каждый PR и push в `main` автоматически проходит полный `make verify` в GitHub Actions; статус-чек `verify` — будущий обязательный чек защиты ветки (TASK-016).

## Контекст

[Решение quality gates](../../engineering/decisions/2026-07-19-quality-gates.md): CI выполняет полный `make verify` без исключений. План утверждён в составе эпика: ubuntu-latest, setup-go 1.24, setup-node (markdownlint), golangci-lint/gofumpt, `make verify`.

## Scope

### Входит

- `.github/workflows/verify.yml` (замена `.gitkeep`).
- Базовый `.markdownlint.jsonc` — без него `make md-lint` в CI падал бы на правилах по умолчанию; уточнение конфига — TASK-017.

### Не входит

- Защита ветки (TASK-016); полный конфиг markdownlint и mermaid-cli (TASK-017); проверка Conventional Commits (TASK-018).

## Критерии приёмки

- [ ] Прогон verify на PR завершается успешно (зелёный чек).
- [ ] Workflow запускается на pull_request и push в main.
- [ ] Выполняются все шаги `make verify`, включая markdownlint (npx доступен).

## План реализации

По утверждённому плану эпика (п. 2): checkout → setup-go 1.24 → setup-node LTS → установка gofumpt (@latest) и golangci-lint (v2.12.2, как локально) → `make verify`. Плюс базовый `.markdownlint.jsonc` (отключены MD013/MD033/MD040, MD024 siblings_only) — обоснование в Scope.

## Затрагиваемые модули и документы

- `.github/workflows/`, `.markdownlint.jsonc` (корень).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-20 — Architect — сигнал на выполнение TASK-013 (план эпика утверждён ранее).
2026-07-20 — Claude Code (Developer) — задача взята в работу (in-progress), ветка `chore/TASK-013-verify-workflow`.

## Отчёт о выполнении

(будет дополнен после зелёного прогона CI)
