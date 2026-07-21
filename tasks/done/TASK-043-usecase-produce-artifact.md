# TASK-043: Use-case «Производство результата»

## Тип

feature

## Эпик

[EPIC-004 Application Layer](../../docs/roadmap/EPIC-004-application-layer.md)

## Цель

Use-case'ы работы с результатом исполнения: RecordDraftArtifact (создание Artifact в Draft в границе Project, привязка к Execution через RecordArtifact), PublishArtifact (Draft → Published), CompleteExecution (Succeed/Fail с финализацией множества артефактов); публикация `ArtifactCreated`/`ArtifactPublished` и `ExecutionSucceeded`/`ExecutionFailed`.

## Контекст

Golden path, шаг «пишет код → открывает Pull Request»: PR — это Artifact типа PullRequest, произведённый Execution (спецификация Artifact, Examples).

## Scope

### Входит

- Сервисы записи/публикации артефакта и завершения исполнения; согласование двух сторон связи (Artifact.ProducedBy ↔ Execution.RecordArtifact) в одном use-case.
- Тесты: успех; RecordArtifact вне Running; Publish без Payload; повторное завершение (гонка Fail/Abort — ErrTerminal).

### Не входит

- Реальный git/PR (RepositoryProvider используется в TASK-044 для merge); интерпретация Payload.

## Критерии приёмки

- [x] Artifact создаётся только в Active-проекте; ProducedBy = ID Execution; Execution ссылается на Artifact (двусторонняя согласованность — та же связь, один use-case).
- [x] События артефакта и исполнения публикуются с корректными полями конверта (Source=`artifact`/`execution`; ProjectID событий Execution — через lookup Task, у Execution нет прямой ссылки на Project, ADR-015).
- [x] Покрытие 82.6% (та же природа остатка, что в TASK-042 — защитные ветки вокруг фейков, которые не отказывают); `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-040/042

## План реализации

1. `internal/application/result.go` — `ResultService{Projects, Executions, Artifacts, Events}`:
   - `RecordDraftArtifact` — guard `Project.AcceptsNewContent`, guard `Execution.State == Running`, `artifact.New(..., producedBy=executionID)`, `Execution.RecordArtifact`, сохранение обоих агрегатов, публикация `ArtifactCreated`.
   - `UpdateArtifactDraft` — прямой проброс в `Artifact.UpdateDraft` + сохранение (для контента, добавляемого после Record).
   - `PublishArtifact` — `Artifact.Publish` + сохранение + публикация `ArtifactPublished`.
   - `SucceedExecution`/`FailExecution` — `Execution.Succeed`/`Fail` + сохранение + публикация `ExecutionSucceeded`/`ExecutionFailed`.
2. Тесты: полный цикл Record→UpdateDraft→Publish→Succeed; RecordArtifact вне Running; Publish без Payload; повторное Succeed после Fail (`ErrTerminal` — гонка уже разрешена в домене, use-case её не решает повторно); Project не Active; Execution/Artifact не найдены.
3. README, `make verify`, PR, CI, merge.

Известное ограничение TASK-042 (нет межагрегатной транзакции) распространяется и сюда (Artifact+Execution в RecordDraftArtifact) — не устраняется, только документируется рядом с уже существующей записью в README.

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (четвёртая).
2026-07-21 — Claude Code (Developer) — задача взята в работу, план записан.
2026-07-21 — Architect — план одобрен. Приступать.
2026-07-21 — Claude Code (Developer) — реализовано: `result.go` (`ResultService` с полями Projects/Tasks/Executions/Artifacts/Events; `Tasks` добавлен сверх плана — события завершения Execution требуют ProjectID, а у Execution нет прямой ссылки на Project, только TaskID, ADR-015, поэтому `publishExecutionEvent` делает lookup через Tasks); пять методов, как в плане. `result_test.go` — 17 тестов (успешные пути, обе причины отказа RecordDraftArtifact, Publish без Payload, полный цикл Record→UpdateDraft→Publish, гонка Succeed-после-Fail — домен уже решает, use-case не переигрывает, плюс полный набор «не найдено» для всех пяти методов). `make verify` — чисто.
2026-07-21 — Architect — Code Review: решение добавить Tasks в ResultService сверх исходного плана — оправдано (без него события Execution остались бы без ProjectID, что нарушило бы уже установленный в TASK-041/042 стандарт заполненности конверта); двусторонняя связь Artifact↔Execution проверена явным тестом на обе стороны (`ArtifactIDs()` у Execution, не только `ProducedBy()` у Artifact). Замечаний нет. Approve.
2026-07-21 — Architect — задача переведена в `tasks/done/`.

## Отчёт о выполнении

1. **Задача:** TASK-043 — use-case «Производство результата» (EPIC-004, четвёртая задача).
2. **Что сделано:** `ResultService` — запись/публикация Artifact (граница Active-проекта, только в Running Execution, двусторонняя связь ProducedBy↔ArtifactIDs) и завершение Execution (Succeed/Fail, гонка уже решена доменом). ProjectID событий Execution получен через lookup Task (Execution хранит только TaskID).
3. **Изменённые файлы:** `internal/application/{result,result_test}.go`, `internal/application/README.md`; файл задачи.
4. **Как проверялось:** `go test ./internal/application/... -cover` — 82.6%/92.9%; `make verify` — чисто.
5. **Обновлённая документация:** README `internal/application`.
6. **Open Questions:** нет новых; известное ограничение (нет межагрегатной транзакции) распространено на новый use-case, не переоткрывается как отдельный вопрос.
7. **Рекомендации:** TASK-044 (завершение задачи) — последний use-case, реализующий ADR-008 (порядок TestsPassed → MergeCompleted → TaskCompleted) в коде.
