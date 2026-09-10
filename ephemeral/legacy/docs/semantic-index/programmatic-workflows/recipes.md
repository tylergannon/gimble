# Retrieval recipes

For the Go API, start with the [recovered POC workflow source](sources/poc-workflows.md):
`program.Loop`, `Codergen[T]`, `Command`, `Validate`. That is the approved shape and it is
ported into `program/`. The later stub sketches were removed on 2026-09-10.

Start with the route that matches the question, then follow that narrow topic
node to its source leaves.

## Understand the proposed replacement

- Follow [go-authorship](routes/go-authorship/index.md) for the core case and the
  original claim/corrections.

## Understand what must be preserved from graphs

- Follow [legibility](routes/legibility/index.md) for the value of DAG progression,
  pseudocode-like source, and derived visual explanation.

## Understand context engineering and the semantic index

- Follow [context-and-indexing](routes/context-and-indexing/index.md) for one context,
  on-disk information, retrieval, and research/indexing.

## Separate reusable method from task data

- Follow [program-input](routes/program-input/index.md) for the short source statement
  and its synthesized consequences for reuse and evaluation.

## Understand roles, diagrams, and telemetry

- Follow [roles-and-evaluation](routes/roles-and-evaluation/index.md) for action roles,
  supervision, five-arts tradeoffs, and bad-run telemetry.
