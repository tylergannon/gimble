# Imperative orchestration and typed agent outputs: bounded research synthesis

Scope: source-only research for the 2026-09-08 programmatic-workflow brief. Four repositories were shallow-cloned, selectively copied, and pinned in `corpus/imperative/manifest.json`; no fetched code ran. Read the semantic leaves first: `index/leaves/imperative-orca.md`, `index/leaves/imperative-genkit-go.md`, `index/leaves/imperative-pydantic-graph.md`, and `index/leaves/imperative-adk-go.md`.

## What the sources actually support

The author's desired separation is viable in source terms: ordinary Go can own nesting, loops, conditionals, routing, and lexical scoping, while a small runtime library owns named agent calls, typed decode/validation, events, and validation invocation. Genkit Go is direct evidence of the first half: `DefineFlow` registers a generic Go callback and `Run` introduces named trace steps inside that callback, with no explicit author-facing edge graph. `corpus/imperative/genkit/go/genkit/genkit.go.txt:465-489`, `corpus/imperative/genkit/go/core/flow.go.txt:75-128`.

Typed output is a separate design choice from imperative control flow. Orca's `resultAs[O]` is the strongest copied pattern: derive a schema before the call, pass it to the backend, parse locally, and use a corrective retry for malformed autonomous output. `corpus/imperative/orca/tools/AgentCall.scala:109-147`, `corpus/imperative/orca/tools/AgentCall.scala:189-295`. Genkit validates typed Go action input/output but does not establish model-wire decoding; ADK Go configures an output schema but its current output-state path preserves raw concatenated text and carries an explicit validation/unmarshal TODO. `corpus/imperative/genkit/go/core/action.go.txt:240-336`, `corpus/imperative/adk-go/agent/llmagent/llmagent.go.txt:520-559`.

The proposed lightweight recovery contract is also independent. Orca proves that robust skip/replay needs much more than a loop: a stable hierarchical stage ID, saved typed JSON, branch/prompt binding, commit/log atomicity, capability constraints, and leftovers/idempotence rules. `corpus/imperative/orca/adr/0018-stage-bound-flow-runtime.md:61-166`, `corpus/imperative/orca/flow/Flow.scala:40-142`. That is evidence for retaining a *fresh validation on restart* contract rather than accidentally reimplementing durable computation. It does not say saved typed stage results are required when the product intentionally re-derives a plan from repository state.

Pre-execution visualization is the constraint that changes the answer. Pydantic Graph can render Mermaid because authors explicitly construct typed nodes and edges; ADK renders DOT because it knows a declared agent object tree. `corpus/imperative/pydantic-ai/pydantic_graph/graph_builder.py:363-386`, `corpus/imperative/adk-go/server/adkrest/internal/services/agentgraphgenerator.go.txt:273-340`. Genkit Go offers trace/reflection metadata, not a static Go-flow renderer. A Go source preview can show lexical loops, both branch bodies, condition text, statically resolved calls, and statically resolvable closures. Unknown runtime values prevent prediction of the chosen path or concrete fan-out instances; unresolved dynamic dispatch, reflection, and unresolvable calls should be labeled uncertain. That is still different from an exact execution graph, without making a separate full declared model the only route to a useful preview.

## Candidate narrow POC

Build a package-local experiment around a single canonical loop, without changing production workflows:

```go
type Review struct { Ready bool `json:"ready"`; Reason string `json:"reason"` }

review, err := Call[Task, Review](ctx, reviewer, task, WithSchema[Review]())
if err != nil { return err }
if review.Ready { return Validate(ctx, goal) }
```

`Call` must require a Go result type, generate/accept its JSON schema, request structured output from the adapter, locally decode and validate the raw payload, retain raw evidence, and return a classified schema/decode error. `Validate` remains declarative and fresh on every restart. The example expresses the recommended boundary; it is not claimed to compile against an existing Tractor API.

The POC's preview should operate on source/AST or explicit runtime registrations before execution and output a conservative Mermaid/DOT document. It should distinguish `call`, `validate`, `loop`, `branch`, source conditions, statically resolved closure calls, and unresolved dynamic regions; it must not manufacture concrete future fan-out instances or claim the runtime branch outcome. This only tests extraction cost and usefulness, not production migration.

## Falsifiers and risks

- Stop if a useful preview of the canonical loop requires whole-program analysis, reflection, or annotations duplicating most workflow policy. That means the extraction cost defeats the stated authoring benefit.
- Stop if every actual agent adapter cannot supply schema-constrained output plus raw-payload capture and local decoding. Plain prompt JSON is not a typed routing guarantee.
- Stop if fresh validation cannot determine success/failure from repository state without replaying expensive or destructive phases. At that point introduce a narrowly scoped checkpoint artifact, rather than a general runtime.
- Stop if nesting needs concurrent workspace mutation or commits without a deterministic ownership rule. Orca demonstrates why a shared Git index turns this into runtime machinery, not an ordinary `go` statement.
- Do not infer runtime proof from these sources. The corpus establishes maintained status and design/code feasibility at pinned revisions; it does not demonstrate Tractor binary behavior, native-agent adapter conformance, restart behavior, or preview correctness.

## Recommendation (agent proposal, not a user decision)

Authorize a small POC only if it measures both unknowns: (1) a typed per-call output wrapper against a real Tractor adapter and (2) a deliberately conservative preview of the one canonical Go loop. Keep declarative goal/validator inputs and fresh validation. Do not port Orca's persistence/session/commit runtime or Pydantic's explicit graph model unless the POC falsifies weak restart or source preview. The main architectural seam is `Call[In,Out]` plus effect/evidence events, not a replacement graph executor.
