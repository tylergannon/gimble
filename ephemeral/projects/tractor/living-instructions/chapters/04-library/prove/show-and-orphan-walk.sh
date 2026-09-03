#!/bin/sh
# Proves: show, render test, orphan walk.
#
# `show --raw` is compared against the sprint 1 snapshot, which was
# captured from Build before any prompt left Go. That binds show to
# Build; comparing show with its own output would prove nothing.
set -eu
root="$(git rev-parse --show-toplevel)"
cd "$root"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# The snapshot's fixed parameters (see workflow/testdata/README.md and
# TestBuildMatchesSnapshot): project demo, workdir /tmp/demo, executable
# /opt/tractor/bin/tractor, seed /tmp/demo/seed.md. The script uses real
# paths of its own and substitutes them back before comparing.
snap_work=/tmp/demo
snap_exe=/opt/tractor/bin/tractor
mkdir -p "$tmp/bin" "$tmp/demo"
go build -o "$tmp/bin/tractor" ./cmd/tractor
printf 'Print a greeting.\n' > "$tmp/demo/seed.md"
canon() { sed -e "s#$tmp/bin/tractor#$snap_exe#g" -e "s#$tmp/demo#$snap_work#g"; }

show_raw() { # workflow node [seed]
  if [ "$#" -eq 3 ]; then
    "$tmp/bin/tractor" workflow show "$1" --project demo --seed "$3" --workdir "$tmp/demo" --node "$2" --raw
  else
    "$tmp/bin/tractor" workflow show "$1" --project demo --workdir "$tmp/demo" --node "$2" --raw
  fi
}

# Every snapshot file is a node; show --raw must equal it byte for byte.
for snap in workflow/testdata/*/*.txt; do
  wf="$(basename "$(dirname "$snap")")"
  node="$(basename "$snap" .txt)"
  if [ "$wf" = plan ]; then
    show_raw "$wf" "$node" "$tmp/demo/seed.md" | canon > "$tmp/got.txt"
  else
    show_raw "$wf" "$node" | canon > "$tmp/got.txt"
  fi
  cmp -s "$snap" "$tmp/got.txt" || { echo "show $wf --node $node differs from $snap"; diff "$snap" "$tmp/got.txt" | head -20; exit 1; }
done

# The headed form names every node and prints no frame.
out="$("$tmp/bin/tractor" workflow show plan --project demo --seed "$tmp/demo/seed.md" --workdir "$tmp/demo")"
echo "$out" | grep -q '^== planner (codergen)' || { echo "show plan: planner missing"; exit 1; }
echo "$out" | grep -q '^== validate (tool)' || { echo "show plan: validate missing"; exit 1; }
echo "$out" | grep -q '<iterate' && { echo "show printed a frame"; exit 1; }
out="$("$tmp/bin/tractor" workflow show large --project demo --workdir "$tmp/demo")"
for n in chapters plan sprints implement; do
  echo "$out" | grep -q "^== $n (" || { echo "show large: $n missing"; exit 1; }
done

# --stage: a stage directory whose prompt.md is the engine's frame shape
# followed by the SNAPSHOT (not show's output), with the script's real
# paths; expect no diff, then a diff after one byte is appended.
mkdir -p "$tmp/stage"
{
  cat ephemeral/projects/tractor/living-instructions/chapters/04-library/fixtures/frame-preamble.txt
  printf '<iterate loop="chapters" checklist="x.md" index="1" count="1" lap="1">\nname: x\ncheck: y\n</iterate>\n\n'
  sed -e "s#$snap_exe#$tmp/bin/tractor#g" -e "s#$snap_work#$tmp/demo#g" workflow/testdata/plan/planner.txt
} > "$tmp/stage/prompt.md"
"$tmp/bin/tractor" workflow show plan --project demo --seed "$tmp/demo/seed.md" --workdir "$tmp/demo" \
  --node planner --stage "$tmp/stage" || { echo "show --stage reported a diff"; exit 1; }
printf 'x' >> "$tmp/stage/prompt.md"
if "$tmp/bin/tractor" workflow show plan --project demo --seed "$tmp/demo/seed.md" --workdir "$tmp/demo" \
  --node planner --stage "$tmp/stage" >/dev/null 2>&1; then
  echo "show --stage missed a diff"; exit 1
fi

# The two tests must run and pass; a -run with no match exits 0.
go test -v -run 'TestLibraryRendersAll|TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/test.log" 2>&1 || { cat "$tmp/test.log"; exit 1; }
grep -q -- '--- PASS: TestLibraryRendersAll' "$tmp/test.log" || { echo "TestLibraryRendersAll did not run"; exit 1; }
grep -q -- '--- PASS: TestLibraryNoOrphans' "$tmp/test.log" || { echo "TestLibraryNoOrphans did not run"; exit 1; }
echo "show-and-orphan-walk.sh: ok"
