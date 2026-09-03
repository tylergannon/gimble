# P3: every promise is marked done in the validation ledger after an independent review routed pass

Archetype: universal over the promises of a run. Checked exhaustively;
no holdout.

## Story

Any `plan` run that reaches `COMPLETED` (P1's or P10's run serves).

## Evidence

- `validation/ledger.md` from the package: one item per promise in
  `promises.md`, all `done: true`.
- `timeline.jsonl`: for each item, a `review` stage that routed to the
  loop node, and the `LoopValidated` event that marked it.
- `events/NNNNNN-review.jsonl` for each review stage: the model and
  provider the harness actually used.
- `events/NNNNNN-design.jsonl`: same for the design stages.

## Validator

`command`: `prove/p3-independent-review.sh`: promise ids in
`promises.md` equal item names in the validation ledger; every item is
`done: true`; for every item the marking `LoopValidated` is immediately
preceded by a `review` stage whose chosen route is the loop node; every
`review` event log reports a provider different from every `design`
event log's provider; every `review` stage's session id is new (no
session reuse across reviews).

No `infer`.

## Not proven

That the designs are good. Only that none was ratified by its author's
provider, and that the engine, not an agent, marked each.
