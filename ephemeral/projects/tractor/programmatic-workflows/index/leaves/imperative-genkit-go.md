# Genkit Go: ordinary function flows, typed action boundaries, trace steps

Purpose: Genkit Go is direct evidence that an AI framework can call a flow an ordinary generic Go function rather than an explicit edge graph. Snapshot: Apache-2.0, commit `192f84a4b7532d815fd0bc37b8dc845d7296f420`; GitHub API reported non-archived on 2026-09-08. The latest repository release was `py/v0.11.0`, while the latest `go/*` release was `go/v1.13.1`, both published 2026-09-03. Corpus and hashes: `imperative/manifest.json`.

## Findings

- `genkit.DefineFlow` takes `func(context.Context, In) (Out, error)`, creates a `core.Flow`, and registers it. The example itself says the body can add `genkit.Run` steps. This is normal Go branching/loops/closures inside the callback, not a user-authored edge graph. `imperative/genkit/go/genkit/genkit.go.txt:465-489`, `imperative/genkit/go/core/flow.go.txt:29-59` ([pinned source](https://github.com/genkit-ai/genkit/blob/192f84a4b7532d815fd0bc37b8dc845d7296f420/go/genkit/genkit.go#L465-L489)).
- `core.Run`/`RunWithContext` establish named runtime steps and trace spans, and require a flow context. The comment says a step result is cached for restart, but the examined implementation itself only opens a span; storage/restart wiring is outside this small surface. That is a precise lead to inspect before adopting Genkit recovery claims. `imperative/genkit/go/core/flow.go.txt:75-128` ([pinned source](https://github.com/genkit-ai/genkit/blob/192f84a4b7532d815fd0bc37b8dc845d7296f420/go/core/flow.go#L75-L128)).
- Generic `Action[In,Out,Stream]` carries inferred or supplied JSON schema. Action execution validates input, invokes the typed Go function, then validates output, returning a typed `Out` to Go callers. This is a useful shape for a Tractor call/effect library, although it validates Go values at its action boundary rather than proving a model emitted a schema-conforming wire payload. `imperative/genkit/go/core/action.go.txt:43-95`, `imperative/genkit/go/core/action.go.txt:240-336` ([pinned source](https://github.com/genkit-ai/genkit/blob/192f84a4b7532d815fd0bc37b8dc845d7296f420/go/core/action.go#L43-L336)).
- No Go Mermaid/DOT/flow-graph renderer was found in the inspected Go source. The runtime starts a reflection server for registered action metadata in development, which is observability/introspection rather than a source-level pre-execution plan. `imperative/genkit/go/genkit/genkit.go.txt:220-231`.

## Representative API shape (source-accurate)

```go
flow := genkit.DefineFlow(g, "review", func(ctx context.Context, in Input) (Verdict, error) {
    return core.RunWithContext(ctx, "validate", func(ctx context.Context) (Verdict, error) {
        return validate(ctx, in)
    })
})
```

The signatures and naming come from `imperative/genkit/go/genkit/genkit.go.txt:471-488` and `imperative/genkit/go/core/flow.go.txt:91-105`. The `validate` function is illustrative.

## What to copy / counterevidence

Copy the small, explicit function wrapper plus named steps and schema-aware typed boundary. Do not describe it as a complete solution for pre-execution visualization: Genkit's source supports run-time tracing, not extraction of an imperative function's control-flow graph. Also do not equate "Go generic return type" with an LLM structured-output guarantee; the provider adapter must receive schema, collect output, and locally decode/validate it.

Unknowns: the exact backing store and scope of the `Run` restart cache, and whether its model plugins expose a provider-independent strict JSON schema path, were not established here. No fetched code was run.
