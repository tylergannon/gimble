#!/usr/bin/env bash
# Runs both closeout-gap proof commands in ephemeral/proof/closeout-gaps/ and
# checks their exit statuses against the requested mode. These two gaps were
# identified by worklog decision #10
# (ephemeral/worklog/202608241000-system-instability-audit.md) as surviving
# the first Tractor validation despite all four established-bug regressions
# passing green:
#
#   1. installing the new binary cannot remove an already-orphaned staged
#      bundle, because the persisted staged path is no longer in watcher
#      memory once a fresh process starts with w.prepared nil.
#   2. active-status deduplication does not reset across an intervening
#      current-state branch, so a genuinely new pending period for a newer
#      candidate can be silently swallowed if the same work is still active.
#
#   baseline: both commands must FAIL (nonzero exit) — proves both gaps are
#             still present in the current production implementation.
#   fixed:    both commands must PASS (zero exit) — proves both gaps have
#             been corrected.
#
# Unlike check-established-bugs.sh's directory scan, this script requires
# exactly the two named closeout-gap commands below and checks each one
# independently, one at a time; neither command's result stands in for or
# masks the other's.
set -uo pipefail

usage() {
  echo "usage: $0 {baseline|fixed}" >&2
  exit 2
}

[[ $# -eq 1 ]] || usage
mode="$1"
case "$mode" in
  baseline|fixed) ;;
  *) usage ;;
esac

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
gaps_dir="$script_dir/closeout-gaps"

commands=(
  "$gaps_dir/01-orphan-residue-with-no-in-memory-bookkeeping.sh"
  "$gaps_dir/02-active-work-dedup-does-not-reset-across-current-cycle.sh"
)

for command in "${commands[@]}"; do
  if [[ ! -x "$command" ]]; then
    echo "check-closeout-gaps: required command missing or not executable: $command" >&2
    exit 1
  fi
done

failures=0
passes=0
for command in "${commands[@]}"; do
  name="$(basename "$command")"
  echo "==> [$mode] $name"
  log="$(mktemp)"
  if "$command" >"$log" 2>&1; then
    status=0
  else
    status=$?
  fi
  cat "$log"
  rm -f "$log"

  if [[ "$mode" == "baseline" ]]; then
    if [[ $status -eq 0 ]]; then
      echo "==> FAIL: $name exited 0 (gap not reproduced) in baseline mode"
      failures=$((failures + 1))
    else
      echo "==> ok: $name exited $status as expected in baseline mode"
      passes=$((passes + 1))
    fi
  else
    if [[ $status -ne 0 ]]; then
      echo "==> FAIL: $name exited $status (gap still present) in fixed mode"
      failures=$((failures + 1))
    else
      echo "==> ok: $name exited 0 as expected in fixed mode"
      passes=$((passes + 1))
    fi
  fi
done

echo
echo "check-closeout-gaps [$mode]: $passes/${#commands[@]} commands as expected, $failures unexpected"
if [[ $failures -ne 0 ]]; then
  exit 1
fi
exit 0
