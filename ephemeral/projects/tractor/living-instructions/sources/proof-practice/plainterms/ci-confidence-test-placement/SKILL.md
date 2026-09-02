---
name: ci-confidence-test-placement
description: >
  Decide where changed-behavior proof belongs. Use before implementation,
  before test deletion, or before E2E assertion relocation to choose unit,
  Storybook/Vitest, E2E, or manual proof without editing files.
---

# CI Confidence Test Placement

Use this for the coverage-design request of the CI Confidence Pass. Apply the
stance and layer-ownership rules in
`.agents/workspace-agents/ci-confidence-engineer/CHARTER.md`. Name the
behavioral invariants that matter, put each one at the cheapest layer that still
proves it, and say when no new test is needed. Prefer unit/Vitest for pure logic,
Storybook/Vitest for rendered component states and local interactions, and E2E
only for browser plus app plus data or external-service boundaries. Return a
small ledger of invariant, risk, best layer, existing coverage, and
recommendation; do not edit files.
