# TASK-023: Рекомендуемые настройки VS Code

## Тип

chore

## Эпик

EPIC-002.6 ([docs/roadmap/EPIC-002.6-developer-experience.md](../../docs/roadmap/EPIC-002.6-developer-experience.md))

## Цель

Новый разработчик, открыв проект в VS Code, получает подсказку установить нужные расширения и автоматически включённое форматирование/линтинг Go, согласованное с CI (gofumpt, golangci-lint).

## Scope

### Входит

- `.vscode/extensions.json`: `golang.go`, `editorconfig.editorconfig`, `davidanson.vscode-markdownlint`, `bierner.markdown-mermaid` (ровно набор, указанный в scope эпика — Go, EditorConfig, markdownlint, Mermaid; лишнего не добавлял).
- `.vscode/settings.json`: `files.eol: \n`; для `.go` — format on save + organize imports; `gopls.formatting.gofumpt: true` (тот же форматтер, что в CI); `go.lintTool: golangci-lint` (тот же линтер, что в CI); markdown — без авто-форматирования (конфликтовало бы с ручной правкой и `markdownlint-cli2 --fix`).

### Не входит

- `.vscode/tasks.json` — не запрошен архитектором; `make` уже даёт единый интерфейс команд для терминала.

## Критерии приёмки

- [x] `extensions.json`/`settings.json` — валидный JSON.
- [x] `gopls.formatting.gofumpt: true` — тот же инструмент форматирования, что в CI и в git-хуках (без расхождений «в редакторе одно, в CI другое»).
- [ ] Прогон verify в PR — зелёный.

## План реализации

Стандартные workspace-файлы VS Code; настройки форматирования/линтинга Go зеркалят зафиксированный toolchain (ADR-009), а не вводят собственный.

## Затрагиваемые модули и документы

- `.vscode/extensions.json` (новый), `.vscode/settings.json` (новый).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-20 — Architect — EPIC-002.6 утверждён к исполнению.
2026-07-20 — Claude Code (Developer) — ready → in-progress → выполнена, переведена в review (PR — следующий).

## Отчёт о выполнении

1. **Задача:** TASK-023 — рекомендуемые настройки VS Code.
2. **Что сделано:** `.vscode/extensions.json` (4 расширения по scope эпика), `.vscode/settings.json` (Go format-on-save через gofumpt, lint через golangci-lint — согласовано с CI/хуками).
3. **Изменённые файлы:** `.vscode/extensions.json` (новый), `.vscode/settings.json` (новый), файл задачи.
4. **Как проверялось:** валидность JSON проверена парсером Node (`node -e "JSON.parse(...)"`); сверка инструментов форматирования/линтинга с `.golangci.yml`/`verify.yml` — совпадают.
5. **Обновлённая документация:** не требуется отдельно.
6. **Open Questions:** нет.
7. **Рекомендации:** нет.
