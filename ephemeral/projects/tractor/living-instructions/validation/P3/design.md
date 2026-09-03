# P3: every promise is marked done in the validation ledger after an independent review routed pass

Archetype: universal over the promises of a run. Checked exhaustively;
no holdout. Lap 12; answers `review-11.md`.

## Story

Any `plan` run the check makes itself at check time
(`validation/lib/run-plan.sh seeds/greeter.md accept-all.rules`).

## Evidence

- `validation/ledger.md` from the package: each item's `name` and `doc`.
- `tractor workflow show plan` output for the run's parameters: per
  node id, the `checklist` path (loops) and the `provider` and `thread`
  (`none` or the thread id) the materialized graph carries (chapter 5
  sprint 1 makes `show` print them). `show` and `run` call the same
  `Build` with the same parameters in the same binary (P8), so this is
  the graph the engine walked.
- `timeline.jsonl`: `LoopItemSelected(node, item)`, `LoopValidated(node,
  item, passed)`, `StageCompleted(name, next)`, `StageFailed`.
- `stages/<seq>-<loop>/validation.json` for every validation loop turn.
- Observer copies of `validation/<id>/design.md` at every stage
  boundary (which stage wrote each design).
- For each item's reviewing stage: `prompt.md` and `response.md`, and
  the observer copy of that item's `design.md` at the `StageCompleted`
  before the reviewing stage (the design as the reviewer saw it; a
  later plan-review lap may rewrite the design after the mark, and P3
  is about the review that marked it).
- `checkpoint.json` `sessions`: the entries under each reviewing node's
  and each designing node's binding keys (the NUL-prefixed `none:<id>`
  for a no-thread node, else its thread id).

## Validator

`command`: `prove/p3-independent-review.sh`: the validation loop node's
`checklist`, as `show` prints it, is the package's
`validation/ledger.md`; each promise id in `promises.md` is bound to
exactly one item, where an item is bound to id X when its `doc` path
lies under `validation/X/` or its name is X followed by a non-word
character or the end of the name (so `P1` never claims `P10`), and
every item is bound to some id (one item per promise, decision 47;
naming is free); every item is `done: true` in the final ledger; for
every item there is exactly one `LoopValidated` with `passed: true`
naming it on the validation loop and one loop stage whose
`validation.json` names it; for every item, let R(item) be the node of
the last completed codergen stage between its last `LoopItemSelected`
and its `LoopValidated` (a `StageFailed` stage with no `response.md` is
a retried attempt and is skipped): that stage's `StageCompleted` `next`
leads to the loop (the review turn belonging to this item; different
items may have different reviewing nodes); let D(item) be the set of
nodes across whose stages that item's `design.md` was created or
changed before the item's reviewing stage began (from the observer
copies before and after each stage; a change after the review or after
the mark, by a plan-review lap, does not count, and a design that
existed before the first stage of the run makes D(item) empty): R(item) is not in
D(item), D(item) is not empty, the materialized graph gives R(item) a
provider different from every node in D(item), and `checkpoint.json`
records, under R(item)'s binding key and under each D(item) node's
key, harnesses that are the routes of those providers, R(item)'s
differing from each; every reviewing stage has `prompt.md`,
`response.md`, and a segment.

`infer` (files: for every item, its reviewing stage's `prompt.md` and
`response.md`; `promises.md`; the copy of each design as its reviewer
saw it): "For each item's reviewing
turn: was it given that item's promise statement and that item's
design, and asked whether a coder could satisfy the design while the
promise is false? Do its notes answer that for that design, and does
the verdict in the notes agree with the `next` in the front matter?
Fail for any item whose turn was not asked about it, did not answer,
or whose words disagree with its route."

## Not proven

That the designs are good. The model within a provider (the run records
the harness). Whether each reviewing turn had fresh context; P3 asks
for another provider, and a persistent session on that provider
satisfies it. Per-stage provider stamps: the engine records the harness
per binding key, not per stage; the check binds key to node through the
graph the same binary materialized. The design-authorship copies race
by model latency (ledger rules).
