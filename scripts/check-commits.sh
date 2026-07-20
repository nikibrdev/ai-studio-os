#!/usr/bin/env bash
# Проверка Conventional Commits для диапазона коммитов PR.
# Используется шагом CI (.github/workflows/verify.yml, событие pull_request);
# можно запускать вручную: bash scripts/check-commits.sh <base> <head>
#
# Правила — docs/development/git-workflow.md: "<type>(<scope>): <описание>",
# type из фиксированного списка; scope необязателен.
set -u

base="${1:-}"
head="${2:-HEAD}"

if [ -z "$base" ]; then
	echo "usage: check-commits.sh <base-ref> [head-ref]" >&2
	exit 2
fi

types='feat|fix|docs|refactor|test|chore|ci'
pattern="^(${types})(\([a-zA-Z0-9_./-]+\))?: .+"

errors=0
total=0

while IFS=$'\t' read -r sha subject; do
	[ -z "$sha" ] && continue
	# Мерж-коммиты (например, после разрешения конфликтов с main) не проверяем.
	case "$subject" in
	"Merge "*) continue ;;
	esac
	total=$((total + 1))
	if ! echo "$subject" | grep -qE "$pattern"; then
		echo "INVALID: ${sha:0:7} $subject"
		errors=$((errors + 1))
	fi
done < <(git log --format='%H%x09%s' "${base}..${head}")

echo "check-commits: проверено $total коммит(ов), нарушений: $errors"
[ "$errors" -eq 0 ]
