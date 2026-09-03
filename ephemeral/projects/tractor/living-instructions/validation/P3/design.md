# P3: every promise is marked done in the validation ledger after an independent review routed pass

Archetype: universal over the promises of a run. Checked exhaustively;
no holdout. Lap 7; answers `review-6.md`.

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
- `timeline.jsonl`: `LoopValidated(node, item, passed)`,
  `StageCompleted(name, next)`, `StageFailed`.
- `stages/<seq>-<loop>/validation.json` for every validation loop turn.
- `observer/` copies of `validation/<item>/design.md` at every stage
  boundary (which stage wrote each design).
- For the reviewing node R: `prompt.md` and `response.md` of every
  completed stage.
- `checkpoint.json` `sessions`: the entries under R's and D's binding
  keys (the NUL-prefixed `none:<id>` for a no-thread node, else its
  thread id).

## Validator

`command`: `prove/p3-independent-review.sh`: the validation loop node's
`checklist`, as `show` prints it, is the package's
`validation/ledger.md`; every id in `promises.md` is the prefix of
exactly one item name and every item name begins with an id (each
promise is represented once; naming beyond the id is free); every item
is `done: true` in the final ledger; for every item there is exactly
one `LoopValidated` with `passed: true` naming it on the validation
loop and one loop stage whose `validation.json` names it; for each such
loop stage, let R be the node of the last completed codergen stage
before it (tool stages may intervene; a `StageFailed` stage with no
`response.md` is a retried attempt and is skipped): R is the same node
for every item and R's `StageCompleted` `next` leads to the loop; let D
be the set of nodes across whose stages any `validation/<item>/design.md`
changed (from the observer copies before and after each stage): R is
not in D (the reviewer authored no design), and D is not empty; every
completed R stage has `prompt.md`, `response.md`, and a segment; the
materialized graph gives R a provider different from every node in D,
and `checkpoint.json` records, under R's binding key and under each D
node's key, harnesses that are the routes of those providers, R's
differing from every D node's.

`infer` (files: every completed R stage's `prompt.md` and
`response.md`, `promises.md`, the designs): "For each R turn: was it
given the promise's statement and the design under review, and asked
whether a coder could satisfy the design while the promise is false?
Do its notes answer that, and does the verdict in the notes agree with
the `next` in the front matter? Fail for any turn that was not asked,
did not answer, or whose words disagree with its route."

## Not proven

That the designs are good. The model within a provider (the run records
the harness). Whether each R turn had fresh context; P3 asks for
another provider, and a persistent session on that provider satisfies
it. Per-stage provider stamps: the engine records the harness per
binding key, not per stage; the check binds key to node through the
graph the same binary materialized. The design-authorship copies race
by model latency (ledger rules).
