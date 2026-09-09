# Static leaf: Tractor graph, runtime, and embedding seam

## Purpose

This leaf is a source-backed map of the current Tractor implementation at commit `07c04ff4c62ff91c625d2e27b6427cd594f67f39`. It separates guarantees supplied by Go types and the parser from graph-policy lint, model/harness contracts, runtime recovery, and the public versus internal embedding boundary.

## Graph shape and nested scope

- `graph.Graph` is a serialized pipeline with name, goal, defaults, start ID, and a heterogeneous `[]Node`; node types are a closed Go interface currently including agent, fan-out, fan-in, command, supervisor, and loop. `corpus/static/tractor/graph/graph.go.txt:16-40,107-112,128-171,249-313`; pinned source: https://github.com/tylergannon/tractor/blob/07c04ff4c62ff91c625d2e27b6427cd594f67f39/graph/graph.go
- Routing is part of each node: ordinary agent/fan-in choices carry `[]Edge{To, Condition}`, commands carry success/error edges, fan-out carries branch roots, and loops carry explicit loop/exit edges. `corpus/static/tractor/graph/graph.go.txt:101-105,128-161,194-211,249-309,369-400`.
- A loop's checklist is optional only when the loop lies in an enclosing loop body; its comments describe a lap as validation of the previous item followed by an evaluator choice. That is the current nested scope contract, not a generic Go lexical scope. `corpus/static/tractor/graph/graph.go.txt:277-313`; `corpus/static/tractor/internal/workflows/chapter-loop.yaml:14-25,31-50`; pinned source: https://github.com/tylergannon/tractor/blob/07c04ff4c62ff91c625d2e27b6427cd594f67f39/internal/workflows/chapter-loop.yaml
- `lint` derives a loop body as the nodes reachable from `edges.loop` without passing through the loop node, and exposes `LoopBodyNodes`, `EnclosingLoops`, `OutermostLoop`, and fan-out ownership queries. `corpus/static/tractor/lint/analysis.go.txt:186-229`; `corpus/static/tractor/lint/topology.go.txt:10-31,33-63,65-92`.
- The built-in YAML demonstrates arbitrary nesting within one graph/run: chapter loop -> sprint loop -> coding/review, with the inner loop omitting `checklist` and returning to the outer loop. `corpus/static/tractor/internal/workflows/chapter-loop.yaml:14-25,31-82`; the execution rules call this “one loop shape at every level, one run.” `corpus/static/tractor/workflow-designer/rules.md:106-123`.

## Parse and lint boundary

- Parsing is stronger than unmarshalling: JSON preflight rejects duplicate members and multiple values, generated JSON/YAML schema validation runs, fan-out object branches are expanded into synthesized agent nodes, duplicate/reserved IDs are rejected, defaults are applied, and service paths are checked. `corpus/static/tractor/graph/parse.go.txt:16-61,116-201,204-225,339-398`.
- `lint.Validator.Validate` runs a fixed sequence of graph-policy checks and then caller rules; `ValidateOrError` blocks only when any diagnostic is error severity. `corpus/static/tractor/lint/lint.go.txt:12-18,24-49,51-88,100-126`; the rule sequence includes reachability, parallel convergence/disjointness, loop body entry/return/exit/checklist/nesting, edge conditions, supervisor cycles, fidelity, thread collisions/harness consistency, visit bounds, and prompts. `corpus/static/tractor/lint/analysis.go.txt:52-91`.
- Loop lint makes the intended scope explicit: body nodes may only be entered from their own body; the body must route back to the loop; body nodes cannot route directly to `success`; an inner loop may omit its checklist only when it is inside an enclosing body; loops in fan-out branches are rejected. `corpus/static/tractor/lint/rules.go.txt:322-442`.
- Parallel lint is semantic graph analysis rather than a compiler check: every branch must converge on one fan-in without terminal/dead regions, branch node sets must be disjoint, thread keys cannot be shared concurrently or cross worktree boundaries, and the fan-in/branch entry boundaries are checked. `corpus/static/tractor/lint/analysis.go.txt:127-183,232-311`; `corpus/static/tractor/lint/rules.go.txt:142-321`.
- Choice conditions are authoring policy. The linter requires non-empty conditions on multi-edge agent/fan-in routes, but it does not evaluate their natural-language truth; the model is later offered the route descriptions. `corpus/static/tractor/lint/rules.go.txt:489-506`; `corpus/static/tractor/engine/codergen.go.txt:222-265`.

## Typed per-call output and route selection

- The provider-neutral `harness.AgentTurn` includes node/role, message parts, exact JSON output schema, resolved model/provider/effort, fidelity/thread key, workdir, run log, and timeout. The semantic `harness.Outcome` is only `{next?, notes}`. `corpus/static/tractor/harness/contract.go.txt:51-83,115-119`.
- `choiceSchema` always requires `notes`; when more than one successor is offered it adds a `next` string enum containing exactly the offered target IDs and sets `additionalProperties:false`. Route descriptions use the edge condition, then target label, then target ID. `corpus/static/tractor/engine/codergen.go.txt:204-248,251-265`.
- Exact response enforcement is in `harness.NewResultValidator`/`Validate`: the schema must be a JSON object, one JSON value is decoded, the result must be an object, and the compiled caller schema validates it without replacement. `corpus/static/tractor/harness/result.go.txt:12-73,75-90`; pinned source: https://github.com/tylergannon/tractor/blob/07c04ff4c62ff91c625d2e27b6427cd594f67f39/harness/result.go
- The public `HarnessBackend.RunResult` already returns the generic, exact-schema `harness.Result` before any pipeline-specific `{next, notes}` decoding; `Run` is the narrower convenience wrapper that calls `decodeOutcome`. `corpus/static/tractor/harness/backend.go.txt:86-103,95-152`; `run-prompt` uses `RunResult` for arbitrary caller schemas and emits the generic object. `corpus/static/tractor/cmd-run_prompt.go.txt:163-190`. The existing test names this boundary explicitly and exercises it. `corpus/static/tractor/harness/backend_test.go.txt:591-640`.
- The agent handler expands `$goal`, prepends the in-memory loop frame, writes `prompt.md`, validates the resolved turn, calls the backend, and writes `response.md` with optional `next` frontmatter and notes. A nil backend simulates completion. `corpus/static/tractor/engine/codergen.go.txt:24-49,51-119,268-279`.
- Runtime route selection is closed over the offered edges: one offered edge can be taken without `next`; multiple require a matching target; an unknown or unoffered target is terminally rejected. Fan-out has a special fan-in successor check. `corpus/static/tractor/engine/runner.go.txt:789-812`.
- This schema guarantees shape and route membership, not truth of an ordinary-language verdict. The rules say validation is engine-owned and checks run before routing onward, but that is a workflow policy encoded in prompts/loop handlers rather than a Go compiler guarantee. `corpus/static/tractor/workflow-designer/rules.md:124-141`; `corpus/static/tractor/internal/workflows/sprint-execute.yaml:21-29,61-77`.

## Public versus internal embedding seam

- The engine exposes a usable public library seam: `NewRunner`/`ResumeRunner`, `RunnerConfig` (including a caller-supplied `Validate` function and backend), `Registry.Register`, `Handler`, `ExecutionScope`, `Stop`, and `Run`. `corpus/static/tractor/engine/runner.go.txt:22-69,103-139,150-196,199-242,253-262`.
- Shipped workflows are a different seam. `internal/workflows` uses `//go:embed *.yaml`, stores a private catalogue, and exposes `List`, `Lookup`, `Names`, and `Read` only inside the parent module's `internal` import boundary. `corpus/static/tractor/internal/workflows/workflows.go.txt:1-13,15-37,62-95`; pinned source: https://github.com/tylergannon/tractor/blob/07c04ff4c62ff91c625d2e27b6427cd594f67f39/internal/workflows/workflows.go
- The CLI imports that internal catalogue, resolves a file first and a built-in name second, tags provenance as `file:<absolute>` or `builtin:<name>`, and exposes `workflows list/show`. `corpus/static/tractor/internal/workflows/workflows.go.txt:86-95`; source URL: https://github.com/tylergannon/tractor/blob/07c04ff4c62ff91c625d2e27b6427cd594f67f39/cmd/tractor/workflows.go (the selective corpus snapshot intentionally focuses on the package and YAML).
- Inference: an external consumer can embed the public engine and supply a `graph.Graph`, validator, handlers, and backend, but cannot import the built-in YAML catalogue from outside the module. Making built-ins a public extension point would be a new API decision; it is not present in this snapshot.

## Events, checkpoints, and recovery

- `Checkpoint` persists timestamp, current/next node, completed nodes, visit/attempt counters, stage sequence, retry flag, last stage/response, and harness session bindings. `corpus/static/tractor/engine/state.go.txt:12-25`; pinned source: https://github.com/tylergannon/tractor/blob/07c04ff4c62ff91c625d2e27b6427cd594f67f39/engine/state.go
- Checkpoints are atomically replaced through a temporary file; timeline events append JSONL with UTC timestamps; stages are numbered directories with `outcome.json` or `error.json` and a `latest/<node>` symlink. `corpus/static/tractor/engine/store.go.txt:34-115,137-200`.
- A normal run writes `PipelineStarted`, an initial checkpoint, stage start/complete/failure/retry events, checkpoints after each node, and terminal pipeline events. `ResumeRunner` loads and validates the checkpoint; resume can return immediately for final pseudo-targets or continue from `NextNode`. `corpus/static/tractor/engine/runner.go.txt:182-196,261-361,414-460,482-509`.
- Loop frame state is intentionally in-memory. On resume, the runner reconstructs enclosing frames from checklist ledgers and may rewind to the innermost enclosing loop, writing `ResumeRewound`; the comment explicitly says the innermost loop selects its first open item. `corpus/static/tractor/engine/runner.go.txt:296-317,630-675`.
- Retry and failure replay are stage-level. A retry allocates a fresh stage and reruns the handler; a retryable harness error is the only automatic retry path, and a failed node is checkpointed with `RetryVisit`. `corpus/static/tractor/engine/runner.go.txt:511-627`; `corpus/static/tractor/engine/state.go.txt:64-82`.
- Therefore current recovery is inexpensive state reconstruction plus possible effect replay, not exact serialized continuation of arbitrary Go or provider execution. Session IDs are persisted as bindings, but the source does not promise provider-side execution replay. This is a source-derived conclusion, not an observed run claim.

## Go compiler versus remaining policy lint

- A Go implementation would get ordinary compile-time type checking, interface satisfaction, and syntax/control-flow structure from the language/toolchain. The existing `graph` package's generated schema and parser currently supply stronger runtime validation than Go alone: closed node types, required route fields, duplicate IDs, defaults, and file policy. `corpus/static/tractor/graph/graph.go.txt:107-112`; `corpus/static/tractor/graph/parse.go.txt:16-61`.
- Go's compiler cannot ensure that a dynamically built graph has a reachable success target, no loop-body escape, convergent/disjoint fan-out, non-colliding thread keys, valid model/harness combinations, checklist evidence, or a validator that actually demonstrates the declared goal. Current Tractor lint explicitly owns these policy checks. `corpus/static/tractor/lint/analysis.go.txt:52-91`; `corpus/static/tractor/lint/rules.go.txt:22-91,142-242,322-442,489-711`.
- A Go CFG can visualize syntax, but a program-to-graph authoring tool must define node boundaries, identify calls with orchestration semantics, preserve runtime-created branches, and attach conditions/promises/artifact scopes. The compiler cannot infer those domain meanings from arbitrary Go. This is an inference supported by the CFG limitations in `corpus/concepts/golang--tools/go/cfg/cfg.go.txt:35-41` and the policy-heavy Tractor linter above.
- Consequently, no source here supports a claim that arbitrary Go can be statically rendered as a complete executable Tractor graph before execution. A narrower design could require an explicit builder/annotation seam and visualize the constructed declarative subset, while keeping ordinary Go for local control flow.

## Themes

- declarative graph contract versus imperative authoring convenience;
- typed route envelopes versus semantic truth and evidence;
- public runtime library versus internal shipped workflow catalogue;
- durable state reconstruction versus exact continuation;
- compiler facts versus domain policy lint and proof.

## Gotchas and counterevidence

- `NewRunner` requires a non-nil `Validate` function and applies it before allocating runtime state; an embedded caller that bypasses the parser can still supply any validator it chooses. `corpus/static/tractor/engine/runner.go.txt:199-219`.
- `graph.Parse` applies defaults and synthesizes fan-out agents, so the authored YAML graph and runtime graph can differ in node count. `corpus/static/tractor/graph/parse.go.txt:116-180`.
- The model can select only the offered route enum, but its notes and the edge-condition interpretation remain untrusted evidence until a checker evaluates the work. The route schema is not a proof mechanism.
- The source snapshot shows no public “build graph from Go source” API and no static visualizer that would make arbitrary program control flow semantically equivalent to the YAML graph.
- The backend seam makes a typed-wrapper/library POC materially narrower than replacing the pipeline contract: arbitrary exact-schema per-call results are already supported by a public backend API. What remains to design is how a Go author declares node boundaries, routes, loop scope, and evidence policy around that backend.

## Retrieval recipes

- For output routing, start at `corpus/static/tractor/engine/codergen.go.txt:222-279`, then `corpus/static/tractor/harness/result.go.txt:14-63` and `corpus/static/tractor/engine/runner.go.txt:789-812`.
- For nested-loop scope, start at `corpus/static/tractor/graph/graph.go.txt:277-313`, `corpus/static/tractor/lint/analysis.go.txt:186-229`, and `corpus/static/tractor/lint/rules.go.txt:322-442`.
- For embeddability/recovery, start at `corpus/static/tractor/engine/runner.go.txt:103-242,482-509`, then `corpus/static/tractor/engine/store.go.txt:51-115` and `:137-221`.

## Status and license

- Snapshot retrieved 2026-09-08 from Tractor commit `07c04ff4c62ff91c625d2e27b6427cd594f67f39`; hashes and the MIT license are in `corpus/static/manifest.json` and `corpus/static/tractor/LICENSE`.
- Findings are source inspection only. No fetched alternative code was executed, and no production code was changed.

## Open questions

- Which explicit Go API should be the semantic authoring seam: a graph builder, typed node constructors, or annotations over a restricted function body?
- Which policy lint must be mandatory before execution, and which can remain warnings when a program dynamically creates work at runtime?
- Should a future visualizer show a static declarative plan, a runtime-expanded plan, or both with uncertainty markers?
