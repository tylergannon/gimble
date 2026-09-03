# P10: two known seeds plan and execute end to end

Archetype: scenario, twice. Lap 19; answers `review-18.md`.

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
(`mktemp -d`, `git init`); writes the seed into it from the pinned
commit (`git show <hash>:ephemeral/projects/tractor/living-instructions/seeds/<name>`) and keeps its own copy of that
text outside the scratch tree (the examples are read from the check's
copy, never from the scratch repository, so a planner that rewrites the
seed it was handed changes nothing the check reads); runs `plan` with `observer.sh` accepting
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
6dec5dc8cd23e6f340e838beb9c0c2a442dffef4:ephemeral/projects/tractor/living-instructions/seeds/<name>.md`
equals each seed as checked out, and that commit is an ancestor of the first commit in
which any chapter 5 sprint item is `done: true` (the seeds predate
chapter 5, as P10 says); for each seed: the plan run's `timeline.jsonl` ends with
`PipelineCompleted` after a `StageCompleted` with `next: success`; the
human gate was passed: the last `QuestionAsked` of the plan run E has a
`tractor ask` `tool_call` before it in the `approve` stage's segment
whose paired `tool_result` contains E's nonce, `answers.log` records
the "Yes." rule for E, and the copy of the whole package, research leaves included, that
the observer takes immediately before that answer (for the approval
question the observer copies everything, not the leafless package it
copies at other events) is byte-identical to the final package outside
`interview/` (what the human saw is what ran; the answer file itself is
the one permitted difference); `validate-plan` accepts the package;
`recommendation.md` names `medium` or `large`; `check.log` shows the
`Next:` line and the identical command executed; the execution run
directory named by that command's `Logs:` line is a different directory
from the plan run's, its `manifest.json` carries a different run id and
a `workflow` name of `medium` or `large` matching the handoff, its
`timeline.jsonl` opens with `PipelineStarted` naming that workflow and
ends with `PipelineCompleted` after a `StageCompleted` with `next:
success` (an execution run of the execution workflow, not the plan run
reported twice); the approved package was what ran: the execution graph's loop node
(`show medium` or `show large` for the run's parameters) names as its
`checklist` the package's own sprint ledger (or chapter ledger whose
items name the sprint ledgers), so the items the engine iterated are
the package's, not a shadow copy; for a LARGE package, the chapter
ledger at the end equals the approved copy except for `done` flags and
for edits made with a reason (decision 44): every other difference
between the two copies (a chapter's `check`, `doc`, or `command`, or a
chapter added or removed) appears across a stage whose `response.md`
states the reason, or after a `QuestionAsked` whose question names the
chapter and whose answer accepts it; a chapter's `checklist` never
changes (the approved sprint ledgers are the ones that run); the judge
below reads every such reason; a
sprint item is identified by its ledger path and name, and for a LARGE
package the events for a chapter's items are those between that
chapter's `LoopItemSelected` on `chapters` and its `LoopValidated`, so
two chapters may reuse a sprint name; every sprint item in A0
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
`plan-review/ledger.md`); the approved and final chapter ledgers for a
LARGE package with, for each difference, the `response.md` of the stage
that made it and every question and answer file of the execution run's
interview directory; each `replan` turn's `response.md` for items it
changed;
the check's own copy of each seed, taken from the pinned
commit, and `probes.log`): "Three judgments. First: did
the approval question put this package in front of the human,
describing its promises, its slices, and its review outcomes as they
actually stand in the package files, and ask whether to approve it?
Fail if the question describes a package other than the one in the
files or does not ask for approval. Second: for every item replan
changed, renamed, or removed, does its response give a reason grounded
in work already done, and for every difference between the approved
and final chapter ledgers, does the stage's response give a reason
grounded in the work, or does a question describe exactly that edit
with an answer accepting it? Naming the chapter is not a reason. Third: run the seed's acceptance examples
yourself against the built program and fail if any does not hold."

## Not proven

That the graph `show` prints afterwards is the graph the run walked:
the run records neither its graph nor its loop's checklist path, and
this design relies, as P2 does, on P8's proof that `show` prints what
`Build` materialized in the same binary for the same parameters. That
no workflow tool node forged run-directory records: every tool
command receives the run directory's path and could write loop stage
files or append events; the observer tree is at a random path the
agents cannot find, but the run directory is theirs to write. The graph
is content the diff review reads, and a forging tool node is a defect
of that review, not of this check (the same concession as P6).
Generality beyond seeds of this size. Whether the sizes match the
expectations noted in the seeds; either size satisfies the promise.
Behaviour the seed's prose names but its examples do not exercise, and
the README's build instructions; P10 promises the examples.
Verify-before-done ordering in the large run; that is P6. Whether the
packages' own checklist commands are strong: that is P3's reviewer's
job and the `verify` node's. How much of the program `plan` scaffolded
before execution; P10 asks that the approved package ran and the
program works. The replan copies race by model latency (ledger rules).
