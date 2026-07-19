# EPIC-002.5: Engineering Platform

## Цель

Превратить репозиторий в полноценную инженерную платформу до начала Domain Layer: Git/GitHub, обязательные проверки, защита основной ветки, шаблоны, автоматизация. После этого эпика ни одно изменение — человеческое или агентское — не попадает в `main` мимо единого процесса проверки.

## Контекст

Решение архитектора (2026-07-19): разработка приостановлена до создания инженерной платформы. Локально уже есть: `make verify`, `scripts/verify-docs.sh`, `.golangci.yml`, шаблон PR. Репозиторий ещё не под git; gh CLI в окружении отсутствует.

## Scope

### Входит

1. Инициализация Git и GitHub-репозитория.
2. GitHub Actions: workflow `verify` (make verify на каждый PR и push в main).
3. CODEOWNERS.
4. Шаблоны Issue (bug / feature / task) и актуализация шаблона PR.
5. Защита `main` (обязательные проверки, запрет прямых пушей).
6. Автоматическая проверка документации и Mermaid в CI (markdownlint + verify-docs).
7. Проверка Conventional Commits в CI.
8. Автогенерация релизных заметок + Dependabot.

### Не входит

- Domain Layer и любой прикладной код; деплой; интеграция агентов с GitHub (v0.3+).

## Декомпозиция

| Задача | Содержание | Зависимости |
|---|---|---|
| TASK-012 | `git init`, первый коммит (Conventional Commits), ветка `main`, создание GitHub-репозитория, push | решения по вопросам 1–3 ниже |
| TASK-013 | `.github/workflows/verify.yml`: make verify (Go 1.24, golangci-lint, gofumpt, node для markdownlint) на PR и push | TASK-012 |
| TASK-014 | CODEOWNERS (владелец — мейнтейнер; на `docs/adr/` и `CONSTITUTION.md` — только мейнтейнер) | вопрос 3 |
| TASK-015 | `.github/ISSUE_TEMPLATE/`: bug_report, feature_request, task; сверка PR-шаблона | — |
| TASK-016 | Защита `main`: обязательный статус verify, обязательное ревью, запрет прямых пушей и force-push | TASK-013, вопрос 2 |
| TASK-017 | CI-проверка документации: markdownlint-cli2 (+ конфиг `.markdownlint.jsonc`), `scripts/verify-docs.sh`; при возможности — mermaid-cli для полной валидации диаграмм | TASK-013 |
| TASK-018 | Проверка Conventional Commits: скрипт `scripts/check-commits.sh` + шаг в CI (диапазон коммитов PR) | TASK-013 |
| TASK-019 | Dependabot: `gomod`, `github-actions` (npm — при появлении dashboard) | TASK-012 |
| TASK-020 | Релизные заметки: `.github/release.yml` (категории по типам Conventional Commits) + процесс тегирования в docs/development/git-workflow.md | TASK-012 |

## План реализации (на утверждение)

1. **TASK-012:** `git init` в корне; `.gitignore` уже готов; первый коммит `chore: bootstrap repository (v0.1 foundation + architecture v0.2 + core contracts)`; репозиторий на GitHub; push `main`. gh CLI в окружении нет — потребуется либо установка gh + `gh auth login` (интерактивно, делает человек), либо создание репозитория вручную в веб-интерфейсе и `git remote add`.
2. **TASK-013:** workflow на `ubuntu-latest`: checkout → setup-go 1.24 → установка golangci-lint/gofumpt → setup-node (для markdownlint) → `make verify`. Один обязательный статус-чек `verify`.
3. **TASK-014–015:** файлы шаблонов и CODEOWNERS по решению вопроса 3.
4. **TASK-016:** защита ветки через настройки GitHub (ruleset): требуемый чек `verify`, ≥1 approve, запрет прямых пушей; выполняется после первого зелёного прогона CI.
5. **TASK-017–020:** конфиги в `.github/` и `scripts/`; каждый — отдельный PR через новый процесс (план в задаче → утверждение → код → ревью).
6. После завершения: правка git-workflow.md (релизный процесс), ADR-009 (замена локального имени модуля `ai-studio-os` на `github.com/<owner>/ai-studio-os` — отдельной задачей, механическая).

## Решения по вопросам (приняты 2026-07-19, [запись](../../engineering/decisions/2026-07-19-github-repository.md))

1. **Имя:** `ai-studio-os`, личный аккаунт мейнтейнера, без организации.
2. **Видимость:** Public.
3. **GitHub-логин мейнтейнера:** определяется из URL репозитория после его создания (нужен для CODEOWNERS, TASK-014).
4. **Создание:** пустой репозиторий через веб-интерфейс (без README/LICENSE/.gitignore), затем git init → commit → remote add → push.

## Проверка после первого push (обязательна до EPIC-003)

По указанию архитектора, до перехода к Domain Layer проверить:

- [ ] GitHub Actions проходит (полный `make verify` зелёный).
- [ ] Все проверки работают (fmt, lint, vet, test, markdownlint, docs).
- [ ] Mermaid-диаграммы корректно отображаются на GitHub.
- [ ] README отображается корректно.
- [ ] Предупреждений линтера нет.
- [ ] Шаблон PR работает правильно.

## Критерии завершения

- [ ] Репозиторий на GitHub; `main` защищена; прямой push невозможен.
- [ ] PR без зелёного `verify` не сливается; ревью обязательно.
- [ ] Шаблоны Issue/PR, CODEOWNERS, Dependabot, release.yml действуют.
- [ ] Чек-лист «Проверка после первого push» закрыт.
- [ ] CHANGELOG и git-workflow.md обновлены.

## Статус

Утверждён; выполняется (TASK-012 начат: локальная часть готова, push ожидает URL репозитория от мейнтейнера)

## Последнее обновление

2026-07-19
