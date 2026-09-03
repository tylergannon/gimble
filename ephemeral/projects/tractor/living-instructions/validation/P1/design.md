# P1: elicitation adds a promise; a declined promise becomes an exclusion

Archetype: scenario. Lap 2; answers `review-1.md`.

## Story

1. The check makes an empty scratch repository and a fresh run directory
   (`validation/lib/run-plan.sh`). Seed `seeds/three-features.md`: a CLI
   `wordkit` with exactly three named features, `count` (words in a
   file), `reverse` (lines), and `banner` (print a word large), and
   nothing about error handling, encodings, configuration, stdin, or exit
   codes.
2. `plan` runs with `answerer.sh` on the run directory. Rules, in order:
   - decline: the first `Promise:` line in a question file that contains
     none of the tokens `count`, `reverse`, `banner`; answer for that
     candidate "No. Record it under exclusions." The `Promise:` line is
     the brief prompt's question skeleton (`templates/question.md`,
     chapter 5 sprint 3), one line per candidate promise.
   - accept: every other `Promise:` line, "Yes."
   - default: a question with no `Promise:` line, "Proceed with your
     recommendation."
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `QuestionAsked(question)` events, `StageStarted` and
  `StageCompleted` for `brief`.
- `interview/NNNN.md` and `.answer.md`, and the answerer's copies under
  `answers/<id>/` with `answers.log` (id, rule, timestamp, matched line).
- The `brief` stage segments under `events/`: `tool_call` arguments.
- `brief.md` from the package.

## Validator

`command`: `prove/p1-elicitation.sh`: the run completed
(`PipelineCompleted` present, last `StageCompleted` has `next: success`);
`answers.log` has at least one decline line; that line's question id has
a `QuestionAsked` event whose `question` path is that file, emitted
between a `brief` `StageStarted` and its `StageCompleted` (so the ask
went through `tractor ask` and blocked a real brief turn); the
`.answer.md` for that id contains the decline answer text; the same
`brief` stage's segment has a `tool_call` whose arguments name `brief.md`
with a `ts` later than the decline's timestamp in `answers.log` (the
brief was written after the answer was read); `brief.md` has an
`## Exclusions` section with at least one entry. The script fails with
"no Promise: line asked" if no question used the skeleton; that is a
planner defect, not an inconclusive run.

`infer` (files: the seed, the declined question file, `brief.md`): "Read
the `Promise:` line the decline rule matched (it is quoted in
answers.log). Name the capability it describes. Fail unless the seed
does not mention that capability and an entry under Exclusions in
brief.md declines it, in any wording."

## Not proven

That the anticipated promise was a good one, or that the planner would
anticipate what a real user cares about. Only that anticipation happens,
that the ask is a real `tractor ask`, and that declining lands in
exclusions. The token rule depends on the question skeleton, which is
library content; a planner that abandons the skeleton fails here by
design.
