# P3: every promise is marked done in the validation ledger after an independent review routed pass

Archetype: universal over the promises of a run. Checked exhaustively;
no holdout. Lap 4; answers `review-3.md`.

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
  `StageCompleted(name, next)`.
- `stages/<seq>-<loop>/validation.json` for every validation loop turn.
- `stages/<seq>-review/prompt.md` and `response.md` for every review
  turn; `stages/<seq>-design/prompt.md` likewise.
- `checkpoint.json` `sessions`: harness for `review` and for `design`.

## Validator

`command`: `prove/p3-independent-review.sh`: the validation loop node's
`checklist`, as `show` prints it for the run's parameters, is the
package's `validation/ledger.md`; the set of item names equals the set
of ids in `promises.md`; every item is `done: true`; for every item
there is exactly one `LoopValidated` with `passed: true` naming it on
the validation loop and one loop stage whose `validation.json` names it
(the engine marked it; a hand-marked item has neither); for each such
loop stage, the nearest preceding `StageCompleted` is a `review` stage
whose `next` is the loop node and whose `response.md` front matter says
the same; every review stage has `prompt.md`, `response.md`, and a
segment (a codergen turn); in `checkpoint.json` the `review` harness
differs from the `design` harness.

`infer` (files: every review stage's `prompt.md` and `response.md`,
`promises.md`, the designs): "For each review turn: was the reviewer
given the promise's statement and the design under review, and asked
whether a coder could satisfy the design while the promise is false?
Do its notes answer that, and does the verdict in the notes agree with
the `next` in the front matter? Fail for any review turn that was not
asked, did not answer, or whose words disagree with its route."

## Not proven

That the designs are good. The model within a provider (the run records
the harness; `models.yaml` is static content). Whether each review turn
had fresh context; P3 asks for another provider, and a persistent
review session on that provider satisfies it.
