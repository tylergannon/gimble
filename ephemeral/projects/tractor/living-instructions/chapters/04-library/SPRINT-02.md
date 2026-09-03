# Sprint 2: `show`, the render test, the orphan walk

Three things that make the library safe to edit.

## `tractor workflow show <name>`

Same flags as `run` (`--project`, `--seed`, `--workdir`), no `--logs`.
Calls `Build` exactly as `run` does and prints, for every node in the
graph in file order:

```
== <node id> (<type>) <library file, for codergen nodes>
<prompt text, or tool_command, or checklist path>
```

The library file in the header is the path under `workflow/library/`
of the prompt template the node was rendered from (for example
`prompts/plan/planner.md`). The proof script reads it to know which
file's sentinel each node must carry.

Nothing else: no frames, no `$goal` expansion, no banner claiming
equality with a run. Register it beside `list` and `run`
(`cmd/tractor/workflow.go:31`); reuse the `Parameters` construction at
`:108-127`. `workflow list` output and the root help set are asserted
exactly by existing tests and must not change.

`--node <id> --raw` prints one node's prompt (or command, or checklist
path) with no header. The proof script compares this against a program
of its own that calls `Build`, so `show` must print what `Build`
returns and nothing it renders on its own.

`--values` prints the template data struct for the given parameters as
`Field: value` lines, one per field the README lists, and nothing else.
The proof script uses it to know every data value a rendered prompt
may contain.

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
`doctrine/` must be named by a `doctrine` action in at least one file
under `prompts/`, `supervisors/`, or `passes/`. All three directories
are prompts in P8's sense: text the library sends to an agent.
Skeletons under `templates/` are not in the walk; P8 does not promise
that every skeleton is used, and sprint 3's script checks its own
skeletons are cited. Prove the walk fails: the test constructs a
synthetic `fs.FS` with one uncited page and asserts the walk reports it
by name. This needs the `BuildFrom(fs.FS)` or equivalent seam from
sprint 1. The proof script injects its own uncited page into a copy of
the tree as well, so the walk is proven against the real embedded tree,
not only the test's synthetic one.

The walk is written from scratch; no surveyed tool has one
(`research/prompt-libraries/goose.md` has the inverse and still drifted).

## Proof script

`prove/show-and-orphan-walk.sh`, in a copy of the tracked tree it makes
itself. Nodes: the headed `show` lists exactly the set of `id`/`type`
pairs the workflow YAML declares, in any order. Equality: the script writes a small Go
program into the copy that calls `workflow.Build` and prints one node's
prompt, command, or checklist; `show --raw` must equal it for every
node of every workflow, and a stage directory built from the committed
frame preamble (`fixtures/frame-preamble.txt`), a synthetic iterate
block, and that program's output makes `--stage` exit 0, and 1 after
one byte is appended. Content: with a sentinel line appended to every
file under `prompts/`, `supervisors/`, and `passes/` and to one doctrine
page, every codergen node's `show --raw` prints its file's sentinel and
every prompt naming the page prints the page's sentinel. Data only: with every data value (from `--values`) removed from each
rendered line and every template action removed from each library
line, every rendered line must equal some library line. Unused: every
file under `prompts/` and `supervisors/` is in the include closure of
some node's header file. Orphans: an
injected `doctrine/zz-uncited.md` makes `go test -run
TestLibraryNoOrphans` fail naming it; rendering: an unclosed action
appended to a doctrine page (using the delimiters the README states)
makes `TestLibraryRendersAll` fail naming the page; in the real tree the two tests
run by name with `-v` and their `--- PASS:` lines are required, since
`go test -run` with no matching test exits 0. If the graph API names in
the script's program differ from the sprint's, fix the program, never
the comparison.

## Ask the reviewer

- Whether `show` should also print supervisor briefs once supervisors
  exist (chapter 5). Recommend yes, as `== <id> (supervisor)`; nothing to
  do now.
