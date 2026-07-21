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

- [ ] Artifact создаётся только в Active-проекте; ProducedBy = ID Execution; Execution ссылается на Artifact (двусторонняя согласованность — та же связь, один use-case).
- [ ] События артефакта и исполнения публикуются с корректными полями конверта (Source=`artifact`/`execution`).
- [ ] Покрытие ≥ 85%; `make verify` — чисто.

## Затрагиваемые модули и документы

- `internal/application/`.

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны — зависит от TASK-040/042

## План реализации

<Заполняется при взятии задачи в работу.>

## История

2026-07-21 — Architect — EPIC-004 открыт; задача поставлена в очередь (четвёртая).
