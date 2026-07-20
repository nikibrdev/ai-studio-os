# TASK-028: ADR-016 — Artifact как самостоятельный Aggregate Root

## Тип

docs

## Эпик

Вне эпика — последнее фундаментальное архитектурное решение архитектора проекта (2026-07-20) перед стартом EPIC-003; закрывает открытый вопрос, честно зафиксированный в [ADR-005](../../docs/adr/ADR-005-executor-contract.md) («где размещается тип Artifact»).

## Цель

Принят ADR-016: Artifact — самостоятельный Aggregate Root, не часть Execution/Task/Project по владению данными; доменная модель и связанная документация отражают это до старта EPIC-003.

## Контекст

Решение архитектора (дословно, для ADR):

- Artifact может существовать до задачи (импортированный документ), во время выполнения, после завершения задачи, независимо от конкретного исполнителя — поэтому не может быть частью Execution/Task/Project.
- Модель: `Project ├── Task └── Artifact`; `Task` создаёт `Execution`; `Execution` использует `Executor`; `Execution` производит `Artifact`.
- Определение: «Artifact — это любое долговременное инженерное произведение, созданное или изменённое системой исполнения» (commit, PR, source file, markdown, ADR, спецификация, test report, build report, diagram, screenshot, Figma-файл, release note; в будущем — video, audio, dataset, prompt, knowledge entry).
- НЕ Artifact: временный лог, прогресс выполнения, heartbeat, токен LLM, внутреннее сообщение агента — это относится к Execution, не является результатом системы.
- Artifact состоит из Metadata (ID, Type, Author, CreatedAt, ProducedByExecution, Version — то, что знает платформа) и Payload (сами данные — Markdown/Git Commit/PDF/JSON/Binary — содержимое зависит от типа, платформа его не интерпретирует).
- Правило: **Execution никогда не владеет Artifact** — Execution только знает, что произвёл конкретный Artifact (ссылка, не контейнер).

Соответствует критериям ADR из CLAUDE.md: вводится/меняется структура доменных модулей (новый Aggregate Root, новый модуль `artifact`) и меняется ранее принятое размещение (ADR-005 предполагало `execution` как вероятного владельца Artifact — теперь заменено).

## Scope

### Входит

- `docs/adr/ADR-016-artifact-aggregate-root.md` — новый ADR, статус Принято.
- `docs/architecture/domain-model.md`: Artifact — Aggregate Root, отдельный владелец данных (модуль `artifact`); Metadata/Payload описаны; связи обновлены (`Project o-- Artifact`, `Execution ..> Artifact : производит` — не владение; `Execution --> Executor : использует`); примеры и границы (что НЕ Artifact) добавлены; сводная таблица владения обновлена (`execution` теряет Artifact, появляется `artifact`).
- `docs/adr/ADR-005-executor-contract.md`: раздел «Открытый вопрос» закрывается ссылкой на ADR-016 (сам ADR-005 не редактируется по существу решения — только фиксация, что вопрос решён, аналогично прежним прецедентам синхронизации терминологии).
- `docs/domain/ubiquitous-language.md`: определение Artifact расширено под новое, более широкое определение архитектора; добавлены примеры и границы (что НЕ Artifact).
- `docs/domain/bounded-contexts.md`: `artifact` добавлен в список сквозных модулей (наравне с `execution`/`executor`/`tool`) в «Отображение на доменные модули».
- `docs/architecture/core.md`, `docs/architecture/components.md`, `docs/architecture/module-boundaries.md`: список доменных модулей — 10 → 11 (добавлен `artifact`).
- `PROJECT_MANIFEST.md`/`CHANGELOG.md` — синхронизация (Last ADR → ADR-016).

### Не входит

- Реализация модуля `artifact` (Go) — EPIC-003.
- Полная схема Metadata/Payload на уровне Go-типов — задача будущей спецификации `docs/specifications/domain/artifact.md` (EPIC-003, Этап 1).

## Критерии приёмки

- [x] ADR-016 создан, статус Принято, отражает решение дословно (определение, границы, Metadata/Payload, правило владения).
- [x] `domain-model.md` согласован: Artifact — Aggregate Root, отдельный модуль, связи без ложного владения Execution.
- [x] Открытый вопрос ADR-005 закрыт ссылкой на ADR-016.
- [x] `bash scripts/verify-docs.sh` — 0 битых ссылок; `npx markdownlint-cli2` — 0 issues.
- [x] `PROJECT_MANIFEST.md`/`CHANGELOG.md` синхронизированы.

## План реализации

1. `docs/adr/ADR-016-artifact-aggregate-root.md` — по шаблону ADR-000, статус Принято, дата 2026-07-20.
2. `docs/architecture/domain-model.md`: обновить Mermaid classDiagram и раздел «Сущности» (Artifact, Execution), сводную таблицу владения, добавить пояснение Metadata/Payload и границы (что НЕ Artifact) со ссылкой на ADR-016.
3. `docs/adr/ADR-005-executor-contract.md`: закрыть открытый вопрос.
4. `docs/domain/ubiquitous-language.md`, `docs/domain/bounded-contexts.md`: синхронизировать определение и карту модулей.
5. `docs/architecture/core.md`, `docs/architecture/components.md`, `docs/architecture/module-boundaries.md`: 10 → 11 модулей.
6. `PROJECT_MANIFEST.md`/`CHANGELOG.md`: синхронизация.
7. Локальная верификация (`verify-docs.sh`, `markdownlint-cli2`, `go build`/`vet`/`lint`/`gofumpt` — Go не менялся, но проверяется по обычному пайплайну).
8. Отчёт, commit → push → PR → CI → review → merge → `tasks/done/`.

## Затрагиваемые модули и документы

`docs/adr/ADR-016-*` (новый), `docs/adr/ADR-005-executor-contract.md`, `docs/architecture/domain-model.md`, `docs/architecture/core.md`, `docs/architecture/components.md`, `docs/architecture/module-boundaries.md`, `docs/domain/ubiquitous-language.md`, `docs/domain/bounded-contexts.md`, `PROJECT_MANIFEST.md`, `CHANGELOG.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — решение дано архитектором дословно, открытых вопросов нет

## История

2026-07-20 — Architect — решение принято дословно (Artifact — Aggregate Root, определение, Metadata/Payload, правило владения); задача создана.
2026-07-20 — Claude Code (Developer) — план записан, задача взята в работу.
2026-07-20 — Claude Code (Developer) — реализовано, локальная верификация пройдена, переведена в review.

## Отчёт о выполнении

1. **Задача:** TASK-028 — ADR-016, Artifact как самостоятельный Aggregate Root.
2. **Что сделано:**
   - Создан `docs/adr/ADR-016-artifact-aggregate-root.md` (статус Принято): модель `Project ├── Task └── Artifact`, определение Artifact (широкое, с примерами и границами — что НЕ Artifact), разделение Metadata/Payload, правило «Execution никогда не владеет Artifact».
   - `docs/architecture/domain-model.md`: classDiagram обновлён (`Project o-- Artifact`, `Execution ..> Artifact : производит (ссылка, не владение)`, `Execution --> Executor : использует`); разделы Artifact и Execution переписаны под новое определение и правило владения; сводная таблица владения — `artifact` выделен в отдельный модуль (был частью `execution`).
   - `docs/adr/ADR-005-executor-contract.md`: открытый вопрос о размещении Artifact закрыт ссылкой на ADR-016 (сам текст решения ADR-005 не менялся).
   - `docs/domain/ubiquitous-language.md`: определение Artifact расширено (примеры, границы, Aggregate Root, правило невладения).
   - `docs/domain/bounded-contexts.md`: `artifact` добавлен в список сквозных модулей рядом с `execution`/`executor`/`tool`.
   - `docs/architecture/core.md`, `docs/architecture/components.md`, `docs/architecture/module-boundaries.md`: концептуальный список доменных модулей — 10 → 11 (`artifact` добавлен).
   - `docs/adr/DECISIONS_INDEX.md`: добавлена строка ADR-016 (Accepted), счётчик Принято — 8 → 9 (задача не была явно в scope, добавлена как необходимое следствие принятия нового ADR — по правилу «индекс обновляется при каждом изменении статуса любого ADR, в том же PR»).
   - `PROJECT_MANIFEST.md`: Last ADR → ADR-016. `CHANGELOG.md`: добавлена запись Unreleased/Added.
3. **Изменённые файлы:**
   - `docs/adr/ADR-016-artifact-aggregate-root.md` (новый)
   - `docs/architecture/domain-model.md`
   - `docs/adr/ADR-005-executor-contract.md`
   - `docs/domain/ubiquitous-language.md`
   - `docs/domain/bounded-contexts.md`
   - `docs/architecture/core.md`
   - `docs/architecture/components.md`
   - `docs/architecture/module-boundaries.md`
   - `docs/adr/DECISIONS_INDEX.md`
   - `PROJECT_MANIFEST.md`
   - `CHANGELOG.md`
   - `tasks/review/TASK-028-artifact-aggregate-root.md` (эта задача)
4. **Как проверялось:**
   - `go build ./...`, `go vet ./...`, `gofumpt -l .`, `golangci-lint run ./...` — чисто (Go-код не менялся, задача документационная).
   - `bash scripts/verify-docs.sh` — 760 ссылок, 11 mermaid-блоков, 0 ошибок.
   - `npx markdownlint-cli2` — 153 файла, 0 issues.
5. **Обновлённая документация:** см. «Изменённые файлы» — вся работа документационная.
6. **Open Questions:** нет. Решение дано архитектором дословно, без пробелов, требующих уточнения.
7. **Рекомендации:** EPIC-003 (Domain Layer) готов к открытию в режиме «Domain Specifications First» (без Go), как явно указал архитектор — спецификации `Artifact`, `Execution`, `Executor`, `Task`, `Project` в `docs/specifications/domain/` должны быть написаны и утверждены до начала реализации.
