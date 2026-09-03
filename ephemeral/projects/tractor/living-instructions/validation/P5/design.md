# P5: all seven review passes are engine-marked after a fresh reviewer on another provider routed pass

Archetype: universal over the passes of a run. Exhaustive; no holdout.
Lap 17 (re-opened after the plan-review loop's routing was corrected;
marked on lap 4); answers `review-3.md`.

## Story

Any `plan` run the check makes itself at check time (the same run P3
uses, through `validation/lib/run-plan.sh`).

## Evidence

- `plan-review/ledger.md` from the package, and `show`'s `checklist`
  path for the plan-review loop.
- `timeline.jsonl`: `LoopItemSelected`, `LoopValidated`, `StageStarted`,
  `StageCompleted(next)` for `reviewer` and the owning nodes;
  `validation.json` of every plan-review loop turn.
- `stages/<seq>-reviewer/prompt.md` and `response.md` for every
  reviewer turn.
- `checkpoint.json` `sessions`: harness and key for `reviewer`, harness
  for every planner node (`intake`, `brief`, `decompose`, `design`,
  `assemble`).

## Validator

`command`: `prove/p5-seven-passes.sh`: the plan-review loop's
`checklist` is the package's `plan-review/ledger.md`; its items include
the seven built-in passes by name, all `done: true`; for each pass
there is exactly one `LoopValidated` with `passed: true` naming it on
the plan-review loop and one loop stage whose `validation.json` names
it; for each such loop stage the nearest preceding `StageCompleted` is
a `reviewer` stage whose `next` is the loop node and whose `response.md`
front matter agrees; for
every `reviewer` stage whose `next` is `brief`, `decompose`, or
`design`, the next `StageStarted` is that node and the stage after that
node is a `reviewer` stage for the same pass, with no `LoopValidated`
for the pass in between (the owning node returns to the reviewer, not
to the loop; a commandless item would otherwise pass on return); every reviewer
stage has `prompt.md`, `response.md`, and a segment; in
`checkpoint.json` the `reviewer` harness differs from the harness of
every planner node (harness names are the routes of providers), and the
`reviewer` key carries the no-thread marker, the engine's record that
each turn was a fresh session.

`infer` (files: every reviewer turn's `prompt.md` and `response.md`,
the seven pass files): "For each reviewer turn: was it given one pass
question and the package, and nothing that answers the question for it?
Do its notes answer that question, and does the verdict in the notes
agree with the `next` in the front matter? Fail for any turn that was
not asked, did not answer, or whose words disagree with its route."

## Not proven

That the passes catch every defect, or that seven is the right number.
The order the passes ran in; P5 does not promise one. Project-added
passes: a run cannot complete with one open. The model within the
provider (see P3). That the harness honoured the no-thread mode with a
genuinely new native session: the harness is out of scope (ledger
rules); the check reads its record.
