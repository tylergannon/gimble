# P1: elicitation adds a promise; a declined promise becomes an exclusion

Archetype: scenario.

## Story

1. Scratch repository, empty. Seed `seeds/three-features.md`: a
   one-paragraph product with exactly three named features and nothing
   about, for example, error handling, auth, or data loss.
2. Run `tractor workflow run plan --project p1 --seed seeds/three-features.md`
   with `answerer.sh` on the interview directory and rules: decline the
   first question whose promise text names a capability absent from the
   seed (rule: match on the planner's "do you promise" phrasing, answer
   "No, and record it as an exclusion"); accept everything else; answer
   "begin" to any open-ended prompt.
3. Wait for `COMPLETED`.

## Evidence

- `interview/NNNN.md` and `.answer.md` for the whole run.
- `answers.log`: which rule fired on which question.
- `brief.md` and `promises.md` from the package.
- `timeline.jsonl`.

## Validator

`command`: `prove/p1-elicitation.sh`: the run completed; at least one
question file exists whose id appears in `answers.log` under the decline
rule; `brief.md` has an exclusions section containing the first line of
that question's promise text (the answerer copies it into the log).

`infer` (files: the seed, every question file): "Name every promise the
planner asked about. Fail unless at least one of them names a capability
the seed does not mention."

## Not proven

That the anticipated promise was a good one, or that the planner would
anticipate what a real user cares about. Only that anticipation happens
and declining lands in exclusions.
