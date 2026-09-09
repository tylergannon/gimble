# Semantic index: should Tractor workflows become Go programs?

Start here to retrieve evidence for this particular decision. This is a compact human-readable routing tree, not an embedding database and not a generic project catalog. Read [the conversation tie-off](../TIE-OFF.md) for the current direction and user corrections. The [original decision report](../DECISION.md) preserves the initial synthesis; these routes lead to source evidence, not an approved architecture or implementation plan.

## Corpus and citation convention

The token cache is `../corpus/` relative to this directory. It contains selected, pinned copies of source, docs, licenses, and metadata, not complete runnable installations. A leaf citation `imperative/orca/README.md:100` resolves under that corpus root; a citation prefixed `corpus/` resolves from this research directory's parent of index/. `.go.txt` files are verbatim Go source with a safe filename; manifests retain the upstream `.go` path. Pinned GitHub URLs permit independent source comparison. URL-only docs are explicitly weaker collection coverage.

Read one route, one relevant leaf, then its cited source. Do not read the entire cache. Root → route → leaf → source normally takes four file reads including this entrypoint. A leaf may answer orientation questions directly, but check source before relying on a claim.

## Pick a question

| Retrieval intent | Next route |
|---|---|
| Ordinary code versus graph builders; nested loops; who tried it | [Authoring](routes/authoring.md) |
| Per-call typed results; fixed next/notes; existing Tractor embedding seam | [Types and embedding](routes/types-and-embedding.md) |
| Seeing loops/conditions before effects; complete graph versus scenario trace | [Pre-run visibility](routes/pre-run-visibility.md) |
| Restart from repository state; durable replay; external-effect ambiguity | [Recovery](routes/recovery.md) |
| What Go checks; validation authority; what keeping the graph buys | [Guarantees](routes/guarantees.md) |
| Freshness, licenses, provenance, collection/indexing method | [Method and status](routes/method-and-status.md) |

## Housekeeping and limits

Built 2026-09-08 by two Terra researchers and one Luna researcher, with root integration and source spot checks. Latest structural metrics and corpus fingerprint: [.semantic-index/state.json](.semantic-index/state.json). Retrieval questions and expected routes: [.semantic-index/evals.jsonl](.semantic-index/evals.jsonl). The benchmark separates link reachability from a lower-model source retrieval exercise; neither proves lower latency than an unindexed search.

Known debt: selective coverage; upstream programs not executed; no comparative agent-authoring experiment; no actual Go workflow preview; no production adoption measurements; some external prose is URL-only. `diffusioninc/skills` had no root LICENSE available in the inspected snapshot, and automatic license classification for the Temporal sample was inconclusive: local research copies are not assumed to grant code-reuse permission. Future refresh must collect separately, update manifests, then rebuild affected leaves; do not silently change pinned source under old citations.
