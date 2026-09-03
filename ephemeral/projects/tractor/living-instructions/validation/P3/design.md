# P3: every promise is marked done in the validation ledger after an independent review routed pass

Archetype: universal over the promises of a run. Checked exhaustively;
no holdout. Lap 8; answers `review-7.md`.

## Story

Any `plan` run the check makes itself at check time
(`validation/lib/run-plan.sh seeds/greeter.md accept-all.rules`).

## Evidence

- `validation/ledger.md` from the package.
- `tractor workflow show plan` output for the run's parameters: per
  node id, the `checklist` path (loops) and the `provider` and `thread`
  (`none` or the thread id) the materialized graph carries (chapter 5
  sprint 1 makes `show` print them). `show` and `run` call the same
  `Build` with the same parameters in the same binary (P8), so this is
  the graph the engine walked.
- `timeline.jsonl`: `LoopItemSelected(node, item)`, `LoopValidated(node,
  item, passed)`, `StageCompleted(name, next)`, `StageFailed`.
- `stages/<seq>-<loop>/validation.json` for every validation loop turn.
- `observer/` copies of `validation/<item>/design.md` at every stage
  boundary (which stage wrote each design).
- For the reviewing node R: `prompt.md` and `response.md` of every
  completed stage.
- `checkpoint.json` `sessions`: the entries under R's and each D node's
  binding keys (the NUL-prefixed `none:<id>` for a no-thread node, else
  its thread id).

## Validator

`command`: `prove/p3-independent-review.sh`: the validation loop node's
`checklist`, as `show` prints it, is the package's
`validation/ledger.md`; every id in `promises.md` is the prefix of
exactly one item name and every item name begins with an id (each
promise is represented once; naming beyond the id is free); every item
is `done: true` in the final ledger; for every item there is exactly
one `LoopValidated` with `passed: true` naming it on the validation
loop and one loop stage whose `validation.json` names it; for every
item, between its last `LoopItemSelected` and its `LoopValidated`, there
is a completed codergen stage of one node R (the same node for every
item) whose `StageCompleted` `next` leads to the loop and which is the
last completed codergen stage before that `LoopValidated` (a review
turn belonging to this item, not a stale one; a `StageFailed` stage
with no `response.md` is a retried attempt and is skipped); let D be
the set of nodes across whose stages any `validation/<item>/design.md`
changed (from the observer copies before and after each stage): R is
not in D (the reviewer authored no design), and D is not empty; every
completed R stage has `prompt.md`, `response.md`, and a segment; the
materialized graph gives R a provider different from every node in D,
and `checkpoint.json` records, under R's binding key and under each D
node's key, harnesses that are the routes of those providers, R's
differing from every D node's.

`infer` (files: for every item, the R stage identified above: its
`prompt.md` and `response.md`; `promises.md`; the designs): "For each
item's R turn: was it given that item's promise statement and that
item's design, and asked whether a coder could satisfy the design
while the promise is false? Do its notes answer that for that design,
and does the verdict in the notes agree with the `next` in the front
matter? Fail for any item whose turn was not asked about it, did not
answer, or whose words disagree with its route."

## Not proven

That the designs are good. The model within a provider (the run records
the harness). Whether each R turn had fresh context; P3 asks for
another provider, and a persistent session on that provider satisfies
it. Per-stage provider stamps: the engine records the harness per
binding key, not per stage; the check binds key to node through the
graph the same binary materialized. The design-authorship copies race
by model latency (ledger rules).
