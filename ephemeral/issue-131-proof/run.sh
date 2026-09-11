#!/bin/zsh

# Issue #131 characterization proof. Run from this repository checkout.
set -o pipefail

proof_dir=${0:A:h}
repo_dir=${proof_dir:h:h}
cd "$repo_dir" || exit 1

git rev-parse HEAD | tee "$proof_dir/head.txt"

go test -count=1 -run '^TestIssue131' -v . 2>&1 | tee "$proof_dir/raw-issue131.txt"
issue131_exit=${pipestatus[1]}

go test -count=1 -run '^(TestRunLogCanBeRead|TestSuperviseCancelsAndJoinsLook)$' -v . 2>&1 | tee "$proof_dir/raw-baseline.txt"
baseline_exit=${pipestatus[1]}

go test -race -count=1 -run '^TestIssue131' . 2>&1 | tee "$proof_dir/raw-race.txt"
race_exit=${pipestatus[1]}

print -r -- "issue131_exit=$issue131_exit baseline_exit=$baseline_exit race_exit=$race_exit" | tee "$proof_dir/status.txt"
if (( issue131_exit != 0 || baseline_exit != 0 || race_exit != 0 )); then
	return 1
fi
