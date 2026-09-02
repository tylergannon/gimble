#!/bin/sh
# Live acceptance for the public plan -> printed MEDIUM handoff. Both runs use
# the current binary's real Codex harness and fresh default log allocation.
set -eu

root="$(cd "$(dirname "$0")/../../../../../.." && pwd)"
cd "$root"
here="ephemeral/projects/tractor/living-instructions/chapters/03-execution"
tmp="$(mktemp -d)"
active_pid=""
plan_logs=""
execution_logs=""
workdir=""
project_dir=""
interview_dir=""

cleanup() {
	status=$?
	trap - EXIT HUP INT TERM
	if [ -n "$active_pid" ] && kill -0 "$active_pid" 2>/dev/null; then
		kill "$active_pid" 2>/dev/null || true
		wait "$active_pid" 2>/dev/null || true
	fi
	if [ "$status" -ne 0 ]; then
		printf '%s\n' '--- live plan-to-MEDIUM proof diagnostics ---' >&2
		for output in "$tmp/plan.out" "$tmp/execution.out"; do
			if [ -f "$output" ]; then
				printf '\n--- %s tail ---\n' "$output" >&2
				tail -120 "$output" >&2
			fi
		done
		for logs in "$plan_logs" "$execution_logs"; do
			if [ -n "$logs" ] && [ -f "$logs/timeline.jsonl" ]; then
				printf '\n--- %s timeline tail ---\n' "$logs" >&2
				tail -80 "$logs/timeline.jsonl" >&2
			fi
		done
		if [ -n "$interview_dir" ] && [ -d "$interview_dir" ]; then
			for artifact in "$interview_dir"/*; do
				[ -f "$artifact" ] || continue
				printf '\n--- %s ---\n' "$artifact" >&2
				cat "$artifact" >&2
			done
		fi
		for artifact in "$project_dir/brief.md" "$project_dir/checklist.md" \
			"$project_dir/recommendation.md"; do
			if [ -f "$artifact" ]; then
				printf '\n--- %s ---\n' "$artifact" >&2
				cat "$artifact" >&2
			fi
		done
		if [ -n "$workdir" ] && [ -d "$workdir/.git" ]; then
			printf '\n--- git log ---\n' >&2
			git -C "$workdir" log --oneline --decorate -12 >&2 || true
			printf '\n--- git status ---\n' >&2
			git -C "$workdir" status --short >&2 || true
		fi
		if [ -n "$execution_logs" ] && [ -d "$execution_logs/stages" ]; then
			for record in "$execution_logs"/stages/*-sprints/validation.json; do
				[ -f "$record" ] || continue
				printf '\n--- %s ---\n' "$record" >&2
				cat "$record" >&2
				log="$(dirname "$record")/validation.log"
				if [ -f "$log" ]; then
					printf '\n--- %s ---\n' "$log" >&2
					cat "$log" >&2
				fi
			done
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

bin_dir="$tmp/bin"
bin="$bin_dir/tractor"
workdir="$tmp/workdir"
project="live-medium-proof"
project_dir="$workdir/ephemeral/projects/$project"
interview_dir="$project_dir/interview"
seed="seed.md"
mkdir -p "$bin_dir" "$workdir"
go build -o "$bin" ./cmd/tractor
cp -R "$here/sprint-06-fixture/." "$workdir/"
cp "$here/seed-sprint-06.md" "$workdir/$seed"

git -C "$workdir" init -q
git -C "$workdir" config user.name 'Tractor Proof'
git -C "$workdir" config user.email 'proof@tractor.invalid'
git -C "$workdir" add .
git -C "$workdir" commit -q -m 'Add ambiguous two-package seed'
base_sha="$(git -C "$workdir" rev-parse HEAD)"

if (cd "$workdir" && go test ./... >"$tmp/initial-tests.out" 2>&1); then
	printf '%s\n' 'fixture Go tests unexpectedly pass before planning' >&2
	exit 1
fi

wait_for_logs() {
	output=$1
	pid=$2
	label=$3
	waited=0
	while :; do
		logs="$(sed -n 's/^Logs: //p' "$output" 2>/dev/null | head -1)"
		if [ -n "$logs" ] && [ -d "$logs" ]; then
			printf '%s\n' "$logs"
			return
		fi
		if ! kill -0 "$pid" 2>/dev/null; then
			printf '%s workflow exited before printing its default log path\n' "$label" >&2
			return 1
		fi
		waited=$((waited + 1))
		if [ "$waited" -ge 60 ]; then
			printf '%s workflow did not print its default log path within 60 seconds\n' "$label" >&2
			return 1
		fi
		sleep 1
	done
}

answer_run() {
	pid=$1
	logs=$2
	phase=$3
	answer=$4
	max_questions=$5
	max_wait=$6
	answered_count=0
	waited=0
	while kill -0 "$pid" 2>/dev/null; do
		if [ -d "$interview_dir" ]; then
			for question in "$interview_dir"/*; do
				[ -f "$question" ] || continue
				case "$question" in
					*.answer.md) continue ;;
				esac
				answer_path="${question%.*}.answer.md"
				[ ! -e "$answer_path" ] || continue
				events="$(grep -c '"type":"QuestionAsked"' "$logs/timeline.jsonl" 2>/dev/null || true)"
				[ "$events" -gt "$answered_count" ] || continue
				[ -s "$question" ] || {
					printf 'empty %s question: %s\n' "$phase" "$question" >&2
					return 1
				}
				answered_count=$((answered_count + 1))
				if [ "$answered_count" -gt "$max_questions" ]; then
					printf '%s workflow exceeded %d questions\n' "$phase" "$max_questions" >&2
					return 1
				fi
				printf '\n## Question %d: %s\n' "$answered_count" "$question" >>"$tmp/$phase-questions.md"
				cat "$question" >>"$tmp/$phase-questions.md"
				printf '\n## Answer %d\n%s\n' "$answered_count" "$answer" >>"$tmp/$phase-answers.md"
				"$bin" answer "$question" "$answer" >/dev/null
			done
		fi
		waited=$((waited + 1))
		if [ "$waited" -ge "$max_wait" ]; then
			printf '%s workflow exceeded %d seconds\n' "$phase" "$max_wait" >&2
			return 1
		fi
		sleep 1
	done

	set +e
	wait "$pid"
	run_status=$?
	set -e
	active_pid=""
	if [ "$run_status" -ne 0 ]; then
		printf '%s workflow exited with status %d\n' "$phase" "$run_status" >&2
		return 1
	fi
}

planning_answer='Use exactly two separate one-turn MEDIUM items, in this order. First, implement and test greeting.Message so greeting.Message("Ada") returns exactly "Hello, Ada!"; its item command must be exactly `go test ./greeting`. Second, implement and test words.Count using Go strings.Fields semantics, so words.Count("one\ttwo\nthree") is 3 and whitespace-only input is 0; its item command must be exactly `go test ./words`. Scope is only those two packages and their tests; do not add dependencies, commands, or features. Success also requires `go test ./...`. Recommend MEDIUM because the independent behaviors require two items, and use the exact standard MEDIUM Next handoff.'

env -u TRACTOR_RUN_DIR -u TRACTOR_INTERVIEW_DIR \
	XDG_STATE_HOME="$tmp/state" PATH="$bin_dir:$PATH" \
	"$bin" workflow run plan \
		--project "$project" \
		--seed "$seed" \
		--workdir "$workdir" >"$tmp/plan.out" 2>&1 &
active_pid=$!
plan_logs="$(wait_for_logs "$tmp/plan.out" "$active_pid" plan)"
answer_run "$active_pid" "$plan_logs" planning "$planning_answer" 8 900
plan_questions=$answered_count

[ "$plan_questions" -ge 1 ] || {
	printf '%s\n' 'planning workflow completed without a blocking question' >&2
	exit 1
}
[ -s "$tmp/planning-questions.md" ] && [ -s "$tmp/planning-answers.md" ] || {
	printf '%s\n' 'planning interview capture is empty' >&2
	exit 1
}
grep -Eiq 'greet|word|contract|behavior|success|scope' "$tmp/planning-questions.md" || {
	printf '%s\n' 'planning questions did not inspect the ambiguous contract' >&2
	exit 1
}

"$bin" workflow validate-plan --project "$project" --workdir "$workdir"
go run "$here/check-sprint-06-inspect.go" plan "$project_dir/checklist.md" "$plan_logs" "$workdir"

for handoff in \
	"Brief: $project_dir/brief.md" \
	"Checklist: $project_dir/checklist.md" \
	"Recommendation: $project_dir/recommendation.md" \
	'Size: MEDIUM'; do
	grep -Fq "$handoff" "$tmp/plan.out" || {
		printf 'missing planning handoff line: %s\n' "$handoff" >&2
		exit 1
	}
done
next_command="$(sed -n 's/^Next: //p' "$tmp/plan.out" | tail -1)"
[ "$next_command" = "tractor workflow run medium --project $project" ] || {
	printf 'unexpected Next command: %s\n' "$next_command" >&2
	exit 1
}

execution_answer='Keep the planned observable contract and item-specific validator unchanged. Implement only the selected package and its tests, run that validator, and commit the sprint.'
env -u TRACTOR_RUN_DIR -u TRACTOR_INTERVIEW_DIR \
	XDG_STATE_HOME="$tmp/state" PATH="$bin_dir:$PATH" \
	sh -c "$next_command --workdir \"\$1\"" sh "$workdir" >"$tmp/execution.out" 2>&1 &
active_pid=$!
execution_logs="$(wait_for_logs "$tmp/execution.out" "$active_pid" execution)"
answer_run "$active_pid" "$execution_logs" execution "$execution_answer" 4 1200

grep -Fq "Completed logs: $execution_logs" "$tmp/execution.out" || {
	printf '%s\n' 'execution output did not report the discovered default log path' >&2
	exit 1
}
go run "$here/check-sprint-06-inspect.go" execution "$project_dir/checklist.md" "$execution_logs" "$workdir"

probe_dir="$workdir/proofprobe"
mkdir -p "$probe_dir"
printf '%s\n' \
	'package main' \
	'' \
	'import (' \
	'  "fmt"' \
	'  "example.com/tractor-medium-proof/greeting"' \
	'  "example.com/tractor-medium-proof/words"' \
	')' \
	'' \
	'func main() {' \
	'  fmt.Println(greeting.Message("Ada"))' \
	'  fmt.Println(words.Count("one\ttwo\nthree"))' \
	'  fmt.Println(words.Count(" \t\n "))' \
	'}' >"$probe_dir/main.go"
probe_output="$(cd "$workdir" && go run ./proofprobe)"
[ "$probe_output" = "Hello, Ada!
3
0" ] || {
	printf 'independent behavior probe returned:\n%s\n' "$probe_output" >&2
	exit 1
}
rm -rf "$probe_dir"
(cd "$workdir" && go test ./...)

commit_count="$(git -C "$workdir" rev-list --count HEAD)"
[ "$commit_count" -ge 3 ] || {
	printf 'git history has %d commits, want seed plus two workflow commits\n' "$commit_count" >&2
	exit 1
}
git -C "$workdir" diff --quiet "$base_sha" HEAD -- greeting/greeting.go || greeting_changed=true
git -C "$workdir" diff --quiet "$base_sha" HEAD -- words/words.go || words_changed=true
[ "${greeting_changed:-false}" = true ] && [ "${words_changed:-false}" = true ] || {
	printf '%s\n' 'workflow commits did not change both package implementations' >&2
	exit 1
}
git -C "$workdir" diff --quiet HEAD -- greeting words || {
	printf '%s\n' 'implemented package changes were not committed by the workflow' >&2
	exit 1
}

cat "$tmp/planning-questions.md"
cat "$tmp/planning-answers.md"
cat "$tmp/plan.out"
cat "$tmp/execution.out"
git -C "$workdir" log --oneline --decorate
printf '%s\n' 'check-sprint-06: ok'
