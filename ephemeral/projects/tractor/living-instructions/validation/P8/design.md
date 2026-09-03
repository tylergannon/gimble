# P8: the library is content

Archetype: universal over the library's files. Exhaustive; no holdout.
Lap 13; answers `review-12.md`.

Reading of the promise: "prompt" means any text the library sends to an
agent, so files under `prompts/`, `supervisors/`, and `passes/` all
count as prompts for the orphan walk; the walk covers doctrine pages,
which P8 names, and not skeletons. `show` prints every node of the
graph, in any order, with the node's type as the graph declares it
(`parallel.fan_in` included), what `Build` materialized for each (prompt for any
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
- The graph files `workflow/library/workflows/*.yaml`: node ids and
  types, read by the check itself.
- `Build`'s output, captured by the check itself through a throwaway
  program it writes into a copy of the tree, for every node that
  carries a prompt, command, or checklist.
- The include closure of each header file: the files named by
  `include` and `doctrine` actions in it, transitively, read by the
  check from the files themselves; and the doctrine and template files
  those name.
- The data values for the check's parameters, computed by the check
  itself from the parameters it supplied and the derivations the README
  states, in raw, Go-quoted (`quote`), and shell-quoted (`shell`) form;
  `show --values` output for the same parameters, which must agree.
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

## Validator

`command`: the three sprint scripts, then at chapter 5
`prove/p8-passes-are-files.sh`, then at chapter 6
`prove/p8-show-stage.sh`.

- `prove/prompts-are-library-files.sh`: the sprint's tripwire (library
  files exist, no planner sentence in Go, snapshot test runs and
  passes). Not the proof.
- `prove/show-equals-build.sh` (sprint 2) and
  `prove/orphan-walk-and-render.sh` (sprint 3). Nodes: the set of
  `id`/`type` pairs
  the headed `show` prints equals the set the workflow YAML declares
  (order free). Equality: for every node of any kind that carries a
  prompt, command, or checklist, `show --raw` equals the check's own
  `Build` dumper; `--stage` on a stage built from the frame preamble
plus the dumper's output exits 0, and 1 after one random byte at a
random offset within the prompt body (past the frame the check itself
wrote) is changed (a comparator that recognises a fixed suffix cannot
pass; a byte in the frame is never the probe, since a correct `--stage`
ignores the frame).
  Content: each prompt-bearing node's header names its library file;
  with a distinct sentinel appended to every file under `prompts/`,
  `supervisors/`, and `passes/`, each node's `show --raw` contains the
  sentinel of the file its header names and contains no sentinel of a
  prompt file outside that file's include closure (a conditional
  include that does not fire for the check's parameters is allowed to
  be absent; a file outside the closure is not allowed to appear); with a sentinel appended to a
  doctrine page, every prompt whose closure names that page prints it.
  No unused prompt file: every file under `prompts/` and `supervisors/`
  is in the closure of some node's header file in some workflow (a file
  no node renders is not a prompt, so it cannot be the citation that
  keeps a doctrine page alive); every file under `passes/` is covered
  by the chapter 5 leg. Text from the closure only: the check computes
  the data values itself (the parameters it passed and the
  README-stated derivations, in raw, `quote`d, and `shell`-quoted
  forms) and requires `show --values` to agree with them; then, for
  each node, it removes every data value form from each rendered line,
  removes every template action and template comment from each line of
  the node's closure (its header file, the prompt files it includes,
  and the doctrine and template files any of them name), and requires
  every rendered residue line to equal some residue line of that
  closure. Text parked in another file, in a comment, or supplied by Go
  fails; a `quote`d or `shell`-quoted value passes; a fragment included
  twice passes. Orphans: an injected uncited page, named at random per run, makes
  `TestLibraryNoOrphans` fail naming it, and so does removing every
  citation of one existing page chosen at random (a test that allows
  the original filenames and rejects only additions fails the second
  probe). Rendering: an unclosed action
  with random text appended to a doctrine page chosen at random makes
  `TestLibraryRendersAll` fail naming the page. In the real tree both
  tests run and pass.
- `prove/p8-passes-are-files.sh` (chapter 5): in a copy of the tree
  with a distinct sentinel appended to every file under `passes/`, a
  `plan` run's generated `plan-review/ledger.md` has one item per pass
  file carrying that file's sentinel in its `check` or `doc`, and every
  reviewer stage's `prompt.md` contains its pass's sentinel; a pass
  carried as a Go string, or one pass file-backed among six Go strings,
  cannot pass this.
- `prove/p8-show-stage.sh` (chapter 6): runs `plan` on
  `seeds/greeter.md` into a fresh run directory, then for the first
  completed stage of each prompt-bearing node runs `show plan --node
  <n> --stage <that stage dir>` with the parameters the script passed
  to `run`, and expects exit 0.

`infer` (files: the two proof scripts and their logs under
`prove/last-run/`, which include the `go test` output of every probe
run copied out of the scratch directory before it is removed;
`workflow/library_test.go`; the non-test Go files under `workflow/`): "For each probe the log
names (the random orphan page, the page whose citation was removed, the
random broken action, the random stage perturbation, the per-file
sentinels), did the test or check fail for the injected reason and name
the injected file, and does the test's code walk the tree rather than
recognise probe names? Read the Go under `workflow/`: does any Go code
supply prompt text beyond the data values the README lists and the
four functions, or rely on a template branch that never renders? Fail
if any probe's failure is generic, names the wrong file, the test
special-cases probes, or Go carries prompt text." This is the model
judgment over recorded evidence that decision 41 requires; the scripts
record, the judge decides.

## Not proven

That the content is good. That `show` reproduces frames (by design;
research F1). That a data value carries no instruction; the README's
field list is the contract, the check recomputes every value, and the
code review reads the struct. Text a template parks in a branch that never renders and Go re-emits:
the residue check cannot tell it from rendering, so the infer judge
reads the Go under `workflow/` for exactly that.
Skeleton text: templates are cited by prompts (sprint 4's script) and
rendered by the render test; what a planner writes from them is its
own output, not the library's.
