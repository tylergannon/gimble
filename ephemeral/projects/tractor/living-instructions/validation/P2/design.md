# P2: the brief/research loop halts through the tool node, and a finding is asked, not applied

Archetype: scenario. Lap 14; answers `review-13.md`.

## Story

1. The check makes a scratch repository whose `research/` token cache
   already holds one planted leaf, `research/customer/integration-guide.md`,
   presented as the customer's integration guide, saying the consumer
   accepts only JSON. Seed `seeds/csv-export.md` says the tool "exports
   CSV". Research reads the cache before it searches, so the leaf is
   indexed on the first research lap and contradicts a stated promise.
2. `plan` runs with `observer.sh`. Rules: the finding rule matches a
   question, asked after `research/findings.md` first holds the
   finding, that mentions any of: the planted leaf's path, its title
   ("integration guide"), a finding id (`F<n>`), or the export's
   format in any of the words JSON, CSV, comma, delimited, encoding,
   format, export; it answers "Yes. Change the promise: the export is
   JSON, not CSV." Other `Promise:` lines "Yes."; default "Proceed
   with your recommendation." Every answer ends with an `observer:`
   nonce line. A question that paraphrases the finding past every
   listed word matches no rule and makes the run inconclusive, never a
   false pass; the infer judge, not the rule, decides citation.
3. Wait for `COMPLETED`.
4. The check then exercises the halt predicate itself: it takes the
   halt node's command as `show plan --node halt --raw` prints it for
   this project and runs it in the finished package with the run's own
   environment reproduced (`TRACTOR_RUN_DIR` set to the run directory,
   the same working directory the tool node had, the same variables the
   engine exports), so a command that behaves differently inside a run
   is exercised as it ran, under three states: one open finding in `research/findings.md` (expect non-zero);
   findings empty but one open entry in `research/plan.md` (expect
   non-zero); both clear (expect zero).

## Evidence

- The graph as `show plan` prints it: the `halt` node's command and
  routes, the `brief` node's successor, and the research nodes. The run
  records neither the graph nor the command a tool node executed (only
  `tool.log`), so the graph evidence is what `show` prints; P8's check,
  in the same binary, is what binds `show` to `Build`.
- `timeline.jsonl`: every `halt` `StageCompleted(next)`; stages of the
  research branches and fan-in; `QuestionAsked`.
- `stages/<seq>-halt/tool.log` for every halt turn; the check's own
  transcript of running the halt command under the three states.
- The research branches' and fan-in's `prompt.md`, `response.md`, and
  segments (what they read and wrote).
- Every segment's `tractor ask` `tool_call` and paired `tool_result`.
- Observer copies: at every stage boundary; immediately before the
  citing question's answer. `answers.log` with nonces.
- The question and answer files; the final `promises.md` and `brief.md`.

## Validator

`command`: `prove/p2-halt.sh`: the graph is the loop decision 46
describes: `halt` routes to `decompose` on success and to `brief` on
error, and `brief`'s successor is the research node; the halt command
run by the check exits non-zero under each of the two open states and
zero under the clear state; every `halt` stage directory contains
`tool.log` (a tool node); the last `halt`
`StageCompleted` has `next: decompose`, and in the observer copy at the
`StageCompleted` of the stage before that halt (the state halt started
from; halt is a tool node that writes nothing, so nothing changes
between that copy and its decision), `research/findings.md` holds no
open finding and `research/plan.md` no open entry (a halt that routed
on with a finding still open fails here whatever its command says); at least one question matched
the finding rule (`answers.log`); call the first E; E lies between a
`brief` stage's `StageStarted` and `StageCompleted` and precedes the
`StageStarted` of the halt that routed to `decompose` (the brief asked
before the loop ended, as decision 46 orders; a question asked by
`decompose` afterwards does not count); some `tractor ask`
`tool_call` has `ts` before E and a paired `tool_result` with `ts`
after the answer timestamp, with E's nonce appearing in that segment at
or after the result (the call waited on this question; the asking turn
blocked on the human); the finding was research's: `research/findings.md` first
contains it in the copy at a research branch or fan-in
`StageCompleted`, and not in the copy at the `StageCompleted`
preceding that research stage; no copy taken before the answer has
`promises.md` or `brief.md` saying the export is JSON (neither the
promise nor the brief was changed before the human said so; a brief
that writes them only after the answer is fine); no promise row of
`promises.md` and no sentence of `brief.md` differs, whitespace aside,
between the copies before and after every research branch and fan-in
stage (research applied nothing; a whitespace normalisation is not an
application).

`infer` (files: the research branches' and fan-in's `prompt.md`,
`response.md`, and segments; the `brief` stage's `prompt.md` and
segment that asked E; the matched question, its answer, the
`findings.md` copy at the question, the `promises.md` and `brief.md`
copies before the answer and at the end): "Two judgments. Did the
research stages derive the finding from the planted leaf (some research
segment shows the leaf read, and the finding in findings.md rests on
it)? Does the question cite that finding, in any wording, and ask
whether to apply it to the promise, rather than report a change already
made in the promise list or the brief? Fail on any no."

## Not proven

A finding cited in words the rule does not list; that run is
inconclusive and is rerun with the rule widened. Convergence on
arbitrary seeds; how many laps the loop takes; whether
the first halt sent the loop back (a run in which research resolves the
finding before the first halt satisfies P2); whether and where the
accepted change was applied afterwards (P2 promises the ask, not the
application). That the command `show` prints is the command the run
executed: the run does not record it, and this design relies on P8's
proof that `show` prints what `Build` materialized in the same binary.
Exhaustion of `max_visits` needs no check: it fails the run. The
snapshot race (ledger rules) applies to the research-stage identity
checks; the ask-then-block check and the halt predicate check do not
depend on it.
