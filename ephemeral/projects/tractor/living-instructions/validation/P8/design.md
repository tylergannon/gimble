# P8: the library is content

Archetype: universal over the library's files. Exhaustive; no holdout.
Lap 3; answers `review-2.md`.

Reading of the promise: "prompt" means any text the library sends to an
agent, so files under `prompts/`, `supervisors/`, and `passes/` all
count as prompts for the orphan walk; the walk covers doctrine pages,
which P8 names, and not skeletons, which it does not. `show` prints
what `Build` materialized; it never claims to reproduce frames.

## Story

Chapter 4 sprints 1 and 2; their proof scripts are the story. Sprint 3's
`prove/doctrine-pages.sh` demonstrates that sprint (which pages exist
and that each skeleton is cited); it is not part of P8's proof.

## Evidence

- The repository at chapter 4's end: `workflow/library/`,
  `workflow/library.go`, `workflow/testdata/`.
- `workflow/testdata/<workflow>/<node>.txt`: what `Build` returned for
  fixed parameters, captured by sprint 1 before any prompt left Go.
- A mutated copy of the source tree the check builds itself: every
  library prompt file with a sentinel line appended, one doctrine page
  with a sentinel appended, and one uncited doctrine page added.
- Output of `prove/prompts-are-library-files.sh` and
  `prove/show-and-orphan-walk.sh`.
- At chapter 6: a `plan` run the check makes itself, and its
  `stages/<seq>-<node>/prompt.md`.

## Validator

`command`: the two sprint scripts, then at chapter 6
`prove/p8-show-stage.sh`.

- `prove/prompts-are-library-files.sh`: library files exist; the prompt
  functions and the planner's opening sentence are absent from Go; the
  snapshot files exist and `TestBuildMatchesSnapshot` runs and passes
  (`-v`, `--- PASS:` line required).
- `prove/show-and-orphan-walk.sh`, in three parts. Content: in a copy
  of the tracked tree, append a sentinel line to every file under
  `prompts/`, `supervisors/`, and `passes/` and to one doctrine page,
  build, and require that `show --raw` for every node of every workflow
  prints its file's sentinel, and that every prompt citing the page
  prints the page's sentinel; a prompt still living in Go cannot pass.
  Equality: in the real tree, for every node header the headed `show`
  prints (so an omitted node is visible), `show --raw` with the
  snapshot's fixed parameters (real paths substituted back) is
  byte-equal to the snapshot; `--stage` on a stage built from the frame
  preamble plus the snapshot exits 0 and exits 1 after a byte is
  appended. Orphans: in the copy, add `doctrine/zz-uncited.md` and
  require `go test -run TestLibraryNoOrphans` to fail naming it; in the
  real tree require `TestLibraryRendersAll` and `TestLibraryNoOrphans`
  to run and pass. The injected orphan and sentinels are the check's,
  not the coder's tests'.
- `prove/p8-show-stage.sh`: runs `plan` on `seeds/greeter.md` into a
  fresh run directory, then for the first stage of each codergen node
  runs `show plan --node <n> --stage <that stage dir>` with the
  parameters the script passed to `run`, and expects exit 0.

No `infer`.

## Not proven

That the content is good. That `show` reproduces frames (by design;
research F1). That no prompt text is assembled in Go from fragments
that never reach a library file: the sentinel test proves the library
files are what `Build` renders, and the code review reads the rest.
