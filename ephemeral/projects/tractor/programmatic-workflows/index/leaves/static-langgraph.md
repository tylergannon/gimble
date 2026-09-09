# Static leaf: LangGraph Functional API and the visualization boundary

## Purpose

LangGraph is a direct comparison for ordinary language control flow with durable task results and human-in-the-loop behavior. Its own documentation states the key trade: the Functional API accepts normal control flow, but does not support static graph visualization because the graph is generated at runtime.

## Findings

- The Functional API wraps ordinary functions with `entrypoint` and discrete `task` primitives, preserving `if`, `for`, and function-call control flow instead of requiring an explicit DAG. `corpus/static/langgraphjs-functional_api.md:1-23`; pinned docs: https://github.com/langchain-ai/langgraphjs/blob/bbbdb5aa8a50f7115bdfbb6e3cf020ee239e1842/docs/docs/concepts/functional_api.md
- The implementation's task wrapper is still a semantic boundary: `task` stores a callable, retry/cache/timeout policies, and returns a future-like object when called; the docs require task inputs/outputs to be serializable when checkpointing is enabled. `corpus/static/langgraph/functional.py:59-94,132-169`; `corpus/static/langgraphjs-functional_api.md:355-385,428-448`.
- LangGraph's Functional API persists task results and resumes by reusing saved results; its docs advise encapsulating side effects and randomness in tasks and designing effects to be idempotent because a task may run again if it started but did not complete. `corpus/static/langgraphjs-functional_api.md:61-64,414-448`.
- The documented comparison is explicit: Graph API checkpoints every superstep and is easy to visualize; Functional API saves task results to an entrypoint checkpoint and “does not support visualization as the graph is dynamically generated during runtime.” `corpus/static/langgraphjs-functional_api.md:450-457`; this is the most decision-relevant evidence against promising a complete pre-run picture for free-form control flow.
- The separate graph renderer constructs a `Graph` by simulating Pregel task/channel progression, collecting static and conditional writes, trigger sources, and edges up to a configurable limit. `corpus/static/langgraph/draw.py:42-64,65-115,117-170,171-218`; pinned source: https://github.com/langchain-ai/langgraph/blob/81bf17b23123e4ef8b9d5f49fa09a0122fc2edd1/libs/langgraph/langgraph/pregel/_draw.py
- This means LangGraph offers two distinct authoring modes: explicit Graph API for a graph artifact and Functional API for ordinary control flow plus runtime durability. The graph renderer does not erase the Functional API's documented visualization limitation.

## Decision relevance

- LangGraph validates the proposal's product intuition that users may prefer ordinary control flow, durable typed task results, retries, and human interrupts.
- It also supplies a strong counterexample to the premise that ordinary control flow can be faithfully visualized before execution: the official docs explicitly decline Functional API visualization because the graph is dynamic at runtime.
- If static, lintable topology is a requirement, a Tractor experiment could follow the same split: (a) a builder/graph mode with honest static visualization, and (b) an imperative mode whose visualization is an abstract declared skeleton plus runtime expansion/events. The UI must label uncertainty rather than present a complete graph it cannot derive.

## Counterargument and gotchas

- Counterargument: LangGraph's renderer can draw a graph. Source inspection shows `draw_graph` simulates runtime writes and task progression; it is meaningful for the Graph/Pregel runtime, not evidence that arbitrary Functional API source has a complete static graph.
- Checkpointing depends on JSON-serializable entrypoint/task values and task boundaries. `corpus/static/langgraphjs-functional_api.md:158-169,355-385,428-448`.
- Resumption reuses completed task results, but incomplete tasks may be re-executed; idempotency remains an application responsibility. `corpus/static/langgraphjs-functional_api.md:440-448`.

## Status, license, and retrieval

- GitHub API status observed 2026-09-08: `langchain-ai/langgraph`, main, not archived, MIT, pushed `2026-09-08T15:21:12Z`; `langgraphjs` main is also MIT/not archived, pushed `2026-09-06T02:16:03Z`. Status is time-sensitive evidence, not a recommendation. Snapshot commits/hashes/licenses are in `corpus/static/manifest.json` and adjacent license files.
- For “what can imperative code honestly show?”, start at `corpus/static/langgraphjs-functional_api.md:450-457`, then inspect `corpus/static/langgraph/draw.py` if a renderer mechanism is needed.

## Unknowns

- The snapshot does not establish how much of a specific Functional API run can be rendered incrementally without executing task effects; that would require a runnable experiment, which this research brief forbids.
- It remains open whether Tractor should expose a static builder plus an imperative task API, or one API with an explicit “declared graph” subset and runtime dynamic expansion.
