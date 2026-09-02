#!/usr/bin/env bash
# Runs every established-bug proof command in ephemeral/proof/established-bugs/
# and checks their exit statuses against the requested mode.
#
#   baseline: every command must FAIL (nonzero exit) — proves each bug is
#             still present in the current production implementation.
#   fixed:    every command must PASS (zero exit) — proves each bug has been
#             corrected.
#
# Both modes fail (exit 1) if no bug commands exist under established-bugs/,
# so an empty directory can never be mistaken for "all bugs pass"/"all bugs
# fail". Commands run independently, one at a time; one command's failure
# never masks or stands in for another's result.
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
bugs_dir="$script_dir/established-bugs"

commands=()
if [[ -d "$bugs_dir" ]]; then
  while IFS= read -r -d '' file; do
    commands+=("$file")
  done < <(find "$bugs_dir" -maxdepth 1 -type f -name '*.sh' -print0 | sort -z)
fi

if [[ ${#commands[@]} -eq 0 ]]; then
  echo "check-established-bugs: no bug commands found under $bugs_dir" >&2
  exit 1
fi

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
      echo "==> FAIL: $name exited 0 (bug not reproduced) in baseline mode"
      failures=$((failures + 1))
    else
      echo "==> ok: $name exited $status as expected in baseline mode"
      passes=$((passes + 1))
    fi
  else
    if [[ $status -ne 0 ]]; then
      echo "==> FAIL: $name exited $status (bug still present) in fixed mode"
      failures=$((failures + 1))
    else
      echo "==> ok: $name exited 0 as expected in fixed mode"
      passes=$((passes + 1))
    fi
  fi
done

echo
echo "check-established-bugs [$mode]: $passes/${#commands[@]} commands as expected, $failures unexpected"
if [[ $failures -ne 0 ]]; then
  exit 1
fi
exit 0
