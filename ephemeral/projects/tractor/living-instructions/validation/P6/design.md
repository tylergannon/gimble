# P6: `large` completes and every chapter is marked only after `verify` routed pass

Archetype: scenario. Lap 4; answers `review-3.md`.

## Story

1. A LARGE package produced at check time by `plan` on
   `seeds/ledger-tool.md` (the script runs `plan` itself; when the
   recommendation is not `large`, it exits "inconclusive: package sized
   medium" and the item stays open).
2. `tractor workflow run large --project <build> --logs <fresh dir>` with
   `observer.sh` accepting any question. `models.yaml` puts `implement`
   on one provider and `verify` on another.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `StageStarted`, `StageCompleted(name, next)`,
  `LoopItemSelected(node, item)`, `LoopValidated(node, item, passed)`,
  `PipelineCompleted`; `validation.json` of every `chapters` loop turn.
- `stages/<seq>-verify/`: `prompt.md`, `response.md` (chosen `next`,
  the verifier's notes with its verdict and what it ran).
- `observer/` copies of the chapter ledger and every chapter's
  `sprints.md` at every stage boundary.
- `checkpoint.json` `sessions`: harness for `verify` and `implement`.

## Validator

`command`: `prove/p6-verify-before-done.sh <large run dir>`: the last
`StageCompleted` has `next: success` and `PipelineCompleted` follows it;
every chapter item is `done: true`; for each, exactly one
`LoopValidated` with `passed: true` names it on `chapters` and one
`chapters` loop stage's `validation.json` names it (engine-marked); for
each such item, between its `LoopItemSelected` on `chapters` and its
`LoopValidated` there is a `verify` stage whose `StageCompleted` has
`next: chapters`, whose `response.md` front matter agrees, and which is
the last non-loop stage before that `LoopValidated` (a verify turn
belonging to this chapter, not a stale one); every `verify` stage
directory has `prompt.md`, `response.md`, and a segment (a codergen
turn); `checkpoint.json` records different harnesses for `verify` and
`implement`; no chapter item's `done` flag is `true` in any copy taken
at an `implement`, `replan`, or `verify` `StageCompleted` before that
item's `LoopValidated` (agents never marked a chapter; other edits to
the chapter ledger are allowed); each `replan` stage changes only the
current chapter's `sprints.md` between the copies before and after it.

`infer` (files: every verify turn's `prompt.md` and `response.md`, the
chapter's validation design, the copies of the sprint ledger before and
after each `replan`): "Was each verifier told which chapter to verify
and the design to follow, and do its notes report what it did and a
verdict that agrees with its route? Did any replan change an item that
was `done: true` before it ran? Fail on any no or any such change."

## Not proven

That the software the package describes is good. That the verifier
operated the software rather than reading artifacts: its prompt requires
it and P10's probes run the software independently, but this promise
is about order and marking only. The verifier's catch rate (research
R2). The copies race by model latency (ledger rules); the marking
checks do not depend on them.
