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
2026-07-20 — Claude Code (Developer) — три итерации CI в PR #2: (1) падение установки golangci-lint по контрольной сумме → замена на пинованный `go install`; (2) 252 × MD060 (стиль таблиц) → отключено в базовом конфиге; (3) зелёный прогон. Задача переведена в review.

## Отчёт о выполнении

1. **Задача:** TASK-013 — GitHub Actions workflow verify.
2. **Что сделано:** `.github/workflows/verify.yml` — полный `make verify` на каждый PR и push в main (ubuntu-latest; Go 1.24 без кэша — go.sum ещё нет; Node LTS; gofumpt @latest; golangci-lint v2.12.2 через `go install` — установочный скрипт релиза падал на checksum); базовый `.markdownlint.jsonc` (MD013/MD033/MD040/MD060 off, MD024 siblings_only).
3. **Изменённые файлы:** `.github/workflows/verify.yml` (новый, вместо `.gitkeep`), `.markdownlint.jsonc` (новый), файл задачи.
4. **Как проверялось:** прогон run #3 на коммите 05e1cdb — **success**: fmt-check, golangci-lint (0 issues), go vet, go test, markdownlint (0 ошибок с конфигом), docs-check (580 ссылок, 9 Mermaid-блоков, 0 ошибок).
5. **Обновлённая документация:** файл задачи; конфиги самодокументированы комментариями.
6. **Open Questions:** нормализовать ли разделительные строки таблиц (|---| → | --- |) в 34 файлах и вернуть MD060 — решить в TASK-017. В списке типов коммитов git-workflow.md нет `ci` — использую `chore(ci)`; предложение: добавить тип `ci` (мелкая правка docs).
7. **Рекомендации:** TASK-016 (защита main с обязательным чеком `verify`) теперь разблокирована; следующим PR — TASK-014/015 (CODEOWNERS, шаблоны Issue) или TASK-016.
