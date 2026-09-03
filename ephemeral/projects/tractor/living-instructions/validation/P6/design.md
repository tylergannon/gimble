# P6: `large` completes and every chapter is marked only after `verify` routed pass

Archetype: scenario. Lap 11; answers `review-10.md`.

## Story

1. A LARGE package produced at check time by `plan` on
   `seeds/ledger-tool.md`. The check runs `plan` itself, then requires
   `tractor workflow validate-plan` to accept the package and the
   chapter ledger to hold at least two items (a v2 LARGE package is
   several chapters; `workflow/artifacts.go`); when the recommendation
   is not `large` it exits "inconclusive: package sized medium", and a
   one-chapter or empty LARGE ledger is a fail, not inconclusive. It
   records the chapter ledger as `plan` left it (C0).
2. `tractor workflow run large --project <build>` with `observer.sh`
   accepting any question; the check reads the run directory from the
   `Logs:` line.
3. Wait for `COMPLETED`.

## Evidence

- C0 and the chapter ledger at the end; `validate-plan` output.
- `timeline.jsonl`: `StageCompleted(name, next)`, `StageFailed`,
  `LoopItemSelected(node, item)`, `LoopValidated(node, item, passed)`,
  `QuestionAsked`, `PipelineCompleted`; `validation.json` of every
  `chapters` loop turn.
- Every completed `verify` stage: `prompt.md`, `response.md` (chosen
  `next`, the verifier's notes with its verdict and what it ran), and
  its segment.
- For any chapter that left the ledger during the run: the observer
  copies of the chapter ledger at every stage boundary, the removing
  stage's `response.md`, every agent segment of the run (every
  `tool_call` with its arguments), and the interview files.

## Validator

`command`: `prove/p6-verify-before-done.sh <package> <large run dir>`:
`validate-plan` accepted the package and C0 has at least two items; the
last `StageCompleted` has `next: success` and `PipelineCompleted`
follows it; every chapter named in C0 is either present at the end or
left the ledger honestly: the stage across which it left is a codergen
stage whose `response.md` names the chapter or a question asked during
the run names it (decision 44: the planner or the human may edit the
chapter ledger with a reason; silent disappearance is a fail), and
either the chapter has its own `LoopValidated` `passed: true` and
`validation.json` from before it left (engine-marked after verify, then
restructured away, which P6 allows) or it is not `done: true` in any
copy and no `tool_call` in any agent segment of the run writes `done:
true` for it (never marked, not even transiently inside a turn; the
segments record every write an agent made); the
final ledger is not empty; every chapter item at the end is `done:
true`; for each, at least one `LoopValidated` with `passed: true` names
it on `chapters`, each with a `chapters` loop stage's `validation.json`
naming it (engine-marked; a hand-marked item has neither; a chapter
reopened with a reason and verified again has two, which P6 allows);
for every such `LoopValidated`, between the `LoopItemSelected` on
`chapters` that preceded it and the event itself, the last completed
codergen stage is a `verify` stage whose `StageCompleted` has `next:
chapters` (a verify turn belonging to that marking, not a stale one; a
`StageFailed` attempt with no `response.md` is a retry and is skipped);
no agent segment writes `done: true` for any chapter (agents never
mark, transiently or otherwise); every completed `verify` stage
has `prompt.md`, `response.md`, and a segment (a codergen turn, not a
tool).

`infer` (files: every completed verify turn's `prompt.md`,
`response.md`, and segment; the chapter's validation design; for a
removed chapter, the removing stage's `response.md` and every segment
that names the chapter, or the question and answer): "Was each verifier told which chapter to
verify and the design to follow? Did it operate the software (run it,
not only read files; decision 41 and P6's statement), and do its notes
report what it did and a verdict that agrees with its route? If a
chapter was removed, is a reason recorded for it, and does the removing
turn's segment show no marking of it? Fail on any no."

## Not proven

That the software the package describes is good; the verifier's catch
rate (research R2); whether a removed chapter's promise is still
covered (P6 is about marking, not coverage; pass 1 of the plan review
and P10's probes are where coverage lives). Which providers served
`verify` and `implement`; the graph's `models.yaml` rule puts them
apart, and P3 and P5 prove independence where a promise asks for it.
What `replan` does; chapter 6 sprint 1's own check covers it. No
snapshot timing is used for marking; the marking proof is the engine's
events and records. The removal copies race by model latency (ledger
rules).
