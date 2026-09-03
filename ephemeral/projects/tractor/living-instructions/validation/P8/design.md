# P8: the library is content

Archetype: universal over the library's files. Exhaustive; no holdout.
Lap 2; answers `review-1.md`.

Reading of the promise: "prompt" means any text the library sends to an
agent, so files under `prompts/`, `supervisors/`, and `passes/` all
count as prompts for the orphan walk. `show` prints what `Build`
materialized; it never claims to reproduce frames.

## Story

Chapter 4 sprints 1 and 2; their proof scripts are the story. Sprint 3's
`prove/doctrine-pages.sh` demonstrates that sprint (which pages exist);
it is not part of P8's proof and P8 makes no claim about page names or
count.

## Evidence

- The repository at chapter 4's end: `workflow/library/`,
  `workflow/library.go`, `workflow/testdata/`.
- `workflow/testdata/<workflow>/<node>.txt`: the snapshot of what
  `Build` returned for fixed parameters, captured by sprint 1 before any
  prompt moved out of Go. The fixed parameters are in the snapshot test.
- Output of `prove/prompts-are-library-files.sh` and
  `prove/show-and-orphan-walk.sh`.
- At chapter 6: a `plan` run the check makes itself, and its
  `stages/<seq>-<node>/prompt.md`.

## Validator

`command`: the two sprint scripts, then at chapter 6
`prove/p8-show-stage.sh`.

- `prove/prompts-are-library-files.sh`: the library files exist; the
  four prompt functions and the planner's opening sentence are absent
  from Go; `go:embed` and `text/template` are used; the snapshot test
  `TestBuildMatchesSnapshot` runs and passes (`go test -v -run` with the
  `--- PASS:` line checked, so a missing test is a fail).
- `prove/show-and-orphan-walk.sh`: for every codergen node of every
  workflow, `tractor workflow show <wf> --node <n> --raw` with the
  snapshot's fixed parameters (the script substitutes its own workdir
  and binary paths for the snapshot's fixed ones) is byte-equal to the
  snapshot file. The snapshot was made from `Build` before the
  migration, so `show` is bound to `Build`, not to itself. `--stage` on
  a stage directory built from the frame preamble plus the snapshot
  file exits 0, and exits 1 after one byte is appended.
  `TestLibraryRendersAll` and `TestLibraryNoOrphans` run by name with
  their `--- PASS:` lines checked; the orphan test proves itself on a
  synthetic uncited page.
- `prove/p8-show-stage.sh`: runs `plan` on `seeds/greeter.md` into a
  fresh run directory, then for the first stage of each codergen node
  runs `show plan --node <n> --stage <that stage dir>` with the
  parameters the script itself passed to `run`, and expects exit 0. The
  stage is real and made by the check, not supplied.

No `infer`.

## Not proven

That the content is good. That `show` reproduces frames (it does not,
by design; research F1). That `show` calls `Build` rather than
re-rendering the templates: byte equality with the pre-migration
snapshot leaves an independent renderer nothing to gain, and the code
review reads the call.
