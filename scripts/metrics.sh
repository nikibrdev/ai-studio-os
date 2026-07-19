#!/usr/bin/env bash
# Снимок инженерных метрик AI Studio OS -> engineering/metrics/YYYY-MM-DD.md
# Запуск: make metrics (или bash scripts/metrics.sh)
# Собирает автоматизируемые показатели; показатели, требующие истории PR и
# Task Engine (время задачи, замечания ревью), помечаются как "нет данных".
set -u
cd "$(dirname "$0")/.."

today=$(date +%F)
out="engineering/metrics/${today}.md"

# --- ADR ---
adr_total=0
adr_accepted=0
adr_dr=0
for f in docs/adr/ADR-0*.md; do
	case "$f" in *ADR-000*) continue ;; esac
	adr_total=$((adr_total + 1))
	if grep -q '\*\*Принято\*\*' "$f"; then adr_accepted=$((adr_accepted + 1)); fi
	if grep -q '\*\*Decision Required\*\*' "$f"; then adr_dr=$((adr_dr + 1)); fi
done

# --- Задачи по стадиям ---
task_row() {
	find "tasks/$1" -maxdepth 1 -name "TASK-*.md" 2>/dev/null | wc -l | tr -d ' '
}
t_backlog=$(task_row backlog)
t_ready=$(task_row ready)
t_inprogress=$(task_row in-progress)
t_review=$(task_row review)
t_blocked=$(task_row blocked)
t_done=$(task_row done)
t_archive=$(task_row archive)
t_total=$((t_backlog + t_ready + t_inprogress + t_review + t_blocked + t_done + t_archive))

# --- Пакеты Go и покрытие документацией ---
pkg_dirs=$(find internal pkg apps agents tools -type f -name "*.go" 2>/dev/null | xargs -r -n1 dirname | sort -u)
pkg_count=$(echo "$pkg_dirs" | grep -c . || true)
pkg_with_readme=0
for d in $pkg_dirs; do
	[ -f "$d/README.md" ] && pkg_with_readme=$((pkg_with_readme + 1))
done

# --- Спецификации ---
spec_count=$(find docs/specifications -mindepth 2 -name "*.md" 2>/dev/null | wc -l | tr -d ' ')

# --- TODO ---
todo_count=$(grep -rn "TODO" --include="*.go" internal pkg apps agents tools 2>/dev/null | wc -l | tr -d ' ')

# --- Документы ---
md_count=$(find . -name "*.md" -not -path "./.git/*" -not -path "./node_modules/*" | wc -l | tr -d ' ')

mkdir -p engineering/metrics
cat >"$out" <<EOF
# Метрики: ${today}

## Назначение

Автоматический снимок инженерных метрик проекта (scripts/metrics.sh).

## Содержание

### ADR

| Показатель | Значение |
|---|---|
| Всего ADR | ${adr_total} |
| Принято | ${adr_accepted} |
| Decision Required | ${adr_dr} |

### Задачи

| Стадия | Кол-во |
|---|---|
| backlog | ${t_backlog} |
| ready | ${t_ready} |
| in-progress | ${t_inprogress} |
| review | ${t_review} |
| blocked | ${t_blocked} |
| done | ${t_done} |
| archive | ${t_archive} |
| **всего** | **${t_total}** |

### Код и документация

| Показатель | Значение |
|---|---|
| Go-пакетов | ${pkg_count} |
| Пакетов с README | ${pkg_with_readme} из ${pkg_count} |
| Спецификаций в docs/specifications | ${spec_count} |
| TODO в коде | ${todo_count} |
| Markdown-документов всего | ${md_count} |

### Показатели, ожидающие данных

| Показатель | Источник |
|---|---|
| Среднее время задачи | История PR (после GitHub) и события Task Engine |
| Среднее число замечаний ревью | История PR |

## Статус

Снимок (не редактируется)

## Последнее обновление

${today}
EOF

echo "metrics: снимок записан в ${out}"
