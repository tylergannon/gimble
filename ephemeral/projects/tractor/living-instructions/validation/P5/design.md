# P5: all seven review passes are engine-marked after a fresh reviewer on another provider routed pass

Archetype: universal over the passes of a run. Exhaustive; no holdout.

## Story

Any `plan` run that reaches `COMPLETED`.

## Evidence

- `plan-review/ledger.md`: seven items in the documented order, all
  `done: true`.
- `timeline.jsonl`: seven or more `reviewer` stages; for each pass, the
  marking `LoopValidated` preceded by a `reviewer` stage that routed to
  the loop node.
- `events/NNNNNN-reviewer.jsonl` per stage: provider, model, session id.
- `plan-review/<pass>/` findings files for any pass that failed at least
  once, and the owning node's stage that followed.

## Validator

`command`: `prove/p5-seven-passes.sh`: the ledger names exactly the
seven built-in passes in order plus any project-added passes; all done;
every reviewer event log's provider differs from the planner nodes'
provider; every reviewer stage has a session id not seen in any earlier
stage; for any reviewer stage that routed elsewhere than the loop, the
next stage is the node it named and the pass is re-selected afterwards.

No `infer`.

## Not proven

That the passes catch every defect, or that seven is the right number.
