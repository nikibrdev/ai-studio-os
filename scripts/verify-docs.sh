#!/usr/bin/env bash
# Проверка документации AI Studio OS:
#   1) все относительные markdown-ссылки указывают на существующие файлы/каталоги;
#   2) mermaid-блоки корректно ограждены и начинаются с известного типа диаграммы.
# Используется целью `make docs-check` и CI (make verify).
set -u
cd "$(dirname "$0")/.."

errors=0
total_links=0
mermaid_blocks=0

# --- 1. Относительные ссылки ---
while IFS= read -r f; do
	dir=$(dirname "$f")
	while IFS= read -r link; do
		target="${link#*](}"
		target="${target%)}"
		case "$target" in
		http* | \#* | mailto*) continue ;;
		esac
		target="${target%%#*}"
		[ -z "$target" ] && continue
		total_links=$((total_links + 1))
		if [ ! -e "$dir/$target" ]; then
			echo "BROKEN LINK: $f -> $target"
			errors=$((errors + 1))
		fi
	done < <(grep -oE '\[[^]]*\]\([^)]+\)' "$f" 2>/dev/null || true)
done < <(find . -name "*.md" -not -path "*/node_modules/*" -not -path "*/.next/*" -not -path "./.git/*")

# --- 2. Mermaid-блоки ---
known_types='flowchart|graph|sequenceDiagram|classDiagram|stateDiagram-v2|erDiagram|gantt|pie|journey|timeline|mindmap'
while IFS= read -r f; do
	fences=$(grep -c '^```' "$f")
	if [ $((fences % 2)) -ne 0 ]; then
		echo "UNBALANCED FENCES: $f ($fences)"
		errors=$((errors + 1))
	fi
	while IFS= read -r first; do
		mermaid_blocks=$((mermaid_blocks + 1))
		if ! echo "$first" | grep -qE "^[[:space:]]*($known_types)"; then
			echo "UNKNOWN MERMAID TYPE: $f -> '$first'"
			errors=$((errors + 1))
		fi
	done < <(grep -A1 '^```mermaid' "$f" | grep -v '^```mermaid' | grep -v '^--$' | grep -v '^$' || true)
done < <(grep -rl '```mermaid' --include="*.md" --exclude-dir=node_modules --exclude-dir=.next --exclude-dir=.git . 2>/dev/null || true)

echo "docs-check: ссылок проверено: $total_links; mermaid-блоков: $mermaid_blocks; ошибок: $errors"
[ "$errors" -eq 0 ]
