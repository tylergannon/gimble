# P6: `large` completes and every chapter is marked only after `verify` routed pass

Archetype: scenario.

## Story

1. A small LARGE package (P10's second seed produces one): two
   chapters, two or three sprints each, sprints that a coder finishes in
   one turn (CLI behaviour, checked by transcript and exit code).
2. Run `tractor workflow run large --project <build>` with the coder on
   codex, `verify` on claude, `answerer.sh` accepting any question.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`.
- `stages/NNNNNN-verify/` for each chapter: the verifier's own captured
  transcripts and exit codes under `evidence/`, and the route it chose.
- `stages/NNNNNN-replan/` after every implement lap: the ledger diff it
  made (may be empty).
- The chapter ledger at the end.

## Validator

`command`: `prove/p6-verify-before-done.sh`: `PipelineCompleted` with
next `success`; for every `LoopValidated` on the chapters loop that
passed, the nearest preceding stage of a codergen node is `verify` and
its route was the chapters loop; every `verify` stage's event log shows
at least one command executed against the built software (not only file
reads); every `replan` stage's diff touches only the current chapter's
sprint ledger and only items without `done: true`; the chapter ledger is
never modified by a `replan` or `implement` stage.

`infer` (files: each verify stage's evidence directory and the chapter's
validation design): "Did the verifier follow the story in the design and
capture evidence for each step? Fail if a step has no captured output."

## Not proven

That the software the package describes is good, or the verifier's
catch rate (research R2: no measured precedent; instrument later).
