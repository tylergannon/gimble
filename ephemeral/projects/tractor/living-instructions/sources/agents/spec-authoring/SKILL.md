---
name: spec-authoring
description: >
  Author a new nlspec collaboratively with HITL: grill for goals, propose the
  cell decomposition, harden contracts external-first, and stop when spanning
  is reached. Use when starting a spec from scratch or respecifying a system.
---

# Authoring an nlspec

Load the spec-writing standard first; it defines every term used here. The
output is an nlspec, reviewed with spec-review before use.

## Workflow

1. **Goals.** Establish the needs and goals the project serves. Use /grilling
   to interview HITL — scope, audience, high-level objectives, success
   criteria. Do not ask fine details here. Draft the Overview and Goals
   section; iterate with HITL until they approve it.
2. **Decomposition.** Propose the cells and seams without contract specifics.
   Apply the Parnas test to every proposed seam: does the reason to change
   differ across it? Present the decomposition as a cell list plus a
   file-tree sketch; refine with HITL. Assign each cell a budget.
3. **Contracts.** Harden contracts seam by seam, external seams first, using
   the notation standard. Get HITL approval for each materially new public
   concept.
4. **Claims and exclusions.** Write the definition of done; classify every
   claim cell-local or integration and name the cells it routes through. If
   most claims are integration-only, revisit step 2. Capture exclusions as
   they surface — anything HITL declines is an exclusion candidate.

## Stopping rule

Keep asking questions only while the answers land in the nlspec's own
sections. If the next answer could change a seam, claim, budget, or exclusion,
answer it before stopping. If it would only change cell internals, do not ask
it — that is derivable span, and answering it up front is speculation.

Authoring ends after two consecutive grill rounds that surface only derivable
material. Questions never run out; seam-moving answers do.

Record consciously unanswered hard questions in the Deferred Decisions
section. Spanning is maintained over the spec's lifetime through the two
pumps, not achieved exhaustively at authoring; a deferred question answered
when a slice actually hits it is answered with the construction attempt in
hand, and paid for only if it arises.

## Handoff

Run spec-review on the draft before it governs any work. For material specs,
offer HITL an independent pass via /request-adversarial-review.
