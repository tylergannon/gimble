# What can humans see before execution?

First distinguish an abstract structural view, a complete domain-level route model, a mocked scenario, and an observed run trace. Neither graph nor Go knows actual future agent outcomes; ordinary syntax can show both branch bodies and loop conditions without knowing them.

- [Dagger and LangGraph comparison](../leaves/static-alternatives.md): follow to Dagger's Go query builder or LangGraph's documented Functional API visualization boundary.
- [Go CFG and reconciliation concepts](../leaves/static-concepts.md): per-function blocks and DOT; condition/panic/short-circuit omissions; AST/source information is needed for a readable workflow view.
- [Pydantic Graph](../leaves/imperative-pydantic-graph.md): Mermaid from a materialized graph, not arbitrary Python discovery.
- [ADK Go](../leaves/imperative-adk-go.md): DOT from recognized configured agent composition.
- [Genkit Go](../leaves/imperative-genkit-go.md): named traces and reflected action metadata; no static Go-flow renderer established by this source selection.
- [Prefect](../leaves/durable-prefect-visualization.md): visualization executes non-task code and mock values select dynamic paths. This is the clearest caution against calling a scenario trace a complete static plan.
- [Tractor](../leaves/static-tractor.md): graph lint and author-declared structure; current parser can expand authored nodes.

For why these limits matter to policy checking: [Guarantees](guarantees.md). Source-linked hierarchical outlines with explicit opaque calls are a proposed experiment, not a capability demonstrated here.
