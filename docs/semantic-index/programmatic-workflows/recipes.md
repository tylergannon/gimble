# Retrieval recipes

Start with the route that matches the question, then follow that narrow topic
node to its source leaves.

## Show actual Go programs that can be compiled and run

- Start with [the stubbed source examples](sources/compiling-examples.md) for
  bake-off, critique circle, and sprint execution. Agent and command replies
  are canned; workspaces and integration are symbolic.
- For the prior native implementation, follow
  [go-library](routes/go-library/index.md) to the recovered POC.

## Find the existing Go POC before designing an API

- Follow [go-library](routes/go-library/index.md) for the preserved branch,
  actual iterator-based workflows, and the earlier concurrency sketches.

## Express a bake-off or critique circle in ordinary Go

- Follow [go-library](routes/go-library/index.md) first to the new `errgroup`
  source examples, then to the earlier design sketches. The examples make
  Go control flow executable while leaving native operations stubbed.

## Find argument schemas, roles, and the curated builtin calling convention

- Follow [program-input](routes/program-input/index.md) and
  [go-library](routes/go-library/index.md). The newer roadmap is indexed with
  Tyler's correction: its API indirection needs revision.

## Understand the proposed Go direction

- Follow [go-authorship](routes/go-authorship/index.md) for the core case and the
  original claim/corrections.

## Understand what must be preserved from graphs

- Follow [legibility](routes/legibility/index.md) for the value of DAG progression,
  pseudocode-like source, and derived visual explanation.

## Understand context engineering and the semantic index

- Follow [context-and-indexing](routes/context-and-indexing/index.md) for one context,
  on-disk information, retrieval, and research/indexing.

## Set context, spill large values, and wait for an updated index

- Follow [context-and-indexing](routes/context-and-indexing/index.md) to the
  filesystem design and demo: immutable values, per-value and aggregate
  budgets, and a coherent snapshot before each agent callback.
- Distinguish Tyler's update-trigger direction from the agent's deferred
  indexing proposal. The example uses byte limits and placeholder routes;
  OpenViking and Letta are research leads, not adopted dependencies.

## Separate reusable method from task data

- Follow [program-input](routes/program-input/index.md) for the short source statement
  and its synthesized consequences for reuse and evaluation.

## Understand roles, diagrams, and telemetry

- Follow [roles-and-evaluation](routes/roles-and-evaluation/index.md) for action roles,
  supervision, five-arts tradeoffs, and bad-run telemetry.
