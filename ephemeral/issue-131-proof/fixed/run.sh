#!/bin/zsh

set -o pipefail

fixed_dir=${0:A:h}
repo_dir=${fixed_dir:h:h:h}
cd "$repo_dir" || exit 1

git rev-parse HEAD > "$fixed_dir/head.sha"
git diff --no-ext-diff --binary -- event_persistence.go events.go jsonschema/LifecycleRecord.json jsonschema/LifecycleRecord.json.sum run.go > "$fixed_dir/production.diff"
shasum -a 256 "$fixed_dir/production.diff" > "$fixed_dir/production.diff.sha256"
git hash-object issue131_proof_test.go > "$fixed_dir/issue131_proof_test.sha"
git status --short > "$fixed_dir/worktree.status"

: > "$fixed_dir/status.txt"

run_check() {
  local label=$1
  shift
  "$@" 2>&1 | tee "$fixed_dir/raw-${label}.txt"
  local exit_code=${pipestatus[1]}
  print -r -- "$label=$exit_code" | tee -a "$fixed_dir/status.txt"
  return 0
}

run_check issue131 go test -count=1 -run '^TestIssue131' -v .
run_check baseline go test -count=1 -run '^(TestRunLogCanBeRead|TestSuperviseCancelsAndJoinsLook)$' -v .
run_check race go test -race -count=1 -run '^TestIssue131' -v .
run_check vet go vet ./...
run_check build just build
run_check all go test ./...

if rg -n '=([1-9][0-9]*)$' "$fixed_dir/status.txt" >/dev/null; then
  exit 1
fi
