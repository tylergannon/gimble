# Static leaf: Dagger Go as a runtime graph builder

## Purpose

Dagger is the strongest comparison for “Go authoring with a graph-shaped runtime,” but its source shows a query-building and container/data dependency model rather than a static analyzer for arbitrary Go control flow.

## Findings

- Dagger presents functions as typed Go code that performs API operations and returns typed outputs; the examples build a chain with `dag.Container().From(...).WithDirectory(...).WithExec(...)`. `corpus/static/dagger/functions.mdx:7-18,30-42,46-87`; pinned docs: https://github.com/dagger/dagger/blob/d8811256e5e901e6760f7011431080fedd61fc47/docs/current_docs/introduction/core-concepts/functions.mdx
- Dagger's Go SDK is an embedding seam for custom applications: the SDK is a client library plus tooling for Dagger Functions, and the API is GraphQL underneath. Generated Go module code includes `dagger.gen.go`, an internal typed client, querybuilder, and telemetry. `corpus/static/dagger/clients-sdk.mdx:8-27,70-97`; pinned docs: https://github.com/dagger/dagger/blob/d8811256e5e901e6760f7011431080fedd61fc47/docs/current_docs/getting-started/api/clients-sdk.mdx
- The selected Go querybuilder makes the distinction concrete: `Select`, `Arg`, `Bind`, and persistent `prev` links construct a selection path; `Build` serializes that path into a GraphQL query and `Execute` sends it. This is a graph/query builder exposed through ordinary Go method calls, not a compiler pass over arbitrary Go. `corpus/static/dagger/querybuilder.go.txt:15-82,84-139,169-201`; pinned source: https://github.com/dagger/dagger/blob/d8811256e5e901e6760f7011431080fedd61fc47/sdk/go/querybuilder/querybuilder.go
- Dagger's operator model is client -> session -> runner; the runner executes container pipelines, and sessions freeze local-directory/source resolution for a session. `corpus/static/dagger/operator_manual.md:8-78`; pinned source: https://github.com/dagger/dagger/blob/d8811256e5e901e6760f7011431080fedd61fc47/core/docs/d7yxc-operator_manual.md
- Official Dagger documentation describes lazy evaluation and optimization: graph/query steps are evaluated only when a leaf is requested or execution is forced. This is useful for runtime dependency planning and parallelization, but it means a “graph” is the dependency query produced by API calls and requested outputs, not a complete static picture of every Go branch. Current docs: https://docs.dagger.io/reference/api/internals/ (consulted 2026-09-08).

## Decision relevance

- Dagger supports the embeddable typed-wrapper portion of the proposal: Go code can expose typed functions and build a runtime dependency graph through a narrow API.
- Dagger does not demonstrate static visualization of arbitrary Go control flow. If static, lintable topology is a requirement, an explicit Tractor builder/annotation API whose calls are the graph is a plausible narrower experiment, with ordinary Go allowed around it; a claim that Dagger proves arbitrary-Go pre-run graph extraction would overstate the source.
- Dagger's graph is principally data/container dependency and lazy execution. Tractor additionally needs model turns, ordinary-language route conditions, nested checklist scopes, validator/evidence policy, stage artifacts, and steering. Those semantics would remain Tractor policy/runtime work.

## Counterargument and gotchas

- Counterargument: Dagger's fluent Go API may feel like “free Go.” The querybuilder source is the rebuttal: graph nodes arise from `Selection` calls and are serialized to GraphQL; the host language's arbitrary branches are not themselves represented unless they choose which builder calls to make.
- Dagger's generated files are part of the module and should be regenerated when API/dependencies change; the docs warn that generated files are not hand-edited. `corpus/static/dagger/clients-sdk.mdx:70-97,201-203`.
- Dagger's operator manual documents SDK/CLI/runner version compatibility and automatic CLI provisioning. `corpus/static/dagger/operator_manual.md:80-93,117-163`; this is operational weight beyond a small embedded Tractor library.

## Status, license, and retrieval

- GitHub API status observed 2026-09-08: `dagger/dagger`, main, not archived, Apache-2.0, pushed `2026-09-08T16:10:02Z`. This is current status evidence, not a quality judgment. Snapshot commit and hashes are in `corpus/static/manifest.json`; license text is `corpus/static/dagger/LICENSE`.
- To compare Go graph builders, start at `corpus/static/dagger/querybuilder.go.txt:15-18,46-75,98-139,175-197`, then read `corpus/static/dagger/clients-sdk.mdx:8-27`.

## Unknowns

- The selective snapshot does not establish whether Dagger exposes a stable pre-execution export suitable for Tractor's human editor; the official docs consulted emphasize API/query execution and runtime telemetry, not a static arbitrary-Go graph export.
- It remains to test whether an explicit Tractor builder can preserve nested loops and dynamic fan-out without forcing all ordinary Go into a DSL; no code was executed in this research lane.
