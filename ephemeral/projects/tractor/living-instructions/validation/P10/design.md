# P10: two known seeds plan and execute end to end

Archetype: scenario, twice. Lap 11; answers `review-10.md`.

## Story

The seeds exist now, under `seeds/`, fixed at commit
`6dec5dc8cd23e6f340e838beb9c0c2a442dffef4` (the proof script fails if
either seed differs from that commit's version):

- `seeds/greeter.md`: a CLI that greets by name, with a flag for
  shouting and a file of names, with four acceptance examples and their
  exact output. Expected MEDIUM; the expectation is a note, not a check.
- `seeds/ledger-tool.md`: a CLI that keeps a markdown ledger with add,
  done, list, validate, and export, with six acceptance examples and
  their exact output. Expected LARGE; same.

For each seed the check: makes an empty scratch repository
(`mktemp -d`, `git init`); runs `plan` with `observer.sh` accepting
everything, answering "Proceed with your recommendation." to open
prompts, and answering the approval question "Yes."; captures `plan`'s
stdout to `check.log`; records the approved package's sprint ledger(s)
in full (A0); runs the `Next:` line from that stdout verbatim, with no
flag added (the canonical handoff `validate-plan` requires), with
`observer.sh` attached to that run once its `Logs:` line names the run
directory; waits for that run's completion; then builds the program
(as the scratch repository's `README.md` says; if that fails, with `go
build ./...`, noting it) and runs every acceptance example from the
seed, comparing stdout, stderr, and exit status exactly.

## Evidence

- `check.log`: `plan`'s stdout, the handoff command as executed and the
  `Logs:` line it printed, the build commands; A0.
- Both plan run directories and both execution run directories (from
  the `Logs:` lines), in full, with their observer trees, including
  the copy taken immediately before the approval answer and the sprint
  ledger copies at every execution stage boundary; the `approve`
  stage's `prompt.md`, `response.md`, and segment; the approval
  question and answer files.
- Both packages; the execution run's sprint ledger(s) at the end;
  every `replan` stage's `response.md`.
- The scratch repositories at the end, and `probes.log` (each example's
  command, expected, actual).

## Validator

`command`: `prove/p10-end-to-end.sh`: `git show
6dec5dc8cd23e6f340e838beb9c0c2a442dffef4:<seed>` equals each seed as
checked out; for each seed: the plan run's `timeline.jsonl` ends with
`PipelineCompleted` after a `StageCompleted` with `next: success`; the
human gate was passed: the last `QuestionAsked` of the plan run E has a
`tractor ask` `tool_call` before it in the `approve` stage's segment
whose paired `tool_result` contains E's nonce, `answers.log` records
the "Yes." rule for E, and the package copy taken immediately before
that answer is byte-identical to the final package outside
`interview/` (what the human saw is what ran; the answer file itself
is the one permitted difference); `validate-plan` accepts the package;
`recommendation.md` names `medium` or `large`; `check.log` shows the
`Next:` line and the identical command executed; the execution run
directory named by that command's `Logs:` line is a different directory
from the plan run's, its `manifest.json` carries a different run id and
a `workflow` name of `medium` or `large` matching the handoff, its
`timeline.jsonl` opens with `PipelineStarted` naming that workflow and
ends with `PipelineCompleted` after a `StageCompleted` with `next:
success` (an execution run of the execution workflow, not the plan run
reported twice); the approved package was what ran: every sprint item in A0
is either present in the final sprint ledger(s) with its `name`,
`check`, `command`, and `infer` unchanged and `done: true`, with
exactly one `LoopValidated` `passed: true` naming it in the execution
run's own timeline and a loop stage's `validation.json` in that run
naming it, or was changed, renamed, or removed across
a `replan` stage while still open (not `done: true` in the copy before
that stage) with that stage's `response.md` naming it (decision 44:
replan edits open items with a reason; an item whose fields change
across an `implement` or any other stage, or after being marked,
fails); every item in the final ledger(s) is `done: true` with the same
event and record; the final ledger(s) are not empty; every acceptance
example in the seed matches exactly.

`infer` (files: the approval question file and its answer, the
`approve` turn's `prompt.md` and `response.md`, the package as
approved (`promises.md`, `checklist.md`, chapter or sprint docs,
`plan-review/ledger.md`); each `replan` turn's `response.md` for items
it changed; each seed and `probes.log`): "Three judgments. First: did
the approval question put this package in front of the human,
describing its promises, its slices, and its review outcomes as they
actually stand in the package files, and ask whether to approve it?
Fail if the question describes a package other than the one in the
files or does not ask for approval. Second: for every item replan
changed, renamed, or removed, does its response give a reason grounded
in work already done? Third: run the seed's acceptance examples
yourself against the built program and fail if any does not hold."

## Not proven

Generality beyond seeds of this size. Whether the sizes match the
expectations noted in the seeds; either size satisfies the promise.
Behaviour the seed's prose names but its examples do not exercise, and
the README's build instructions; P10 promises the examples.
Verify-before-done ordering in the large run; that is P6. Whether the
packages' own checklist commands are strong: that is P3's reviewer's
job and the `verify` node's. How much of the program `plan` scaffolded
before execution; P10 asks that the approved package ran and the
program works. The replan copies race by model latency (ledger rules).
