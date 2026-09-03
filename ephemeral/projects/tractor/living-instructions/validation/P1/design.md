# P1: elicitation adds a promise; a declined promise becomes an exclusion

Archetype: scenario. Lap 3; answers `review-2.md`.

## Story

1. The check makes an empty scratch repository and a fresh run directory
   (`validation/lib/run-plan.sh`). Seed `seeds/three-features.md`: a CLI
   `wordkit` with exactly three named features, `count` (words in a
   file), `reverse` (lines), and `banner` (print a word large), and
   nothing about error handling, encodings, configuration, stdin, or exit
   codes.
2. `plan` runs with `observer.sh` on the run directory. Rules, in order:
   - decline: the first `Promise:` line in a question file that contains
     none of the tokens `count`, `reverse`, `banner`; answer for that
     candidate "No. Record it under exclusions."
   - accept: every other `Promise:` line, "Yes."
   - default: a question with no `Promise:` line, "Proceed with your
     recommendation."
   The `Promise:` line is the question-file convention for a promise
   candidate (declaration section 4, the question-file seam: batched
   questions are a convention inside the file). A planner that abandons
   the convention breaks the seam, and the check says so; that is a
   defect, not a stricter reading of P1.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `QuestionAsked(question)`, `StageStarted` and
  `StageCompleted` for `brief`.
- `observer/`: the package snapshot at the declined `QuestionAsked` (in
  particular `brief.md` as it was when the question was asked) and at
  every `brief` `StageCompleted`; `answers.log`.
- `interview/NNNN.md` and `.answer.md`.
- `brief.md` from the package at the end.

## Validator

`command`: `prove/p1-elicitation.sh`: the run completed
(`PipelineCompleted` present, last `StageCompleted` has `next: success`);
`answers.log` has at least one decline line; that line's question id has
a `QuestionAsked` event whose `question` path is that file, emitted
between a `brief` `StageStarted` and its `StageCompleted` (a real
`tractor ask` that blocked a real brief turn); the `.answer.md` for that
id contains the decline answer text; in the observer snapshot taken at
that `QuestionAsked`, `brief.md` either does not exist or has no
Exclusions entry (the exclusion was not written before the answer); in
the snapshot at that brief stage's `StageCompleted`, `brief.md` has an
`## Exclusions` section with at least one entry. If no question file
carried a `Promise:` line, the script fails with "question-file seam
not used". If every candidate named a seed feature, so nothing was
declined, it exits "inconclusive: nothing to decline" and the item
stays open.

`infer` (files: the seed, the declined question file, the `brief.md`
snapshot at the question, the final `brief.md`): "Read the `Promise:`
line the decline rule matched (quoted in answers.log). Name the
capability it describes. Fail unless the seed does not mention that
capability, the snapshot's exclusions did not already decline it, and
the final brief.md declines it under Exclusions, in any wording."

## Not proven

That the anticipated promise was a good one, or that the planner would
anticipate what a real user cares about. Only that anticipation happens,
that the ask is a real `tractor ask`, and that declining lands in
exclusions after the answer. The token rule is scaffolding: an
anticipated promise phrased around a seed feature ("count handles empty
files") is accepted rather than declined, which can only make the run
inconclusive, never a false pass.
