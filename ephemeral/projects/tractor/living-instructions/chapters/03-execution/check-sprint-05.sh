#!/bin/sh
set -eu

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT HUP INT TERM

go build -o "$scratch/tractor" ./cmd/tractor

"$scratch/tractor" workflow list >"$scratch/workflow-list.txt"
"$scratch/tractor" workflow --help >"$scratch/workflow-help.txt"
"$scratch/tractor" workflow list --help >"$scratch/workflow-list-help.txt"
for name in plan medium large; do
	"$scratch/tractor" workflow run "$name" --help >"$scratch/workflow-run-$name-help.txt"
done

for name in plan medium large; do
	grep -Eq "^$name[[:space:]]" "$scratch/workflow-list.txt"
done

for command in \
	'tractor workflow run plan --project demo --seed seed.md' \
	'tractor workflow run medium --project demo' \
	'tractor workflow run large --project demo'; do
	grep -Fq -- "$command" "$scratch/workflow-help.txt"
done

for name in plan medium large; do
	help="$scratch/workflow-run-$name-help.txt"
	for required in \
		'--project' '--seed' '--workdir' '--logs' \
		'foreground runner' 'Tractor state root' 'blocking questions' \
		'Logs: <absolute-path>' 'loop engine validates and marks items'; do
		grep -Fq -- "$required" "$help"
	done
done

docs='README.md docs/spec.md skills/tractor/SKILL.md llms.txt src/content/docs/planning.md src/content/docs/loops.md'
for file in $docs; do
	grep -Fq -- 'tractor workflow run medium' "$file"
	grep -Fq -- 'tractor workflow run large' "$file"
done

if grep -Eiq -- 'pending (handoff|workflow)|only built-in workflow available' $docs; then
	grep -Ein -- 'pending (handoff|workflow)|only built-in workflow available' $docs
	exit 1
fi

for required in 'QuestionAsked' 'validation.log' 'done: true' 'Logs:'; do
	grep -Fq -- "$required" README.md
	grep -Fq -- "$required" docs/spec.md
	grep -Fq -- "$required" skills/tractor/SKILL.md
	grep -Fq -- "$required" llms.txt
	grep -Fq -- "$required" src/content/docs/planning.md src/content/docs/loops.md
done

pnpm verify

printf '%s\n' 'check-sprint-05: ok'
