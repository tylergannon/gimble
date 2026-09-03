# P1: elicitation adds a promise; a declined promise becomes an exclusion

Archetype: scenario. Lap 12; answers `review-11.md`.

## Story

1. The check makes an empty scratch repository and a fresh run directory
   (`validation/lib/run-plan.sh`). It writes the seed itself at check
   time from `seeds/three-features.tmpl`: a one-paragraph CLI whose three
   features are drawn at random from a pool of twelve in the template,
   with nothing about error handling, encodings, configuration, stdin,
   or exit codes. The chosen three are the scenario's tokens.
2. `plan` runs with `observer.sh` on the run directory. The observer's
   tree lives beside the run directory, not under it, so nothing under
   `TRACTOR_RUN_DIR` tells the planner it is being watched. Rules, in
   order:
   - decline: the first `Promise:` line in a question file that names
     none of the three chosen features; answer for that candidate "No.
     Record it under exclusions."
   - accept: every other `Promise:` line, "Yes."
   - default: a question with no `Promise:` line, "Proceed with your
     recommendation."
   The `Promise:` line is part of the question-file seam (declaration
   section 4). When a question numbers its candidates, the observer's
   answer mirrors the numbering (decision 39). Every answer ends with an
   `observer:` nonce line.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: `QuestionAsked(question, ts)`, `PipelineCompleted`.
- Every segment under `events/`: each `tool_call` that invoked `tractor
  ask`, with its `ts`, and its paired `tool_result` (the answer text
  `tractor ask` printed, nonce included) with its `ts`.
- The observer tree: the package copy taken immediately before the
  declined question's answer; `answers.log` with answer timestamps,
  nonces, and, for the decline, the quoted `Promise:` line and its
  number.
- `interview/NNNN.md` and `.answer.md`; the generated seed.
- `brief.md` from the package at the end.

## Validator

`command`: `prove/p1-elicitation.sh`: the run completed
(`PipelineCompleted` present, last `StageCompleted` has `next: success`);
`answers.log` has at least one decline line; that line's question id has
a `QuestionAsked` event E whose `question` path is that file; the ask blocked a turn and its
answer reached an agent: either some segment has a `tractor ask`
`tool_call` with `ts` before E and a paired `tool_result` with `ts`
after the answer timestamp, with E's nonce appearing in that segment at
or after the result (in the result itself, or in a later tool result
or assistant text when the agent captured the answer through command
substitution and read it back), or a tool stage's `StageStarted`
precedes E, its `StageCompleted` follows the answer timestamp, its
`tool.log` contains E's nonce (a tool node running `tractor ask`), and
a later codergen segment's `tool_result` contains the nonce (an agent
read the answer); concurrent asks are fine because the nonce picks the
right one; the `.answer.md` for that id contains
the decline text, under the candidate's number when the question
numbered them; the final `brief.md` has an exclusions section with
content, where an exclusions section is any Markdown heading, at any
level, whose text contains "exclusions" (case-insensitive), followed by
at least one non-empty line (list item, paragraph, or table row) before
the next heading. If no question file carried a `Promise:` line, the
script fails with "question-file seam not used". If every candidate
named a chosen feature, so nothing was declined, it exits "inconclusive:
nothing to decline" and the item stays open.

`infer` (files: the generated seed, the declined question file in full,
its answer file, `answers.log`, the `brief.md` copy before the answer,
the final `brief.md`): "answers.log quotes the `Promise:` line the
decline rule matched and its number. Read the whole question file.
Does it actually ask the human to decide the promise on that line,
with the line belonging to the question rather than standing apart
from it, and does the answer file decline that candidate and no other?
Name the capability the line describes. Fail unless the question asks
about it, the seed does not mention it, the copy taken before the
answer does not already exclude that capability (other exclusions may
exist), and the final brief.md declines it under its exclusions, in any
wording or layout."

## Not proven

That the anticipated promise was a good one, or that the planner would
anticipate what a real user cares about; a planner that always asks one
generic unnamed promise satisfies P1 as written, and one that
special-cases the twelve-feature pool is not caught (the pool is
committed; no holdout in this project). That the planner behaves the
same unobserved: the observer leaves no trace under the run directory,
and the promise is about what `plan` does on this seed, which is what
the run shows. Which node asked. An anticipated promise phrased around
a chosen feature is accepted rather than declined, which can only make
the run inconclusive.
