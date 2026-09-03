# P8: the library is content

Archetype: universal over the library's files. Exhaustive; no holdout.

## Story

Chapter 4's four sprints; their proof scripts are the story.

## Evidence

- The repository at chapter 4's end: `workflow/library/` tree,
  `workflow/library.go`, tests.
- `prove/prompts-are-library-files.sh`,
  `prove/show-and-orphan-walk.sh`, `prove/doctrine-pages.sh` output.
- A recorded stage from any real run of `plan` (P10's run) for the
  `--stage` leg.

## Validator

The three proof scripts under `chapters/04-library/prove/` plus, at
chapter 6, `prove/p8-show-stage.sh`: `workflow show plan --node planner
--stage <recorded stage> --project … --seed … --workdir …` with the
recorded run's parameters exits 0.

`infer` on the doctrine pages as written in the chapter 4 ledger.

## Not proven

That the content is good. That `show` reproduces frames (it does not,
by design; research F1).
