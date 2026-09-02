---
name: ci-confidence-test-maintenance
description: >
  Maintain the actual test portfolio. Use after implementation, during E2E
  refactors, or when tests, stories, fixtures, or helpers need review, deletion,
  relocation, or focused edits.
---

# CI Confidence Test Maintenance

Use this for the test-repair request of the CI Confidence Pass. Apply the
stance and layer-ownership rules in
`.agents/workspace-agents/ci-confidence-engineer/CHARTER.md`. Edit only tests,
Storybook stories, fixtures, and helpers unless the parent agent explicitly
expands the scope. Keep or add coverage only when it protects a named invariant
at the cheapest trustworthy layer. Delete, narrow, or move down tests that are
duplicate, flaky, tautological, snapshot-heavy, fixture echoes, implementation
mirrors, or E2E checks of component-only state. Return the verdict, changed
files, invariants covered, tests intentionally not added, proof run, and
remaining gaps.

If this pass is delegated to a Claude-family subagent through `agent opus`,
`agent sonnet`, `agent haiku`, or `agent fable`, let that subagent finish by
default. Do not cancel it merely because it is slow or the parent agent is ready
to proceed. Cancel only if the user explicitly asks, the process is in a clear
non-progressing failure loop, or continuing would cause a concrete safety or
tooling problem.
