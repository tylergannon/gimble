---
name: cell-lifecycle
description: >
  Govern cells in a live nlspec-managed codebase: contract-guarded
  maintenance, condemn triggers, the harvest-first rebuild protocol, and the
  catalytic converter that elevates tests into claims. Use for maintenance
  work, drift response, or deciding whether a cell should be rebuilt.
---

# Cell Lifecycle

Load the spec-writing standard first. A cell is in exactly one mode:
maintenance (common) or rebuild (rare, deliberate). There is no per-edit
regeneration of specs, and no gap analysis for routine work.

## Maintenance mode

Ordinary contract-guarded agent work. The instruction shape is: "fix this; do
not touch the contract; if the fix forces a contract re-evaluation, stop and
report." The nlspec is a guardrail here, not a build input. The two-pump rule
applies: maintenance never amends the nlspec directly. Maintenance does not
waive the security carve-out: cells handling credentials, authentication or
authorization decisions, irreversible or destructive operations, or data
migrations are human-reviewed regardless of cell locality (see slice-design's
review scope).

## Condemn triggers

The methodology runs on replace economics: budget-bounded cells are cheap to
rebuild from their contracts, so drift needs triggers, not baselines. Condemn
a cell when any of these fires:

- **budget breach** — the primary tripwire; degraded internals show up as
  volume first (duplication, defensive scaffolding, flag soup);
- **contract-check failure** — actual surface no longer matches the stated
  contract, or something reaches past a contract into internals;
- **undemonstrable claim** — a claim routed through the cell can no longer be
  demonstrated;
- **seam spray** — a supposedly cell-local change touches other cells;
- **friction** — a change that ought to be cell-local repeatedly fails to be,
  or a fix takes multiple attempts inside one cell. Drift announcing itself
  before a mechanical wire trips.

Check budgets mechanically on every slice; keep them slightly tight so the
wire trips while the rebuild is still cheap.

## Rebuild protocol

The order is fixed:

1. **Harvest.** Run the catalytic converter over the dying cell's tests. A
   condemned cell's tests are the only record of everything its ephemeral
   slice specs never wrote down.
2. **Condemn.** Mark the cell for rebuild; route pending maintenance around
   it.
3. **Rebuild.** Design the replacement with slice-design, from the nlspec,
   conditioned on the surviving elements — a rebuild driven by requirements
   change keeps what is still expected to succeed; it is never a fresh
   stochastic sample.
4. **Demonstrate.** Prove the routed claims per /proof-of-work before the
   rebuild lands.

Archived slice specs from the cell's history are available as forensics; they
are not load-bearing inputs.

## Catalytic converter

Elevates behavioral evidence from code into claims, up to inclusion in the
nlspec — checkable memory of hard-won decisions, without freezing how they
were implemented.

- **Intake:** test cases only. Never opinions, rationale prose, or
  language-specific "lessons" — the moment the converter accepts those it
  becomes a third authority.
- **Elevation test:** does this test encode intent that would otherwise
  survive only by accident — a boundary decision, an invariant, a regression
  a future rebuild must not reintroduce? Routine coverage stays in the code.
- **Mechanics:** propose elevated claims to HITL as nlspec amendments (this is
  pump 2), classify each cell-local or integration, and record its provenance.

Run the converter routinely at slice exit (see slice-design) and mandatorily
at harvest.
