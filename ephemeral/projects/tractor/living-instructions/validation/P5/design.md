# P5: all seven review passes are engine-marked after a fresh reviewer on another provider routed pass

Archetype: universal over the passes of a run. Exhaustive; no holdout.
Lap 3; answers `review-2.md`.

## Story

Any `plan` run the check makes itself at check time (the same run P3
uses, through `validation/lib/run-plan.sh`).

## Evidence

- `plan-review/ledger.md` from the package, and `show`'s `checklist`
  path for the plan-review loop.
- `timeline.jsonl`: `LoopItemSelected`, `LoopValidated`, `StageStarted`,
  `StageCompleted(next)` for `reviewer` and the owning nodes.
- `observer/` snapshots of `plan-review/ledger.md` at every stage
  boundary.
- `stages/<seq>-reviewer/response.md` for every reviewer turn.
- `checkpoint.json` `sessions`: harness and key for `reviewer`, harness
  for every planner node (`intake`, `brief`, `decompose`, `design`,
  `assemble`).

## Validator

`command`: `prove/p5-seven-passes.sh`: the plan-review loop's
`checklist` is the package's `plan-review/ledger.md`; its first seven
items are the seven built-in passes in the documented order, all
`done: true`; at the `StageCompleted` of `assemble` none is done (the
planner wrote the ledger open); for each pass exactly one loop stage
flips it, with a `LoopValidated` `passed: true` for that pass during
that stage and no change across any other stage; for each such loop
stage the nearest preceding `StageCompleted` is a `reviewer` stage
whose `next` is the loop node, whose `response.md` front matter agrees
and whose body contains `ROUTE: pass`; for every `reviewer` stage whose
`next` is `brief`, `decompose`, or `design`, its `response.md` contains
`ROUTE: fail`, the next `StageStarted` is that node, and a later
`LoopItemSelected` names the same pass; every reviewer stage has
`prompt.md`, `response.md`, and a segment; in `checkpoint.json` the
`reviewer` harness differs from the harness of every planner node
(harness names are the routes of providers: `codex` for `openai`,
`claude` for `anthropic`, `agy` for `gemini`), and the `reviewer` key
carries the no-thread marker, the engine's record that each turn was a
fresh session.

No `infer`.

## Not proven

That the passes catch every defect, or that seven is the right number.
Project-added passes: a run cannot complete with one open. The model
within the provider (see P3). That the harness honoured the no-thread
mode with a genuinely new native session: the harness is out of scope
(ledger rules); the check reads its record.
