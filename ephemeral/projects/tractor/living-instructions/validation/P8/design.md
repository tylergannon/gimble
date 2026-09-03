# P8: the library is content

Archetype: universal over the library's files. Exhaustive; no holdout.
Lap 5; answers `review-4.md`.

Reading of the promise: "prompt" means any text the library sends to an
agent, so files under `prompts/`, `supervisors/`, and `passes/` all
count as prompts for the orphan walk; the walk covers doctrine pages,
which P8 names, and not skeletons. `show` prints every node of the
graph, in any order, and what `Build` materialized for each, with the
library file each prompt came from in the node header; it never claims
to reproduce frames.

## Story

Chapter 4 sprints 1 and 2; their proof scripts are the story. Sprint 3's
`prove/doctrine-pages.sh` demonstrates that sprint; it is not part of
P8's proof.

## Evidence

- The repository at chapter 4's end: `workflow/library/`,
  `workflow/library.go`, `workflow/testdata/`.
- The graph files `workflow/library/workflows/*.yaml`: node ids and
  types, read by the check itself.
- `Build`'s output, captured by the check itself through a throwaway
  program it writes into a copy of the tree.
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
  (order free). Equality: `show --raw` equals the check's own `Build`
  dumper for every node; `--stage` on a stage built from the frame
  preamble plus the dumper's output exits 0, and 1 after a byte is
  appended. Content: each codergen node's header names its library
  file; with a distinct sentinel appended to every file under
  `prompts/`, `supervisors/`, and `passes/`, each node's `show --raw`
  contains the sentinel of the file its header names and no sentinel
  of any other prompt file (a shared appendix cannot pass for a
  prompt); with a sentinel appended to a doctrine page, every prompt
  whose file names that page prints it. Orphans: an injected
  `doctrine/zz-uncited.md` makes `TestLibraryNoOrphans` fail naming it.
  Rendering: an injected broken action in a doctrine page makes
  `TestLibraryRendersAll` fail naming the page. In the real tree both
  tests run and pass. The mutations are the check's, so a hollow test
  under the right name fails here.
- `prove/p8-show-stage.sh`: runs `plan` on `seeds/greeter.md` into a
  fresh run directory, then for the first stage of each codergen node
  runs `show plan --node <n> --stage <that stage dir>` with the
  parameters the script passed to `run`, and expects exit 0.

No `infer`.

## Not proven

That the content is good. That `show` reproduces frames (by design;
research F1). That supervisor briefs and passes are rendered by a run
before chapter 5 exists; in chapter 4 they are files the walk and the
render test cover, and chapter 5's runs exercise them.
