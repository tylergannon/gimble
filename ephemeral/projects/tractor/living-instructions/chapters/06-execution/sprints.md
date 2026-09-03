---
chapter: Execution and the live proof
items: []
---

# Chapter 6 sprints

Empty ledger; written when the chapter is entered. The run that builds
this chapter was materialized before `replan` existed, so within it the
plan node edits this backlog when it runs (decision 44); `replan` first
re-plans in runs started after item 1 lands (the seed runs of items 3
and 4 are the first).

## Backlog sketch, in order

1. `replan` in `medium` and `large`: node, prompt, `models.yaml` role;
   proof that it edits only open items and never the chapter ledger.
2. `verify` in `large` with the holdout handoff: node, prompt, route
   targets; the holdout path recorded at design time and rendered only
   here. Proof: a small scratch package with one universal promise whose
   holdout sample is disclosed to no prompt but `verify`'s (obscure, not
   secret, decision 43: an unsandboxed coder could find it on disk; the
   check reads the coder's segments for any read of the holdout path
   and fails the sprint on one; see interview 0014, question 3).
3. The greeter seed (fixed at commit 6dec5dc) end to end: `plan`, then
   the printed handoff, then the seed's acceptance examples. Proof: the
   greeter half of P10.
4. The ledger-tool seed end to end: `plan`, then `large`, then the
   examples; P6's check on that run; the proof record under
   `proof/planning-v2/` written from both runs. Proof: P6, the
   ledger-tool half of P10, and `chapters/04-library/prove/p8-show-stage.sh`
   (P8's real-stage leg, over the greeter and ledger-tool runs).
5. Docs and skill (P9), including where the interview directory and the
   holdout come from: the docs-only reader runs plan, ask, answer, and
   show. Proof: P9.
