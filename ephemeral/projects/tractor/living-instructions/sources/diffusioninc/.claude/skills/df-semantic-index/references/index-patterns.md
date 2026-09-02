# Semantic Index Patterns

Use these as patterns, not templates. The right shape is the one that shortens the route from retrieval intent to useful citations.

## Compact Markdown Router

Good for a focused prior-art corpus with tens of sources.

- Root `README.md` explains scope, layout, source caveats, and how to navigate.
- `topics.md` maps alphabetical concepts to source leaves.
- `themes.md` synthesizes cross-source patterns.
- `recipes.md` gives task-first routes such as "if you need X, start here."
- `repos/` and `web/` leaves contain dense per-source annotations with `path:line` citations.

This shape is easy for agents to inspect with a few reads and grep calls. It becomes weak when `topics.md` turns into a catch-all or when route files contain too many unrelated concepts.

## Three-Level Theme/Source Tree

Good for broad research corpora with mixed source families.

- Level 1: `README.md` plus `TAXONOMY.md` define corpus halves, route clusters, and citation conventions.
- Level 2: `themes/NN-topic.md` files collect cross-source claims and route to source leaves.
- Level 3: `sources/<family>/<source>.md` files annotate each document, repo, or cluster.

This shape supports breadth while keeping source leaves stable. It needs housekeeping when a family grows much faster than others, when a theme spans too many unrelated leaves, or when source leaves are indexed but not reachable from taxonomy routes.

## Database Or Hybrid Index

Good for very large corpora or frequent incremental updates.

- Keep a human entrypoint that explains how agents query the index.
- Store routes, leaves, citations, fingerprints, and benchmark results in SQLite, JSONL, or another structured format.
- Add generated markdown summaries for the highest-traffic routes when that reduces tool calls.
- Keep scripts small and deterministic so future agents can inspect or rerun them quickly.

Database-backed indexes still need routing semantics. A table of files without task, theme, source-family, or citation fields is an inventory, not a semantic index.
