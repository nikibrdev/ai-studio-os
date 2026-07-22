# TASK-071: README apps/api, синхронизация документации, закрытие EPIC-008

## Тип

docs

## Эпик

[EPIC-008 API Layer](../../docs/roadmap/EPIC-008-api-layer.md)

## Цель

Закрыть EPIC-008: README `apps/api`, синхронизация архитектурной документации, обновление PROJECT_MANIFEST/PROJECT_HEALTH/ROADMAP/CHANGELOG.

## Контекст

Последняя задача эпика — по аналогии с TASK-051/057/063, закрывавшими предыдущие эпики.

## Scope

### Входит

- `apps/api/README.md` — назначение, зависимости (`internal/application`, `internal/infrastructure/wiring`), запуск локально, ссылка на `docs/api/`.
- `docs/architecture/module-boundaries.md` — сверка с фактической реализацией (пакет `httpapi`, точка входа `main.go`).
- `docs/architecture/system-design.md` — если описание API-слоя там расходится с реализацией.
- ROADMAP.md (v0.9 — Завершено, с честным описанием ограничений: без auth, без списковых проекций), PROJECT_MANIFEST.md, PROJECT_HEALTH.md, CHANGELOG.md.

### Не входит

- Открытие эпика Dashboard (v0.8) — отдельная задача после этой.

## Критерии приёмки

- [x] `apps/api/README.md` написан по стандарту README модуля.
- [x] Архитектурная документация, упоминающая API-слой, сверена с фактической реализацией.
- [x] ROADMAP/PROJECT_MANIFEST/PROJECT_HEALTH/CHANGELOG отражают фактический результат, включая честные ограничения (без auth, без списковых проекций сверх `TaskProjection.Get`).
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `apps/api/README.md`, `docs/architecture/module-boundaries.md`, `docs/architecture/system-design.md`, `ROADMAP.md`, `PROJECT_MANIFEST.md`, `PROJECT_HEALTH.md`, `CHANGELOG.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-064…070

## План реализации

1. `apps/api/README.md` — обновить: убрать устаревшее «сейчас реализован только GET /healthz», задокументировать все 15 операций, URL-схему `/projects/{projectId}/tasks/...` (BUGFIX-003) и живую проверку golden path (TASK-070); Статус → Завершён.
2. `docs/architecture/module-boundaries.md` — сверить с фактической реализацией: добавить явное упоминание уже реализованного узкого исключения (`httpapi/errors.go` импортирует `internal/domain/*` только для сравнения sentinel-ошибок).
3. `docs/architecture/system-design.md` — сверено; расхождений с реализацией не найдено, правка не требуется (высокоуровневое описание API уже точно).
4. `docs/roadmap/EPIC-008-api-layer.md` — все критерии завершения отмечены со ссылками на подтверждающие артефакты; в «Риски и зависимости» добавлена запись о материализовавшемся и исправленном риске (BUGFIX-003); Статус → Закрыт.
5. `ROADMAP.md` — v0.9 помечен «Завершено» с честным абзацем результата (включая упоминание BUGFIX-003 и ограничений).
6. `PROJECT_MANIFEST.md` — Version/Current Epic (эпика больше нет), Application/Infrastructure/API в таблице слоёв, Last Review.
7. `PROJECT_HEALTH.md` — Architecture (3 ADR Decision Required, было 4 — ADR-012 принят при открытии эпика, эта строка не обновлялась до сих пор), Implementation/Testing/API/Dashboard под фактический результат (285 unit-тестов).
8. `CHANGELOG.md` — запись о закрытии EPIC-008 в Unreleased/Added, по образцу записей EPIC-004…007, с честным упоминанием BUGFIX-003.
9. `make verify`.

## История

2026-07-22 — Architect — EPIC-008 открыт; задача поставлена в очередь (последняя, закрывает эпик).

2026-07-22 — Developer — задача взята в работу, документация синхронизирована, эпик закрыт (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-071 — README `apps/api`, синхронизация документации, закрытие EPIC-008.

### Что сделано

- `apps/api/README.md` — синхронизирован с фактической реализацией: все 15 операций, URL-схема `/projects/{projectId}/tasks/...` (BUGFIX-003) с объяснением причины, раздел «Проверено вживую» (TASK-070); Статус → Завершён.
- `docs/architecture/module-boundaries.md` — добавлено явное упоминание уже реализованного узкого исключения по доменным sentinel-ошибкам в разделе `apps/api`.
- `docs/architecture/system-design.md` — сверено, расхождений с реализацией не найдено, изменений не потребовалось.
- `docs/roadmap/EPIC-008-api-layer.md` — все семь критериев завершения отмечены со ссылками на подтверждающие файлы/тесты; в «Риски» добавлена запись о материализовавшемся и исправленном риске (BUGFIX-003); Статус → Закрыт.
- `ROADMAP.md` — v0.9 API помечен «Завершено» (2026-07-22) с честным абзацем результата: golden path целиком, BUGFIX-003, ограничения (без auth, без списковых проекций, реальный вызов `RepositoryProvider` не проверен).
- `PROJECT_MANIFEST.md` — Version/Current Epic отражают закрытие; таблица слоёв — Application (5 сервисов), Infrastructure (составной ключ Task), API (Implemented); Last Review указывает на закрытие EPIC-008 целиком (PR #80–#87).
- `PROJECT_HEALTH.md` — Architecture (97%→98%, 3 ADR Decision Required вместо 4 — синхронизировано с принятием ADR-012, пропущенное при открытии эпика), Implementation (70%→80%), Testing (48%→52%, 285 unit-тестов), API (0%→100%), Dashboard — уточнена ссылка на decision о порядке реализации.
- `CHANGELOG.md` — запись о закрытии EPIC-008 в Unreleased/Added, по образцу записей EPIC-004…007, включая честное упоминание BUGFIX-003 и всех ограничений.

### Изменённые файлы

- `apps/api/README.md`, `docs/architecture/module-boundaries.md`, `docs/roadmap/EPIC-008-api-layer.md`, `ROADMAP.md`, `PROJECT_MANIFEST.md`, `PROJECT_HEALTH.md`, `CHANGELOG.md`.

### Как проверялось

- `make verify` — чисто (markdownlint, docs-check — 1291 ссылка проверена, 0 ошибок).

### Обновлённая документация

Вся документация, перечисленная в «Изменённые файлы» — эта задача целиком документационная.

### Open Questions

Нет.

### Рекомендации

- Замечена не связанная с этой задачей неточность в `PROJECT_MANIFEST.md`, строка `Dashboard | Not Started (v0.6)` — по `ROADMAP.md` Dashboard это v0.8, не v0.6 (тот же пункт уже отмечался в отчёте TASK-063, до сих пор не исправлен). Не исправлено и в этом PR (вне scope TASK-071); рекомендуется отдельная точечная задача.
- Следующий шаг по плану — открытие v0.8 Dashboard (первый эпик, зависящий от только что завершённого API), с собственной декомпозицией; вне scope этой задачи.
