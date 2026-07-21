# TASK-050: RepositoryProvider — адаптер GitHub REST API

## Тип

feature

## Эпик

[EPIC-005 Infrastructure Layer](../../docs/roadmap/EPIC-005-infrastructure-layer.md)

## Цель

Реализовать `platform.RepositoryProvider` (`internal/platform/repository.go`) — единственный шлюз платформы к GitHub: CreateBranch, OpenPullRequest, RequestReview, MergePullRequest, ClosePullRequest, PullRequestState. Без новой библиотеки — `net/http` и GitHub REST API v3 напрямую (см. `stack.md`: `go-github` не входит в стек, вводить его — новая зависимость без обоснованной необходимости, тогда как узкий контракт из шести операций не требует полнофункционального клиента).

## Контекст

Контракт уже зафиксирован (`repo` — строка, формат `owner/name`, ADR-013 о managed-проектах не блокирует этот узкий контракт — см. EPIC-005 «Контекст»). Аутентификация — токен из переменной окружения (тот же паттерн, что использовался в этой сессии для `gh` через `GH_TOKEN`, но здесь — прямой HTTP, не обёртка над CLI).

## Scope

### Входит

- `internal/infrastructure/github/provider.go` — реализация шести методов `platform.RepositoryProvider` через GitHub REST API v3 (`https://api.github.com`), токен — заголовок `Authorization`.
- Маппинг состояний PR GitHub → `platform.PullRequestState` (open/merged/closed).
- Структурированные ошибки при неуспехе HTTP-запроса (код статуса, тело ответа) — не паника, не голое `fmt.Errorf` без контекста.
- Юнит-тесты на `httptest.Server` (без реального обращения к GitHub) — каждый метод, включая отказные ответы (404, 422, конфликт при мердже).
- README-раздел с описанием переменной окружения токена и ограничений (rate limit не обрабатывается отдельно в MVP — задокументировать как известное ограничение).

### Не входит

- Интеграционный тест против реального GitHub API (см. риски EPIC-005 — нет решения о хранении секрета в CI); ручная проверка автором задачи на тестовом репозитории документируется в отчёте, но не входит в автоматический `make verify`.
- Webhook-приём событий от GitHub — не в scope контракта `RepositoryProvider` (он — исходящий шлюз, не приёмник).

## Критерии приёмки

- [x] Все шесть методов реализованы, соответствуют `platform.RepositoryProvider` без изменения контракта.
- [x] Юнит-тесты на `httptest` покрывают успешный путь и минимум по одному отказному сценарию на метод.
- [x] Ошибки от GitHub API оборачиваются с достаточным контекстом для диагностики (код статуса, repo, операция) — `*APIError` (Method/Path/StatusCode/Body).
- [x] `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/infrastructure/github/` (новый), README `internal/infrastructure`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — независима от TASK-046…049 (не использует PostgreSQL), может выполняться параллельно

## План реализации

1. `internal/infrastructure/github/provider.go` — `Provider{baseURL, token, client}`, `New()` (токен из `GITHUB_TOKEN`) / `NewWithToken(token)`; общий `do(ctx, method, path, body, out)` — маршалит тело, шлёт заголовки (`Authorization: Bearer`, `Accept: application/vnd.github+json`, `X-GitHub-Api-Version`), декодирует ответ; `*APIError{Method, Path, StatusCode, Body}` при статусе ≥300.
2. Два места, где контракт не полностью однозначен — решения зафиксировать явно (не как молчаливое предположение):
   - `OpenPullRequest` не принимает целевую ветку → doc-комментарий контракта фиксирует её как «main» текстом; реализовано буквально.
   - `RequestReview` не принимает личность ревьюера → ADR-008 говорит, что обязательность ревью обеспечивается стадией Review workflow, а не настройкой GitHub, — значит, нативный GitHub reviewer здесь не нужен; реализовано как публикация комментария к PR (видимый сигнал, не выдуманный параметр).
3. Шесть методов: `CreateBranch` (GET ref → POST refs), `OpenPullRequest` (POST pulls, base="main", возвращает `number` как строку), `RequestReview` (POST issue comment), `MergePullRequest` (PUT merge, `merge_method: "merge"` — ADR-008), `ClosePullRequest` (PATCH pulls, `state: "closed"`), `PullRequestState` (GET pulls → merged/state → `platform.PullRequestState`).
4. Тесты на `httptest.Server`: успешный путь + минимум один отказной сценарий на метод (404/422/405), проверка заголовков запроса, проверка тела запроса (`merge_method`, `base`, `state`).
5. README `internal/infrastructure` — раздел `github`, включая явную фиксацию обоих решений по неоднозначности контракта.
6. `make verify`, PR, CI, merge.

## История

2026-07-21 — Architect — EPIC-005 открыт; задача поставлена в очередь (пятая; не зависит от Postgres-задач).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан, включая два явных решения по неоднозначности контракта (целевая ветка PR; отсутствие личности ревьюера).
2026-07-21 — Architect — план одобрен. Оба решения корректны: doc-комментарий контракта действительно фиксирует "main" текстом (не домысел), а ADR-008 действительно снимает необходимость нативного GitHub-ревьюера. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано по плану. 89.6% покрытия юнит-тестами на `httptest`, без единого обращения к реальному GitHub. Ручная проверка против настоящего репозитория НЕ выполнена (нужен тестовый репозиторий и PAT — не настроены в этой сессии); зафиксировано как Open Question.
2026-07-21 — Architect — Code Review: узкий общий `do()` — оправданное упрощение (шесть методов, одна и та же обвязка); `*APIError` как самостоятельный тип (не sentinel) — правильно, так как несёт контекст (Method/Path/StatusCode/Body), а не просто факт ошибки; решение про `RequestReview` как комментарий — обоснованно ссылкой на ADR-008, не выдумано. Замечаний нет. Approve. Открытый вопрос (ручная проверка на реальном GitHub) — не блокирует эту задачу, отмечен для TASK-051/будущего.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-050 — `RepositoryProvider` через GitHub REST API (пятая задача EPIC-005).
2. **Что сделано:** `internal/infrastructure/github.Provider` реализует все шесть операций `platform.RepositoryProvider` напрямую через `net/http` (без клиентской библиотеки); токен — `GITHUB_TOKEN`; структурированные ошибки — `*APIError`. Два места неоднозначности контракта разрешены явно и обоснованно (см. План): целевая ветка PR всегда «main» (буквально по doc-комментарию контракта); `RequestReview` реализован как комментарий к PR, а не назначение GitHub-ревьюера (ADR-008: обязательность ревью — на уровне workflow, не GitHub).
3. **Изменённые файлы:** `internal/infrastructure/github/{doc.go,provider.go,provider_test.go}` (новые); `internal/infrastructure/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/infrastructure/github/... -cover` — 89.6%, все тесты на `httptest.Server` (успех + отказной сценарий на каждый из шести методов); `make verify` — чисто.
5. **Обновлённая документация:** README `internal/infrastructure` (раздел `github`, с явной фиксацией обоих решений по неоднозначности контракта).
6. **Open Questions:** ручная проверка адаптера против настоящего GitHub API не выполнена в этой сессии (требует тестового репозитория и PAT — не настроены); интеграционный прогон против реального GitHub не входит в scope этой задачи (см. риски EPIC-005 — нет решения о хранении секрета в CI). Рекомендуется выполнить перед использованием адаптера в реальном golden path.
7. **Рекомендации:** TASK-051 (composition root) может собрать все адаптеры вместе; интеграционный golden-path тест — на in-memory/Postgres-адаптерах для остального, `RepositoryProvider` для реального прогона потребует отдельного решения человека о секрете.
