# P10: two known seeds plan and execute end to end

Archetype: scenario, twice. Lap 8; answers `review-7.md`.

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
and their item names (A0); runs the `Next:` line from that stdout
verbatim with a fresh `--logs <dir>` the check names, logging the
command it ran; waits for that run's completion; then builds the
program (as the scratch repository's `README.md` says; if that fails,
with `go build ./...`, noting it) and runs every acceptance example
from the seed, comparing stdout, stderr, and exit status exactly.

## Evidence

- `check.log`: `plan`'s stdout, the handoff command as executed, the
  build commands; A0.
- Both plan run directories and both execution run directories (the
  ones the check named), in full, with their `observer/` trees,
  including the copy taken immediately before the approval answer; the
  `approve` stage's `prompt.md`, `response.md`, and segment; the
  approval question and answer files.
- Both packages; the execution run's sprint ledger(s) at the end; the
  segments of every `implement` stage.
- The scratch repositories at the end, and `probes.log` (each example's
  command, expected, actual).

## Validator

`command`: `prove/p10-end-to-end.sh`: `git show
6dec5dc8cd23e6f340e838beb9c0c2a442dffef4:<seed>` equals each seed as
checked out; for each seed: the plan run's `timeline.jsonl` ends with
`PipelineCompleted` after a `StageCompleted` with `next: success`; the
human gate was passed: the last `QuestionAsked` of the plan run E has a
`tractor ask` `tool_call` before it in the `approve` stage's segment
with its paired `tool_result` after the answer, `answers.log` records
the "Yes." rule for E, and the package copy taken immediately before
that answer is byte-identical to the final package outside
`interview/` (what the human saw is what ran; the answer file itself
is the one permitted difference); `validate-plan` accepts the package;
`recommendation.md` names `medium` or `large`; `check.log` shows the
`Next:` line and the identical command executed; the execution run
directory the check named has a `timeline.jsonl` ending with
`PipelineCompleted` after a `StageCompleted` with `next: success`; the
approved package was what ran: every sprint item name in A0 is present
in the final sprint ledger(s) and `done: true`, with exactly one
`LoopValidated` `passed: true` naming it and a loop stage's
`validation.json` naming it (the execution workflow ran the approved
items and the engine marked each; a ledger emptied or replaced during
execution fails); at least one `implement` stage's segment has a
`tool_call` that wrote a source file of the program (execution, not
planning, built it; scaffolding by `plan` is allowed); every
acceptance example in the seed matches exactly.

`infer` (files: the approval question file and its answer, the
`approve` turn's `prompt.md` and `response.md`, the package as
approved (`promises.md`, `checklist.md`, chapter or sprint docs,
`plan-review/ledger.md`); each seed and `probes.log`): "Two judgments.
First: did the approval question put this package in front of the
human, describing its promises, its slices, and its review outcomes as
they actually stand in the package files, and ask whether to approve
it? Fail if the question describes a package other than the one in the
files or does not ask for approval. Second: run the seed's acceptance
examples yourself against the built program and fail if any does not
hold."

## Not proven

Generality beyond seeds of this size. Whether the sizes match the
expectations noted in the seeds; either size satisfies the promise.
Behaviour the seed's prose names but its examples do not exercise, and
the README's build instructions; P10 promises the examples.
Verify-before-done ordering in the large run; that is P6. Whether the
packages' own checklist commands are strong: that is P3's reviewer's
job and the `verify` node's. How much of the program `plan` scaffolded
before execution; only that execution ran the approved items and wrote
program source.
