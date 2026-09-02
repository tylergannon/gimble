#!/bin/sh
set -eu

test_output="$(mktemp)"
binary_dir="$(mktemp -d)"
trap 'rm -f "$test_output"; rm -rf "$binary_dir"' EXIT HUP INT TERM

if ! go test ./cmd/tractor -run '^(TestWorkflowList|TestWorkflowRun|TestWorkflowRejects|TestWorkflowHandoff)$' -count=1 -v >"$test_output" 2>&1; then
	cat "$test_output"
	exit 1
fi
cat "$test_output"

for test_name in TestWorkflowList TestWorkflowRun TestWorkflowRejects TestWorkflowHandoff; do
	grep -q -- "--- PASS: $test_name" "$test_output"
done

go build -o "$binary_dir/tractor" ./cmd/tractor

list_output="$($binary_dir/tractor workflow list)"
printf '%s\n' "$list_output"
printf '%s\n' "$list_output" | grep -q '^plan[[:space:]]'
if printf '%s\n' "$list_output" | grep -Eq '^(medium|large)[[:space:]]'; then
	printf '%s\n' 'workflow list exposed an execution workflow before chapter 3' >&2
	exit 1
fi

workflow_help="$($binary_dir/tractor workflow --help)"
printf '%s\n' "$workflow_help"
printf '%s\n' "$workflow_help" | grep -q 'tractor workflow list'
printf '%s\n' "$workflow_help" | grep -q 'run[[:space:]]*Run a built-in workflow'

plan_help="$($binary_dir/tractor workflow run plan --help)"
printf '%s\n' "$plan_help"
for required in --project --seed --workdir --logs brief.md checklist.md recommendation.md; do
	printf '%s\n' "$plan_help" | grep -q -- "$required"
done
