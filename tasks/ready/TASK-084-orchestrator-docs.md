# TASK-084: Документация apps/orchestrator

## Тип

docs

## Эпик

[EPIC-010 Orchestrator](../../docs/roadmap/EPIC-010-orchestrator.md)

## Цель

Задокументировать построенный механизм: README модуля, новый архитектурный документ о механизме диспетчеризации и курсорном опросе журнала, синхронизация `module-boundaries.md`/`components.md` с фактической реализацией.

## Контекст

По аналогии с README `agents/claude-code`/`apps/api`/`apps/dashboard` — описание должно позволить читателю понять механизм без чтения кода. `module-boundaries.md` уже содержит раздел `apps/orchestrator`, но написан до реализации (терминология «Core», без упоминания `EventJournal`) — тот же вид уточнения, что уже сделан для `apps/api` в EPIC-008/009.

## Scope

### Входит

- `apps/orchestrator/README.md` — назначение, переменные окружения, механизм (бутстрап исполнителя, курсорный опрос, диспетчеризация Developer), ограничения (курсор в памяти, один исполнитель, без ретраев), запуск локально.
- `docs/architecture/orchestrator.md` (новый, по единой структуре: Заголовок/Назначение/Содержание/Статус/Последнее обновление) — механизм диспетчеризации, курсорный опрос журнала вместо `Subscribe` (со ссылкой на обоснование в EPIC-010), диаграмма последовательности Developer-шага.
- `docs/architecture/module-boundaries.md` — раздел `apps/orchestrator`: уточнение терминологии «Core»→`internal/application` (как для `apps/api`), упоминание порта `EventJournal`.
- `docs/architecture/components.md` — при необходимости уточнить описание Orchestrator'а под фактическую реализацию.
- Ссылка на `docs/architecture/orchestrator.md` из `docs/architecture/README.md` (индекс), если там есть перечень документов раздела.

### Не входит

- Изменение кода — эта задача только документация.

## Критерии приёмки

- [ ] `apps/orchestrator/README.md` написан по стандарту README модуля.
- [ ] `docs/architecture/orchestrator.md` создан, отражает фактическую реализацию (не аспирационное описание).
- [ ] `module-boundaries.md`/`components.md` синхронизированы.
- [ ] `make verify` (включая проверку ссылок и markdownlint) — чисто.

## Затрагиваемые модули и документы

- `apps/orchestrator/README.md`, `docs/architecture/orchestrator.md`, `docs/architecture/module-boundaries.md`, `docs/architecture/components.md`, `docs/architecture/README.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-081…083 (описывает фактически построенное)

## План реализации

<Заполняется исполнителем до начала работы; реализация начинается только после утверждения плана.>

## История

2026-07-23 — Architect — EPIC-010 открыт; задача поставлена в очередь, зависит от TASK-081…083.

## Отчёт о выполнении

<Заполняется исполнителем после завершения.>
