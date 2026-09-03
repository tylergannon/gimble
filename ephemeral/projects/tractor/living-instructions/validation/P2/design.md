# P2: the brief/research loop halts through the tool node, and a finding is asked, not applied

Archetype: scenario. Lap 3; answers `review-2.md`.

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
- `observer/` snapshots of `promises.md` and `research/findings.md` at
  every stage boundary and at the citing question; `answers.log`.
- The question and answer files.

## Validator

`command`: `prove/p2-halt.sh`: the last `halt` stage directory contains
`tool.log` (only a tool node has one) and a success `outcome.json`, and
its `StageCompleted` has `next: decompose`, so the halt was the tool
node's exit status, not an agent's choice; exactly one question file
names the planted leaf or a finding id; the finding was research's:
`research/findings.md` first gains that finding across a research
branch or fan-in stage (absent in the snapshot at its `StageStarted`,
present at its `StageCompleted`) and across no `intake` or `brief`
stage; the promise changed only after the answer: `promises.md` says
CSV in the snapshot at the citing `QuestionAsked` and JSON in the
snapshot at the `StageCompleted` of the `brief` stage that asked; and
`promises.md` and `brief.md` are byte-identical across every research
branch and fan-in stage (research applied nothing).

`infer` (files: the citing question, its answer, the `findings.md`
snapshot at the question, the `promises.md` snapshots before and after
the brief stage): "Does the question ask whether to apply the finding
to the promise, rather than report a change already made? Fail if the
change is described as done, or if the question does not rest on the
finding in findings.md."

## Not proven

Convergence on arbitrary seeds; how many laps the loop takes. Only that
the halt is the tool node's decision and that findings route through the
human. Exhaustion of `max_visits` needs no check: it fails the run, it
cannot route to `decompose`. Whether the question carried a
recommendation; P2 does not ask for one.
