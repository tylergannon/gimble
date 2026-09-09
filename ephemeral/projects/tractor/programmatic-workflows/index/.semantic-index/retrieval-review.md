# Bounded semantic retrieval review

This is a single-model, eight-query smoke check of the routing tree. It follows
`index/README.md` -> one topic route -> one leaf (or the explicitly routed leaf)
-> cited local source ranges. It assesses source match, not merely whether the
expected leaf exists. There is no unindexed baseline and no timing-savings
claim.

## Results

### 1. Does Tractor already provide arbitrary schema-validated results before `next`/`notes` decoding?

- Route: `routes/types-and-embedding.md` -> `leaves/static-tractor.md`.
- Files read for this question: `corpus/static/tractor/harness/backend.go.txt`, `corpus/static/tractor/harness/result.go.txt`.
- Exact source anchors: `corpus/static/tractor/harness/backend.go.txt:86-152`; `corpus/static/tractor/harness/result.go.txt:12-90`.
- Answer: Yes. `HarnessBackend.RunResult` validates the caller-supplied output schema, runs the adapter, validates the returned JSON object, and returns the generic `Result`; `Run` then calls `decodeOutcome`. The generic validated result therefore exists before pipeline-specific `{next, notes}` decoding.
- Retrieval: **succeeded**. The cited implementation directly matches both the ordering and the arbitrary-schema claim.

### 2. What does Temporal demand beyond ordinary Go control flow?

- Route: `routes/authoring.md` -> `leaves/durable-temporal-go.md`.
- Files read for this question: `corpus/durable/temporal-go/workflow.go.txt`.
- Exact source anchors: `corpus/durable/temporal-go/workflow.go.txt:211-309`; `:416-456`; `:526-536`.
- Answer: Temporal’s code-shaped workflow surface carries a deterministic-replay contract. `ExecuteActivity` schedules work and returns a `Future`; local-activity results are serialized into history so replay can avoid rerunning them; `SideEffect` records nondeterministic values for replay; and incompatible workflow changes require version markers. The source supports these dedicated boundaries, not unconstrained ordinary Go effects.
- Retrieval: **succeeded**. The local source directly states scheduling, history serialization, replay behavior, and code-version constraints.

### 3. What is Genkit Go’s typed flow boundary, and what does it not establish about model output?

- Route: `routes/types-and-embedding.md` -> `leaves/imperative-genkit-go.md`.
- Files read for this question: `corpus/imperative/genkit/go/genkit/genkit.go.txt`, `corpus/imperative/genkit/go/core/flow.go.txt`, `corpus/imperative/genkit/go/core/action.go.txt`.
- Exact source anchors: `corpus/imperative/genkit/go/genkit/genkit.go.txt:465-489`; `corpus/imperative/genkit/go/core/flow.go.txt:29-59,75-128`; `corpus/imperative/genkit/go/core/action.go.txt:43-95,240-336`.
- Answer: `DefineFlow[In, Out]` wraps a `func(context.Context, In) (Out, error)` as a typed flow. A generic `Action` resolves input/output JSON schemas, validates the Go input, invokes the typed function, validates the returned Go value, and returns typed `Out`. That establishes a typed action/function boundary; it does not establish that a model emitted schema-conforming wire output. The cited validation happens around `fn`, and no provider model-wire guarantee is shown.
- Retrieval: **succeeded**. The source supports the boundary and its limit; the latter is explicitly a scope conclusion from what the cited code validates.

### 4. Which exact source shows an effect completed before checkpoint write?

- Route: `routes/recovery.md` -> `leaves/durable-dbos-go.md`.
- Files read for this question: `corpus/durable/dbos-go/workflow.go.txt`.
- Exact source anchors: `corpus/durable/dbos-go/workflow.go.txt:2623-2689`; `:2692-2694`.
- Answer: DBOS first invokes `fn(stepCtx)`, then serializes its output, then calls `RecordOperationResult`; the source therefore exposes the crash window in which an external effect inside `fn` can complete before its checkpoint is recorded. `runAsTxn` is the narrower exception where the step body and checkpoint share one database transaction. This is source ordering, not an observed crash run, and it does not make an external HTTP/Git/CI effect atomic.
- Retrieval: **succeeded**. The exact callback/serialize/record sequence and the transactional exception are present in the cited local source.

### 5. Could code preview show both `if` branches before model outputs arrive, and what does Go CFG omit?

- Route: `routes/pre-run-visibility.md` -> `leaves/static-concepts.md`.
- Files read for this question: `corpus/concepts/golang--tools/go/cfg/cfg.go.txt`.
- Exact source anchors: `corpus/concepts/golang--tools/go/cfg/cfg.go.txt:5-14`; `:35-41`; `:220-260`.
- Answer: Yes, a structural preview can show both lexical branches and their successor structure before runtime/model values exist; `go/cfg` can format blocks and emit DOT. Its source says the CFG omits the control statements themselves while retaining their subexpressions/relationships, and does not record conditional-edge conditions, `&&`/`||` short-circuit semantics, or panic control flow. It therefore cannot by itself recover Tractor route predicates or orchestration meaning.
- Retrieval: **succeeded**. The source directly supports the previewable structural facts and the omissions.

### 6. What does current Tractor resume reconstruct from durable data?

- Route: `routes/recovery.md` -> `leaves/static-tractor.md`.
- Files read for this question: `corpus/static/tractor/engine/state.go.txt`, `corpus/static/tractor/engine/runner.go.txt`, `corpus/static/tractor/engine/store.go.txt`.
- Exact source anchors: `corpus/static/tractor/engine/state.go.txt:12-25`; `corpus/static/tractor/engine/store.go.txt:34-115`; `corpus/static/tractor/engine/runner.go.txt:253-317`; `:630-675`.
- Answer: Resume loads checkpointed current/next node, completed nodes, visit/attempt counters, retry/stage/response fields, and session bindings; it resumes at `NextNode` after validation. Because loop frames are in-memory, it rebuilds enclosing frames from durable checklist ledgers, rewinds to the innermost relevant loop, selects its first open item, and records `ResumeRewound`. This is state reconstruction with possible stage replay, not serialized arbitrary execution continuation.
- Retrieval: **succeeded**. The cited checkpoint, load/replace, resume, and ledger-rebuild code matches the requested recovery behavior.

### 7. How does a declared graph enable a faithful pre-run diagram in Pydantic Graph?

- Route: `routes/pre-run-visibility.md` -> `leaves/imperative-pydantic-graph.md`.
- Files read for this question: `corpus/imperative/pydantic-ai/graph.md`, `corpus/imperative/pydantic-ai/pydantic_graph/graph_builder.py`.
- Exact source anchors: `corpus/imperative/pydantic-ai/graph.md:47-61`; `:113-186`; `corpus/imperative/pydantic-ai/pydantic_graph/graph_builder.py:363-386`.
- Answer: `GraphBuilder` explicitly registers nodes and edges; node `run` return annotations determine outgoing edges; `build()` materializes the executable graph. The renderer then passes the materialized `nodes` and `edges_by_source` to Mermaid. The diagram is faithful to that declared graph, while the cited API makes no claim to discover arbitrary Python control flow.
- Retrieval: **succeeded**. The declaration, materialization, and renderer inputs are all source-backed.

### 8. Is semantic indexing here an embeddings requirement or a routing-tree process?

- Route: `routes/method-and-status.md` -> `leaves/method-semantic-index.md`.
- Files read for this question: `corpus/method/df-semantic-index-SKILL.md`.
- Exact source anchors: `corpus/method/df-semantic-index-SKILL.md:9-14`; `:36-46`; `:218-257`.
- Answer: It is a routing-tree process over a local POSIX token cache. The required surfaces are an entrypoint, routing nodes, citation-bearing leaves, housekeeping state, and retrieval evals; Markdown is an accepted format. The method does not require an embedding model, vector store, or similarity service. The benchmark section defines how to measure retrieval, but this smoke check is not such a comparative benchmark.
- Retrieval: **succeeded**. The local method source directly answers the storage/indexing question and its evaluation boundary.

## Deduplicated files read

Entrypoint and routes:

- `index/README.md`
- `index/routes/authoring.md`
- `index/routes/types-and-embedding.md`
- `index/routes/pre-run-visibility.md`
- `index/routes/recovery.md`
- `index/routes/guarantees.md`
- `index/routes/method-and-status.md`

Leaves:

- `index/leaves/static-tractor.md`
- `index/leaves/durable-temporal-go.md`
- `index/leaves/imperative-genkit-go.md`
- `index/leaves/durable-dbos-go.md`
- `index/leaves/static-concepts.md`
- `index/leaves/imperative-pydantic-graph.md`
- `index/leaves/method-semantic-index.md`

Local source files:

- `corpus/static/tractor/harness/backend.go.txt`
- `corpus/static/tractor/harness/result.go.txt`
- `corpus/static/tractor/engine/state.go.txt`
- `corpus/static/tractor/engine/runner.go.txt`
- `corpus/static/tractor/engine/store.go.txt`
- `corpus/durable/temporal-go/workflow.go.txt`
- `corpus/imperative/genkit/go/genkit/genkit.go.txt`
- `corpus/imperative/genkit/go/core/flow.go.txt`
- `corpus/imperative/genkit/go/core/action.go.txt`
- `corpus/durable/dbos-go/workflow.go.txt`
- `corpus/concepts/golang--tools/go/cfg/cfg.go.txt`
- `corpus/imperative/pydantic-ai/graph.md`
- `corpus/imperative/pydantic-ai/pydantic_graph/graph_builder.py`
- `corpus/method/df-semantic-index-SKILL.md`

All eight routes reached a supporting local source citation. No unindexed
comparison was performed, and no timing improvement is inferred.

## Integration note on evaluation limits

The reader reported opening `.semantic-index/evals.jsonl` during directory inspection after retrieval, despite instructions not to consult expected targets. It reported not using those contents for its answers. This is therefore documented as a source-match smoke check, not a blinded retrieval experiment. The additional file access is excluded from no efficiency calculation: none was performed. Root separately checked source-pointer existence/ranges and corpus hashes; those checks do not measure semantic relevance.
