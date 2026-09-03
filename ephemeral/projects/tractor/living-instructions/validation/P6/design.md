# P6: `large` completes and every chapter is marked only after `verify` routed pass

Archetype: scenario. Lap 2; answers `review-1.md`.

## Story

1. A LARGE package produced at check time by `plan` on
   `seeds/ledger-tool.md` (P10's second leg makes it; this script takes
   that run's package, or makes its own when run alone).
2. `tractor workflow run large --project <build> --logs <fresh dir>` with
   the coder on codex, `verify` on claude, `answerer.sh` accepting any
   question.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `StageCompleted(name, next)`, `LoopItemSelected`,
  `LoopValidated(node, item, passed)`, `PipelineCompleted`.
- Segments of every `plan`, `implement`, `replan`, and `verify` stage:
  `tool_call` arguments.
- `stages/<seq>-verify/`: the verifier's `response.md` and whatever it
  captured under the run directory.
- The chapter ledger (`checklist.md`) and each chapter's `sprints.md` at
  the end.

## Validator

`command`: `prove/p6-verify-before-done.sh <large run dir>`: the last
`StageCompleted` has `next: success` and `PipelineCompleted` follows it;
every item in the chapter ledger is `done: true` and has a
`LoopValidated` on the `chapters` loop with `passed: true`; for each such
event, the nearest preceding `StageCompleted` of a node other than the
two loops is a `verify` stage with `next: chapters`; no `plan`,
`implement`, `replan`, or `verify` segment has a `tool_call` whose
arguments name the chapter ledger path; every `replan` segment's writing
`tool_call`s name only the current chapter's `sprints.md` (the chapter
from the enclosing `LoopItemSelected`).

`infer` (files: each `replan` segment, the final `sprints.md` of each
chapter): "Each replan segment's tool calls carry the edits it made. Did
any replan change an item that was already `done: true` when it ran?
Fail if so."

## Not proven

That the software the package describes is good. That the verifier
operated the software rather than reading artifacts: its prompt requires
it and P10's judge runs the software independently, but this promise is
about order and marking only. The verifier's catch rate (research R2).
