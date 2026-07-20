# BUGFIX-002: Закрепить golangci-lint под Go 1.24 и запретить скрытый auto-toolchain

## Тип

fix

## Эпик

Вне эпика — срочное исправление; продолжение BUGFIX-001, обнаружено при обработке Dependabot PR #11.

## Цель

CI действительно собирает и линтит проект на Go 1.24 (как заявлено в ADR-009), без скрытой подмены тулчейна; повторный провал PR #11 после BUGFIX-001 устранён.

## Контекст

BUGFIX-001 закрыл только половину проблемы. Реальная причина глубже: у нас всегда было `GOTOOLCHAIN=auto` (умолчание Go), и когда инструмент требовал более новый Go, чем 1.24, тулчейн **тихо** докачивался и подменялся — это видно в логе успешного (!) прогона TASK-013:

```
go: mvdan.cc/gofumpt@v0.10.0 requires go >= 1.25.0; switching to go1.25.12
go: github.com/golangci/golangci-lint/v2@v2.12.2 requires go >= 1.25.0; switching to go1.25.12
```

То есть **golangci-lint v2.12.2 требовал Go 1.25 с самого TASK-013** — просто это маскировалось auto-toolchain, и наш CI всё это время фактически собирал и линтил на Go 1.25, а не на заявленном в ADR-009 Go 1.24. `actions/setup-go@v7` (Dependabot PR #11) меняет умолчание на `GOTOOLCHAIN=local`, что и вскрыло реальное расхождение — обе проблемы (gofumpt и golangci-lint) были следствием одной и той же маскировки, я исправил только первую.

Проверено через Go module proxy (`.mod` файлы релизов): `golangci-lint v2.8.0` — последняя версия, требующая `go 1.24.0`; `v2.9.0+` требует `go 1.25.0`.

Не поднимаю Go до 1.25 — это означало бы негласно отступить от принятого [ADR-009](../../docs/adr/ADR-009-toolchain.md); версия языка — решение архитектора, не агента.

## Scope

### Входит

- `.github/workflows/verify.yml`: `golangci-lint@v2.12.2` → `@v2.8.0`.
- `env: GOTOOLCHAIN: local` на уровне job — чтобы будущее несоответствие версии инструмента и Go **падало явно**, а не маскировалось auto-download.

### Не входит

- Изменение версии Go в ADR-009.
- Аудит остальных `npx --yes` зависимостей (markdownlint-cli2, js-yaml) на ту же категорию риска — отдельная задача (упомянута в [engineering/decisions/2026-07-20-pin-ci-tool-versions.md](../../engineering/decisions/2026-07-20-pin-ci-tool-versions.md)).

## Критерии приёмки

- [ ] `golangci-lint v2.8.0` собирается и проходит на репозитории локально с `GOTOOLCHAIN=local` (Go 1.24.5).
- [ ] CI зелёный на этом PR (первый честный прогон на реальном Go 1.24 без маскировки).
- [ ] Dependabot PR #11 (setup-go bump) после ребейза тоже проходит.

## План реализации

1. Локально: `GOTOOLCHAIN=local go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0`; прогнать на репозитории.
2. `verify.yml`: пин версии + `env: GOTOOLCHAIN: local` на уровне job (виден в каждом шаге, не только том, что его выставляет неявно).
3. Проверить `make verify` целиком локально с `GOTOOLCHAIN=local`.

## Затрагиваемые модули и документы

- `.github/workflows/verify.yml`; дополнение к [engineering/decisions/2026-07-20-pin-ci-tool-versions.md](../../engineering/decisions/2026-07-20-pin-ci-tool-versions.md).

## Definition of Ready

- [x] Цель и результат сформулированы
- [x] Критерии приёмки определены
- [x] Затрагиваемые модули/документы указаны
- [x] Ограничения и зависимости указаны

## История

2026-07-20 — Claude Code (Developer) — обнаружено при повторном провале Dependabot PR #11 после BUGFIX-001; корневая причина установлена через сравнение логов старого (успешного, но с маскировкой) и нового (честного) прогонов; версия golangci-lint найдена через Go module proxy.

## Отчёт о выполнении

1. **Задача:** BUGFIX-002 — закрепить golangci-lint под Go 1.24 и запретить скрытый auto-toolchain.
2. **Что сделано:**
   - `golangci-lint@v2.12.2` → `@v2.8.0` в `verify.yml` (последняя версия с `go 1.24.0`, подтверждено через Go module proxy).
   - `env: GOTOOLCHAIN: local` на уровне job — реальный Go-тулчейн в CI больше не может незаметно подменяться.
   - Побочная находка при верификации: `golangci-lint v2.8.0` (в отличие от `v2.12.2`) поймал `revive`-предупреждение `var-naming: avoid meaningless package names` на `package shared` — вероятно, из-за разницы дефолтных наборов правил между мажорными версиями golangci-lint. Название пакета `shared` — осознанный, прошедший ревью выбор ([ADR-015](../../docs/adr/ADR-015-internal-layering.md)), поэтому не переименовывал пакет, а добавил точечный `//nolint:revive` с обоснованием на самой строке `package`.
3. **Изменённые файлы:** `.github/workflows/verify.yml`, `internal/domain/shared/types.go` (nolint-комментарий), `engineering/decisions/2026-07-20-pin-ci-tool-versions.md` (дополнение), файл задачи.
4. **Как проверялось:** локально — `GOTOOLCHAIN=local go install .../golangci-lint@v2.8.0` (собрался как `golangci-lint 2.8.0 built with go1.24.5` — честная сборка на 1.24, без подмены); `GOTOOLCHAIN=local golangci-lint run ./...` — 0 issues; `gofumpt -l .`, `go build ./...`, `go vet ./...` — чисто. В CI — следующий PR.
5. **Обновлённая документация:** `engineering/decisions/2026-07-20-pin-ci-tool-versions.md`.
6. **Open Questions:** нет технических. Процессное наблюдение: два инцидента подряд (gofumpt, затем golangci-lint) из-за незакреплённого тулчейна — показывают, что `GOTOOLCHAIN=local` стоило внести ещё в TASK-013; теперь внесено, риск для будущих инструментов закрыт системно, а не точечно.
7. **Рекомендации:** после merge — дождаться повторного прогона Dependabot PR #11 (ребейз уже стоял в очереди) и смержить его как финальное подтверждение исправления.
