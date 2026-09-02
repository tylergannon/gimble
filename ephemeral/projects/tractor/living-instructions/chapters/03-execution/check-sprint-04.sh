#!/bin/sh
set -eu

test_output="$(mktemp)"
binary_dir="$(mktemp -d)"
expected_list="$(mktemp)"
trap 'rm -f "$test_output" "$expected_list"; rm -rf "$binary_dir"' EXIT HUP INT TERM

if ! go test ./cmd/tractor -run '^(TestWorkflowList|TestWorkflowRun|TestWorkflowRejects|TestWorkflowHandoff|TestWorkflowExecutionRun|TestWorkflowDefaultLogs)$' -count=1 -v >"$test_output" 2>&1; then
	cat "$test_output"
	exit 1
fi
cat "$test_output"

for test_name in \
	TestWorkflowList \
	TestWorkflowRun \
	TestWorkflowRejects \
	TestWorkflowHandoff \
	TestWorkflowExecutionRun \
	TestWorkflowDefaultLogs; do
	grep -q -- "--- PASS: $test_name" "$test_output"
done

# These subtests invoke the exact printed Next commands through the injected
# foreground runner. Their lack of --seed and --logs must pass validation.
grep -q -- '--- PASS: TestWorkflowExecutionRun/medium' "$test_output"
grep -q -- '--- PASS: TestWorkflowExecutionRun/large' "$test_output"

go build -o "$binary_dir/tractor" ./cmd/tractor

cat >"$expected_list" <<'EOF'
large	Plan and execute every chapter through nested engine-owned checklists.
medium	Run every sprint in a planning checklist through engine-owned validation.
plan	Interview the caller and write a planning brief, checklist, and size recommendation.
EOF
"$binary_dir/tractor" workflow list | cmp -s - "$expected_list"

workflow_help="$($binary_dir/tractor workflow --help)"
printf '%s\n' "$workflow_help"
for command in \
	'tractor workflow run plan --project demo --seed seed.md' \
	'tractor workflow run medium --project demo' \
	'tractor workflow run large --project demo'; do
	printf '%s\n' "$workflow_help" | grep -Fq -- "$command"
done

for name in plan medium large; do
	run_help="$($binary_dir/tractor workflow run "$name" --help)"
	printf '%s\n' "$run_help"
	for required in --project --seed --workdir --logs 'Tractor state root' interview; do
		printf '%s\n' "$run_help" | grep -Fq -- "$required"
	done
done

printf '%s\n' 'check-sprint-04: ok'
