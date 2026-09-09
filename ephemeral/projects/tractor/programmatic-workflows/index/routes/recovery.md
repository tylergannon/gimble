# Revalidate and continue versus replay

- For desired-state reconciliation, read [concepts](../leaves/static-concepts.md): current state drives the next action; coding agents still need fresh goal validation and are not guaranteed to converge.
- For what Tractor already saves and reconstructs, read [Tractor](../leaves/static-tractor.md): checkpoints, ledger-derived loop frames, ResumeRewound, session bindings. Current recovery is not full serialization of execution.
- For optional Git-coupled stage checkpoints in an imperative agent app, read [Orca](../leaves/imperative-orca.md): typed saved values, stable IDs, commits, shared-index concurrency limits.
- For deterministic command-history replay and code-version restrictions, read [Temporal](../leaves/durable-temporal-go.md).
- For Go handler routing with journaled context operations, read [Restate](../leaves/durable-restate-go.md).
- For the exact crash-after-effect-before-record window and the narrower atomic database exception, read [DBOS](../leaves/durable-dbos-go.md).

The user accepts useful recovery faster than a clean restart, not identical traces. Distinguish surviving child processes, human decisions, stale passes, and external PR/deploy effects from disposable local planning. Add persistence only for a demonstrated need. See [Types and embedding](types-and-embedding.md) for the existing call boundary.
