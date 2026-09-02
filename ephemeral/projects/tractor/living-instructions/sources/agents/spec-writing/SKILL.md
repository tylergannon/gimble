---
name: spec-writing
description: >
  The nlspec standard: what a Natural Language Spec is and must contain —
  seams, cells, contracts, budgets, claims, exclusions, the artifact
  hierarchy, and pseudocode notation. Load before reading, writing,
  reviewing, or executing against an nlspec. For workflows use
  spec-authoring, spec-review, slice-design, and cell-lifecycle.
---

# The nlspec Standard

An nlspec is a **minimally spanning** natural-language specification: a basis
from which the whole software is constructible given the materials plus the
rules of engagement. Spanning: every question a competent implementer could
not settle alone is answered. Minimal: *only* those questions are answered;
anything a capable reader can derive in context is elided.

**Membership test.** A decision belongs in the nlspec iff it is not derivable
from what is already there. Detector: two capable readers would derive
different answers, and the difference matters at a seam, claim, budget, or
exclusion. If they would differ only in cell internals, leave it out.

## Required contents

- **Overview and goals** — intent, needs served, key design principles.
- **Definition of done** — behavioral claims that must be demonstrated
  (these are proof claims in the /proof-of-work sense). Classify each as
  *cell-local* (provable through one cell's contract) or *integration*
  (provable only end-to-end). If most claims are integration-only, the
  decomposition is suspect. Every seam must be accounted for here: each
  contract is exercised by at least one claim. Done means the application
  actually runs and does its real work — live services, real secrets. An
  authenticated mock (see slice-design) may stand in for cell-local claims
  only; integration claims are demonstrated live, and every seam to an
  external service carries at least one integration claim — a seam
  demonstrated only against mocks never reaches done.
- **Exclusions** — what is deliberately not built, to forestall
  over-engineering.
- **Central data types and algorithms** — language-neutral pseudocode per
  [notation.md](notation.md); exact syntax only for public formats that must
  be shown exactly.
- **Seams, cells, contracts** — see below.
- **Deferred decisions** — questions consciously left unanswered. A tripwire
  list, not a dumping ground: an implementer who hits one escalates
  immediately instead of improvising.

## Seams, cells, budgets

The seams partition the codebase into **cells**, each with an explicit
responsibility, boundary, and contract. Everything inside a cell is
**fungible**: if the contract holds and the claims are demonstrated, any
implementation is acceptable.

- **Seam placement (Parnas test):** a seam belongs where the reason to change
  differs on either side. What changes together lives together.
- **Budget:** every cell carries a maximum code volume, stated in the nlspec
  and mechanically checkable. The budget is a **tripwire, not an implementer
  rule**: on breach, stop and report — the spec decides where the split goes.
  Keep budgets slightly tight; they are the primary drift tripwire (internal
  degradation shows up as volume first).
- **Seam integrity checks** (all mechanical): public surface matches contract;
  nothing outside reaches past a contract; every cell under budget; routed
  claims demonstrated. Per diff, the checks also report any public-surface
  change and any new or removed cross-cell call site — the review-routing
  signal consumed by slice-design.

## Authority and the two pumps

The nlspec and the code are the only durable authorities: the nlspec holds
intent, the code holds fact. Exactly two channels write to the nlspec:

1. **Spec-defect escalation.** Implementers never amend the nlspec. On hitting
   an underivable decision they stop and file a spec defect; the adjudicated
   answer lands in the nlspec; work resumes.
2. **Catalytic converter.** Test cases discovered in code may be elevated into
   claims. Intake is behavioral evidence only — never opinions, rationale
   prose, or "lessons."

## Artifact hierarchy

| Artifact | Role | Lifetime |
|---|---|---|
| nlspec | intent, claims, exclusions, seams, budgets | permanent, maintained, sole authority |
| per-slice technical spec | derivable mechanics for one slice (signatures, call trees, file layout) | one slice, then archived in its PR; never maintained |
| code + tests | fact; tests are candidate claims | permanent |

Routing test for any content: *how long must someone agree with this?*
Forever-and-binding → nlspec. This slice → slice spec. In the code → do not
write it down. The slice spec may only contain decisions whose lifetime is at
most the slice; anything a future slice must agree with is seam-level and must
be promoted to the nlspec before the slice ships.

## Proportionality

The apparatus scales with the blast radius of the change, not with habit:

- **Cell-local change** — no seam, claim, budget, or exclusion touched:
  maintenance mode per cell-lifecycle. No slice spec, no gap analysis; hand
  the agent the contract guardrail and go.
- **Single-cell feature** with claim impact: one slice with a minimal slice
  spec — often just the call-stack tree and the claim it demonstrates.
- **Cross-seam or multi-slice work**: full slice-design.
- **No nlspec governs the code at all**: code merged to a maintained branch
  gets one; prototypes, spikes, and scratch work do not. HITL decides
  borderline cases — the implementing agent does not.

Proportionality scales the slice apparatus only. Any change to a seam,
contract, claim, budget, or exclusion reaches the nlspec only through the
two pumps, with HITL, at every size.

## Style

Prose is the canonical contract. Use pseudocode for data types and for
algorithms whose ordering, precedence, or edge cases would be ambiguous in
prose; it defines required semantics, not a prescribed implementation. Do not
prescribe an implementation language. Seek HITL approval before adding a
materially new public concept or contract.
