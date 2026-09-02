#!/bin/sh
# Definition of done for sprint 1: ask/answer round trip on the built binary.
set -eu

root="$(cd "$(dirname "$0")/../../../../../.." && pwd)"
cd "$root"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

bin="$tmp/tractor"
go build -o "$bin" ./cmd/tractor

dir="$tmp/interview"
run="$tmp/run"
mkdir -p "$run"
nonce="ANSWER_$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')"

# 1. Markdown question, answered while ask blocks, with a run dir present.
printf '# Which colour?\n\nRed or blue?\n' > "$tmp/q.md"
TRACTOR_INTERVIEW_DIR="$dir" TRACTOR_RUN_DIR="$run" "$bin" ask "$tmp/q.md" > "$tmp/out1" 2> "$tmp/err1" &
ask_pid=$!
i=0
until [ -f "$dir/0001.md" ]; do
  i=$((i + 1)); [ "$i" -lt 100 ] || { echo "question was not moved"; cat "$tmp/err1"; exit 1; }
  sleep 0.1
done
[ ! -e "$tmp/q.md" ] || { echo "source file still present after move"; exit 1; }
grep -q 'Red or blue' "$dir/0001.md" || { echo "moved question lost its body"; exit 1; }
grep -q '"type":"QuestionAsked"' "$run/timeline.jsonl" || { echo "no QuestionAsked in timeline"; cat "$run/timeline.jsonl" 2>/dev/null; exit 1; }
grep -q '0001.md' "$run/timeline.jsonl" || { echo "timeline event does not name the question"; exit 1; }
sleep 1
kill -0 "$ask_pid" 2>/dev/null || { echo "ask exited before an answer existed"; cat "$tmp/err1"; exit 1; }

"$bin" answer "$dir/0001.md" "$nonce" > "$tmp/ans1"
wait "$ask_pid" || { echo "ask exited nonzero"; cat "$tmp/err1"; exit 1; }
[ "$(cat "$tmp/out1")" = "$nonce" ] || { echo "ask printed: $(cat "$tmp/out1")"; exit 1; }
[ -f "$dir/0001.answer.md" ] || { echo "answer file missing"; exit 1; }

# 2. Refuse to overwrite.
if "$bin" answer "$dir/0001.md" "second" 2>/dev/null; then
  echo "answer overwrote an existing answer"; exit 1
fi
[ "$(cat "$dir/0001.answer.md")" = "$nonce" ] || { echo "answer file changed"; exit 1; }

# 3. HTML question takes the next number and keeps its extension; no run dir
#    means a warning, not a failure. Resume with the moved path does not
#    renumber.
printf '<h1>Second</h1>\n' > "$tmp/q.html"
TRACTOR_INTERVIEW_DIR="$dir" "$bin" ask "$tmp/q.html" > "$tmp/out2" 2> "$tmp/err2" &
ask_pid=$!
i=0
until [ -f "$dir/0002.html" ]; do
  i=$((i + 1)); [ "$i" -lt 100 ] || { echo "html question was not moved"; cat "$tmp/err2"; exit 1; }
  sleep 0.1
done
kill "$ask_pid" 2>/dev/null; wait "$ask_pid" 2>/dev/null || true
TRACTOR_INTERVIEW_DIR="$dir" "$bin" ask "$dir/0002.html" > "$tmp/out3" 2> "$tmp/err3" &
ask_pid=$!
sleep 1
[ ! -e "$dir/0003.html" ] || { echo "resume renumbered the question"; exit 1; }
printf 'via stdin\n' | "$bin" answer "$dir/0002.html" > /dev/null
wait "$ask_pid" || { echo "resumed ask exited nonzero"; cat "$tmp/err3"; exit 1; }
[ "$(cat "$tmp/out3")" = "via stdin" ] || { echo "resumed ask printed: $(cat "$tmp/out3")"; exit 1; }

# 4. Unknown extension is refused.
printf 'x\n' > "$tmp/q.txt"
if TRACTOR_INTERVIEW_DIR="$dir" "$bin" ask "$tmp/q.txt" 2>/dev/null; then
  echo "ask accepted a .txt question"; exit 1
fi

echo "check-sprint-01: ok"
