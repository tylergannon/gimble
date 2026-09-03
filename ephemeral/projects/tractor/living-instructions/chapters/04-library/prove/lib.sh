# Shared setup for the chapter 4 proof scripts. Sourced, not run.
# Provides: $root $tmp $seed, show_raw, show_headed, shown_nodes,
# shown_file, yaml_nodes, closure, build_copy (a copy of the tracked tree
# under $tmp/src with the Build dumper built at $tmp/bin/builddump and the
# binary at $tmp/bin/tractor), $workflows.
set -eu
root="$(git rev-parse --show-toplevel)"
cd "$root"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/demo" "$tmp/src"
printf 'Print a greeting.\n' > "$tmp/demo/seed.md"
seed="$tmp/demo/seed.md"
lib=workflow/library

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
shown_nodes() { show_headed "$1" "$2" | sed -n 's/^== \([^ ]*\).*/\1/p' | sort; }
shown_file() { show_headed "$1" "$2" | sed -n "s/^== $3 \(([a-z_.]*) \)\{0,1\}\(.*\)$/\2/p"; }
yaml_nodes() { awk '/^  - id:/{id=$3} /^    type:/{print id, $2}' "workflow/library/workflows/$1.yaml" | sort; }
yaml_ids() { awk '/^  - id:/{print $3}' "workflow/library/workflows/$1.yaml" | sort; }
# built_ids WORKFLOW: every node id Build returns, from the check's dumper.
built_ids() {
  if [ "$1" = plan ]; then "$tmp/bin/builddump" list "$1" "$tmp/demo" "$tmp/bin/tractor" "$seed"
  else "$tmp/bin/builddump" list "$1" "$tmp/demo" "$tmp/bin/tractor" ""; fi | sort
}
workflows="$(ls workflow/library/workflows/*.yaml | xargs -n1 basename | sed 's/\.yaml$//')"

# closure FILE: prints FILE and every prompt/supervisor/pass file reachable
# from it through include actions (an action naming another library file
# by path or by base name), transitively. Reads from $closure_lib.
closure_lib="$lib"
closure() {
  printf '%s\n' "$1" > "$tmp/cl.txt"
  while :; do
    before="$(wc -l < "$tmp/cl.txt")"
    while read -r f; do
      for cand in $(cd "$closure_lib" && find prompts supervisors passes templates -type f 2>/dev/null); do
        base="$(basename "$cand" .md)"
        if grep -q "include.*\"$base\"\|include.*\"$cand\"" "$closure_lib/$f" 2>/dev/null; then printf '%s\n' "$cand"; fi
      done
    done < "$tmp/cl.txt" >> "$tmp/cl.txt"
    sort -u "$tmp/cl.txt" -o "$tmp/cl.txt"
    [ "$(wc -l < "$tmp/cl.txt")" -eq "$before" ] && break
  done
  cat "$tmp/cl.txt"
}

# build_copy: copy the tracked tree, add the Build dumper, build both.
build_copy() {
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
	params := workflow.Parameters{
		Project: "demo", Workdir: workdir, Executable: exe,
		Plan: workflow.PlanParameters{Seed: seed},
	}
	// "list" mode: print every node id Build returns for the workflow
	// named by the second argument, synthesized branch nodes included.
	if name == "list" {
		g, err := workflow.Build(node, params)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		for _, n := range g.Nodes {
			fmt.Println(n.Base().ID)
		}
		return
	}
	// "render" mode: print the standalone rendering of one library file.
	if name == "render" {
		out, err := workflow.Render(node, params)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Print(out)
		return
	}
	g, err := workflow.Build(name, params)
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
  mkdir -p "$tmp/src/workflow/library/doctrine" "$tmp/src/workflow/library/templates" "$tmp/src/workflow/library/supervisors" "$tmp/src/workflow/library/passes"
  (cd "$tmp/src" && go build -o "$tmp/bin/tractor" ./cmd/tractor && go build -o "$tmp/bin/builddump" ./cmd/builddump)
}
