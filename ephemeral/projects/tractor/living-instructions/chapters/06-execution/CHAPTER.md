# Chapter 6: execution and the live proof

Status: planned

`medium` and `large` gain `replan` after every implement lap; `large`
gains `verify` at chapter exit, reading the holdout from outside the
workdir and routing pass or fail. Two seeds go end to end through `plan`
and then `medium` or `large`, and the proof record is written. Promises
P6, P9, P10, and P8's real-stage leg (`show --stage` against recorded
stages of the seed runs, backlog item 4).

## Pyramid index

- L0: A v2 package executes on the loop node with sprints re-planned
  each lap and chapters proven by a verifier that operates the software.
- L1:
  - `replan`: own node, cheap model, fresh context, edits open sprint
    items only.
  - `verify`: other provider than the coder, fresh context, tools; reads
    the validation design and, for a universal promise with a holdout,
    the path `$XDG_STATE_HOME/tractor/holdouts/<build>.path` records
    (declaration §4, decision 43); routes pass to the
    chapter loop, fail to the sprint loop or a human question.
  - Holdout handoff from the design lap to the `verify` prompt.
  - Two known seeds planned and executed end to end; proof record under
    `proof/planning-v2/`; docs and skill; closeout.
- L2: sprints in `sprints.md`; execution side in `planning-workflow.md`
  §5; seeds under `seeds/`.

## Vector

Decisions 41, 43, 44, 57. Research R3 (capture-format rule; this
project's own software is a CLI, so evidence is transcripts and exit
codes), R2 (an operating verifier has no measured precedent: instrument,
do not cite).

## Review posture

The proof run is the artifact: identity block, claims, scope check,
findings, in the format of `proof/living-instructions-build/README.md`.
A chapter's `done: true` that is not preceded by a `verify` pass in the
timeline fails P6 regardless of anything else.

## Non-goals

A secret holdout location. Screenshot tooling for this project. The web
client. The complexity-signals research (its own project).
