# P10: two known seeds plan and execute end to end

Archetype: scenario, twice. Lap 3; answers `review-2.md`.

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
prompts, and answering the approval question "Yes."; runs the printed
`Next:` handoff verbatim with a fresh `--logs`; waits for `COMPLETED`;
then builds the program as the scratch repository's `README.md` says
and runs every acceptance example from the seed, comparing stdout,
stderr, and exit status exactly.

## Evidence

- Both plan run directories and both execution run directories, in
  full, with their `observer/` trees.
- Both packages.
- The scratch repositories at the end, and the probe transcript
  (`probes.log`: each example's command, expected, actual).

## Validator

`command`: `prove/p10-end-to-end.sh`: `git show
6dec5dc8cd23e6f340e838beb9c0c2a442dffef4:<seed>` equals each seed as
checked out; for each seed: `plan` completed; the package was approved
through the human gate: the last `QuestionAsked` of the plan run was
emitted during the `approve` stage, `answers.log` records the "Yes."
rule for it, and `approve`'s `StageCompleted` has `next: success`;
`validate-plan` accepts the package; `recommendation.md` names `medium`
or `large` and the handoff it prints ran to `COMPLETED`; every
acceptance example in the seed matches exactly.

`infer` (files: each seed, each scratch repository's `README.md`,
`probes.log`): "Build and run the program yourself from the README.
Does it do what the seed asked, beyond the examples the script already
ran? Fail if a named feature is missing, the README's instructions do
not work, or the program does not run."

## Not proven

Generality beyond seeds of this size. Whether the sizes match the
expectations noted in the seeds; either size satisfies the promise.
Verify-before-done ordering in the large run; that is P6, which makes
its own run. Whether the packages' own checklist commands are strong:
that is P3's reviewer's job and the `verify` node's; the acceptance
examples here are the independent probe.
