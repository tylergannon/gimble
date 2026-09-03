# P8: the library is content

Archetype: universal over the library's files. Exhaustive; no holdout.
Lap 6; answers `review-5.md`.

Reading of the promise: "prompt" means any text the library sends to an
agent, so files under `prompts/`, `supervisors/`, and `passes/` all
count as prompts for the orphan walk; the walk covers doctrine pages,
which P8 names, and not skeletons. `show` prints every node of the
graph, in any order, what `Build` materialized for each (prompt for any
node kind that carries one, command for tools, checklist for loops),
and, in the header, the library file each prompt came from (declaration
section 3); it never claims to reproduce frames.

## Story

Chapter 4 sprints 1 and 2; their proof scripts are the story. Sprint 3's
`prove/doctrine-pages.sh` demonstrates that sprint; it is not part of
P8's proof.

## Evidence

- The repository at chapter 4's end: `workflow/library/`,
  `workflow/library.go`, `workflow/testdata/`, and the README's list of
  the data fields templates may read (the template contract, section 4
  of the declaration).
- The graph files `workflow/library/workflows/*.yaml`: node ids and
  types, read by the check itself.
- `Build`'s output, captured by the check itself through a throwaway
  program it writes into a copy of the tree, for every node that
  carries a prompt, command, or checklist.
- Mutated copies of the tree the check makes: every library prompt file
  and one doctrine page with a sentinel naming that file appended; one
  uncited doctrine page added; and, separately, one doctrine page with a
  broken template action.
- Output of `prove/prompts-are-library-files.sh` and
  `prove/show-and-orphan-walk.sh`.
- At chapter 6: a `plan` run the check makes itself, and its
  `stages/<seq>-<node>/prompt.md`.

## Validator

`command`: the two sprint scripts, then at chapter 6
`prove/p8-show-stage.sh`.

- `prove/prompts-are-library-files.sh`: the sprint's tripwire (library
  files exist, no planner sentence in Go, snapshot test runs and
  passes). Not the proof.
- `prove/show-and-orphan-walk.sh`. Nodes: the set of `id`/`type` pairs
  the headed `show` prints equals the set the workflow YAML declares
  (order free). Equality: for every node of any kind that carries a
  prompt, command, or checklist, `show --raw` equals the check's own
  `Build` dumper; `--stage` on a stage built from the frame preamble
  plus the dumper's output exits 0, and 1 after a byte is appended.
  Content: each prompt-bearing node's header names its library file;
  with a distinct sentinel appended to every file under `prompts/`,
  `supervisors/`, and `passes/`, each node's `show --raw` contains the
  sentinel of the file its header names and no sentinel of any other
  prompt file; with a sentinel appended to a doctrine page, every
  prompt whose file names that page prints it. Data only: every line of
  each rendered prompt either appears verbatim in its library file or
  in a doctrine or template file, or is shorter than 200 bytes (a data
  value: a path, a name, a command); a Go-supplied instruction string
  rendered through a thin template fails this. Orphans: an injected
  `doctrine/zz-uncited.md` makes `TestLibraryNoOrphans` fail naming it.
  Rendering: an injected broken action in a doctrine page makes
  `TestLibraryRendersAll` fail naming the page. In the real tree both
  tests run and pass.
- `prove/p8-show-stage.sh`: runs `plan` on `seeds/greeter.md` into a
  fresh run directory, then for the first completed stage of each
  prompt-bearing node runs `show plan --node <n> --stage <that stage
  dir>` with the parameters the script passed to `run`, and expects
  exit 0.

No `infer`.

## Not proven

That the content is good. That `show` reproduces frames (by design;
research F1). That a data value under 200 bytes carries no instruction;
the README's field list is the contract and the code review reads the
struct.
