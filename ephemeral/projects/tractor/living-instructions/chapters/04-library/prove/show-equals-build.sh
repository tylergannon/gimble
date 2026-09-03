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
  yaml_nodes "$wf" > "$tmp/declared.txt"
  cmp -s "$tmp/shown.txt" "$tmp/declared.txt" || { echo "show $wf lists different nodes than workflows/$wf.yaml"; diff "$tmp/declared.txt" "$tmp/shown.txt"; exit 1; }
  while read -r node kind; do
    if [ "$wf" = plan ]; then s="$seed"; else s=""; fi
    if ! "$tmp/bin/builddump" "$wf" "$node" "$tmp/demo" "$tmp/bin/tractor" "$s" > "$tmp/want.txt" 2>"$tmp/dump.err"; then
      cat "$tmp/dump.err"; exit 1
    fi
    show_raw "$tmp/bin/tractor" "$wf" "$node" > "$tmp/got.txt"
    cmp -s "$tmp/want.txt" "$tmp/got.txt" || { echo "show $wf --node $node differs from Build"; diff "$tmp/want.txt" "$tmp/got.txt" | head -20; exit 1; }
  done < "$tmp/declared.txt"
done
show_headed "$tmp/bin/tractor" plan | grep -q '<iterate' && { echo "show printed a frame"; exit 1; }

# ---- --stage against a stage built from Build's own output ------------
mkdir -p "$tmp/stage"
{
  cat ephemeral/projects/tractor/living-instructions/chapters/04-library/fixtures/frame-preamble.txt
  printf '<iterate loop="chapters" checklist="x.md" index="1" count="1" lap="1">\nname: x\ncheck: y\n</iterate>\n\n'
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
byte="$(od -An -N1 -tx1 /dev/urandom | tr -d ' \n')"
printf "\\$(printf '%03o' "0x$byte")" | dd of="$tmp/stage/prompt.md" bs=1 seek="$pos" conv=notrunc 2>/dev/null
if cmp -s "$tmp/stage/prompt.md" "$tmp/stage/prompt.md.orig"; then
  printf 'x' >> "$tmp/stage/prompt.md"; echo "stage probe: random byte equalled the original; appended instead"
else
  echo "stage probe: byte $byte at offset $pos"
fi
if "$tmp/bin/tractor" workflow show plan --project demo --seed "$seed" --workdir "$tmp/demo" \
  --node planner --stage "$tmp/stage" >/dev/null 2>&1; then
  echo "show --stage missed a diff"; exit 1
fi

# ---- --values agrees with the values the check derives itself ----------
# The README's derivations are fixed by sprint 1:
#   Project, Workdir, Executable, Seed, ProjectDir=Workdir/ephemeral/projects/Project,
#   BriefPath=ProjectDir/brief.md, ChecklistPath=ProjectDir/checklist.md,
#   InterviewDir=ProjectDir/interview, QuestionCommand=Executable ask.
projdir="$tmp/demo/ephemeral/projects/demo"
printf '%s\n' demo "$tmp/demo" "$tmp/bin/tractor" "$seed" "$projdir" "$projdir/brief.md" "$projdir/checklist.md" "$projdir/interview" "$tmp/bin/tractor ask" | sort -u > "$tmp/values-raw.txt"
"$tmp/bin/tractor" workflow show plan --project demo --seed "$seed" --workdir "$tmp/demo" --values > "$tmp/values.txt" \
  || { echo "show --values failed"; exit 1; }
sed -n 's/^[A-Za-z_]*: //p' "$tmp/values.txt" | awk 'length($0) > 0' | sort -u > "$tmp/values-shown.txt"
comm -13 "$tmp/values-raw.txt" "$tmp/values-shown.txt" > "$tmp/values-extra.txt" || true
if [ -s "$tmp/values-extra.txt" ]; then
  echo "show --values reports values the check did not derive:"; cat "$tmp/values-extra.txt"; exit 1
fi
comm -23 "$tmp/values-raw.txt" "$tmp/values-shown.txt" > "$tmp/values-missing.txt" || true
if [ -s "$tmp/values-missing.txt" ]; then
  echo "show --values omits values the check derived:"; cat "$tmp/values-missing.txt"; exit 1
fi
echo "values: $(wc -l < "$tmp/values-raw.txt" | tr -d ' ') derived values, all shown, none extra"
echo "show-equals-build.sh: ok"
