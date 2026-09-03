# P2: the brief/research loop halts through the tool node, and a finding is asked, not applied

Archetype: scenario. Lap 6; answers `review-5.md`.

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

- `timeline.jsonl`: every `halt` `StageCompleted(next)`; stages of the
  research branches and fan-in; `QuestionAsked`.
- `stages/<seq>-halt/tool.log` for every halt turn, and the halt
  command itself as `show plan --node halt --raw` prints it.
- The research fan-in's `prompt.md` and `response.md`, and its segment
  (what it read).
- Every segment's `tractor ask` `tool_call` and paired `tool_result`.
- `observer/` copies: at every stage boundary; immediately before the
  citing question's answer. `answers.log`.
- The question and answer files; the final `promises.md`.

## Validator

`command`: `prove/p2-halt.sh`: every `halt` stage directory contains
`tool.log` (a tool node); the last `halt` `StageCompleted` has `next:
decompose`; exactly one question file names the planted leaf or a
finding id, with event E; some `tractor ask` `tool_call` has `ts`
before E and a paired `tool_result` after the answer timestamp (the
asking turn blocked on the human); the finding was research's:
`research/findings.md` first contains it in the copy at a research
branch or fan-in `StageCompleted`, and not in the copy at the
`StageCompleted` preceding that research stage; `promises.md` says CSV
in the copy taken immediately before the answer; `promises.md` and
`brief.md` are byte-identical between the copies before and after
every research branch and fan-in stage (research applied nothing).

`infer` (files: the halt command text and every halt `tool.log`; the
fan-in's `prompt.md`, `response.md`, and segment; the citing question,
its answer, the `findings.md` copy at the question, `promises.md`
before the answer and at the end): "Three judgments. Does the halt
command decide by reading `research/findings.md` and the open entries
of `research/plan.md`, and does each tool.log show that decision, so
that the final route to decompose was the tool's reading of an empty
findings file and no open entry? Did the fan-in derive the finding from
the planted leaf (its segment shows the leaf read and its response
cites it)? Does the question ask whether to apply the finding to the
promise, rather than report a change already made? Fail on any no."

## Not proven

Convergence on arbitrary seeds; how many laps the loop takes; whether
the first halt sent the loop back (a run in which research resolves the
finding before the first halt satisfies P2); whether and where the
accepted change was applied afterwards (P2 promises the ask, not the
application). Exhaustion of `max_visits` needs no check: it fails the
run. The snapshot race (ledger rules) applies to the research-stage
identity checks; the ask-then-block check does not depend on it.
