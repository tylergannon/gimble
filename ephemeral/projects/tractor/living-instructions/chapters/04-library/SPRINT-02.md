# Sprint 2: `show`, the render test, the orphan walk

Three things that make the library safe to edit.

## `tractor workflow show <name>`

Same flags as `run` (`--project`, `--seed`, `--workdir`), no `--logs`.
Calls `Build` exactly as `run` does and prints, for each node in file
order:

```
== <node id> (<type>)
<prompt text, or tool_command, or checklist path>
```

Nothing else: no frames, no `$goal` expansion, no banner claiming
equality with a run. Register it beside `list` and `run`
(`cmd/tractor/workflow.go:31`); reuse the `Parameters` construction at
`:108-127`. `workflow list` output and the root help set are asserted
exactly by existing tests and must not change.

`--node <id> --raw` prints one node's prompt (or command, or checklist
path) with no header. The proof script compares this against the sprint
1 snapshot for the same parameters, so `show` must print what `Build`
returns and nothing it renders on its own.

`--stage <dir>`: read `<dir>/prompt.md`, strip the frame (everything
from the start through the end of the outermost `</iterate>` block and
the preamble before it; see `engine/frames.go:108-146` for the shape),
and print a unified diff against the node's prompt. `--stage` requires
`--node`. Exit 0 on no diff, 1 on diff. This is the tool that chapter
6's proof uses against a run it makes; here it needs a unit test with a
hand-built stage directory.

## Render test

`TestLibraryRendersAll`: for every workflow in the library, `Build` with
representative parameters and assert every codergen prompt is non-empty
and every template under `prompts/` was executed at least once. Also
execute every file under `doctrine/` and `templates/` standalone with
the same data so a syntax error in a page fails here, not at run time.

## Orphan walk

`TestLibraryNoOrphans`: walk the embedded tree; every file under
`doctrine/` and `templates/` must be named by an `include` or `doctrine`
action in at least one file under `prompts/`, `supervisors/`, or
`passes/`. All three directories are prompts in P8's sense: text the
library sends to an agent. Prove it fails: the test constructs a
synthetic `fs.FS` with one uncited page and asserts the walk reports it
by name. This needs the `BuildFrom(fs.FS)` or equivalent seam from
sprint 1.

The walk is written from scratch; no surveyed tool has one
(`research/prompt-libraries/goose.md` has the inverse and still drifted).

## Proof script

`prove/show-and-orphan-walk.sh`: builds the binary; for every file under
`workflow/testdata/<workflow>/`, runs `show <workflow> --node <node>
--raw` with the snapshot's fixed parameters (real paths substituted
back) and requires byte equality with the snapshot, which was captured
from `Build` before the migration; asserts the headed output names every
node and prints no frame; builds a stage directory from the committed
frame preamble (`fixtures/frame-preamble.txt`), a synthetic iterate
block, and the planner snapshot, and asserts `--stage` exits 0 on it and
1 after one byte is appended; runs the two tests by name with `-v` and
requires their `--- PASS:` lines, since `go test -run` with no matching
test exits 0.

## Ask the reviewer

- Whether `show` should also print supervisor briefs once supervisors
  exist (chapter 5). Recommend yes, as `== <id> (supervisor)`; nothing to
  do now.
