#!/bin/sh
# Sprint 2 definition of done: show, render test, orphan walk.
set -eu
root="$(git rev-parse --show-toplevel)"
cd "$root"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

go build -o "$tmp/tractor" ./cmd/tractor
mkdir -p "$tmp/work"
printf 'Print a greeting.\n' > "$tmp/seed.md"

out="$("$tmp/tractor" workflow show plan --project demo --seed "$tmp/seed.md" --workdir "$tmp/work")"
echo "$out" | grep -q '^== planner (codergen)' || { echo "show plan: planner missing"; exit 1; }
echo "$out" | grep -q '^== validate (tool)' || { echo "show plan: validate missing"; exit 1; }
echo "$out" | grep -q 'ephemeral/projects/demo' || { echo "show plan: project path missing"; exit 1; }
echo "$out" | grep -q '<iterate' && { echo "show printed a frame"; exit 1; }

out="$("$tmp/tractor" workflow show large --project demo --workdir "$tmp/work")"
for n in chapters plan sprints implement; do
  echo "$out" | grep -q "^== $n (" || { echo "show large: $n missing"; exit 1; }
done

# --stage: build a stage directory whose prompt.md is a frame followed by
# exactly the prompt show prints for the node, and expect no diff. The
# frame shape is the engine's (engine/frames.go); the preamble text is
# whatever framePreamble says at HEAD, copied here by the sprint.
mkdir -p "$tmp/stage"
"$tmp/tractor" workflow show plan --project demo --seed "$tmp/seed.md" --workdir "$tmp/work" --node planner --raw > "$tmp/prompt.txt"
{
  cat ephemeral/projects/tractor/living-instructions/chapters/04-library/fixtures/frame-preamble.txt
  printf '<iterate loop="chapters" checklist="x.md" index="1" count="1" lap="1">\nname: x\ncheck: y\n</iterate>\n\n'
  cat "$tmp/prompt.txt"
} > "$tmp/stage/prompt.md"
"$tmp/tractor" workflow show plan --project demo --seed "$tmp/seed.md" --workdir "$tmp/work" \
  --node planner --stage "$tmp/stage" || { echo "show --stage reported a diff"; exit 1; }

# And a diff is a diff: perturb one byte, expect exit 1.
printf 'x' >> "$tmp/stage/prompt.md"
if "$tmp/tractor" workflow show plan --project demo --seed "$tmp/seed.md" --workdir "$tmp/work" \
  --node planner --stage "$tmp/stage" >/dev/null 2>&1; then
  echo "show --stage missed a diff"; exit 1
fi

go test -run 'TestLibraryRendersAll|TestLibraryNoOrphans' ./workflow/... -count=1
echo "check-sprint-02: ok"
