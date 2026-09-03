# P2: the brief/research loop halts through the tool node, and a finding is asked, not applied

Archetype: scenario. Lap 4; answers `review-3.md`.

## Story

1. The check makes a scratch repository whose `research/` token cache
   already holds one planted leaf, `research/customer/integration-guide.md`,
   presented as the customer's integration guide, saying the consumer
   accepts only JSON. Seed `seeds/csv-export.md` says the tool "exports
   CSV". Research reads the cache before it searches, so the leaf is
   indexed on the first research lap and contradicts a stated promise.
2. `plan` runs with `observer.sh`. Rules: a question that names the
   planted leaf's path or a finding id (`F<n>` from
   `research/findings.md`) is answered "Yes. Change the promise: the
   export is JSON, not CSV."; other `Promise:` lines "Yes."; default
   "Proceed with your recommendation."
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: stages of `intake`, `brief`, the research branches
  and fan-in, `halt`, `decompose`; `QuestionAsked`.
- `stages/<seq>-halt/`: `tool.log` and `outcome.json`.
- The asking `brief` stage's segment: the `tractor ask` `tool_call` and
  its paired `tool_result`.
- `observer/` copies: at every stage boundary; at the citing
  `QuestionAsked`; immediately before its answer. `answers.log`.
- The question and answer files; the final `promises.md`.

## Validator

`command`: `prove/p2-halt.sh`: the last `halt` stage directory contains
`tool.log` (only a tool node has one) and a success `outcome.json`, and
its `StageCompleted` has `next: decompose`, so the halt was the tool
node's exit status; exactly one question file names the planted leaf or
a finding id; a `brief` segment's `tractor ask` `tool_call` for that
question has its paired `tool_result` after the answer timestamp (the
turn blocked on the human); the finding was research's:
`research/findings.md` first contains it in the copy at a research
branch or fan-in `StageCompleted`, and not in the copy at the
`StageCompleted` preceding that research stage; the promise changed only
after the answer: `promises.md` says CSV in the copy taken immediately
before the answer and JSON in the final package; `promises.md` and
`brief.md` are byte-identical between the copies before and after every
research branch and fan-in stage (research applied nothing).

`infer` (files: the citing question, its answer, the `findings.md` copy
at the question, `promises.md` before the answer and at the end): "Does
the question ask whether to apply the finding to the promise, rather
than report a change already made? Fail if the change is described as
done, or if the question does not rest on the finding in findings.md."

## Not proven

Convergence on arbitrary seeds; how many laps the loop takes; in which
later stage the accepted change was written. Only that the halt is the
tool node's decision and that findings route through the human.
Exhaustion of `max_visits` needs no check: it fails the run, it cannot
route to `decompose`. The snapshot race (ledger rules) applies to the
research-stage identity checks; the ask-then-block check does not
depend on it.
