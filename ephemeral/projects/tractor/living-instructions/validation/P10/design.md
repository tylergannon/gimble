# P10: two known seeds plan and execute end to end

Archetype: scenario, twice. Lap 5; answers `review-4.md`.

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
stdout to `check.log`; runs the `Next:` line from that stdout verbatim
with a fresh `--logs`, logging the command it ran; waits for
`COMPLETED`; then builds the program as the scratch repository's
`README.md` says and runs every acceptance example from the seed,
comparing stdout, stderr, and exit status exactly.

## Evidence

- `check.log`: `plan`'s stdout, the handoff command as executed, the
  build commands.
- Both plan run directories and both execution run directories, in
  full, with their `observer/` trees, including the copy taken
  immediately before the approval answer; the `approve` stage's
  `prompt.md`, `response.md`, and segment; the approval question and
  answer files.
- Both packages.
- The scratch repositories at the end, and `probes.log` (each example's
  command, expected, actual).

## Validator

`command`: `prove/p10-end-to-end.sh`: `git show
6dec5dc8cd23e6f340e838beb9c0c2a442dffef4:<seed>` equals each seed as
checked out; for each seed: `plan` completed; the human gate was
passed: the last `QuestionAsked` of the plan run E has a `tractor ask`
`tool_call` before it in the `approve` stage's segment with its paired
`tool_result` after the answer, `answers.log` records the "Yes." rule
for E, `approve`'s `StageCompleted` has `next: success`, and the package
copy taken immediately before that answer is byte-identical to the
final package (what the human saw is what ran); `validate-plan` accepts
the package; `recommendation.md` names `medium` or `large`; `check.log`
shows the `Next:` line and the identical command executed, and that run
reached `COMPLETED`; every acceptance example in the seed matches
exactly.

`infer` (files: the approval question file and its answer, the
`approve` turn's `prompt.md` and `response.md`; each seed, each scratch
repository's `README.md`, `probes.log`): "Two judgments. First: did the
approval question put the package (its promises, chapters or sprints,
and review outcomes) in front of the human and ask whether to approve
it, so that the Yes was an approval of this package? Second: build and
run the program yourself from the README and run the seed's acceptance
examples; fail if the README's instructions do not work or an example
does not hold when you run it."

## Not proven

Generality beyond seeds of this size. Whether the sizes match the
expectations noted in the seeds; either size satisfies the promise.
Behaviour the seed's prose names but its examples do not exercise
(greeter's `--shout` with `--names`); P10 promises the examples.
Verify-before-done ordering in the large run; that is P6. Whether the
packages' own checklist commands are strong: that is P3's reviewer's
job and the `verify` node's.
