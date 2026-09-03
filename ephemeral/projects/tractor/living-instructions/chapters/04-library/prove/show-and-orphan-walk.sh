#!/bin/sh
# Proves: show, render test, orphan walk. Three parts:
#   nodes    - headed `show` lists exactly the nodes the workflow YAML
#              declares, with their types.
#   content  - in a copy of the tree, a dumper the script writes calls
#              workflow.Build directly; `show --raw` must equal it for
#              every node. Then a sentinel appended to every library
#              prompt file must show up in `show --raw`; a prompt still
#              living in Go cannot pass this.
#   orphans  - an uncited page injected in the copy makes the orphan
#              test fail by name; the real tree's tests run and pass.
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
shown_nodes() { show_headed "$1" "$2" | sed -n 's/^== \([^ ]*\) (\([a-z]*\)).*/\1 \2/p'; }
yaml_nodes() { # workflow: "id type" per node, file order
  awk '/^  - id:/{id=$3} /^    type:/{print id, $2}' "workflow/library/workflows/$1.yaml"
}

# ---- copy of the tracked tree, with a Build dumper --------------------
git ls-files -z | tar --null -T - -cf - | tar -xf - -C "$tmp/src"
mkdir -p "$tmp/src/cmd/builddump"
cat > "$tmp/src/cmd/builddump/main.go" <<'EOF'
// builddump prints what workflow.Build materialized for one node. It is
// written by the proof script, not the coder, so the comparison with
// `show --raw` is bound to Build itself.
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
			fmt.Print(v.PromptValue())
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

# ---- nodes and equality with Build ------------------------------------
for wf in plan medium large; do
  shown_nodes "$tmp/bin/tractor" "$wf" > "$tmp/shown.txt"
  yaml_nodes "$wf" > "$tmp/declared.txt"
  cmp -s "$tmp/shown.txt" "$tmp/declared.txt" || { echo "show $wf lists different nodes than workflows/$wf.yaml"; diff "$tmp/declared.txt" "$tmp/shown.txt"; exit 1; }
  while read -r node kind; do
    case "$kind" in codergen|tool|loop) ;; *) continue ;; esac
    if [ "$wf" = plan ]; then s="$seed"; else s=""; fi
    "$tmp/bin/builddump" "$wf" "$node" "$tmp/demo" "$tmp/bin/tractor" "$s" > "$tmp/want.txt"
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

# ---- content: sentinels in the copy -----------------------------------
lib="$tmp/src/workflow/library"
sentinel="SENTINEL-$(date +%s)-$$"
find "$lib/prompts" "$lib/supervisors" "$lib/passes" -type f 2>/dev/null | while read -r f; do
  printf '\n%s %s\n' "$sentinel" "$(basename "$f")" >> "$f"
done
page="$(find "$lib/doctrine" -type f -name '*.md' 2>/dev/null | head -1 || true)"
if [ -n "$page" ]; then printf '\n%s doctrine %s\n' "$sentinel" "$(basename "$page")" >> "$page"; fi
printf '# uncited\n\nNo prompt cites this page.\n' > "$lib/doctrine/zz-uncited.md"
(cd "$tmp/src" && go build -o "$tmp/bin/mutated" ./cmd/tractor)
for wf in plan medium large; do
  yaml_nodes "$wf" | while read -r node kind; do
    [ "$kind" = codergen ] || continue
    show_raw "$tmp/bin/mutated" "$wf" "$node" > "$tmp/mut.txt"
    grep -q "^$sentinel " "$tmp/mut.txt" || { echo "content: $wf/$node prompt does not come from a library file"; exit 1; }
    if [ -n "$page" ] && grep -q "$(basename "$page" .md)" "$tmp/mut.txt"; then
      grep -q "^$sentinel doctrine " "$tmp/mut.txt" || { echo "content: $wf/$node names $(basename "$page") but does not render it"; exit 1; }
    fi
  done
done

# ---- orphans ------------------------------------------------------------
(cd "$tmp/src" && go test -run 'TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/orphan.log" 2>&1) \
  && { echo "orphans: injected zz-uncited.md was not reported"; cat "$tmp/orphan.log"; exit 1; }
grep -q 'zz-uncited' "$tmp/orphan.log" || { echo "orphans: test failed but did not name zz-uncited.md"; cat "$tmp/orphan.log"; exit 1; }
go test -v -run 'TestLibraryRendersAll|TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/test.log" 2>&1 || { cat "$tmp/test.log"; exit 1; }
grep -q -- '--- PASS: TestLibraryRendersAll' "$tmp/test.log" || { echo "TestLibraryRendersAll did not run"; exit 1; }
grep -q -- '--- PASS: TestLibraryNoOrphans' "$tmp/test.log" || { echo "TestLibraryNoOrphans did not run"; exit 1; }
echo "show-and-orphan-walk.sh: ok"
