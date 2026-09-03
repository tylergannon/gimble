#!/bin/sh
# Proves: `workflow show` prints every node the YAML declares and, for
# each, exactly what Build materialized; `--stage` diffs a stage's
# prompt.md minus its frame; `--values` agrees with the data values the
# check derives itself.
. "$(dirname "$0")/lib.sh"
build_copy

# ---- nodes and equality with Build ------------------------------------
for wf in $workflows; do
  shown_nodes "$tmp/bin/tractor" "$wf" > "$tmp/shown.txt"
  built_ids "$wf" > "$tmp/built-ids.txt"
  cmp -s "$tmp/shown.txt" "$tmp/built-ids.txt" || { echo "show $wf lists different nodes than Build returns"; diff "$tmp/built-ids.txt" "$tmp/shown.txt"; exit 1; }
  echo "nodes: show $wf lists every node Build returns ($(wc -l < "$tmp/built-ids.txt" | tr -d ' '))"
  while read -r node; do
    if [ "$wf" = plan ]; then s="$seed"; else s=""; fi
    if ! "$tmp/bin/builddump" "$wf" "$node" "$tmp/demo" "$tmp/bin/tractor" "$s" > "$tmp/want.txt" 2>"$tmp/dump.err"; then
      cat "$tmp/dump.err"; exit 1
    fi
    show_raw "$tmp/bin/tractor" "$wf" "$node" > "$tmp/got.txt"
    cmp -s "$tmp/want.txt" "$tmp/got.txt" || { echo "show $wf --node $node differs from Build"; diff "$tmp/want.txt" "$tmp/got.txt" | head -20; exit 1; }
    echo "equality: $wf/$node equals Build"
  done < "$tmp/built-ids.txt"
done
show_headed "$tmp/bin/tractor" plan | grep -q '<iterate' && { echo "show printed a frame"; exit 1; }
# The headed form carries every payload: the body under each header
# equals the node's --raw output.
for wf in $workflows; do
  show_headed "$tmp/bin/tractor" "$wf" > "$tmp/headed.txt"
  built_ids "$wf" | while read -r node; do
    awk -v id="$node" 'BEGIN{p=0} /^== /{ if (p) exit; p = ($2==id) ; next } p{print}' "$tmp/headed.txt" > "$tmp/headed-body.txt"
    show_raw "$tmp/bin/tractor" "$wf" "$node" > "$tmp/raw-body.txt"
    # Exact comparison: the headed body is the raw payload plus the one
    # blank separator line the format adds after it.
    printf '\n' >> "$tmp/raw-body.txt"
    cmp -s "$tmp/headed-body.txt" "$tmp/raw-body.txt" || { echo "headed: $wf/$node body differs from --raw"; exit 1; }
    echo "headed: $wf/$node body equals --raw"
  done
done

# ---- --stage against a stage built from Build's own output ------------
mkdir -p "$tmp/stage"
{
  cat ephemeral/projects/tractor/living-instructions/chapters/04-library/fixtures/frame-preamble.txt
  printf '<iterate loop="chapters" checklist="x.md" index="1" count="2" lap="1">\nname: x\ncheck: y\n</iterate>\n'
  printf '<iterate loop="sprints" checklist="y.md" index="1" count="3" lap="1">\nname: z\ncheck: w\n</iterate>\n\n'
} > "$tmp/stage/frame.txt"
{
  cat "$tmp/stage/frame.txt"
  "$tmp/bin/builddump" plan planner "$tmp/demo" "$tmp/bin/tractor" "$seed"
} > "$tmp/stage/prompt.md"
cp "$tmp/stage/prompt.md" "$tmp/stage/prompt.md.orig"
"$tmp/bin/tractor" workflow show plan --project demo --seed "$seed" --workdir "$tmp/demo" \
  --node planner --stage "$tmp/stage" || { echo "show --stage reported a diff"; exit 1; }
echo "stage: no diff on the unperturbed stage"
# Perturb one random byte at a random offset, never a fixed suffix.
size="$(wc -c < "$tmp/stage/prompt.md" | tr -d ' ')"
frame="$(wc -c < "$tmp/stage/frame.txt" | tr -d ' ')"
body=$(( size - frame ))
[ "$body" -gt 0 ] || { echo "stage probe: empty body"; exit 1; }
pos=$(( frame + $(od -An -N2 -tu2 /dev/urandom | tr -d ' ') % body ))
# A printable ASCII byte (0x21..0x7e) so the stage stays valid text.
byte="$(printf '%02x' $(( 33 + $(od -An -N1 -tu1 /dev/urandom | tr -d ' ') % 94 )))"
printf "\\$(printf '%03o' "0x$byte")" | dd of="$tmp/stage/prompt.md" bs=1 seek="$pos" conv=notrunc 2>/dev/null
if cmp -s "$tmp/stage/prompt.md" "$tmp/stage/prompt.md.orig"; then
  printf 'x' >> "$tmp/stage/prompt.md"; echo "stage probe: random byte equalled the original; appended instead"
else
  echo "stage probe: byte $byte at offset $pos"
fi
set +e
"$tmp/bin/tractor" workflow show plan --project demo --seed "$seed" --workdir "$tmp/demo" \
  --node planner --stage "$tmp/stage" > "$tmp/stage-diff.txt" 2>"$tmp/stage-diff.err"
rc=$?
set -e
[ "$rc" -eq 1 ] || { echo "show --stage exited $rc on a perturbed stage (want 1)"; cat "$tmp/stage-diff.err"; exit 1; }
[ -s "$tmp/stage-diff.txt" ] || { echo "show --stage printed no diff"; exit 1; }
echo "stage probe: exit 1 with a diff; first differing lines:"
grep '^[-+]' "$tmp/stage-diff.txt" | grep -v '^[-+][-+]' | head -4

# ---- --values agrees with the values the check derives itself ----------
# The README's derivations are fixed by sprint 1:
#   Project, Workdir, Executable, Seed, ProjectDir=Workdir/ephemeral/projects/Project,
#   BriefPath=ProjectDir/brief.md, ChecklistPath=ProjectDir/checklist.md,
#   InterviewDir=ProjectDir/interview, QuestionCommand=Executable ask.
projdir="$tmp/demo/ephemeral/projects/demo"
printf '%s\n' demo "$tmp/demo" "$tmp/bin/tractor" "$seed" "$projdir" "$projdir/brief.md" "$projdir/checklist.md" "$projdir/interview" "$tmp/bin/tractor ask" | sort -u > "$tmp/values-raw.txt"
"$tmp/bin/tractor" workflow show plan --project demo --seed "$seed" --workdir "$tmp/demo" --values > "$tmp/values.txt" \
  || { echo "show --values failed"; exit 1; }
# Exact format: one "Field: value" line per README-listed field, each
# once, with the value the check derived; nothing else.
{
  printf 'Project: %s\n' demo
  printf 'Workdir: %s\n' "$tmp/demo"
  printf 'Executable: %s\n' "$tmp/bin/tractor"
  printf 'Seed: %s\n' "$seed"
  printf 'ProjectDir: %s\n' "$projdir"
  printf 'BriefPath: %s\n' "$projdir/brief.md"
  printf 'ChecklistPath: %s\n' "$projdir/checklist.md"
  printf 'InterviewDir: %s\n' "$projdir/interview"
  printf 'QuestionCommand: %s\n' "$tmp/bin/tractor ask"
} | sort > "$tmp/values-expected.txt"
sort "$tmp/values.txt" > "$tmp/values-shown.txt"
cmp -s "$tmp/values-expected.txt" "$tmp/values-shown.txt" || { echo "show --values differs from the derived Field: value lines"; diff "$tmp/values-expected.txt" "$tmp/values-shown.txt"; exit 1; }
[ "$(wc -l < "$tmp/values.txt" | tr -d ' ')" -eq 9 ] || { echo "show --values printed $(wc -l < "$tmp/values.txt") lines, want 9"; exit 1; }
echo "values: nine Field: value lines, exactly the derived ones"
echo "show-equals-build.sh: ok"
