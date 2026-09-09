# Static visualization and program-to-graph research

Date: 2026-09-08. This memo is source inspection only; no fetched code was executed and no production code was changed.

## Decision-relevant conclusion

Tractor already has two of the hard seams a Go-authored POC would need: a public engine library (`NewRunner`, `ResumeRunner`, `Registry`, `RunnerConfig.Validate`) and a public harness backend whose `RunResult` returns arbitrary exact-schema objects before the pipeline-specific `{next, notes}` wrapper. The missing seam is an authoring representation that preserves nested loop scope, route/evidence policy, artifact boundaries, and a truthful pre-run visualization.

The external evidence supports a conditional builder/annotation counterproposal, not a claim that arbitrary Go can be converted into a complete static Tractor graph. The assumption is that a static, lintable topology is required; the sources do not establish that requirement. Dagger shows a typed Go API whose calls build a runtime GraphQL/dependency query. LangGraph shows ordinary control flow plus durable task results, while explicitly documenting that its Functional API cannot be visualized as a graph before runtime. In parallel, the ordinary-Go question can be tested directly over Tractor's existing `RunResult` seam, with runtime-created structure labeled as dynamic.

## Current Tractor mechanism

The current graph is a closed typed node interface with explicit edges, loop edges, branch metadata, and optional nested-loop checklist inheritance. `corpus/static/tractor/graph/graph.go.txt:16-40,101-171,222-313,369-400`. Parsing performs schema validation, duplicate/reserved ID checks, fan-out synthesis, defaults, and service path policy. `corpus/static/tractor/graph/parse.go.txt:16-61,116-201,204-225`.

The linter is where graph semantics live: reachability, terminal success, loop-body entry/return/exit, checklist inheritance, no loops in fan-out, parallel convergence/disjointness, thread/worktree boundaries, conditions, supervisor cycles, model/fidelity/thread policy, and prompt requirements. `corpus/static/tractor/lint/analysis.go.txt:52-91,127-183,186-229`; `corpus/static/tractor/lint/rules.go.txt:22-91,142-242,322-442,489-711`. A Go compiler or generic CFG supplies neither the domain node boundaries nor these promises.

The route contract is already typed per call. `choiceSchema` requires notes and, for multiple successors, a `next` enum restricted to currently offered IDs with `additionalProperties:false`. `corpus/static/tractor/engine/codergen.go.txt:204-265`. Generic exact-schema validation is public and independent of pipeline Outcome decoding: `HarnessBackend.RunResult` validates and returns `harness.Result`, while `Run` decodes the narrower Outcome. `corpus/static/tractor/harness/backend.go.txt:86-152`; `corpus/static/tractor/harness/result.go.txt:14-73`; `corpus/static/tractor/cmd-run_prompt.go.txt:163-190`.

The public embedding seam is therefore closer than a YAML replacement suggests: a consumer can supply a graph, validation function, registry, backend, stop signal, and logs to the public engine. `corpus/static/tractor/engine/runner.go.txt:22-69,103-139,150-242`. Built-in YAML workflows remain an internal embedded catalogue (`//go:embed *.yaml`) and are not an external extension API. `corpus/static/tractor/internal/workflows/workflows.go.txt:1-13,62-95`.

Recovery is already state reconstruction: atomic checkpoints persist current/next node, counters, last response, stage sequence, and session bindings; timeline JSONL and stage artifacts preserve evidence; resume reconstructs loop frames from ledgers and may rewind to the innermost loop. `corpus/static/tractor/engine/state.go.txt:12-25`; `corpus/static/tractor/engine/store.go.txt:51-115,137-221`; `corpus/static/tractor/engine/runner.go.txt:296-317,482-509,630-675`. The proposal can keep this inexpensive recovery shape while deciding whether a Go-authored declared graph should be persisted alongside it.

## Alternative comparison

| Alternative | What source actually demonstrates | Pre-run visualization claim | Fit to Tractor |
| --- | --- | --- | --- |
| Dagger Go SDK | Typed Go functions call a fluent API; querybuilder links selections and serializes them to GraphQL; a client/session/runner executes the dependency query. `corpus/static/dagger/functions.mdx:7-18,30-42`; `corpus/static/dagger/querybuilder.go.txt:15-18,46-82,98-139,175-197`; `corpus/static/dagger/operator_manual.md:36-78` | Honest for the explicit API/query graph; no evidence that arbitrary host-language branches are statically recovered. | Good model for an embeddable typed builder and runtime dependency graph; incomplete model for agent/evidence/loop policy. |
| LangGraph Functional API | `entrypoint` and `task` preserve ordinary `if`/`for`/call control flow with checkpointed task results, retries, interrupts, and serializability constraints. `corpus/static/langgraphjs-functional_api.md:9-23,355-385,428-448` | Official docs explicitly say Functional API visualization is unsupported because the graph is dynamically generated at runtime; Graph API visualization is the static mode. `corpus/static/langgraphjs-functional_api.md:450-457` | Strong evidence for imperative control flow + durable task outcomes, and strong warning to label dynamic visualization as partial/observed rather than complete. |

## Go compiler and CFG boundary

`golang.org/x/tools/go/cfg` can expose basic blocks, successors, implicit returns, and DOT. `corpus/concepts/golang--tools/go/cfg/cfg.go.txt:5-14,52-69,138-185,220-260`. It explicitly omits conditional edge conditions, short-circuit semantics, and panic flow. `corpus/concepts/golang--tools/go/cfg/cfg.go.txt:35-41`. Thus a program visualizer can show syntax and perhaps a declared subset, but policy remains outside compiler facts: what counts as an agent node, what evidence proves a promise, which calls are replayable, and how dynamic fan-out/checklists scope.

Go's compile-time type checking can make typed result wrappers and builder APIs safe. It cannot guarantee a dynamically assembled graph reaches success, respects loop scope, converges fan-out, avoids thread collisions, runs a meaningful validator, or interprets natural-language conditions. Those are current Tractor lint/runtime policies, not language properties. This is an inference grounded in the source, not a claim that Go cannot analyze any control flow.

## Recommendation and falsifiers

If static, lintable topology is required, a small builder POC can make the semantic boundary explicit:

1. Build a typed Go authoring API whose node/loop/branch calls produce the existing `graph.Graph` (or a versioned equivalent) before execution.
2. Reuse the existing `lint.Validator`, `choiceSchema`, `HarnessBackend.RunResult`, and public `engine.Runner` seam rather than inventing a second result or backend contract.
3. Emit a static graph for declared nodes and mark runtime-created/dynamic branches as such; render loop scope and policy lint findings directly.
4. Keep ordinary Go for local preparation and result handling, while requiring orchestration calls/annotations for node boundaries, route conditions, artifact declarations, and checklist/evidence promises.

This is the narrower builder counterproposal. It does not replace the separate experiment of testing ordinary Go routing over `HarnessBackend.RunResult` with a source outline and explicit dynamic/runtime markers.

Falsify this POC if a representative nested loop cannot produce a stable declared graph without executing agent/tool effects, if route/evidence policy cannot be linted before execution, or if the visualizer must silently guess dynamic nodes to look complete. LangGraph's documented Functional API limitation is the warning case; Dagger's querybuilder is the positive case for an explicit builder seam.

## Status evidence and licenses

GitHub API status observed 2026-09-08: Dagger, LangGraph, LangGraphJS, and Tractor had active, non-archived default branches. These are time-sensitive status observations, not quality scores. Snapshot commit/file hashes and license texts are recorded in `corpus/static/manifest.json`; concept corpus commit/status/license metadata is in `corpus/concepts/manifest.json`.

## Unknowns

- No fetched alternative was executed, so runtime rendering, checkpoint behavior, and version compatibility beyond source/docs claims remain unverified here.
- The POC still needs a decision on whether dynamic work is represented as an explicit expansion node, an opaque runtime region, or a separate runtime event stream.
- The human visualization contract is unresolved: static declared plan only, runtime-expanded plan, or both with uncertainty markers. The source evidence supports “both, clearly labeled” but does not make that a user decision.
