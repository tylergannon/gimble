# Static leaf: alternative source family

## Purpose

This is the compact router for the two selected source-backed alternatives. The detailed findings live in the source-specific leaves so retrieval can target the relevant comparison.

## Findings

- **Dagger Go:** read [`static-dagger.md`](static-dagger.md) for the typed fluent query builder, lazy runtime dependency graph, and the crucial distinction between graph-building calls and arbitrary host-language control flow. Its snapshot sources are `corpus/static/dagger/` and are pinned in `corpus/static/manifest.json`.
- **LangGraph Functional API:** read [`static-langgraph.md`](static-langgraph.md) for ordinary control flow plus explicit durable task boundaries, runtime-derived rendering, and the official limitation on complete pre-run visualization. Its snapshot sources are `corpus/static/langgraph/` and `corpus/static/langgraphjs-functional_api.md`.
- Together, these alternatives frame a conditional design choice: if static lintable topology is a product requirement, an explicit builder or annotation boundary is a plausible narrower experiment; if ordinary Go routing is the question, test it directly against Tractor's existing public `HarnessBackend.RunResult` seam and mark any runtime-created structure as observed/dynamic.

## Status and license

- Dagger is pinned at commit `d8811256e5e901e6760f7011431080fedd61fc47` under Apache-2.0; LangGraph at `81bf17b23123e4ef8b9d5f49fa09a0122fc2edd1` under MIT; LangGraphJS documentation at `bbbdb5aa8a50f7115bdfbb6e3cf020ee239e1842` under MIT. File hashes, URLs, and license snapshots are in `corpus/static/manifest.json`.
- GitHub API status observed 2026-09-08 records active, non-archived default branches; this is time-sensitive metadata, not a quality claim.

## Retrieval recipes

- For a Go graph-builder comparison: `static-dagger.md` -> `corpus/static/dagger/querybuilder.go.txt:15-82,84-139,169-201`.
- For imperative control flow and its visualization boundary: `static-langgraph.md` -> `corpus/static/langgraphjs-functional_api.md:450-457`.
