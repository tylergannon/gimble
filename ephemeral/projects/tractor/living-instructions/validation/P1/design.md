# P1: elicitation adds a promise; a declined promise becomes an exclusion

Archetype: scenario. Lap 7; answers `review-6.md`.

## Story

1. The check makes an empty scratch repository and a fresh run directory
   (`validation/lib/run-plan.sh`). It writes the seed itself at check
   time from `seeds/three-features.tmpl`: a one-paragraph CLI whose three
   features are drawn at random from a pool of twelve (word count, line
   reverse, banner, rot13, line numbering, trailing-space trim, CSV to
   TSV, duplicate-line removal, hex dump, word frequency, line sort,
   paragraph wrap), with nothing about error handling, encodings,
   configuration, stdin, or exit codes. The chosen three are the
   scenario's tokens.
2. `plan` runs with `observer.sh` on the run directory. Rules, in order:
   - decline: the first `Promise:` line in a question file that names
     none of the three chosen features; answer for that candidate "No.
     Record it under exclusions."
   - accept: every other `Promise:` line, "Yes."
   - default: a question with no `Promise:` line, "Proceed with your
     recommendation."
   The `Promise:` line is part of the question-file seam (declaration
   section 4).
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `QuestionAsked(question, ts)`, `PipelineCompleted`.
- Every segment under `events/`: each `tool_call` that invoked `tractor
  ask`, with its `ts`, and its paired `tool_result` with its `ts`.
- `observer/`: the package copy taken immediately before the declined
  question's answer; `answers.log` with answer timestamps (the
  observer's own record of what it was shown and what it answered).
- `interview/NNNN.md` and `.answer.md`; the generated seed.
- `brief.md` from the package at the end.

## Validator

`command`: `prove/p1-elicitation.sh`: the run completed
(`PipelineCompleted` present, last `StageCompleted` has `next: success`);
`answers.log` has at least one decline line; that line's question id has
a `QuestionAsked` event E whose `question` path is that file; some
`tractor ask` `tool_call` in some segment has `ts` before E and a paired
`tool_result` (same `call_id`) with `ts` after the answer timestamp in
`answers.log` (a turn asked through `tractor ask` and blocked until the
human answered; concurrent asks are allowed, any bracketing call
qualifies); the `.answer.md` for that id contains the decline text; in
the copy taken immediately before the answer, `brief.md` either does
not exist or has no exclusions section with an entry; the final
`brief.md` has an exclusions section with at least one entry, where an
exclusions section is any Markdown heading, at any level, whose text
contains "exclusions" (case-insensitive), followed by at least one list
item or paragraph before the next heading. If no question file carried
a `Promise:` line, the script fails with "question-file seam not used".
If every candidate named a chosen feature, so nothing was declined, it
exits "inconclusive: nothing to decline" and the item stays open.

`infer` (files: the generated seed, the declined question file in full,
the `brief.md` copy before the answer, the final `brief.md`): "Read the
whole question file. Does it actually ask the human to decide the
promise on the `Promise:` line the decline rule matched (quoted in
answers.log), with that line belonging to the question rather than
standing apart from it? Name the capability the line describes. Fail
unless the question asks about it, the seed does not mention it, the
copy's exclusions did not already decline it, and the final brief.md
declines it under its exclusions, in any wording."

## Not proven

That the anticipated promise was a good one, or that the planner would
anticipate what a real user cares about; a planner that always asks one
generic unnamed promise satisfies P1 as written. Which node asked. An
anticipated promise phrased around a chosen feature is accepted rather
than declined, which can only make the run inconclusive.
