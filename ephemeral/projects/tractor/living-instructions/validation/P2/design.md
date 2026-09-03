# P2: the brief/research loop halts through the tool node, and a finding is asked, not applied

Archetype: scenario. Lap 2; answers `review-1.md`.

## Story

1. The check makes a scratch repository whose `research/` token cache
   already holds one planted leaf, `research/customer/integration-guide.md`,
   presented as the customer's integration guide, saying the consumer
   accepts only JSON. Seed `seeds/csv-export.md` says the tool "exports
   CSV". Research reads the cache before it searches, so the leaf is
   indexed on the first research lap and contradicts a stated promise.
2. `plan` runs with `answerer.sh`. Rules: a question that names the
   planted leaf's path or a finding id (`F<n>` from
   `research/findings.md`) is answered "Yes. Change the promise: the
   export is JSON, not CSV."; other `Promise:` lines "Yes."; default
   "Proceed with your recommendation." The answerer copies `promises.md`
   and `research/findings.md` at the moment of every `QuestionAsked`.
3. Wait for `COMPLETED`.

## Evidence

- `timeline.jsonl`: stages of `brief`, the research branches and fan-in,
  `halt`, `decompose`.
- `stages/<seq>-halt/`: `tool.log` and `outcome.json`.
- `answers/<id>/` snapshots and the question and answer files.
- Segments of every research branch and the research fan-in (node ids
  from the plan graph's research block).
- `promises.md` at the end.

## Validator

`command`: `prove/p2-halt.sh`: the last `halt` stage directory contains
`tool.log` (only a tool node has one) and a success `outcome.json`, and
its `StageCompleted` has `next: decompose`, so the halt was the tool
node's decision; exactly one question file names the planted leaf or a
finding id; the answerer's `promises.md` snapshot taken at that question
still says CSV and the final `promises.md` says JSON (the change followed
the answer); no research segment contains a `tool_call` whose arguments
name `promises.md` or `brief.md` (research wrote nothing outside
`research/`).

`infer` (files: the citing question, its answer, the `findings.md`
snapshot, the final `promises.md`): "Does the question ask whether to
apply the finding to the promise, with a recommendation, rather than
report a change already made? Fail if the change is described as done,
or if the finding in findings.md was not the source of the question."

## Not proven

Convergence on arbitrary seeds; how many laps the loop takes. Only that
the halt is the tool node's decision and that findings route through the
human. Exhaustion of `max_visits` needs no check: it fails the run, it
cannot route to `decompose`.
