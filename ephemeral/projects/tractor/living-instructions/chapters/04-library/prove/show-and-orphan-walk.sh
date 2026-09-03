#!/bin/sh
# Proves: show, render test, orphan walk. Parts:
#   nodes     - headed `show` lists the same id/type set the YAML declares.
#   equality  - a dumper the script writes calls workflow.Build directly;
#               `show --raw` must equal it for every node that carries a
#               prompt, command, or checklist, whatever its kind.
#   content   - each prompt-bearing node's header names its library file;
#               a per-file sentinel appended in a copy of the tree must
#               appear in that node's output and no other file's may;
#               every rendered line is library text or a short data value.
#   orphans   - an injected uncited page fails the orphan test by name.
#   rendering - an injected broken action fails the render test by name.
set -eu
root="$(git rev-parse --show-toplevel)"
cd "$root"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/demo" "$tmp/src"
printf 'Print a greeting.\n' > "$tmp/demo/seed.md"
seed="$tmp/demo/seed.md"

show_raw() { # bin workflow node
  if [ "$2" = plan ]; then "$1" workflow show "$2" --project demo --seed "$seed" --workdir "$tmp/demo" --node "$3" --raw
  else "$1" workflow show "$2" --project demo --workdir "$tmp/demo" --node "$3" --raw; fi
}
show_headed() { # bin workflow
  if [ "$2" = plan ]; then "$1" workflow show "$2" --project demo --seed "$seed" --workdir "$tmp/demo"
  else "$1" workflow show "$2" --project demo --workdir "$tmp/demo"; fi
}
# Header form: "== <id> (<type>) [<library file>]"; the file is present for
# every node that carries a prompt.
shown_nodes() { show_headed "$1" "$2" | sed -n 's/^== \([^ ]*\) (\([a-z_]*\)).*/\1 \2/p' | sort; }
shown_file() { show_headed "$1" "$2" | sed -n "s/^== $3 ([a-z_]*) \(.*\)$/\1/p"; }
yaml_nodes() { awk '/^  - id:/{id=$3} /^    type:/{print id, $2}' "workflow/library/workflows/$1.yaml" | sort; }

# ---- copy of the tracked tree, with a Build dumper --------------------
git ls-files -z | tar --null -T - -cf - | tar -xf - -C "$tmp/src"
mkdir -p "$tmp/src/cmd/builddump"
cat > "$tmp/src/cmd/builddump/main.go" <<'EOF'
// builddump prints what workflow.Build materialized for one node. It is
// written by the proof script, not the coder, so the comparison with
// `show --raw` is bound to Build itself. Exit 3 means the node carries
// nothing to print (a plain parallel or fan-in without a prompt).
package main

import (
	"fmt"
	"os"

	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/workflow"
)

func main() {
	name, node, workdir, exe, seed := os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5]
	g, err := workflow.Build(name, workflow.Parameters{
		Project: "demo", Workdir: workdir, Executable: exe,
		Plan: workflow.PlanParameters{Seed: seed},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	for _, n := range g.Nodes {
		if n.Base().ID != node {
			continue
		}
		switch v := n.(type) {
		case *graph.CodergenNode:
			fmt.Print(v.PromptValue(v.DisplayLabel()))
		case *graph.SupervisorNode:
			fmt.Print(v.Prompt)
		case *graph.FanInNode:
			fmt.Print(v.PromptValue(v.DisplayLabel()))
		case *graph.ParallelNode:
			fmt.Print(v.PromptValue(v.DisplayLabel()))
		case *graph.ToolNode:
			fmt.Print(v.ToolCommand)
		case *graph.LoopNode:
			fmt.Print(v.Checklist.Value)
		default:
			fmt.Fprintf(os.Stderr, "node %s: unsupported kind %T\n", node, n)
			os.Exit(2)
		}
		return
	}
	fmt.Fprintf(os.Stderr, "node %s not in %s\n", node, name)
	os.Exit(2)
}
EOF
(cd "$tmp/src" && go build -o "$tmp/bin/tractor" ./cmd/tractor && go build -o "$tmp/bin/builddump" ./cmd/builddump)
workflows="$(ls workflow/library/workflows/*.yaml | xargs -n1 basename | sed 's/\.yaml$//')"

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
  "$tmp/bin/builddump" plan planner "$tmp/demo" "$tmp/bin/tractor" "$seed"
} > "$tmp/stage/prompt.md"
"$tmp/bin/tractor" workflow show plan --project demo --seed "$seed" --workdir "$tmp/demo" \
  --node planner --stage "$tmp/stage" || { echo "show --stage reported a diff"; exit 1; }
printf 'x' >> "$tmp/stage/prompt.md"
if "$tmp/bin/tractor" workflow show plan --project demo --seed "$seed" --workdir "$tmp/demo" \
  --node planner --stage "$tmp/stage" >/dev/null 2>&1; then
  echo "show --stage missed a diff"; exit 1
fi

# closure FILE: prints FILE and every prompt/supervisor/pass file reachable
# from it through include actions (an action naming another library file
# by path or by base name), transitively.
closure() {
  printf '%s\n' "$1" > "$tmp/cl.txt"
  while :; do
    before="$(wc -l < "$tmp/cl.txt")"
    while read -r f; do
      for cand in $(cd "${lib:-workflow/library}" && find prompts supervisors passes -type f 2>/dev/null); do
        base="$(basename "$cand" .md)"
        if grep -q "include.*\"$base\"\|include.*\"$cand\"" "$lib/$f" 2>/dev/null; then printf '%s\n' "$cand"; fi
      done
    done < "$tmp/cl.txt" >> "$tmp/cl.txt"
    sort -u "$tmp/cl.txt" -o "$tmp/cl.txt"
    [ "$(wc -l < "$tmp/cl.txt")" -eq "$before" ] && break
  done
  cat "$tmp/cl.txt"
}
# ---- text from the closure only ----------------------------------------
# The check computes every data value itself from the parameters it
# passed and the README's stated derivations, in raw, Go-quoted, and
# shell-quoted forms, and requires `show --values` to agree. Then, per
# node: remove every value form from each rendered line; remove every
# template action and template comment from each line of the node's
# closure (header file, included prompt files, and the doctrine and
# template files any of them name); every rendered residue must equal
# some closure residue, and there may be no more rendered residue lines
# than closure residue lines. Text parked in another file, in a comment,
# or supplied by Go fails; quoted values pass.
lib_real=workflow/library
delim_open="$(sed -n 's/^[Dd]elimiters:[[:space:]]*\([^[:space:]]*\)[[:space:]]*\([^[:space:]]*\).*/\1/p' "$lib_real/README.md" | head -1)"
delim_close="$(sed -n 's/^[Dd]elimiters:[[:space:]]*\([^[:space:]]*\)[[:space:]]*\([^[:space:]]*\).*/\2/p' "$lib_real/README.md" | head -1)"
test -n "$delim_open" && test -n "$delim_close" || { echo "closure-text: README does not state both delimiters"; exit 1; }
# Independent data values. The README's derivations are fixed by sprint 1:
#   Project, Workdir, Executable, Seed, ProjectDir=Workdir/ephemeral/projects/Project,
#   BriefPath=ProjectDir/brief.md, ChecklistPath=ProjectDir/checklist.md,
#   InterviewDir=ProjectDir/interview, QuestionCommand=Executable ask.
project=demo; workdir="$tmp/demo"; exe="$tmp/bin/tractor"
projdir="$workdir/ephemeral/projects/$project"
{
  printf '%s\n' "$project" "$workdir" "$exe" "$seed" "$projdir" "$projdir/brief.md" "$projdir/checklist.md" "$projdir/interview" "$exe ask"
} | sort -u > "$tmp/values-raw.txt"
"$tmp/bin/tractor" workflow show plan --project demo --seed "$seed" --workdir "$tmp/demo" --values > "$tmp/values.txt" \
  || { echo "closure-text: show --values failed"; exit 1; }
sed -n 's/^[A-Za-z_]*: //p' "$tmp/values.txt" | awk 'length($0) > 0' | sort -u > "$tmp/values-shown.txt"
comm -13 "$tmp/values-raw.txt" "$tmp/values-shown.txt" > "$tmp/values-extra.txt" || true
if [ -s "$tmp/values-extra.txt" ]; then
  echo "closure-text: show --values reports values the check did not derive:"; cat "$tmp/values-extra.txt"; exit 1
fi
# Quoted forms: Go strconv.Quote and single-quoted shell form.
python3 - "$tmp/values-raw.txt" > "$tmp/values-all.txt" <<'PYEOF'
import json, sys
vals = [l.rstrip("\n") for l in open(sys.argv[1]) if l.strip()]
out = set()
for v in vals:
    out.add(v)
    out.add(json.dumps(v))                       # Go quote for plain strings
    out.add("'" + v.replace("'", "'\\''") + "'")  # shell single-quoted
for v in sorted(out, key=len, reverse=True):
    print(v)
PYEOF
strip_values() { # stdin -> stdout, every data value form removed, longest first
  cmd="sed"
  while IFS= read -r v; do
    esc="$(printf '%s' "$v" | sed 's/[.[\*^$\/&|]/\\&/g')"
    cmd="$cmd -e 's|$esc||g'"
  done < "$tmp/values-all.txt"
  eval "$cmd"
}
strip_actions() { sed -e "s/$delim_open\/\*.*\*\/$delim_close//g" -e "s/$delim_open[^>]*$delim_close//g"; }
doc_closure() { # FILE... -> doctrine and template files named by any of them
  for f in "$@"; do
    for cand in $(cd "$lib_real" && find doctrine templates -type f 2>/dev/null); do
      base="$(basename "$cand" .md)"; base2="$(basename "$cand")"
      if grep -q "\"$base\"\|\"$base2\"\|\"$cand\"" "$lib_real/$f" 2>/dev/null; then printf '%s\n' "$cand"; fi
    done
  done | sort -u
}
lib="$lib_real"
for wf in $workflows; do
  yaml_nodes "$wf" | while read -r node kind; do
    file="$(shown_file "$tmp/bin/tractor" "$wf" "$node")"
    [ -n "$file" ] || continue
    closure "$file" > "$tmp/node-closure.txt"
    doc_closure $(cat "$tmp/node-closure.txt") >> "$tmp/node-closure.txt"
    : > "$tmp/closure-residue.txt"
    while read -r f; do strip_actions < "$lib_real/$f" >> "$tmp/closure-residue.txt"; done < "$tmp/node-closure.txt"
    awk 'length($0) > 0' "$tmp/closure-residue.txt" | sed 's/[[:space:]]*$//' | sort > "$tmp/closure-residue.txt.s"
    show_raw "$tmp/bin/tractor" "$wf" "$node" | strip_values | sed 's/[[:space:]]*$//' | awk 'length($0) > 0' | sort > "$tmp/rendered-residue.txt"
    while IFS= read -r line; do
      grep -qxF -- "$line" "$tmp/closure-residue.txt.s" \
        || { echo "closure-text: $wf/$node renders a line that is neither closure text nor data: $line"; exit 1; }
    done < "$tmp/rendered-residue.txt"
    r="$(wc -l < "$tmp/rendered-residue.txt")"; c="$(wc -l < "$tmp/closure-residue.txt.s")"
    [ "$r" -le "$c" ] || { echo "closure-text: $wf/$node renders $r lines from a closure of $c"; exit 1; }
  done
done
lib="$tmp/src/workflow/library"

# ---- content: one sentinel per file, in the copy ------------------------
lib="$tmp/src/workflow/library"
stamp="SENTINEL-$(date +%s)-$$"
find "$lib/prompts" "$lib/supervisors" "$lib/passes" -type f 2>/dev/null | while read -r f; do
  rel="${f#$lib/}"
  printf '\n%s FILE %s\n' "$stamp" "$rel" >> "$f"
done
page="$(find "$lib/doctrine" -type f -name '*.md' 2>/dev/null | head -1 || true)"
if [ -n "$page" ]; then printf '\n%s DOCTRINE %s\n' "$stamp" "$(basename "$page" .md)" >> "$page"; fi
printf '# uncited\n\nNo prompt cites this page.\n' > "$lib/doctrine/zz-uncited.md"
(cd "$tmp/src" && go build -o "$tmp/bin/mutated" ./cmd/tractor)
for wf in $workflows; do
  yaml_nodes "$wf" | while read -r node kind; do
    file="$(shown_file "$tmp/bin/mutated" "$wf" "$node")"
    case "$kind" in
      codergen|supervisor) test -n "$file" || { echo "content: $wf/$node header names no library file"; exit 1; } ;;
    esac
    [ -n "$file" ] || continue
    show_raw "$tmp/bin/mutated" "$wf" "$node" > "$tmp/mut.txt"
    grep -q "^$stamp FILE $file\$" "$tmp/mut.txt" || { echo "content: $wf/$node does not render $file"; exit 1; }
    # The include closure of the header file: every prompt file it names
    # through an include action, transitively. Sentinels of files in the
    # closure are expected; any other prompt file's sentinel is a fail.
    closure "$file" > "$tmp/closure.txt"
    grep "^$stamp FILE " "$tmp/mut.txt" | sed "s/^$stamp FILE //" | while read -r seen; do
      grep -qxF -- "$seen" "$tmp/closure.txt" || { echo "content: $wf/$node renders $seen, outside the include closure of $file"; exit 1; }
    done
    while read -r inc; do
      grep -q "^$stamp FILE $inc\$" "$tmp/mut.txt" || { echo "content: $wf/$node includes $inc but does not render it"; exit 1; }
    done < "$tmp/closure.txt"
    if [ -n "$page" ] && grep -q "$(basename "$page" .md)" "$lib/$file"; then
      grep -q "^$stamp DOCTRINE " "$tmp/mut.txt" || { echo "content: $wf/$node names $(basename "$page") but does not render it"; exit 1; }
    fi
  done
done

# ---- no unused prompt file --------------------------------------------
: > "$tmp/used.txt"
for wf in $workflows; do
  yaml_nodes "$wf" | while read -r node kind; do
    file="$(shown_file "$tmp/bin/mutated" "$wf" "$node")"
    [ -n "$file" ] || continue
    closure "$file"
  done
done | sort -u > "$tmp/used.txt"
(cd "$lib" && find prompts supervisors -type f | sort) | while read -r f; do
  grep -qxF -- "$f" "$tmp/used.txt" || { echo "unused: no node renders $f"; exit 1; }
done

# ---- orphans: the injected page must be reported by name ---------------
(cd "$tmp/src" && go test -run 'TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/orphan.log" 2>&1) \
  && { echo "orphans: injected zz-uncited.md was not reported"; cat "$tmp/orphan.log"; exit 1; }
grep -q 'zz-uncited' "$tmp/orphan.log" || { echo "orphans: test failed but did not name zz-uncited.md"; cat "$tmp/orphan.log"; exit 1; }

# ---- rendering: a broken action must be reported by name ---------------
rm -f "$lib/doctrine/zz-uncited.md"
if [ -n "$page" ]; then
  delim="$(sed -n 's/^[Dd]elimiters:[[:space:]]*\([^[:space:]]*\).*/\1/p' "$lib/README.md" | head -1)"
  test -n "$delim" || { echo "rendering: README does not state the delimiters"; exit 1; }
  printf '\n%s broken-action-never-closed\n' "$delim" >> "$page"
  (cd "$tmp/src" && go test -run 'TestLibraryRendersAll' ./workflow/ -count=1 > "$tmp/render.log" 2>&1) \
    && { echo "rendering: broken page was not reported"; cat "$tmp/render.log"; exit 1; }
  grep -q "$(basename "$page")" "$tmp/render.log" || { echo "rendering: test failed but did not name $(basename "$page")"; cat "$tmp/render.log"; exit 1; }
fi

# ---- the real tree's tests run and pass --------------------------------
go test -v -run 'TestLibraryRendersAll|TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/test.log" 2>&1 || { cat "$tmp/test.log"; exit 1; }
grep -q -- '--- PASS: TestLibraryRendersAll' "$tmp/test.log" || { echo "TestLibraryRendersAll did not run"; exit 1; }
grep -q -- '--- PASS: TestLibraryNoOrphans' "$tmp/test.log" || { echo "TestLibraryNoOrphans did not run"; exit 1; }
echo "show-and-orphan-walk.sh: ok"
