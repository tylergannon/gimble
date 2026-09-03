# P8: the library is content

Archetype: universal over the library's files. Exhaustive; no holdout.
Lap 26; answers `review-25.md` and pass 4's fourth lap.

Reading of the promise: "prompt" means any text the library sends to an
agent, so files under `prompts/`, `supervisors/`, and `passes/` all
count as prompts for the orphan walk; the walk covers doctrine pages,
which P8 names, and not skeletons; a page is referenced only if a `doctrine` action in an agent-facing
file (a prompt body, supervisor brief, or pass; never another doctrine
page) renders it for some node under at least one of the representative
parameter sets the library README lists (the sets the render test
uses), so a `doctrine` action in a branch that never renders is no
reference and a hub page that includes the others does not reference
them. The
README's list is part of the template contract (declaration section
4): an author who adds a page rendered only under some parameter adds
a set that renders it, or the page is an orphan by the library's own
rule; skeletons are library files that
prompts include, and the sentinel test covers them like any included
file. `show` prints every node of the
graph, in any order, optionally with its type, what `Build` materialized for each (prompt for any
node kind that carries one, command for tools, checklist for loops),
and, in the header, the library file each prompt came from (declaration
section 3); a prompt may be composed from that file and the files it
includes, and may include a file more than once. `show` never claims to
reproduce frames.

## Story

Chapter 4 sprints 1 to 3; their proof scripts are the story. Sprint 4's
`prove/doctrine-pages.sh` demonstrates that sprint; it is not part of
P8's proof. Passes and the plan-review ledger arrive in chapter 5; the
pass leg of the check below is empty until then and runs then.

## Evidence

- The repository at chapter 4's end: `workflow/library/`,
  `workflow/library.go`, `workflow/testdata/`, and the README's list of
  the data fields templates may read (the template contract, section 4
  of the declaration) with how each derives from `Parameters`.
- The node census of each graph as `Build` returns it (`g.Nodes`,
  synthesized branch nodes included), printed by the check's own
  program in its `list` mode; the YAML is not the universe, the built
  graph is.
- `Build`'s output, captured by the check itself through a throwaway
  program it writes into a copy of the tree, for every node that
  carries a prompt, command, or checklist; and, through the same
  program, the standalone rendering of the node's header file
  (`workflow.Render`, the seam sprint 1 exports for the render test).
- The include closure of each header file: the prompt, supervisor,
  pass, and template files named by `include` actions in it,
  transitively, read by the check from the files themselves. Doctrine
  pages are checked separately (each renders under some listed
  parameter set); they are not part of the sentinel closure.
- The data values for the check's parameters, computed by the check
  itself from the parameters it supplied and the derivations the README
  states; `show --values` output for the same parameters, which must
  list exactly them.
- Mutated copies of the tree the check makes: every library prompt file
  and one doctrine page with a sentinel naming that file appended; one
  uncited doctrine page added under a name the check draws at random
  each run; and, separately, one doctrine page chosen at random with a
  broken template action whose text the check draws at random. A test
  that recognises a fixed probe cannot pass twice.
- At chapter 5 and after: a `plan` run the check makes itself from a
  copy of the tree in which every file under `passes/` carries its own
  sentinel; the `plan-review/ledger.md` it generated and every reviewer
  stage's `prompt.md`.
- At chapter 6: the same run's `stages/<seq>-<node>/prompt.md`.

## Verifier

`command`: the three sprint scripts, then at chapter 5
`prove/p8-passes-are-files.sh`, then at chapter 6
`prove/p8-show-stage.sh`.

- `prove/prompts-are-library-files.sh`: at sprint 1, every node's
  `show --raw` equals the payload the pre-migration `Build` produced,
  built by the check from the commit recorded in `prove/base-commit.txt`
  (an immutable baseline the coder cannot update). From sprint 4 on,
  when the prompts legitimately change to include doctrine, the script
  prints the unified diff between that baseline and the current output
  for every node into its log instead of failing, and the ledger item's
  judge reads the diffs: every difference is text rendered from a library file (a moved
  passage, an include, or a rewording that lives in the library),
  never text from Go; what the words say is not P8's concern (the
  declaration excludes content quality), and whether sprint 4 lost
  anything is sprint 4's own judge's question, not this one's. The coder's snapshot test is a tripwire, not the
  baseline. The judge also reads the Go for prompt text in strings.
- `prove/show-equals-build.sh` (sprint 2) and
  `prove/orphan-walk-and-render.sh` (sprint 3). Nodes: the set of
node ids the headed `show` prints equals the set of ids in `g.Nodes`
as the check's dumper lists them from `Build` (order free; synthesized
branch nodes included; the type in the header is a convenience, not a
requirement), and the body the headed output prints under each header
equals that node's `--raw` output (the headed form is the primary
output and carries every payload, not only headers). Equality: for every node of any kind that carries a
  prompt, command, or checklist, `show --raw` equals the check's own
  `Build` dumper; `--stage` on a stage built from the frame preamble plus two nested
  `<iterate>` blocks (an outer chapters frame and an inner sprints
  frame, the shape `large` records) plus the dumper's output exits 0,
  and, after one random byte at a random offset within the prompt body
  (past the frame the check itself wrote) is changed, exits exactly 1 with a diff on stdout that names the perturbed line
  (the perturbing byte is a printable ASCII character different from
  the original, so the stage stays valid text; an exit of 2, a panic,
  or silence is a fail; the log records the offset, the exit
  status, and the diff; a comparator that recognises a fixed suffix
  cannot pass; a byte in the frame is never the probe, since a correct
  `--stage` ignores the frame).
  Content: each prompt-bearing node (a node whose kind, as the check's
  dumper lists it from `Build`, is codergen, supervisor, fan-in, or
  parallel-with-prompt; never a whitespace heuristic) has a header that
  names its library file;
  with a distinct sentinel appended to every file under `prompts/`,
  `supervisors/`, `passes/`, and `templates/`, each node's `show --raw`
  contains the sentinel of the file its header names and contains no
  sentinel of a file outside that file's include closure, and every
  skeleton the library README lists as the planner's (the list sprint 4
  writes) has its sentinel in some node's output (each listed skeleton
  is a library file some prompt includes; an inert file listed beside
  an inline skeleton fails; a file under `templates/` not in the list
  may exist unused) (a conditional
  include that does not fire for the check's parameters is allowed to
  be absent; a file outside the closure is not allowed to appear); with a sentinel appended to every doctrine page, each page's
  sentinel appears in some node's `show --raw` under at least one of
  the README's representative parameter sets (the check runs `show`
  once per listed set; a page no listed set renders is an orphan by
  the library's own definition, whatever text names it).
  A citation counts for the orphan walk only from a file some node
  renders (the walk's own rule, sprint 3), and the render test requires
  every file under `prompts/`, `supervisors/`, and `passes/` to render
  under some listed parameter set, so an unused agent-facing file fails
  the render test by name. Rendering is the whole of `Build`: `show --raw` for each node equals
  the check's own standalone `workflow.Render` of the node's header
  file, so `Build` adds nothing to what the library file renders; the
  check computes the data values itself and requires `show --values`
  to list exactly them; what `Render` itself could add beyond those
  values is Go the judge reads (below). Orphans: an injected uncited page, named at random per run, makes
  `TestLibraryNoOrphans` fail naming it, and so does removing every citation of one existing page chosen at
  random from the agent-facing files (a test that allows the original
  filenames and rejects only additions fails the second probe; a page
  cited only from another doctrine page is an orphan, so a hub cannot
  hide one). Rendering: an unclosed action
  with random text appended to a doctrine page chosen at random makes
  `TestLibraryRendersAll` fail naming the page. In the real tree both
  tests run and pass.
- `prove/p8-passes-are-files.sh` (chapter 5): in a copy of the tree
  with a distinct sentinel appended to every file under `passes/`, a
  `plan` run's generated `plan-review/ledger.md` has one item per pass
  file carrying that file's sentinel in its `check` or `doc`, and every
  reviewer stage's `prompt.md`, frame stripped, passes the closure-text
  check against its pass file and the reviewer prompt's closure (every
  line is library text plus data); a pass carried as a Go string, one
  pass file-backed among six, or a Go body with a tiny library file
  concatenated, cannot pass this.
- `prove/p8-show-stage.sh` (chapter 6): runs `plan` on
  `seeds/greeter.md` into a fresh run directory, then for the first
  completed stage of each codergen node (a fan-in's `prompt.md` carries
  the branch results the engine appends at run time and is not diffed;
  a supervisor has no stage) runs `show plan --node <n> --stage <that
  stage dir> --goal <the run's goal from manifest.json>` with the parameters the script passed to `run`, and
  expects exit 0 (`--stage` strips the frame and, given `--goal`,
  expands `$goal` as the engine does, so a library prompt that uses
  `$goal` is not a false diff); does the same for the first `implement`
  stage of P10's `medium` run and for the first `implement` stage of
  P10's `large` run, whose `prompt.md` carries the
  nested chapters-and-sprints frame, so a `--stage` that strips only
  one frame fails there.

`infer` (files: the proof tooling that produced the evidence
(`validation/observer.sh`, `validation/lib/`, the proof script named
above) and the segment of every turn this design judges; every proof script
this design names, `prove/lib.sh`, `prove/base-commit.txt` (whose hash
the ledger command itself carries, so a moved pin fails the command
before the judge sees anything), `SPRINT-04.md`, and each script's log
under `prove/last-run/` (chapter 4's two, chapter 5's pass leg with the
generated `plan-review/ledger.md` and every reviewer stage's
`prompt.md` of its run, chapter 6's stage leg with the stage
directories it diffed), the library files under `workflow/library/`,
the non-test Go of every package in this module that `workflow/` or
`cmd/tractor/` imports, transitively (the script writes `go list -deps`
for both into the log, so a helper package that implements `include`
is in scope wherever it lives): each ledger command captures its script's stdout,
which names every probe and its outcome (the stage perturbation's byte
and offset, each node's sentinel result, each skeleton's rendering),
to `prove/last-run/<script>.log`, and the orphan script copies the `go
test` output of every probe run beside it before the scratch directory
is removed;
`workflow/library_test.go`): "For each probe the logs name (the random orphan page, the page whose
citation was removed, the random broken action, the random stage
perturbation, the per-file sentinels, the per-pass sentinels in the
generated review ledger and reviewer prompts, and the real stages
diffed), did the test or check fail or pass for the reason the script
names and name the right file, and does the test's code walk the tree rather than
recognise probe names? Read the Go under `workflow/` and `cmd/tractor/workflow.go`: does any
Go code supply prompt text beyond the data values the README lists and
the four functions, rely on a template branch that never renders, or
alter any node, supervisor briefs included, anywhere between `Build`
and `engine.NewRunner`? Fail
if any probe's failure is generic, names the wrong file, the test
special-cases probes, or Go carries prompt text." This is the model
judgment over recorded evidence that decision 41 requires; the scripts
record, the judge decides.

## Not proven

That a Go package outside this module (a dependency) supplies prompt
text; the judge reads this module. That the content is good. That `show` reproduces frames (by design;
research F1). That a data value carries no instruction; the README's
field list is the contract, the check recomputes every value, and the
code review reads the struct. Text a template parks in a branch that never renders and Go re-emits:
the residue check cannot tell it from rendering, so the infer judge
reads the Go under `workflow/` for exactly that.
What a planner writes from a skeleton is its own output, not the
library's; the skeleton itself is covered by the sentinel test.
