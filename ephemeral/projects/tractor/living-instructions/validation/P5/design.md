# P5: all seven review passes are engine-marked after a fresh reviewer on another provider routed pass

Archetype: universal over the passes of a run. Exhaustive; no holdout.
Lap 2; answers `review-1.md`.

## Story

Any `plan` run the check makes itself at check time (the same run P3
uses; `prove/p3-independent-review.sh` and this script share it through
`validation/lib/run-plan.sh`).

## Evidence

- `plan-review/ledger.md` from the package.
- `timeline.jsonl`: `LoopItemSelected`, `LoopValidated`, and
  `StageCompleted(next)` for `reviewer` and for the owning nodes.
- `checkpoint.json` `sessions`: harness and thread-mode marker for
  `reviewer` and for every planner node (`intake`, `brief`, `decompose`,
  `design`, `assemble`).
- Segments of every planner stage: `tool_call` arguments.
- `plan-review/<pass>/` findings files for any pass that failed at least
  once.

## Validator

`command`: `prove/p5-seven-passes.sh`: `plan-review/ledger.md` names the
seven built-in passes in the documented order as its first seven items,
all `done: true`; each has a `LoopValidated` event on the plan-review
loop with `passed: true`; for each, the nearest preceding
`StageCompleted` is a `reviewer` stage with `next:` the loop node; for
every `reviewer` `StageCompleted` whose `next` is `brief`, `decompose`,
or `design`, the next `StageStarted` is that node and a later
`LoopItemSelected` names the same pass again; in `checkpoint.json` the
`reviewer` harness differs from the harness of every planner node and
the `reviewer` key carries the no-thread marker; no planner segment has
a `tool_call` whose arguments name `plan-review/ledger.md`.

No `infer`.

## Not proven

That the passes catch every defect, or that seven is the right number.
Project-added passes: a run cannot complete with one open, so they need
no check. The model within the provider (see P3).
