# Evidence freshness and indexing method

- [Semantic-index method](../leaves/method-semantic-index.md): local token cache, task routes, citation leaves, housekeeping, and honest retrieval checks. Embeddings are not a requirement.
- [Orca](../leaves/imperative-orca.md), [Genkit Go](../leaves/imperative-genkit-go.md), [Pydantic Graph](../leaves/imperative-pydantic-graph.md), [ADK Go](../leaves/imperative-adk-go.md): repository pins, release dates, language distinctions and licenses in corpus/imperative/manifest.json. A Python Genkit release must not be used as the Go release.
- [Temporal](../leaves/durable-temporal-go.md), [Restate](../leaves/durable-restate-go.md), [DBOS](../leaves/durable-dbos-go.md), [Prefect](../leaves/durable-prefect-visualization.md): corpus/durable/status.json and MANIFEST.md contain API-observed status and source hashes. UI/runtime behavior was not exercised.
- [Dagger and LangGraph](../leaves/static-alternatives.md): pinned Go builder and functional-task sources, non-archived repository status, licenses in corpus/static/manifest.json.
- [Tractor](../leaves/static-tractor.md) and [concepts](../leaves/static-concepts.md): pinned local implementation, language-tooling and controller sources. Do not treat prior worktree notes as current proof.

Source manifests, copied licenses, and repository metadata live alongside each family. Missing license classification is an open reference/reuse question, not evidence that a project is abandoned. A recent release indicates current availability, not maturity or demonstrated fit. The index's own structural and retrieval results live under .semantic-index/ at its root.
