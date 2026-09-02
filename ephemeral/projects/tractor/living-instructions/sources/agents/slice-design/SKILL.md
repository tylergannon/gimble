---
name: slice-design
description: >
  Turn an accepted nlspec plus the current code into vertical work slices,
  each with an ephemeral technical spec and an implementer handoff contract.
  Use when planning implementation work governed by an nlspec.
---

# Slice Design

Load the spec-writing standard first. The inputs are the nlspec and the
current code; the outputs are vertical slices, each carrying an ephemeral
slice spec. The slice spec is authoritative for its slice only, then archived
in the PR — never maintained. It may contain nothing with cross-slice
lifetime (the lifetime rule).

## Slicing

Slice vertically: each slice crosses the stack thinly and yields something
exercisable early — curl-able, browsable, claim-demonstrable. Reject
horizontal stack-order plans (migrations → services → API → frontend); that
is the default model tic, and under it nothing is touchable until the end.
Sequence slices so each demonstrates at least one claim or de-risks a seam,
and so no slice cuts across a seam mid-contract.

Within a slice, order steps so each leaves the slice exercisable. Typically
that is middle-out: the seam contract serving mock data first (curl-able),
the consumer against the mock (browsable), then real behavior behind the
contract, then persistence — but a claim about durability or idempotency
pulls persistence forward. Error behavior named in a contract, claim, or the
nlspec is required work and lands within the slice, before acceptance; only
unspecified failure modes are drift.

## The slice spec

Derive it fresh from the nlspec and the code as it is today, conditioned on
what already exists. Contents, using the artifacts from the notation standard:

- call-stack tree for the control-flow change (diff syntax);
- file-tree diff, respecting cell boundaries — the countable preview of "is
  this change about to spray across seams";
- types and signatures for key new functions;
- the existing functions and types to modify;
- which claims this slice demonstrates, and how;
- the enumerated work items for the slice — the acceptance denominator.

Content that is derivable by the implementer in context is omitted; content a
future slice must agree with is promoted to the nlspec before this slice
ships, via HITL.

## Mocks

A mock is a claim about the real service, and it is untrusted until
**authenticated**. Authentication is evidentiary, not diligence: a response
recorded from the real service with real credentials, archived as a proof
artifact, that the mock demonstrably reproduces — plus a review of the mock
against the stated contract and the service's documented behavior, confirming
it suffices to validate the cell's implementation of the contract. Reading
the docs and declaring agreement is not authentication. Where no real access
exists, escalate to HITL; never self-downgrade the claim to a mock
demonstration.

An authenticated mock may stand in for a cell-local claim only, and only
until the integration claim covering the same seam is demonstrated against
the live service. A demonstration against an unauthenticated mock is an
assertion, not a demonstration.

## Implementer handoff

Every slice ships with the handoff contract:

- Contracts are inviolable. If the work forces a contract re-evaluation, stop
  and report; do not amend the nlspec.
- An underivable decision, or any flagged deferred decision, escalates
  immediately as a spec defect. Improvisation at a seam is a defect in itself.
- A budget breach stops work; the spec decides the split.
- Claims are demonstrated per /proof-of-work, not asserted. Removing,
  skipping, or weakening a failing test violates this contract.
- Build only what the spec requires. Gold plating, belt-and-suspenders
  defenses, and fixes for undemonstrated hypothetical failures are drift, not
  diligence; error behavior named in a contract, claim, or the nlspec is
  required work, never gold plating.

## Acceptance

A slice is accepted when the exit valves pass and all three hold:

1. **The main point is proven.** Every routed claim is demonstrated per
   /proof-of-work. This core is 100% — never subject to the percentage below.
2. **No known demonstrated bugs.** A bug blocks acceptance when demonstrated
   by a failing unit test or a broken example program showing genuinely
   broken software. Acceptance bounds scope, not verification: curtailing
   tests on already-written code to avoid discovering bugs, or removing,
   skipping, or weakening a failing test, is a handoff-contract violation. A
   suspicion is discharged by a demonstration attempt or a written dismissal
   filed with the residue; a bug demonstrated after acceptance reopens the
   slice. Issues touching no routed claim or contract do not block.
3. **Slice-spec work items are 90–95% done.** The denominator is the slice
   spec's enumerated work items, not the claims. Work explicitly named by a
   contract, claim, or the nlspec is never eligible residue. Enumerate the residue and
   file each item as follow-up work before accepting; residue that exposes
   an underivable decision goes through spec-defect escalation instead.

Meeting the threshold is an instruction to accept.

## Review scope

HITL review is spent where lifetime lives: the slice spec before
implementation begins, seam-touching diffs, and every proposed nlspec
amendment. Seam-touching is not self-reported: a diff requires review when
the seam-integrity checks' per-diff signal (see spec-writing) reports a
public-surface change or a new or removed cross-cell call site — and
additionally in the cases below. Cells handling credentials, authentication
or authorization decisions, irreversible or destructive operations, or data
migrations are human-reviewed regardless of cell locality — tripwires cannot
see those failure classes.

Other cell internals are not line-reviewed: they are fungible, and the
budget and claim tripwires are the quality instrument there — a deliberate
divergence from per-slice full code review. Every rule here is exploitable
by the agent it governs; the tripwires bound drift, and the check on
rule-gaming is independent multi-model review. The model that did the work
never adjudicates its own compliance.

## Exit valves

Before a slice's PR merges, verify:

1. **Promotion** — nothing with cross-slice lifetime remains in the slice
   spec; promoted decisions are in the nlspec.
2. **Conversion** — new tests worth keeping as behavior are run through the
   catalytic converter (see cell-lifecycle) as claim candidates.
3. **Archive and budgets** — the slice spec is archived in the PR, and every
   touched cell's budget is re-checked mechanically.
4. **Review** — every review the Review scope section requires (slice spec,
   routed seam-touching diffs, security-class cells) is complete.
