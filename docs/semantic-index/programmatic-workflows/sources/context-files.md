# Source leaf: filesystem-backed workflow context

## Purpose

Route to the [context design note](../../../../ephemeral/projects/gimble/programmatic-workflows/CONTEXT-FILES.md)
and [filesystem demo](../../../../examples/go-workflows/context/README.md).
Separate Tyler's context-update direction from the agent's proposed timing,
the current deterministic example, and external systems worth inspecting.

## Key concepts

- Tyler's direction: filesystem-backed `SetContext`, small values inline,
  large values indexed, aggregate budget rebalance, and waiting for a coherent
  update: `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-FILES.md:8-21`.
- Agent proposal: setters dirty the view; the next agent boundary coalesces
  updates and prepares one snapshot. No background worker or global pause:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-FILES.md:23-41`.
- Prototype APIs, byte thresholds, immutable values, versioned indices, and
  largest-first spilling:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-FILES.md:43-64`.
- Real filesystem writes, stubbed agents, placeholder routes, and unresolved
  relevance/pinning policy:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-FILES.md:66-71`.
- OpenViking's tiered filesystem and completion barrier; Letta MemFS's context
  projection, recompile boundary, and budget maintenance. Neither is an
  adopted dependency or a demonstrated complete match:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-FILES.md:73-98`.

## Implementation entrypoints

- Context demo with short step instructions and three data-growth stages:
  `examples/go-workflows/context/main.go:15-44`.
- Store creation and immutable-value setter:
  `examples/go-workflows/internal/program/context.go:45-112`.
- Snapshot publication and bounded context projection:
  `examples/go-workflows/internal/program/context.go:114-232`.
- Automatic snapshot before invoking an agent callback:
  `examples/go-workflows/internal/program/runtime.go:37-68`.

## Retrieval recipes

- For "what happens after SetContext?", read the timing proposal and the
  `Codergen` entrypoint. The setter persists data; indexing waits until a
  snapshot is requested, automatically at the next agent call.
- For "why did a small value leave the prompt?", inspect aggregate spilling
  and run `go run ./examples/go-workflows/context` from the repository root.
  The demo prints the delivered prompts and retains its context files.
- For "what semantic-index dependency do we use?", read the limitations and
  research leads. This example uses a deterministic JSON route index only.

## Themes

Filesystem context; aggregate prompt budget; inline versus external data;
coherent snapshots; deferred indexing; immutable source files; agent-entry wait;
plain-text routes; semantic indexing research.
