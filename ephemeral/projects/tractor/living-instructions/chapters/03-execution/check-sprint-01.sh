#!/bin/sh
set -eu

output="$(mktemp)"
trap 'rm -f "$output"' EXIT HUP INT TERM

if ! go test ./workflow -run '^(TestBuiltInPlan|TestPlanArtifacts|TestPlanExecutionShapes|TestRecommendation)$' -count=1 -v >"$output" 2>&1; then
	cat "$output"
	exit 1
fi
cat "$output"

for test_name in TestBuiltInPlan TestPlanArtifacts TestPlanExecutionShapes TestRecommendation; do
	grep -q -- "--- PASS: $test_name" "$output"
done
