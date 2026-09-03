---
chapter: Execution and the live proof
items: []
---

# Chapter 6 sprints

Empty ledger; written when the chapter is entered. The run that builds
this chapter was materialized before `replan` existed, so within it the
plan node edits this backlog when it runs (decision 44); `replan` first
re-plans, and `verify` first gates a chapter mark, in runs started after
items 1 and 2 land (the seed runs of items 3 and 4 are the first); this
chapter's own mark is the engine's, on its command, like chapters 4 and 5.

## Backlog sketch, in order

1. `replan` in `medium` and `large`: node, prompt, `models.yaml` role;
   proof, from the ledger diff across each `replan` turn, that it edits
   only open items, never the item the lap's `LoopItemSelected` names,
   and never the chapter ledger; and that on the `verify` fail edge it
   either appends exactly one open item or asks the human (a
   `QuestionAsked` in its segment), never neither.
2. `verify` in `large` with the holdout handoff: node, prompt, route
   targets (pass to `chapters`, fail to `replan`); the holdout path recorded at design time and rendered only
   here. Proof: a small scratch package with one universal promise whose
   holdout sample is disclosed to no prompt but `verify`'s (obscure, not
   secret, decision 43: an unsandboxed coder could find it on disk; the
   check reads the coder's segments for any read of the holdout path
   and fails the sprint on one; see interview 0014, question 3).
3. The greeter seed (fixed at commit 6dec5dc) end to end: `plan`, then
   the printed handoff, then the seed's acceptance examples, with the
   `infer` judge P10's design names over the tooling's logs, both run
   directories, and the built program's transcripts. Proof: the greeter
   half of P10.
4. The ledger-tool seed end to end: `plan`, then `large`, then the
   examples with the same `infer` judge; P6's check and judge on that
   run; the proof record under `proof/planning-v2/` written from both
   runs. Proof: P6, the ledger-tool half of P10, and `chapters/04-library/prove/p8-show-stage.sh`
   (P8's real-stage leg, over the greeter and ledger-tool runs).
5. Docs and skill (P9), including where the interview directory and the
   holdout come from: the docs-only reader runs plan, ask, answer, and
   show. Proof: P9.
