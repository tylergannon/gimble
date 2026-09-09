# Go-authored workflow research

**Shape and input:** [Tyler's clarification](INPUT-SEPARATION-VERBATIM.md) makes a distinct separation between program shape and starting task data a high-level principle. Input includes available information and index structure/quality; both can evolve during execution. [The direction document](../../../../docs/direction.md#separate-program-shape-from-its-input) connects this to reuse and separate evaluation of program shapes versus information conditions.

**Latest clarification:** [Agent-assisted diagrams, a small action vocabulary, and research/indexing as companions](ACTIONS-AND-KNOWLEDGE-VERBATIM.md) are now incorporated into [Tractor's direction](../../../../docs/direction.md). Static extraction can anchor a visual explanation that an agent arranges and enriches conceptually. Preparing local, searchable task knowledge is part of the workflow; the copied semantic-index method was briefly revisited, not rebuilt or benchmarked again.

**Broader product direction:** [Tractor's direction](../../../../docs/direction.md) connects this work to the experience of authoring workflows, the legibility to preserve from graphs, Go that reads like pseudocode, and future telemetry for inspecting context and steering. [Tyler's statement](DIRECTION-CLARIFICATION-VERBATIM.md) is preserved verbatim. This extends the direction without expanding the first POC's implementation scope.

**The central claim:** [Workflows as programs](../../../../docs/workflows-as-programs.md) records Tyler's architectural direction: author orchestration in Go and organize work and information together around agent success. Goals, scoped state, the semantic index, and on-disk material form one context; workflows distribute attention and responsibility across agents and time. [The original statement](KEY-CLAIM-VERBATIM.md) and [the later context-engineering clarification](CONTEXT-REFRAMING-VERBATIM.md) are preserved verbatim. The clarification supersedes the earlier framing of files as context overflow and the program metaphor as the governing objective.

**A narrow Go POC was implemented in the source experimental branch and local commits.** See [the runnable guide and live results](POC.md) for its `program` CLI, checklist iterator, the three Go workflow counterparts, and the limits of this slice. This documentation PR preserves its record but does not add that runtime to `main`. The notes below preserve the preceding research journey.

**Start with [the conversation tie-off](TIE-OFF.md).** It captures the latest user corrections, decisions versus proposals, the full journey, source findings, unresolved questions, and a continuation handoff. Current direction: program-owned declarative JSON schemas; ordinary Go control flow and typed agent results; runtime observer visibility; comparative evals across orchestration shapes; useful static projections for supported structures. Naming and linter tricks are deferred.

[The original decision report](DECISION.md) and [the subsequent reconsideration](research/established-shapes.md) preserve earlier thinking. Neither is the current implementation plan. The latter overemphasized a maintained routine catalog and a linter-centered first experiment; the tie-off corrects that.

- [Semantic index](index/README.md): question-first routes to source evidence and counterarguments.
- [Earlier POC proposal](research/poc-contract.md): superseded as a starting point. Tyler's subsequent implementation request and the actual narrow slice are recorded in [POC.md](POC.md).
- [Source collection](corpus/): selected pinned source/docs, copied licenses where available, and source manifests. Upstream programs were not run.
- [Researcher memos](research/): imperative, durable, and static-analysis lanes; the conditional graph-builder counterproposal is retained.
- [Index audit](index/.semantic-index/state.json): corpus integrity, local citation ranges, link reachability, and known limits.
- [Retrieval smoke check](index/.semantic-index/retrieval-review.md): eight source-backed questions, with no baseline or efficiency claim.

Research source head for Tractor: `07c04ff4c62ff91c625d2e27b6427cd594f67f39`. Collected 2026-09-08. The tie-off distinguishes Tyler's stated directions from unratified agent proposals. The later POC introduces an experimental public API; it does not establish a migration or final naming scheme.

To refresh the local structural audit after changing index material:

```sh
python3 ephemeral/projects/tractor/programmatic-workflows/research/audit-index.py
```

The audit reads the corpus; it never executes it. Copied Go source uses `.go.txt` so ordinary parent-repository package discovery cannot incorporate upstream examples.
