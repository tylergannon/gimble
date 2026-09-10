# Source leaf: earlier Go concurrency sketches

## Purpose

Recovered original design for critique circles and bake-offs, using ordinary
goroutines/errgroup with synchronous typed operations. These are proposed
algorithms; the recovered POC was sequential.

## Key concepts

- Tyler's pseudocode-first request, ordinary-Go starting point, and the
  2026-09-10 status note (tactics inline, no `workflows.BakeOff`, no
  workspace primitive):
  `ephemeral/projects/gimble/programmatic-workflows/CONCURRENCY-SHAPES.md:1-29`.
- Shared-proposal critique, adjudication, revision, and competing-draft variant:
  `ephemeral/projects/gimble/programmatic-workflows/CONCURRENCY-SHAPES.md:31-82`.
- Isolated bake-off trials, failed candidates as data, joining, and explicit
  selection separate from integration:
  `ephemeral/projects/gimble/programmatic-workflows/CONCURRENCY-SHAPES.md:84-123`.
- Withdrawn "supplies underneath" list (scoped context, observation scopes,
  workspace isolation) and the old POC's cancellation limitations:
  `ephemeral/projects/gimble/programmatic-workflows/CONCURRENCY-SHAPES.md:125-144`.
- Historical comparison to Temporal/Genkit and the proposed next implementation:
  `ephemeral/projects/gimble/programmatic-workflows/CONCURRENCY-SHAPES.md:146-180`.

## Retrieval recipes

- For “do we need fan-out/fan-in or Future primitives?”, read lines 6-29,
  then the concrete algorithms at lines 31-123.
- Do not implement lines 125-144; they are withdrawn. The rule is AGENTS.md,
  "No wrappers".
- For provenance and current status, use [POC recovery](poc-recovery.md).
