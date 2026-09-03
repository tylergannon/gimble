# P2: the brief/research loop halts through the tool node, and a finding is asked, not applied

Archetype: scenario. Lap 7; answers `review-6.md`.

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
4. The check then exercises the halt predicate itself: it takes the
   halt node's command as `show plan --node halt --raw` prints it for
   this project, runs it in the finished package with
   `research/findings.md` holding one open finding (expect a non-zero
   exit) and with it empty and `research/plan.md` holding no open entry
   (expect zero). The command is what the engine ran (P8 binds `show`
   to `Build`); running it under both states is what proves it decides
   by the files.

## Evidence

- The graph as `show plan` prints it: the `halt` node's command and its
  two routes (`decompose` on success, `brief` on error).
- `timeline.jsonl`: every `halt` `StageCompleted(next)`; stages of the
  research branches and fan-in; `QuestionAsked`.
- `stages/<seq>-halt/tool.log` for every halt turn; the check's own
  transcript of running the halt command under both states.
- The research fan-in's `prompt.md` and `response.md`, and its segment
  (what it read).
- Every segment's `tractor ask` `tool_call` and paired `tool_result`.
- `observer/` copies: at every stage boundary; immediately before the
  citing question's answer. `answers.log`.
- The question and answer files; the final `promises.md`.

## Validator

`command`: `prove/p2-halt.sh`: the graph routes `halt` to `decompose`
on success and to `brief` on error (the loop exists as decision 46
describes; a straight line has no error route back); the halt command
run by the check exits non-zero with an open finding and zero with
none; every `halt` stage directory contains `tool.log` (a tool node);
the last `halt` `StageCompleted` has `next: decompose`; exactly one
question file names the planted leaf or a finding id, with event E;
some `tractor ask` `tool_call` has `ts` before E and a paired
`tool_result` after the answer timestamp (the asking turn blocked on
the human); the finding was research's: `research/findings.md` first
contains it in the copy at a research branch or fan-in
`StageCompleted`, and not in the copy at the `StageCompleted`
preceding that research stage; no copy taken before the answer has
`promises.md` saying JSON (the promise was not changed before the
human said so; a brief that writes `promises.md` only after the answer
is fine); `promises.md` and `brief.md` are byte-identical between the
copies before and after every research branch and fan-in stage
(research applied nothing).

`infer` (files: the fan-in's `prompt.md`, `response.md`, and segment;
the citing question, its answer, the `findings.md` copy at the
question, the `promises.md` copies before the answer and at the end):
"Two judgments. Did the fan-in derive the finding from the planted
leaf (its segment shows the leaf read and its response cites it)? Does
the question ask whether to apply the finding to the promise, rather
than report a change already made? Fail on any no."

## Not proven

Convergence on arbitrary seeds; how many laps the loop takes; whether
the first halt sent the loop back (a run in which research resolves the
finding before the first halt satisfies P2); whether and where the
accepted change was applied afterwards (P2 promises the ask, not the
application). Exhaustion of `max_visits` needs no check: it fails the
run. The snapshot race (ledger rules) applies to the research-stage
identity checks; the ask-then-block check and the halt predicate check
do not depend on it.
