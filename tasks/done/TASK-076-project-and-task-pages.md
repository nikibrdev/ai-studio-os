# TASK-076: Страница проекта (список задач) и страница задачи (детали)

## Тип

feature

## Эпик

[EPIC-009 Dashboard](../../docs/roadmap/EPIC-009-dashboard.md)

## Цель

Завершить навигационный путь Dashboard первой версии: проект → его задачи → детали одной задачи.

## Контекст

Поверх TASK-072/073 (`GET /projects/{id}/tasks`) и уже существующего `GET /projects/{projectId}/tasks/{id}` (EPIC-008).

## Scope

### Входит

- `/projects/[id]` — список задач проекта: ID (`TASK-NNN`), заголовок, состояние; ссылка на страницу задачи.
- `/projects/[id]/tasks/[taskId]` — детали задачи: заголовок, тип, scope, критерии приёмки, состояние, время последнего обновления.
- Компонентные/страничные тесты на моках HTTP-ответов обеих операций.

### Не входит

- Формы действий (plan/start/review/testing) — сознательно вне scope этой версии (см. EPIC-009 «Не входит»).

## Критерии приёмки

- [x] Полный путь навигации (список проектов → задачи проекта → детали задачи) проходит вживую против настоящего `apps/api`.
- [x] Неизвестный проект/задача — читаемое сообщение об ошибке, не необработанное падение страницы.
- [x] `pnpm test` — тесты обеих страниц зелёные.

## Затрагиваемые модули и документы

- `apps/dashboard/src/app/projects/[id]/page.tsx`, `apps/dashboard/src/app/projects/[id]/tasks/[taskId]/page.tsx` и тесты.
- `internal/application/{task_planning.go,projection.go}` и тесты, `apps/api/httpapi/tasks.go` и тесты, `docs/api/tasks.md` — расширение (см. Отчёт).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-072, 073, 074, 075

## План реализации

1. `src/app/projects/[id]/page.tsx` — `listProjectTasks(id)`, список задач (ID, заголовок, состояние), ссылка на страницу задачи; `export const dynamic = "force-dynamic"` (та же причина, что TASK-075).
2. `src/app/projects/[id]/tasks/[taskId]/page.tsx` — `getTask(id, taskId)`, детали (заголовок, тип, scope, критерии приёмки, состояние, время обновления).
3. **Обнаружено при проектировании (до написания кода)**: `GET /projects/{projectId}/tasks/{id}` отдаёт только `{id, projectId, state, updatedAt}` — заголовка/типа/scope/критериев приёмки в ответе нет, потому что `TaskProjection` (TASK-045, EPIC-004) намеренно хранит минимум, достаточный для состояния, а не полное описание задачи. Решение архитектора (чекпойнт): расширить `TaskProjection`/`TaskView` этими полями, заполняя их из данных события `TaskCreated` через `Envelope.WithData` — тот же механизм, что уже используется для `ReviewCompleted`'s `to` (TASK-044) — вместо прямого обращения `apps/api` к `TaskStore` (что нарушило бы ADR-014: `TaskProjection` — единственный путь чтения Task).
4. `internal/application/task_planning.go`: `CreateTask` прикрепляет `title`/`type`/`scope`/`acceptanceCriteria` (JSON-строкой — `WithData` несёт только `map[string]string`) к `TaskCreated` вместо использования общего `publish`-хелпера (у него нет доступа к `WithData`).
5. `internal/application/projection.go`: `TaskView` — новые поля; `Handle` — `applyCreatedData` (вынесенный хелпер) заполняет их **только** при `e.Type() == event.TaskCreated`, чтобы последующие события (`TaskPlanned` и т.д.) не затирали их пустыми значениями.
6. `apps/api/httpapi/tasks.go`: `taskViewResponse` — новые поля, общий конструктор `taskViewResponseFrom` для `handleGetTask`/`handleListTasks` (не дублировать сборку в двух местах).
7. `docs/api/tasks.md`: обе операции чтения задачи дополнены новыми полями ответа.
8. Тесты на всех трёх уровнях: `internal/application` (публикация данных, устойчивость к последующим переходам), `apps/api/httpapi` (HTTP-ответ содержит поля), `apps/dashboard` (обе страницы на моках `lib/api`).
9. Живая проверка вживую (полный стек — Postgres → `apps/api` → Dashboard): создание проекта и задачи через API (с диагностикой и обходом артефакта кодировки инструмента тестирования — inline `curl -d` на этой Windows-машине портит кириллицу, `--data-binary @file` — нет; сам API/приложение кириллицу не портят), Playwright — полный путь навигации (список проектов → задачи проекта → детали задачи) с проверкой `page.url()` на каждом шаге, все поля отображаются корректно; отдельно проверено `/projects/no-such-project/tasks/TASK-999` — читаемая ошибка вместо падения.
10. `make verify`, полный набор интеграционных тестов (`-tags=integration`) на реальном PostgreSQL — без регрессий, включая `TestGoldenPath_HTTP`.

## История

2026-07-22 — Architect — EPIC-009 открыт; задача поставлена в очередь.

2026-07-23 — Developer — задача взята в работу; обнаружен пробел в API (описательные поля задачи отсутствуют в `TaskProjection`) — зафиксирован чекпойнтом, решение получено (расширить проекцию); реализовано и проверено вживую (см. Отчёт).

## Отчёт о выполнении

### Задача

TASK-076 — страница проекта (список задач) и страница задачи (детали).

### Что сделано

- `src/app/projects/[id]/page.tsx` — список задач проекта со ссылками на детали; `src/app/projects/[id]/tasks/[taskId]/page.tsx` — детали задачи. Обе — `dynamic = "force-dynamic"`.
- **Расширение `TaskProjection` (по решению чекпойнта)**: `TaskView` дополнен `Title`/`Type`/`Scope`/`AcceptanceCriteria`, заполняемыми один раз из `TaskCreated` через `Envelope.WithData` — без обращения `apps/api` к `TaskStore` напрямую (сохраняет ADR-014: `TaskProjection` остаётся единственным путём чтения Task). `TaskPlanningService.CreateTask` прикрепляет эти данные при публикации `TaskCreated`.
- `apps/api/httpapi`: `GET /projects/{projectId}/tasks/{id}` и `GET /projects/{projectId}/tasks` теперь отдают все четыре новых поля; `docs/api/tasks.md` синхронизирован.
- Тесты на всех трёх уровнях (Application/inmemory затронуты не были, httpapi, dashboard) — новые сценарии для описательных полей, включая проверку, что поля переживают последующие переходы состояния (`PlanTask` не стирает `Title`).

### Изменённые файлы

- `apps/dashboard/src/app/projects/[id]/page.tsx`, `page.test.tsx`, `apps/dashboard/src/app/projects/[id]/tasks/[taskId]/page.tsx`, `page.test.tsx` (новые).
- `apps/dashboard/src/lib/api.ts` — `TaskView` дополнен новыми полями.
- `internal/application/task_planning.go`, `task_planning_test.go` — `WithData` для `TaskCreated`.
- `internal/application/projection.go`, `projection_test.go` — новые поля `TaskView`, `applyCreatedData`; `reflect.DeepEqual` вместо `!=` в тесте на пересборку (структура со срезом `[]string` больше не сравнима оператором).
- `apps/api/httpapi/tasks.go`, `tasks_test.go` — `taskViewResponseFrom`, новые поля ответа.
- `docs/api/tasks.md`, `internal/application/README.md` — синхронизированы.

### Как проверялось

- `go test ./... -tags=integration` (реальный PostgreSQL) — всё зелёное, включая `TestGoldenPath_HTTP`, без регрессий.
- `pnpm lint`/`format:check`/`test`/`build` (`apps/dashboard`) — чисто (8/8 тестов).
- Живая проверка вживую: реальный `apps/api` + PostgreSQL + Dashboard. Создан проект и две задачи (одна — намеренно с испорченной кириллицей из-за артефакта инструмента тестирования, не приложения — диагностировано и подтверждено сравнением `curl -d` inline vs `--data-binary @file`). Playwright: полный путь навигации с проверкой `page.url()` на каждом переходе — список проектов → задачи проекта → детали задачи, все поля (заголовок, тип, состояние, scope, критерии приёмки, время обновления) корректны, кириллица отображается верно. Отдельно проверено: несуществующие проект/задача — читаемая ошибка «Не удалось загрузить данные: ... 404 Not Found» вместо падения страницы.
- `make verify` (корень репозитория) — чисто.

### Обновлённая документация

- `docs/api/tasks.md`, `internal/application/README.md`.

### Open Questions

Нет.

### Рекомендации

- Оформление страницы деталей задачи (`<dl>/<dt>/<dd>` без визуального разделения меток и значений) — функционально корректно, но минимально; косметическое улучшение — по решению архитектора, не блокирует MVP наблюдения.
- TASK-078 (закрытие эпика) может сослаться на это расширение `TaskProjection` как на прецедент: любое будущее поле, которое должен увидеть Dashboard, добавляется тем же путём (через `Envelope.WithData` на создающем событии), а не прямым чтением хранилища.
