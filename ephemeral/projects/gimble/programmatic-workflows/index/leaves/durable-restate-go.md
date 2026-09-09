# Restate Go: regular handler plus journaled context actions

**Purpose.** Restate Go is the closest inspected runtime to normal Go control flow with durable checkpoints: handler code can branch and loop normally, while selected `Context` actions and `Run` results are journaled. It demonstrates a lighter authoring model than Temporal, while still being a durable-execution runtime with its own server.

**Findings.**

- The Go SDK documents `Run` as storing final results (including terminal errors) durably in a journal; transient error causes the invocation to be retried, and replay returns the same value (`durable/restate-go/run.go.txt:12-25`; [pinned source](https://github.com/restatedev/sdk-go/blob/3f7852328e1fc1f5b0716d8a1e556a614d83dc97/run.go#L12-L25)). This is per-call memoization/journaling, not automatic persistence of arbitrary local variables between calls.
- The same API permits the non-deterministic operation inside `Run` but forbids use of the handler `Context`/nested durable actions there (`durable/restate-go/run.go.txt:18-29`; [pinned source](https://github.com/restatedev/sdk-go/blob/3f7852328e1fc1f5b0716d8a1e556a614d83dc97/run.go#L18-L29)). The boundary is explicit: keep durable control actions outside; put non-deterministic external work in `Run`.
- The official parallel example has an ordinary `range` loop that dynamically creates async durable runs, then waits for their results (`durable/restate-go/parallelizework-main.go.txt:24-53`; [pinned source](https://github.com/restatedev/sdk-go/blob/3f7852328e1fc1f5b0716d8a1e556a614d83dc97/examples/parallelizework/main.go#L24-L53)). It is good evidence that Go-defined fan-out can depend on runtime input; it is also direct counterevidence to any promise of a fully known pre-run graph.
- A `RunAsyncFuture.Result` must not be called from a goroutine; the source directs multi-result waits through `Context.Select` (`durable/restate-go/run.go.txt:94-99`; [pinned source](https://github.com/restatedev/sdk-go/blob/3f7852328e1fc1f5b0716d8a1e556a614d83dc97/run.go#L94-L99)). Therefore “normal Go” has material context API restrictions at durable coordination points.
- Restate's official request-lifecycle document says the server journals Context operations and replays them after a handler crash ([lifecycle guide](https://docs.restate.dev/guides/request-lifecycle)); its Go durable-steps guide confirms `Run` is for database/HTTP/UUID nondeterminism ([durable steps](https://docs.restate.dev/develop/go/durable-steps)). These are vendor claims not independently executed in this research.

**Recovery and ambiguity.** Once a `Run` result is durably journaled, retry skips it. The source shows the journal result arrives *after* the callback returns; it does not prove an atomic transaction with an arbitrary remote API. So a crash after “remote system accepted PR/deploy/email” and before the result journal commits is unresolved from this source alone; design the remote request with a stable idempotency key, query the remote state on recovery, or compensate. This is a precise counterexample to zero records only when the external action cannot be safely discovered from repository/provider state.

**Visualization.** Restate documentation advertises an execution UI/trace after invocation, but no static Go graph extractor was found in the inspected Go SDK. Dynamic `range` fan-out prevents a pre-run view from naming the concrete future nodes, but a source-derived outline can represent the loop and its possible effect calls. Whether a separate declared graph/plan is worthwhile is a product decision; this source only establishes that a runtime trace is not complete static topology.

**Themes.** checkpointed calls; regular Go routing; dynamic fan-out; runtime trace; idempotency at external boundary.

**Gotchas.** Do not put durable context actions inside `Run`; do not assume Go goroutines are interchangeable with Restate's waiting APIs; do not convert the vendor's “exactly-once invocation” language into exactly-once arbitrary third-party side effects without checking that provider's idempotency protocol.

**Retrieval recipe.** To design a durable effect wrapper, open `durable/restate-go/run.go.txt:12`; to test whether a code graph can be static, open `durable/restate-go/parallelizework-main.go.txt:24`.

**Status and license.** API evidence on 2026-09-08: active/non-archived, `main` `3f7852328e1fc1f5b0716d8a1e556a614d83dc97`, commit date 2026-08-21, latest release `v1.0.4` (2026-08-21), MIT; see `durable/status.json` and `durable/MANIFEST.md`.

**Unknowns and counterevidence.** No Restate server/UI was run. The public documentation source was not obtainable through GitHub API in this collection, so its prose was linked rather than copied; `run.go` is the local implementation evidence. A separate spike would be required to establish Go loop/retry behavior, replay compatibility across code changes, and exact failure ordering at an external effect boundary.
