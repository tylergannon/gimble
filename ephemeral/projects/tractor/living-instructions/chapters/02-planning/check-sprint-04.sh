#!/bin/sh
# Live definition of done for the built-in planning workflow: a real Codex
# harness turn blocks in tractor ask, consumes the answer, and writes a
# mechanically valid one-sprint plan.
set -eu

root="$(cd "$(dirname "$0")/../../../../../.." && pwd)"
cd "$root"
here="ephemeral/projects/tractor/living-instructions/chapters/02-planning"
tmp="$(mktemp -d)"
run_pid=""

cleanup() {
	status=$?
	trap - EXIT HUP INT TERM
	if [ -n "$run_pid" ] && kill -0 "$run_pid" 2>/dev/null; then
		kill "$run_pid" 2>/dev/null || true
		wait "$run_pid" 2>/dev/null || true
	fi
	if [ "$status" -ne 0 ]; then
		printf '%s\n' '--- live planning proof diagnostics ---' >&2
		tail -80 "$tmp/run.out" 2>/dev/null >&2 || true
		for artifact in "$tmp/questions.md" "$tmp/answers.md" \
			"$tmp/workdir/ephemeral/projects/live-planning-proof/brief.md" \
			"$tmp/workdir/ephemeral/projects/live-planning-proof/checklist.md" \
			"$tmp/workdir/ephemeral/projects/live-planning-proof/recommendation.md"; do
			if [ -f "$artifact" ]; then
				printf '\n--- %s ---\n' "$artifact" >&2
				cat "$artifact" >&2
			fi
		done
		if [ -f "$tmp/logs/timeline.jsonl" ]; then
			printf '\n--- timeline tail ---\n' >&2
			tail -40 "$tmp/logs/timeline.jsonl" >&2
		fi
	fi
	rm -rf "$tmp"
	exit "$status"
}
trap cleanup EXIT HUP INT TERM

command -v codex >/dev/null 2>&1 || {
	printf '%s\n' 'the supported Codex harness is not installed' >&2
	exit 1
}

bin="$tmp/tractor"
workdir="$tmp/workdir"
logs="$tmp/logs"
project="live-planning-proof"
project_dir="$workdir/ephemeral/projects/$project"
interview_dir="$project_dir/interview"
seed="seed.md"

go build -o "$bin" ./cmd/tractor
mkdir -p "$workdir"
cp "$here/seed-sprint-04.md" "$workdir/$seed"
git -C "$workdir" init -q
git -C "$workdir" add "$seed"
git -C "$workdir" -c user.name='Tractor Proof' -c user.email='proof@tractor.invalid' \
	commit -q -m 'Add planning seed'

env -u TRACTOR_RUN_DIR -u TRACTOR_INTERVIEW_DIR \
	"$bin" workflow run plan \
		--project "$project" \
		--seed "$seed" \
		--workdir "$workdir" \
		--logs "$logs" >"$tmp/run.out" 2>&1 &
run_pid=$!

answer='Choose `Hello, Tractor!`. The root-level executable `greet.sh` must print exactly `Hello, Tractor!` followed by one newline, accept no arguments, and use POSIX sh. No other behavior or files are in scope. Success is the exact shell check `test "$(./greet.sh)" = "Hello, Tractor!"`.'
question_count=0
waited=0
max_wait=900
max_questions=6

while kill -0 "$run_pid" 2>/dev/null; do
	if [ -d "$interview_dir" ]; then
		for question in "$interview_dir"/*; do
			[ -f "$question" ] || continue
			case "$question" in
				*.answer.md) continue ;;
			esac
			answer_path="${question%.*}.answer.md"
			[ ! -e "$answer_path" ] || continue
			grep -q 'QuestionAsked' "$logs/timeline.jsonl" 2>/dev/null || continue
			[ -s "$question" ] || {
				printf 'empty planning question: %s\n' "$question" >&2
				exit 1
			}
			question_count=$((question_count + 1))
			if [ "$question_count" -gt "$max_questions" ]; then
				printf 'planner exceeded %d questions\n' "$max_questions" >&2
				exit 1
			fi
			printf '\n## Question %d\n' "$question_count" >>"$tmp/questions.md"
			cat "$question" >>"$tmp/questions.md"
			printf '\n## Answer %d\n%s\n' "$question_count" "$answer" >>"$tmp/answers.md"
			"$bin" answer "$question" "$answer" >/dev/null
		done
	fi
	waited=$((waited + 1))
	if [ "$waited" -ge "$max_wait" ]; then
		printf 'planning workflow exceeded %d seconds\n' "$max_wait" >&2
		exit 1
	fi
	sleep 1
done

set +e
wait "$run_pid"
run_status=$?
set -e
run_pid=""
if [ "$run_status" -ne 0 ]; then
	printf 'planning workflow exited with status %d\n' "$run_status" >&2
	exit 1
fi

[ "$question_count" -ge 1 ] || {
	printf '%s\n' 'planning workflow completed without asking a question' >&2
	exit 1
}
[ -s "$tmp/answers.md" ] || {
	printf '%s\n' 'planning workflow did not receive a non-empty answer' >&2
	exit 1
}
grep -Eiq 'greet|greeting|output|Hello, Tractor|Hello, world' "$tmp/questions.md" || {
	printf '%s\n' 'planning question did not address the seed material choice' >&2
	exit 1
}
grep -q '"type":"QuestionAsked"' "$logs/timeline.jsonl"
grep -q '"type":"PipelineCompleted"' "$logs/timeline.jsonl"
grep -Eq '"harness":[[:space:]]*"codex"' "$logs/checkpoint.json"

for handoff in \
	"Brief: $project_dir/brief.md" \
	"Checklist: $project_dir/checklist.md" \
	"Recommendation: $project_dir/recommendation.md" \
	'Size: SIMPLE' \
	'Next: Execute the plan yourself.'; do
	grep -Fq "$handoff" "$tmp/run.out" || {
		printf 'missing CLI handoff line: %s\n' "$handoff" >&2
		exit 1
	}
done

"$bin" workflow validate-plan --project "$project" --workdir "$workdir"
grep -Fq 'greet.sh' "$project_dir/brief.md"
grep -Fq 'Hello, Tractor!' "$project_dir/brief.md"
go run "$here/check-sprint-04-inspect.go" "$project_dir/checklist.md"

grep -Fxq 'Size: SIMPLE' "$project_dir/recommendation.md"
grep -Fxq 'Next: Execute the plan yourself.' "$project_dir/recommendation.md"

cat "$tmp/questions.md"
cat "$tmp/answers.md"
cat "$tmp/run.out"
printf '%s\n' 'check-sprint-04: ok'
