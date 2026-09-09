# What the compiler replaces, and what remains

- [Tractor](../leaves/static-tractor.md) maps parser/schema checks, loop-body reachability, parallel ownership, thread conflicts, route enums, runtime checks, and public/internal boundaries. Read it for the difference between Go lexical scope and the engine's completion authority.
- [Go CFG concepts](../leaves/static-concepts.md) explains why syntactic structure is useful but not a proof of application intent or effect ownership.
- [Pydantic Graph](../leaves/imperative-pydantic-graph.md) is the strongest compact counterexample: code can retain a materialized graph and a faithful diagram, at the cost of explicit topology.
- [Orca](../leaves/imperative-orca.md) has type-level capability distinctions around shared-Git mutation; do not import its whole checkpoint runtime merely to express a loop.
- [ADK Go](../leaves/imperative-adk-go.md) illustrates why a schema option is weaker than validated, decoded output.

A Go program gets variable scope, ordinary types and syntax checking. It does not automatically get exhaustive domain switches, fresh validation, independent reviewers, safe concurrency, or no early success. A graph's route membership also does not make a model's judgment true. See [Pre-run visibility](pre-run-visibility.md) for static inspection and [Recovery](recovery.md) for stale-state authority.
