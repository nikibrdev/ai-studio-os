# TASK-073: Спецификации списковых операций (Documentation First)

## Тип

docs

## Эпик

[EPIC-009 Dashboard](../../docs/roadmap/EPIC-009-dashboard.md)

## Цель

Описать `GET /projects` и `GET /projects/{id}/tasks` в `docs/api/` до того, как Dashboard (TASK-075/076) начнёт их вызывать — тот же принцип Documentation First, что и TASK-066.

## Контекст

Операции реализованы в TASK-072; эта задача документирует их точную форму (запрос, ответ, коды ошибок) по шаблону [API.md](../../.claude/templates/API.md), дополняя уже существующие `docs/api/projects.md` и `docs/api/tasks.md`.

## Scope

### Входит

- `docs/api/projects.md` — операция `GET /projects`.
- `docs/api/tasks.md` — операция `GET /projects/{id}/tasks`.

### Не входит

- Пагинация/фильтрация — не описывается, поскольку не реализована (TASK-072).

## Критерии приёмки

- [ ] Обе операции описаны в точности по фактически реализованному в TASK-072 (не наоборот).
- [ ] `make verify` (markdownlint, docs-check) — чисто.

## Затрагиваемые модули и документы

- `docs/api/projects.md`, `docs/api/tasks.md`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-072

## План реализации

## История

2026-07-22 — Architect — EPIC-009 открыт; задача поставлена в очередь.

## Отчёт о выполнении
