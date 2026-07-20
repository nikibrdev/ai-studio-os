# TASK-029: Спецификация домен-модуля Artifact

## Тип

docs

## Эпик

[EPIC-003 Domain Layer](../../docs/roadmap/EPIC-003-domain-layer.md), этап 1 (Domain Specifications First)

## Цель

Полная, утверждённая спецификация `docs/specifications/domain/artifact.md` по шаблону [Specification-Domain.md](../../.claude/templates/Specification-Domain.md) (19 обязательных разделов — [Domain Specification Review](../../engineering/decisions/2026-07-20-domain-specification-review.md), [Three-Pass Review](../../engineering/decisions/2026-07-20-domain-specification-three-pass-review.md), [Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)) — техническое задание для будущей реализации `internal/domain/artifact` (этап 2, отдельная задача, не начинается без утверждения этой спецификации). Пишется тремя отдельными PR (фундамент → поведение → завершение), см. «План реализации».

## Контекст

Artifact — первый модуль в порядке проектирования Domain Layer ([domain-model.md](../../docs/architecture/domain-model.md), [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md)): самостоятельный Aggregate Root, не часть Execution/Task/Project. Концептуальное описание уже есть в ADR-016 и domain-model.md — задача переводит его в полную спецификацию по всем 19 разделам Specification-Domain.md, а не изобретает решение заново. Как первая спецификация Domain Layer, PR 1 этой задачи задаёт прецедент для TASK-030…033.

## Scope

### Входит

- `docs/specifications/domain/artifact.md`, все 19 обязательных разделов, тремя PR:
  - **PR 1 — фундамент**: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle, Relationships (владение/ссылки/создание/удаление, не определяет сущность), Alternative Interpretations Considered. Ни одного упоминания Go.
  - **PR 2 — поведение**: Domain Events, Commands, Queries, Examples (не код — содержательные примеры конкретных Artifact).
  - **PR 3 — завершение и ревью**: Acceptance Criteria, Future Extensions, Anti-Responsibilities, Non-Goals, Removal Test, Decision Log (ADR-016 и другие решения), Open Questions; три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md).
- Явное описание разделения Metadata/Payload (по ADR-016) в разделах Responsibilities/Invariants — что обязано хранить ядро (Metadata), что не интерпретирует (Payload).
- Обновление `internal/domain/README.md` (при необходимости) — ссылка на новую спецификацию, без изменения кода.

### Не входит

- Реализация Go-пакета `internal/domain/artifact` — отдельная задача этапа 2, после утверждения.
- Спецификации Execution/Executor/Task/Project — отдельные задачи (TASK-030…033).

## Критерии приёмки

- [x] Спецификация содержит все 19 обязательных разделов Specification-Domain.md.
- [x] Пройдены три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md) (Internal Consistency, Cross-domain Consistency, Future-proof Review) — с письменными ответами на диагностические вопросы (Open Questions).
- [x] Непротиворечива с ADR-016, ADR-005, domain-model.md.
- [x] Статус спецификации — «Reference» (после явного подтверждения архитектора; выше, чем «Утверждена» — первая эталонная доменная спецификация проекта).
- [x] `bash scripts/verify-docs.sh`, `npx markdownlint-cli2` — чисто.

## Затрагиваемые модули и документы

`docs/specifications/domain/artifact.md` (новый).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## План реализации

PR 1 («фундамент») пишется в порядке One Sentence → Identity → Purpose → Responsibilities → Invariants (Structural/Behavioral) → Lifecycle → Relationships → Alternative Interpretations Considered ([Model First](../../engineering/decisions/2026-07-20-domain-specification-model-first.md)). Ни одного упоминания Go/структур/реализации. Источники: [ADR-016](../../docs/adr/ADR-016-artifact-aggregate-root.md) (Aggregate Root, Metadata/Payload, невладение со стороны Execution) и [domain-model.md](../../docs/architecture/domain-model.md) — используются как материал, а не переписываются дословно; там, где ADR-016 не даёт прямого ответа (точный набор состояний Lifecycle, полный список инвариантов, устройство Metadata/Payload на уровне спецификации), формулировка предлагается и явно помечается как черновая, для утверждения архитектором, а не выдаётся за решённый вопрос.

Статус документа после PR 1 — «Черновик» (полный набор разделов появится только к PR 3). Открытые вопросы, если возникнут в процессе (например, по поводу Lifecycle или конкретных инвариантов), фиксируются как Open Questions прямо в PR, а не решаются самостоятельно.

После утверждения PR 1 — PR 2 (Domain Events, Commands, Queries, Examples), затем PR 3 (Acceptance Criteria, Future Extensions, Anti-Responsibilities, Non-Goals, Removal Test, Decision Log, Open Questions + три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md)).

## История

2026-07-20 — Architect — EPIC-003 открыт в режиме Domain Specifications First; задача поставлена в очередь (первая по порядку проектирования).
2026-07-20 — Architect — введён Domain Specification Review (12 обязательных разделов, Specification-Domain.md); задача синхронизирована с новым шаблоном перед стартом.
2026-07-20 — Architect — введён Three-Pass Review (+4 раздела до 16, три прохода проверки, сознательный темп — несколько PR допустимы); задача официально открыта к работе.
2026-07-20 — Architect — утверждён трёхэтапный процесс (PR 1/2/3); введён Model First — PR 1 начинается с One Sentence/Identity, Invariants разделены на Structural/Behavioral, добавлен Alternative Interpretations Considered (итого 19 разделов); сформулирован главный принцип EPIC-003 — цель не закончить спецификацию, а исключить неправильные трактовки.
2026-07-20 — Claude Code (Developer) — задача взята в работу; план PR 1 записан; начинается написание содержания.
2026-07-20 — Claude Code (Developer) — PR 1 (`docs/specifications/domain/artifact.md`, черновик) написан: One Sentence, Identity, Purpose, Responsibilities, Invariants (Structural/Behavioral), Lifecycle, Relationships, Alternative Interpretations Considered. Локальная верификация пройдена. В отличие от предыдущих PR этой сессии, содержание не продиктовано архитектором дословно — это самостоятельная доменная модель, требующая реального ревью, а не самоапрува; PR отправлен на рассмотрение, merge не выполняется до обратной связи.
2026-07-20 — Architect — первое содержательное ревью PR 1: **Changes Requested** (не отклонение — направление признано сильным). Замечания: (1) One Sentence слишком завязан на «содержание», нужен акцент на «долговременный инженерный результат»; (2) Identity — критерий «можно открыть и прочитать» не универсален (бинарники, архивы, Figma); (3) Metadata как инвариант нужно уточнить минимально обязательным набором полей; (4) Author и Origin — разные понятия, развести; (5) Lifecycle — проверить, является ли Produced состоянием или событием (тест: может ли длиться часы/дни/недели); (6) Relationships — без замечаний; (7) Alternative Interpretations — усилить вариантом «Artifact как производное Event»; (8) Open Questions — добавить кардинальность Execution↔Artifact; (9) главное — не хватает явного ответа, зачем предметной области вообще нужен Artifact (переиспользование, ревью, релизы, накопление знаний, трассируемость).
2026-07-20 — Claude Code (Developer) — все девять замечаний обработаны: One Sentence и Identity переписаны без «контента» и без критерия читаемости; Structural Invariants уточнены минимальным набором полей (Identifier/Type/CreatedAt/Author/Origin), Author и Origin разведены как отдельные понятия; Lifecycle упрощён до Draft → Published → Archived — Produced исключён с явным объяснением через тест состояние-или-событие; в Purpose добавлено объяснение практической ценности (переиспользование/ревью/релизы/знания/трассируемость); Alternative Interpretations дополнен вариантом «производное Event»; кардинальность Execution↔Artifact зафиксирована как Behavioral Invariant (не просто открытый вопрос) с пояснением вывода. Локальная верификация пройдена повторно, PR обновлён новым коммитом (не amend).
2026-07-20 — Architect — второе ревью: **Approve with one non-blocking architectural note**. Согласие с исключением Produced, но по более глубокой причине (Produced описывал степень доверия процесса к артефакту, а не состояние самого Artifact — это ответственность Review/Workflow/Task, не Artifact). Уточнено значение Draft («не признан частью инженерной истории», не «не закончен») и потребован явный инвариант «Archived ≠ удалён». Единственное замечание к смержу: инвариант «не более одного producing Execution» — не делать жёстким доменным инвариантом v1, так как это может закрыть путь к будущей модели истории изменений (например, если один Artifact дополняют несколько разных Execution/Executor последовательно); смягчить до «ссылка на первое вводящее Execution» и зафиксировать кандидата на будущую сущность Revision. Разрешение смержить после смягчения.
2026-07-20 — Claude Code (Developer) — инвариант о producing Execution смягчён (ссылка только на первое вводящее Execution, вопрос истории нескольких Execution явно оставлен открытым, не запрещён); добавлен явный инвариант «Archived не означает удалённый»; Draft переопределён через «не признан частью инженерной истории проекта»; будущий кандидат Revision зафиксирован в Open Questions. Локальная верификация пройдена, PR обновлён.
2026-07-20 — Claude Code (Developer) — PR 1 смержен (PR #34), задача остаётся в работе. Начат PR 2 («поведение»): Domain Events (ArtifactCreated/ArtifactPublished/ArtifactArchived — только на переходах Lifecycle, правки Draft событий не порождают; Artifact не потребляет чужие события), Commands (Create/UpdateDraft/Publish/Archive — Type и Origin зафиксированы уже при Create, не входят в UpdateDraft), Queries (по Identifier/Project/Type/Author/вводящему Execution/состоянию), Examples (Markdown-спецификация, Pull Request, Figma-дизайн, Test Report, ADR — каждый с явными Type/Origin/Author). По ходу работы уточнена Structural Invariant 3 (Origin, как и Type, фиксируется при создании) — малая правка ради внутренней согласованности с новым разделом Commands, не новое решение. Локальная верификация пройдена. Содержание вновь не продиктовано архитектором — PR отправлен на рассмотрение, merge не выполняется до обратной связи.
2026-07-20 — Architect — ревью PR 2: **Approve with comments**. Замечания: (1) «Draft edits don't emit events» сформулировано слишком жёстко — смешивает внутренние доменные факты и публичные интеграционные события, ослабить до «только переходы Lifecycle порождают обязательные публичные Domain Events»; (2) «Artifact never consumes events» — убрать из спецификации целиком: это архитектурное правило о взаимодействии модулей, а не природа Artifact (пример на будущее: Git Import → ArtifactImported → Artifact); (3) Commands — без замечаний, хороший пример «инварианты как следствие команд»; (4) Queries — терминология «introducing Execution» заменить на «producing»/«originating» (связь про происхождение, не про последующие изменения); (5) пример Pull Request — отметить, что после merge неизменяем именно набор изменений, а не сопутствующая коммуникация; (6) подумать (не реализовывать) о будущем сценарии производных артефактов (Clone/Derive); (7) главное — зафиксировать в Open Questions мысль о будущей модели авторитетности (Authority/Approval), вокруг которой уже сейчас вращаются Publish/Draft/immutability/Review, не вводя эту сущность сейчас.
2026-07-20 — Claude Code (Developer) — все замечания обработаны: формулировка о Draft-событиях смягчена; утверждение о непотреблении событий удалено; терминология унифицирована на «породившее Execution» (Behavioral Invariant 3, Relationships, ArtifactPublished, Queries, Open Questions); пример Pull Request дополнен уточнением про коммуникацию вне Artifact; в Open Questions добавлены два новых кандидата — производные артефакты (Clone/Derive) и модель авторитетности (Authority/Approval). Локальная верификация пройдена, PR обновлён.
2026-07-20 — PR 2 смержен (PR #35). Architect — задал рамку для PR 3: не «хвост» спецификации, а экзамен модели («выдержит ли развитие проекта в течение нескольких лет?»). Детальные требования: Acceptance Criteria — критерии качества модели, не факта заполнения разделов; Future Extensions — разделить на «ожидаемые» и «возможные»; Anti-Responsibilities — максимально строгий явный список (включая «не определяет политику публикации»); Non-Goals — как сознательные ограничения версии, не список отсутствующих функций; Removal Test — короткий, один абзац; Decision Log — таблица «решение/основание», не список ADR; каждый из трёх проходов DomainSpecificationReview.md должен завершаться письменным ответом на диагностический вопрос, а не статусом «OK». Также анонсировано (вступит в силу только после утверждения TASK-029): правило Delta Review для TASK-030 и далее — ревью новой спецификации должно явно проверять, не вынуждает ли она пересматривать уже утверждённые.
2026-07-20 — Claude Code (Developer) — PR 3 написан по всем требованиям: Acceptance Criteria (7 критериев с самопроверкой), Future Extensions (Ожидаемые: Revision/Authority/Clone-Derive/новые типы Payload; Возможные: ACL/совместное редактирование/распределённое хранение/цифровые подписи/репликация), Anti-Responsibilities (5 пунктов, включая «не определяет политику публикации»), Non-Goals (5 пунктов как сознательные ограничения версии), Removal Test (один абзац), Decision Log (таблица, 11 строк), Open Questions сокращён до трёх подлинно открытых вопросов (кандидаты на расширение перенесены в Future Extensions) плюс письменные ответы на все три диагностических вопроса Three-Pass Review (самое слабое место — незакрытый перечень Origin; сильнее всего влияющая сущность — Execution; наиболее вероятный пересмотр — появление Revision). Обновлён .claude/checklists/DomainSpecificationReview.md — каждый Pass теперь требует письменного ответа, не только чек-бокс. Статус документа — «Черновик», ожидает финального решения архитектора об утверждении (не выставлен самостоятельно). Локальная верификация пройдена, PR отправлен на рассмотрение.
2026-07-20 — Architect — **Final Architecture Review: Approve** (не «Approve with comments»). Единственный необязательный комментарий: добавить раздел Stability Assessment (Current assessment / Confidence / Known revision triggers) как самый последний раздел документа — не основание задерживать merge. После merge — три решения: (1) статус **Reference** — четвёртый статус документа, «эта спецификация — образец для следующих», Artifact заслуживает его; (2) утверждённые/Reference спецификации после этого меняются только через новый ADR, отдельный Domain Revision PR или обоснованное изменение с Delta Review — не напрямую; (3) **Delta Review** для TASK-030 и далее — три вопроса относительно всех утверждённых спецификаций (требует ли пересмотра существующей модели; использует ли принятые понятия единообразно; не вводит ли дублирующее понятие под другим именем). Также предложено зафиксировать в ROADMAP.md внутренний архитектурный рубеж «First Reference Domain Specification» — не пользовательскую версию.
2026-07-20 — Claude Code (Developer) — Stability Assessment добавлен как 20-й, финальный раздел шаблона Specification-Domain.md и заполнен для Artifact (Stable for v1, Confidence: High, известные триггеры пересмотра — Revision/история Execution/Authority). Статус документа выставлен в **Reference**, с пояснением защищённого процесса изменения. Создано решение [2026-07-20-domain-specification-reference-status.md](../../engineering/decisions/2026-07-20-domain-specification-reference-status.md), объединяющее все три «после merge» решения. DomainSpecificationReview.md дополнен блоком Delta Review в Pass 2 (обязателен с TASK-030). ROADMAP.md дополнен разделом «Внутренние архитектурные вехи» с записью First Reference Domain Specification. EPIC-003, specifications/README.md, PROJECT_MANIFEST.md, CHANGELOG.md синхронизированы. Локальная верификация пройдена, PR обновлён — готов к merge по итогам Approve.

## Отчёт о выполнении

1. **Задача:** TASK-029 — спецификация домен-модуля Artifact, первая спецификация Domain Layer (EPIC-003, этап 1).
2. **Что сделано:** написана и утверждена полная спецификация `docs/specifications/domain/artifact.md` — 20 разделов, тремя отдельными PR (фундамент → поведение → завершение), каждый прошёл содержательное ревью архитектора (не самоапрув) с реальными правками между раундами. Итоговый статус — **Reference**: первая доменная спецификация проекта, официально признанная образцом для последующих. По ходу работы возникли и были формализованы шесть уровней методологии, применимых ко всем будущим спецификациям Domain Layer: Domain Specification Review (12 разделов), Three-Pass Review (+4 раздела, три прохода проверки), Model First (+3 раздела, структура PR 1, главный принцип), и Reference Status (+1 раздел Stability Assessment, статус Reference, защита утверждённых спецификаций, Delta Review для TASK-030+).
3. **Изменённые файлы:**
   - `docs/specifications/domain/artifact.md` (новый, статус Reference)
   - `.claude/templates/Specification-Domain.md` (создан и трижды расширен по ходу работы)
   - `.claude/checklists/DomainSpecificationReview.md` (создан и дважды расширен)
   - `engineering/decisions/2026-07-20-domain-specification-review.md`
   - `engineering/decisions/2026-07-20-domain-specification-three-pass-review.md`
   - `engineering/decisions/2026-07-20-domain-specification-model-first.md`
   - `engineering/decisions/2026-07-20-domain-specification-reference-status.md`
   - `docs/roadmap/EPIC-003-domain-layer.md`
   - `docs/specifications/README.md`
   - `ROADMAP.md` (новый раздел «Внутренние архитектурные вехи»)
   - `PROJECT_MANIFEST.md`, `CHANGELOG.md`
   - `internal/domain/README.md` (ссылка на новую спецификацию)
4. **Как проверялось:** на каждом из трёх PR — `go build/vet/gofumpt/golangci-lint` (чисто, Go-код не менялся), `bash scripts/verify-docs.sh` (0 ошибок), `npx markdownlint-cli2` (0 issues); финально — три прохода [DomainSpecificationReview.md](../../.claude/checklists/DomainSpecificationReview.md) с письменными ответами на диагностические вопросы каждого прохода (зафиксированы в Open Questions спецификации).
5. **Обновлённая документация:** см. «Изменённые файлы» — задача целиком документационная, реализация Go (`internal/domain/artifact`) — отдельная задача этапа 2, не входит в scope.
6. **Open Questions:** три подлинно открытых вопроса зафиксированы в самой спецификации (полный перечень значений Origin; история участия нескольких Execution в одном Artifact; полный состав Metadata сверх минимума) — не решены самостоятельно, требуют решения архитектора при дальнейшей работе (спецификация Execution, TASK-030, или отдельный ADR).
7. **Рекомендации:** приступить к TASK-030 (спецификация Execution) с учётом установленного эталона — включая обязательный Delta Review относительно Artifact; учитывать в Relationships/Invariants Execution, что связь с Artifact уже зафиксирована как мягкая (ссылка на породившее Execution, не владение).
