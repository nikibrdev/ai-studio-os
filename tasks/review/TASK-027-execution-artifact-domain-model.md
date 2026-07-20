# TASK-027: Execution — не Bounded Context, Artifact — первичная сущность, порядок Domain Layer

## Тип

docs

## Эпик

Вне эпика — архитектурные решения архитектора проекта, 2026-07-20 (тот же разговор, что и ADR-005/TASK-026); эта задача — оставшаяся часть, намеренно вынесенная из TASK-026, чтобы не смешивать переименование кода с широкой доке-перестройкой в одном PR.

## Цель

Документация синхронизирована с четырьмя решениями архитектора, ещё не отражёнными в коде/доках:

1. Execution — не Bounded Context, а сквозная техническая возможность, координируемая Application Layer; контекстов — четыре: Planning → Development → Review → Knowledge (переименован из Memory).
2. Artifact — первичная сущность результата работы (не Result, не Output, не Response); домен-модель отражает это.
3. Порядок проектирования Domain Layer (EPIC-003): Artifact → Execution → Executor → Task → Project — зафиксирован явно, а не подразумевается.
4. Закрывающая формулировка VISION.md усилена дословной цитатой архитектора.

## Контекст

Решение принято в одном разговоре с ADR-005 (TASK-026, уже смержен, PR #24/#25). TASK-026 закрыл только код (`internal/platform`) и терминологическую часть (Agent/Executor). Эта задача — оставшаяся, более широкая доке-перестройка: границы контекстов, доменная модель, видение.

Ключевая цитата архитектора (дословно, для VISION.md):

> AI Studio OS — это операционная система исполнения инженерной работы. LLM, человек, Claude Code, Codex, OpenHands — это просто разные исполнители. Задачи — это способ организовать работу. А артефакты — это настоящая ценность, которую производит система.

## Scope

### Входит

- `docs/domain/bounded-contexts.md`: четыре контекста вместо пяти (Execution убран из таблицы «Контексты и владение», Memory → Knowledge); таблица «Отображение на доменные модули» и Mermaid-карта обновлены; открытый вопрос про границу Execution — снят (заменён новой явной формулировкой из решения архитектора); новый открытый вопрос — куда относится состояние Task `Testing` (QA), раз Execution больше не контекст (см. «Открытые вопросы» ниже).
- `docs/domain/ubiquitous-language.md`: короткая синхронизация — контекст, к которому относилась Memory, теперь называется Knowledge (без переименования самой сущности/модуля `memory` — см. «Открытые вопросы»).
- `docs/architecture/domain-model.md`:
  - `Result` как отдельная сущность убирается: Execution напрямую производит Artifact (`Execution "1" o-- "0..*" Artifact : производит`) и несёт `ExecutionStatus` — согласовано с уже принятым ADR-005 (`Accept`/`Artifacts`/`Status`/`Finish`, без Result).
  - `Agent` → `Executor` (сущность, Mermaid class, связи, владение) — согласовано с ADR-005/ubiquitous-language.md.
  - Новый раздел «Порядок проектирования Domain Layer» с обоснованием Artifact → Execution → Executor → Task → Project.
  - Сводная таблица владения данными и Decision Required — обновлены под перечисленные изменения.
- `docs/architecture/core.md` и `docs/architecture/components.md`: концептуальный список 10 доменных модулей — `agent` → `executor` (тот же список, то же число модулей).
- `docs/architecture/module-boundaries.md`: точечно — только если там остались строки со словом «agent» как именем будущего модуля (проверить при реализации).
- `VISION.md`: раздел «Ядро видения» — добавлена дословная цитата архитектора.
- Новая запись `engineering/decisions/2026-07-20-execution-artifact-domain-order.md` — фиксирует все четыре пункта решения.
- `PROJECT_MANIFEST.md` / `CHANGELOG.md` — синхронизация (дата, при необходимости).

### Не входит

- Реализация модулей `execution`, `executor`, `artifact` — Domain Layer (EPIC-003, отдельный старт по сигналу архитектора).
- Переименование кода/модуля `memory` → `knowledge` — не запрошено явно; см. «Открытые вопросы».
- Решение об отдельном модуле `artifact` vs владение модулем `execution` — уже зафиксировано как открытый вопрос в самом ADR-005, TASK-027 его не решает, только не противоречит.

## Открытые вопросы — согласованы архитектором 2026-07-20

1. **Task `Testing` (QA) вне контекстов.** Подтверждено: `Testing` не привязывается ни к одному из четырёх контекстов — показывается как переход через сквозную возможность Execution (Application Layer), примыкающую к Review→Done, роли Reviewer и QA Engineer не смешиваются.
2. **Memory → Knowledge — только название контекста.** Подтверждено: сущность/код-модуль остаются `Memory`/`memory` до отдельного решения; переименован только Bounded Context.
3. **Домен-модуль `agent` → `executor`.** Подтверждено: понятие Agent не становится отдельным модулем (выражается связкой Role + Executor); список 10 модулей обновляется `agent` → `executor`.

## Критерии приёмки

- [x] `docs/domain/bounded-contexts.md` — четыре контекста, Mermaid-карта и таблицы согласованы, старый открытый вопрос про Execution закрыт.
- [x] `docs/domain/ubiquitous-language.md` — синхронизирован без противоречий с bounded-contexts.md.
- [x] `docs/architecture/domain-model.md` — Result убран, Executor вместо Agent, добавлен порядок проектирования Domain Layer, диаграмма и таблицы согласованы.
- [x] `docs/architecture/core.md`, `docs/architecture/components.md` — список модулей согласован (executor вместо agent).
- [x] `VISION.md` — цитата добавлена дословно.
- [x] Новая запись в `engineering/decisions/`.
- [x] `bash scripts/verify-docs.sh` — 0 битых ссылок; `npx markdownlint-cli2` — 0 issues.
- [x] `PROJECT_MANIFEST.md`/`CHANGELOG.md` синхронизированы (CHANGELOG дополнен; PROJECT_MANIFEST проверен — изменений, требующих правки, не нашлось: версия/эпик/слои не менялись).

## План реализации

1. Получить подтверждение по трём открытым вопросам выше (или скорректированные ответы).
2. `docs/domain/bounded-contexts.md`: переписать таблицы «Контексты и владение», «Отображение на доменные модули», Mermaid-диаграмму, раздел «Открытые вопросы» (закрыть старый про границу Execution, отразить решение архитектора; добавить новый про Testing, если не решён вопросом 1).
3. `docs/domain/ubiquitous-language.md`: синхронизировать упоминания Memory/Knowledge с bounded-contexts.md (объём — по ответу на вопрос 2).
4. `docs/architecture/domain-model.md`: убрать Result, переименовать Agent → Executor (класс, связи, таблица владения), добавить раздел «Порядок проектирования Domain Layer», обновить Decision Required при необходимости.
5. `docs/architecture/core.md`, `docs/architecture/components.md`, точечно `module-boundaries.md`: `agent` → `executor` в списке 10 модулей.
6. `VISION.md`: добавить дословную цитату в разделе «Ядро видения».
7. Новая запись `engineering/decisions/2026-07-20-execution-artifact-domain-order.md`.
8. `PROJECT_MANIFEST.md`/`CHANGELOG.md`: синхронизация.
9. Локальная верификация: `bash scripts/verify-docs.sh`, `npx markdownlint-cli2`.
10. Заполнить отчёт, commit → push → PR → CI → review → merge → `tasks/done/` (тот же пайплайн, что TASK-026).

## Затрагиваемые модули и документы

`docs/domain/bounded-contexts.md`, `docs/domain/ubiquitous-language.md`, `docs/architecture/domain-model.md`, `docs/architecture/core.md`, `docs/architecture/components.md`, `docs/architecture/module-boundaries.md` (точечно), `VISION.md`, `engineering/decisions/`, `PROJECT_MANIFEST.md`, `CHANGELOG.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Открытые вопросы согласованы с архитектором (см. выше)

## История

2026-07-20 — Claude Code (Developer) — план записан по итогам решения архитектора (тот же разговор, что ADR-005); задача остановлена на согласовании открытых вопросов, реализация не начата.
2026-07-20 — Architect — три открытых вопроса согласованы (рекомендованные варианты приняты); реализация разрешена.
2026-07-20 — Claude Code (Developer) — реализовано, локальная верификация пройдена, переведена в review.

## Отчёт о выполнении

1. **Задача:** TASK-027 — Execution не Bounded Context, Artifact — первичная сущность, порядок Domain Layer, цитата в VISION.md.
2. **Что сделано:**
   - `docs/domain/bounded-contexts.md`: контекстов — четыре (Planning, Development, Review, Knowledge); Execution описан как сквозная возможность Application Layer, не контекст; таблица «Отображение на доменные модули» и Mermaid-карта перестроены (Execution вынесен из подграфов, Task: Testing показан как переход через сквозную возможность, не часть Review); старый открытый вопрос про границу Execution закрыт, добавлено пояснение оставшихся вопросов проектирования EPIC-003.
   - `docs/domain/ubiquitous-language.md`: термины Memory и Execution синхронизированы с bounded-contexts.md (Bounded Context называется Knowledge, сама сущность/модуль Memory не переименованы; Execution явно помечен как не-контекст).
   - `docs/architecture/domain-model.md`: сущность `Result` убрана — Execution напрямую производит Artifact и несёт `ExecutionStatus` (согласовано с уже принятым ADR-005); `Agent` → `Executor` (класс, связи, определение, владение); добавлен раздел «Порядок проектирования Domain Layer» (Artifact → Execution → Executor → Task → Project) с обоснованием; сводная таблица владения данными обновлена.
   - `docs/architecture/core.md`, `docs/architecture/components.md`: домен-модуль `agent` → `executor` в списке 10 модулей и на диаграммах; попутные упоминания контракта (`Agent` → `Executor`, «Agent Runtime» → «Executor Runtime» со сноской о дословной формулировке ADR-014) приведены в соответствие с ADR-005.
   - `docs/architecture/module-boundaries.md`: проверен — упоминаний домен-модуля `agent` по имени не найдено, правка не потребовалась.
   - `VISION.md`: в разделе «Ядро видения» добавлена дословная цитата архитектора и следствие про Task как организационный механизм, а Artifact/Execution — центральные сущности.
   - Новая запись `engineering/decisions/2026-07-20-execution-artifact-domain-order.md` — фиксирует все четыре решения и согласованные открытые вопросы.
   - `CHANGELOG.md`: добавлена запись Unreleased/Added. `PROJECT_MANIFEST.md`: проверен, правка не потребовалась (версия/эпик/состояние слоёв не изменились).
3. **Изменённые файлы:**
   - `docs/domain/bounded-contexts.md`
   - `docs/domain/ubiquitous-language.md`
   - `docs/architecture/domain-model.md`
   - `docs/architecture/core.md`
   - `docs/architecture/components.md`
   - `VISION.md`
   - `CHANGELOG.md`
   - `engineering/decisions/2026-07-20-execution-artifact-domain-order.md` (новый)
   - `tasks/review/TASK-027-execution-artifact-domain-model.md` (эта задача)
4. **Как проверялось:**
   - `go build ./...`, `go vet ./...`, `gofumpt -l .`, `golangci-lint run ./...` — чисто (изменений Go-кода не было, проверено на всякий случай — задача документационная).
   - `bash scripts/verify-docs.sh` — 743 ссылки, 11 mermaid-блоков, 0 ошибок.
   - `npx markdownlint-cli2` — 151 файл, 0 issues.
5. **Обновлённая документация:** см. «Изменённые файлы» выше — вся работа документационная.
6. **Open Questions:** один — уже зафиксирован в самом ADR-005 (не новый, не решается этой задачей): точное место типа `Artifact` в коде (отдельный домен-модуль `artifact` vs владение модулем `execution`) — требует архитектора при проектировании модуля Artifact, EPIC-003.
7. **Рекомендации:** после начала EPIC-003 — при написании спецификаций модулей ([docs/specifications/domain/](../../docs/specifications/domain/)) явно следовать порядку Artifact → Execution → Executor → Task → Project, зафиксированному в domain-model.md, а не начинать с Task по умолчанию.
