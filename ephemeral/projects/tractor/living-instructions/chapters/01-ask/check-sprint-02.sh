#!/bin/sh
# Definition of done for sprint 2: a live run whose agent asks a question
# through tractor ask, sees it answered, and completes.
set -eu

root="$(cd "$(dirname "$0")/../../../../../.." && pwd)"
cd "$root"
here="ephemeral/projects/tractor/living-instructions/chapters/01-ask"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

bin="$tmp/tractor"
go build -o "$bin" ./cmd/tractor

logs="$tmp/run"
ask_dir="$tmp/interview"
out="$tmp/secret.txt"
nonce="SECRET_$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')"

# Do not inherit an enclosing run's directory; the run under test must set its own.
env -u TRACTOR_RUN_DIR ASK_PROOF_DIR="$ask_dir" ASK_PROOF_OUT="$out" \
  "$bin" run "$here/proof-pipeline.yaml" --workdir "$root" --logs "$logs" > "$tmp/run.out" 2>&1 &
run_pid=$!

i=0
until [ -f "$logs/timeline.jsonl" ] && grep -q '"type":"QuestionAsked"' "$logs/timeline.jsonl"; do
  i=$((i + 1))
  if [ "$i" -gt 900 ] || ! kill -0 "$run_pid" 2>/dev/null; then
    echo "no QuestionAsked within the wait, or the run died"; tail -50 "$tmp/run.out"; exit 1
  fi
  sleep 1
done

question="$(ls "$ask_dir"/0001.* 2>/dev/null | grep -v answer | head -1)"
[ -n "$question" ] || { echo "no question file in $ask_dir"; ls -la "$ask_dir" 2>/dev/null; exit 1; }
"$bin" answer "$question" "$nonce" > /dev/null

wait "$run_pid" || { echo "run exited nonzero"; tail -50 "$tmp/run.out"; exit 1; }
grep -q '"type":"PipelineCompleted"' "$logs/timeline.jsonl" || { echo "run did not complete"; tail -50 "$tmp/run.out"; exit 1; }
[ -f "$out" ] || { echo "agent did not write $out"; exit 1; }
[ "$(tr -d '[:space:]' < "$out")" = "$nonce" ] || { echo "agent wrote: $(cat "$out")"; exit 1; }

echo "check-sprint-02: ok"
