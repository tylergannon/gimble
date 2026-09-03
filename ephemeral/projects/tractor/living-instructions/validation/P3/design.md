# P3: every promise is marked done in the validation ledger after an independent review routed pass

Archetype: universal over the promises of a run. Checked exhaustively;
no holdout. Lap 3; answers `review-2.md`.

## Story

Any `plan` run the check makes itself at check time
(`validation/lib/run-plan.sh seeds/greeter.md accept-all.rules`).

## Evidence

- `validation/ledger.md` from the package; item names are promise ids
  (the ledger is generated from `templates/ledger.md` with `name: <id>`
  per row of `promises.md`).
- `tractor workflow show plan` output for the run's parameters: the
  validation loop's `checklist` path.
- `timeline.jsonl`: `LoopValidated(node, item, passed)`,
  `StageStarted`, `StageCompleted(name, next)`.
- `observer/` snapshots of `validation/ledger.md` at every stage
  boundary.
- `stages/<seq>-review/response.md` for every review turn: the chosen
  `next` in its front matter and the reviewer's notes.
- `checkpoint.json` `sessions`: harness for `review` and for `design`.

## Validator

`command`: `prove/p3-independent-review.sh`: the validation loop node's
`checklist`, as `show` prints it for the run's parameters, is the
package's `validation/ledger.md` (the inspected ledger is the one the
loop ran); the set of item names equals the set of ids in
`promises.md`; every item is `done: true`; for every item, exactly one
loop stage flips it: not done in the snapshot at that stage's
`StageStarted`, a `LoopValidated` for the item with `passed: true`
during the stage, done at its `StageCompleted`, and no change across
any other stage; for each such loop stage, the nearest preceding
`StageCompleted` is a `review` stage whose `next` is the loop node, and
that review's `response.md` front matter says the same `next` and its
body contains the verdict line `ROUTE: pass` the review prompt asks
for; every review stage directory has `prompt.md` and `response.md` and
a segment in `events/index.jsonl` (a codergen turn, not a tool); in
`checkpoint.json`, the `review` session's harness differs from the
`design` session's harness.

No `infer`.

## Not proven

That the designs are good. The model within a provider (the run records
the harness; `models.yaml` is static content). Whether each review turn
had fresh context; P3 asks for another provider, and a persistent
review session on that provider satisfies it.
