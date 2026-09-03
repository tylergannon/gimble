# P3: every promise is marked done in the validation ledger after an independent review routed pass

Archetype: universal over the promises of a run. Checked exhaustively;
no holdout. Lap 5; answers `review-4.md`.

## Story

Any `plan` run the check makes itself at check time
(`validation/lib/run-plan.sh seeds/greeter.md accept-all.rules`).

## Evidence

- `validation/ledger.md` from the package; item names are promise ids.
- `tractor workflow show plan` output for the run's parameters: per
  node, the `checklist` path (loops), and the `provider` and `thread`
  (`none` or the thread id) the materialized graph carries (chapter 5
  sprint 1 makes `show` print them).
- `timeline.jsonl`: `LoopValidated(node, item, passed)`,
  `StageCompleted(name, next)`.
- `stages/<seq>-<loop>/validation.json` for every validation loop turn.
- `stages/<seq>-review/prompt.md` and `response.md` for every review
  turn.
- `checkpoint.json` `sessions`: the entry whose key is `review`'s
  binding key as `show` reports it (the NUL-prefixed `none:review` for
  a no-thread node, else its thread id), and likewise for `design`.

## Validator

`command`: `prove/p3-independent-review.sh`: the validation loop node's
`checklist`, as `show` prints it, is the package's
`validation/ledger.md`; the set of item names equals the set of ids in
`promises.md`; every item is `done: true`; for every item there is
exactly one `LoopValidated` with `passed: true` naming it on the
validation loop and one loop stage whose `validation.json` names it;
for each such loop stage, the last codergen stage before it (tool
stages may intervene) is a `review` stage whose `next` leads to the
loop node and whose `response.md` front matter says the same; every
review stage has `prompt.md`, `response.md`, and a segment; the
materialized graph gives `review` and `design` different providers, and
`checkpoint.json` records, under each node's own binding key, harnesses
that are the routes of those providers and differ from each other (the
provider the graph names is what the engine routed, and the engine's
record of that route is bound to the node by the key it derives from
the node's thread mode).

`infer` (files: every review stage's `prompt.md` and `response.md`,
`promises.md`, the designs): "For each review turn: was the reviewer
given the promise's statement and the design under review, and asked
whether a coder could satisfy the design while the promise is false?
Do its notes answer that, and does the verdict in the notes agree with
the `next` in the front matter? Fail for any review turn that was not
asked, did not answer, or whose words disagree with its route."

## Not proven

That the designs are good. The model within a provider (the run records
the harness). Whether each review turn had fresh context; P3 asks for
another provider, and a persistent review session on that provider
satisfies it.
