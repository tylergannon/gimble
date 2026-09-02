---
name: spec-review
description: >
  Mechanically review an nlspec for consistency, claim traceability, seam
  quality, and spanning. Use after authoring, after any amendment via the two
  pumps, or when reconciling a spec against related documents.
---

# Reviewing an nlspec

Load the spec-writing standard first. Read the complete spec and its
referenced documents before reporting anything. This is the pass models do
reliably and humans do badly: specs die of internal contradiction more than of
missing detail.

## Passes

Run all of them; report findings per pass.

1. **Consistency.** Check terminology and every repeated rule across prose,
   tables, examples, and pseudocode. Any concept with two names, or rule
   stated twice with drift, is a defect.
2. **Claim traceability — both directions.** Every definition-of-done claim
   must trace to specified behavior, and every specified behavior must serve
   some claim or stated goal. Behavior serving nothing is speculative scope;
   flag it against the exclusions.
3. **Seam audit.** For each cell: responsibility, boundary, contract, and
   budget all present; contract expressible in the notation standard; the
   Parnas test holds across each seam. Flag any pair of cells that would
   change for the same reason (merge candidates) and any cell with two
   unrelated reasons to change (split candidates).
4. **Claim classification.** Verify each claim's cell-local vs. integration
   label and the cells it routes through. If most claims are integration-only,
   report the decomposition itself as the defect.
5. **Derivability spot-check.** Sample decisions the spec does answer: would
   two capable readers have derived the same answer without it? If yes, it is
   span — recommend elision. Then sample what it does not answer: would two
   readers diverge somewhere that matters at a seam, claim, budget, or
   exclusion? If yes, it is a spanning failure — recommend the missing basis
   vector or a Deferred Decisions entry.
6. **Pump hygiene.** Confirm amendments since the last review arrived via
   spec-defect escalation or the catalytic converter, and that converter
   additions are behavioral claims, not opinions or rationale prose.

## Reporting

Report defects with location, the pass that caught each, and a proposed
amendment. Do not silently fix: amendments go through HITL. For material
specs, offer an independent pass via /request-adversarial-review; use
/consensus if findings need adjudication.
