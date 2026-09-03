#!/bin/sh
# Proves: show, render test, orphan walk. Three parts:
#   content  - a sentinel appended to every library prompt file, in a
#              copy of the tree, shows up in `show --raw`; a prompt still
#              living in Go cannot pass this.
#   equality - `show --raw` equals the sprint 1 snapshot, captured from
#              Build before any prompt left Go; `--stage` diffs work.
#   orphans  - an uncited page injected in the copy makes the orphan
#              test fail by name; the real tree's tests run and pass.
set -eu
root="$(git rev-parse --show-toplevel)"
cd "$root"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

snap_work=/tmp/demo
snap_exe=/opt/tractor/bin/tractor
mkdir -p "$tmp/bin" "$tmp/demo" "$tmp/src"
printf 'Print a greeting.\n' > "$tmp/demo/seed.md"

show_raw() { # bin workflow node
  if [ "$2" = plan ]; then
    "$1" workflow show "$2" --project demo --seed "$tmp/demo/seed.md" --workdir "$tmp/demo" --node "$3" --raw
  else
    "$1" workflow show "$2" --project demo --workdir "$tmp/demo" --node "$3" --raw
  fi
}
show_headed() { # bin workflow
  if [ "$2" = plan ]; then
    "$1" workflow show "$2" --project demo --seed "$tmp/demo/seed.md" --workdir "$tmp/demo"
  else
    "$1" workflow show "$2" --project demo --workdir "$tmp/demo"
  fi
}
nodes() { show_headed "$1" "$2" | sed -n 's/^== \([^ ]*\) (\([a-z]*\)).*/\1 \2/p'; }

# ---- content: mutated copy ------------------------------------------
git ls-files -z | tar --null -T - -cf - | tar -xf - -C "$tmp/src"
lib="$tmp/src/workflow/library"
sentinel="SENTINEL-$(date +%s)-$$"
for f in $(find "$lib/prompts" "$lib/supervisors" "$lib/passes" -type f 2>/dev/null); do
  printf '\n%s %s\n' "$sentinel" "$(basename "$f")" >> "$f"
done
page="$(find "$lib/doctrine" -type f -name '*.md' 2>/dev/null | head -1 || true)"
if [ -n "$page" ]; then printf '\n%s doctrine %s\n' "$sentinel" "$(basename "$page")" >> "$page"; fi
printf '# uncited\n\nNo prompt cites this page.\n' > "$lib/doctrine/zz-uncited.md"
(cd "$tmp/src" && go build -o "$tmp/bin/mutated" ./cmd/tractor)
for wf in plan medium large; do
  nodes "$tmp/bin/mutated" "$wf" | while read -r node kind; do
    [ "$kind" = codergen ] || continue
    show_raw "$tmp/bin/mutated" "$wf" "$node" | grep -q "^$sentinel " \
      || { echo "content: $wf/$node prompt does not come from a library file"; exit 1; }
  done
done
if [ -n "$page" ]; then
  cited="$(grep -rl "$(basename "$page" .md)" "$lib/prompts" "$lib/supervisors" "$lib/passes" 2>/dev/null | head -1 || true)"
  if [ -n "$cited" ]; then
    # find a workflow/node whose prompt file is $cited: any node whose raw
    # output carries that file's sentinel must also carry the page's.
    for wf in plan medium large; do
      nodes "$tmp/bin/mutated" "$wf" | while read -r node kind; do
        [ "$kind" = codergen ] || continue
        out="$(show_raw "$tmp/bin/mutated" "$wf" "$node")"
        if echo "$out" | grep -q "^$sentinel $(basename "$cited")\$"; then
          echo "$out" | grep -q "^$sentinel doctrine " || { echo "content: $wf/$node cites $(basename "$page") but does not render it"; exit 1; }
        fi
      done
    done
  fi
fi
(cd "$tmp/src" && go test -run 'TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/orphan.log" 2>&1) \
  && { echo "orphans: injected zz-uncited.md was not reported"; cat "$tmp/orphan.log"; exit 1; }
grep -q 'zz-uncited' "$tmp/orphan.log" || { echo "orphans: test failed but did not name zz-uncited.md"; cat "$tmp/orphan.log"; exit 1; }

# ---- equality: real tree against the snapshot ----------------------
go build -o "$tmp/bin/tractor" ./cmd/tractor
canon() { sed -e "s#$tmp/bin/tractor#$snap_exe#g" -e "s#$tmp/demo#$snap_work#g"; }
for wf in plan medium large; do
  nodes "$tmp/bin/tractor" "$wf" > "$tmp/nodes.txt"
  test -s "$tmp/nodes.txt" || { echo "show $wf printed no nodes"; exit 1; }
  while read -r node kind; do
    snap="workflow/testdata/$wf/$node.txt"
    test -f "$snap" || { echo "no snapshot for $wf/$node"; exit 1; }
    show_raw "$tmp/bin/tractor" "$wf" "$node" | canon > "$tmp/got.txt"
    cmp -s "$snap" "$tmp/got.txt" || { echo "show $wf --node $node differs from $snap"; diff "$snap" "$tmp/got.txt" | head -20; exit 1; }
  done < "$tmp/nodes.txt"
done
show_headed "$tmp/bin/tractor" plan | grep -q '<iterate' && { echo "show printed a frame"; exit 1; }

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

# ---- the coder's tests must run and pass in the real tree ----------
go test -v -run 'TestLibraryRendersAll|TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/test.log" 2>&1 || { cat "$tmp/test.log"; exit 1; }
grep -q -- '--- PASS: TestLibraryRendersAll' "$tmp/test.log" || { echo "TestLibraryRendersAll did not run"; exit 1; }
grep -q -- '--- PASS: TestLibraryNoOrphans' "$tmp/test.log" || { echo "TestLibraryNoOrphans did not run"; exit 1; }
echo "show-and-orphan-walk.sh: ok"
