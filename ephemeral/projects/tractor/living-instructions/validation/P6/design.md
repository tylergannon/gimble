# P6: `large` completes and every chapter is marked only after `verify` routed pass

Archetype: scenario. Lap 6; answers `review-5.md`.

## Story

1. A LARGE package produced at check time by `plan` on
   `seeds/ledger-tool.md`. The check runs `plan` itself, then requires
   `tractor workflow validate-plan` to accept the package and the
   chapter ledger to hold at least two items (a v2 LARGE package is
   several chapters; `workflow/artifacts.go`); when the recommendation
   is not `large` it exits "inconclusive: package sized medium", and a
   one-chapter or empty LARGE ledger is a fail, not inconclusive. It
   records the chapter ledger as `plan` left it (C0).
2. `tractor workflow run large --project <build> --logs <fresh dir>` with
   `observer.sh` accepting any question.
3. Wait for `COMPLETED`.

## Evidence

- C0 and the chapter ledger at the end; `validate-plan` output.
- `timeline.jsonl`: `StageCompleted(name, next)`, `StageFailed`,
  `LoopItemSelected(node, item)`, `LoopValidated(node, item, passed)`,
  `PipelineCompleted`; `validation.json` of every `chapters` loop turn.
- Every completed `verify` stage: `prompt.md`, `response.md` (chosen
  `next`, the verifier's notes with its verdict and what it ran), and
  its segment.

## Validator

`command`: `prove/p6-verify-before-done.sh <package> <large run dir>`:
`validate-plan` accepted the package and C0 has at least two items; the
last `StageCompleted` has `next: success` and `PipelineCompleted`
follows it; the set of chapter item names at the end equals the set in
C0 (no chapter vanished); every chapter item is `done: true`; for each,
exactly one `LoopValidated` with `passed: true` names it on `chapters`
and one `chapters` loop stage's `validation.json` names it
(engine-marked; a hand-marked item has neither); for each such item,
between its `LoopItemSelected` on `chapters` and its `LoopValidated`,
the last completed codergen stage is a `verify` stage whose
`StageCompleted` has `next: chapters` (a verify turn belonging to this
chapter, not a stale one; a `StageFailed` attempt with no `response.md`
is a retry and is skipped); every completed `verify` stage has
`prompt.md`, `response.md`, and a segment (a codergen turn, not a tool).

`infer` (files: every completed verify turn's `prompt.md`,
`response.md`, and segment; the chapter's validation design): "Was each
verifier told which chapter to verify and the design to follow? Did it
operate the software (run it, not only read files), and do its notes
report what it did and a verdict that agrees with its route? Fail on
any no."

## Not proven

That the software the package describes is good; the verifier's catch
rate (research R2). Which providers served `verify` and `implement`;
the graph's `models.yaml` rule puts them apart, and P3 and P5 prove
independence where a promise asks for it. What `replan` does; chapter
6 sprint 1's own check covers it. No snapshot timing is used here; the
marking proof is the engine's events and records.
