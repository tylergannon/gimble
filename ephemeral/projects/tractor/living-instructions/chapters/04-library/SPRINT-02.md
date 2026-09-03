# Sprint 2: `tractor workflow show`

One command that prints what `Build` materialized, so an editor can see
the effect of a content change without running a pipeline. Exercisable
at the end: `tractor workflow show plan --project demo --seed s.md`
prints every node of the plan graph with its prompt.

## Read first

- `cmd/tractor/workflow.go:31` (subcommand registration) and `:108-127`
  (the `Parameters` construction `run` uses; reuse it).
- `engine/frames.go:108-146` (the frame shape `--stage` strips).
- `research/workflow-package-inventory/migration-inventory.md` §4.

## The work

`tractor workflow show <name>` shares `run`'s parameter flags
(`--project`, `--seed`, `--workdir`) and has no `--logs`. It calls `Build`
exactly as `run` does and prints, for every node `Build` returns
(synthesized branch nodes included), in any stable order:

```
== <node id> [(<type>)] <library file, for nodes that carry a prompt>
<prompt text, or tool_command, or checklist path>
```

The library file is the path under `workflow/library/` of the template
the node was rendered from (for example `prompts/plan/planner.md`).
Nothing else: no frames, no `$goal` expansion, no banner claiming
equality with a run. `workflow list` output and the root help set are
asserted exactly by existing tests and must not change.

`--node <id> --raw` prints one node's prompt (or command, or checklist
path) with no header.

`--values` prints the template data struct for the given parameters as
`Field: value` lines, one per field the README lists, and nothing else.

`--stage <dir>` (requires `--node`): read `<dir>/prompt.md`, strip the
frame (everything from the start through the end of the outermost
`</iterate>` block and the preamble before it), expand `$goal` in the
node's prompt when `--goal <text>` is given (the engine's own
replacement, so a library prompt that uses `$goal` diffs clean against
its stage), and print a unified diff against the result. Exit 0 on no diff, 1 on diff. Unit test
with a hand-built stage directory.

## Definition of done

The ledger item's command runs `prove/show-equals-build.sh`, which
proves: the set of node ids `show` prints equals the set `Build`
returns, as a program the script writes lists them; for every node `show --raw` is byte-equal to a
program the script writes into a copy of the tree that calls
`workflow.Build` directly; `--stage` exits 0 on a stage built from the
frame preamble (`fixtures/frame-preamble.txt`) plus that program's
output and 1 after one byte is appended; `--values` reports no value
the script did not derive itself from the parameters and the README's
derivations. If the graph API names in the script's program differ
from the sprint's, fix the program, never the comparison.

## Not in this sprint

The render test, the orphan walk, sentinel and mutation proofs
(sprint 3). Doctrine pages (sprint 4). Docs (sprint 5). Printing
provider and model per node (chapter 5 sprint 1 adds it).

## Later

Supervisor briefs print as `== <id> (supervisor) <file>` once supervisor
nodes exist; chapter 5's supervisor sprint adds that, and this sprint's
node loop already handles any node kind that carries a prompt.
