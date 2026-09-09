# Go program proof of concept

Tyler authorized this implementation on 2026-09-08: aim for a working narrow
slice in 30 minutes, with 90 minutes as the ceiling. The requested primitives
are codergen, validation, command execution, and a Go custom iterator over the
existing Markdown/YAML checklist. Fan-out/fan-in is follow-up work.

**Placement note (2026-09-09):** the implementation and proof documented here
remain in the [experimental source branch](https://github.com/tylergannon/tractor/tree/codex/programmatic-workflow-research)
at `4906a98c44930763597ebe015d5b996c120bc1eb` and its earlier commits. This
documentation PR preserves the record only; it does not add the `tractor
program` CLI or runtime to `main`. The commands and artifact links below are
historical. Paths under `/var/folders` and `/tmp` were on the original machine
and are not portable proof access.

## The first choice

**Agent implementation decision:** `program.Loop` owns item validation, ledger
reconciliation, and optional whole-ledger evaluation. The ordinary Go body
owns agent calls and review routing. This keeps `done` tied to checks while
letting the program choose each call's result shape and use Go conditions.
The validation callback can be an ordinary function; it is not required to
be an agent call. `Runtime.Validate` supplies the existing command-plus-infer
behavior as a convenience.

```go
for iteration, err := range program.Loop(ctx, ledger, program.LoopOptions{
    Validate: runtime.Validate,
    MaxIterations: 8,
}) {
    if err != nil {
        return err
    }
    // Ordinary Go: call agents, inspect typed replies, branch, call helpers.
    // The iterator validates the resulting workspace when the body returns.
    _ = iteration
}
```

The iterator re-reads the existing ledger on each lap and validates every
listed item, including items already marked done. Failed checks reopen work.
The iteration limit bounds body executions and still permits their final
validation. Breaking out stops iteration; it does not assert goal completion.
If all item checks pass but the optional whole-ledger evaluator rejects the
goal, the first item is reopened with its feedback. That deterministic fallback
is a POC choice, not a proposed general replanning policy.

## Try the CLI

The existing graph CLI remains available. The new path is explicit:

```sh
go build -o /tmp/tractor-program ./cmd/tractor
/tmp/tractor-program program schema sprint-execute
/tmp/tractor-program program run sprint-execute \
  --input /absolute/path/task.json \
  --workdir /absolute/path/application \
  --logs /absolute/path/new-run-directory
```

Example sprint task:

```json
{
  "goal": "Complete the planned sprints and demonstrate their promises.",
  "checklist": "docs/sprints/ledger.md",
  "implement_model": "gpt-5.6-terra",
  "review_model": "gpt-5.6-sol",
  "evaluate_model": "gpt-5.6-luna",
  "max_iterations": 8
}
```

`program/workflows` contains Go counterparts of `sprint-execute`,
`chapter-loop`, and `delivery-loop`. Each publishes its own input JSON Schema.
The chapter routine nests the sprint iterator using each chapter item's
`checklist` field. The delivery routine plans, critiques, updates, then runs
the coding/review/check loop. Review returns a domain-shaped material-defect
answer; Go determines whether to code again or proceed to validation.

These are the central control flows, not full compatibility replacements.
Supervisor patrols, native-session fidelity across laps, exact resume,
fan-out/fan-in, and a static Go diagram extractor are outside this slice.
The default implementer, reviewer and evaluator use different models of the
same provider; cross-provider review can be configured. No comparative
quality or cost claim follows from that default.

## What to inspect

The code is separated so the algorithm is visible:

- `program/workflows/workflows.go`: sequence, nested iteration, and review/repair routing.
- `invoke.go`: JSON validation, decoding, and named-program dispatch.
- `config.go`: typed inputs, input guards, defaults, and path preparation.
- `prompts.go`: instructions and prompt construction.
- `steps.go`: typed agent calls and their result schemas.

This separation followed Tyler's review of the initial POC. Prompt wording,
public signatures, and control-flow order were preserved. The native proof
below identifies the earlier frozen candidate; the subsequent extraction was
checked with source comparison, ordinary tests, and linting.

`program` is the embeddable package. `Codergen[T]` accepts an exact JSON Schema
for that call and returns its decoded Go value. `Runtime.Command` reports a
command's exit code and output separately from infrastructure errors.
`Runtime.Validate` combines command results and evidence judgment. The
iterator accepts callbacks, so programmatic and mixed validations fit without
a new declarative action language.

The run directory holds operation start/end records and native agent logs.
These are the first observation primitives; a runtime map and spending UI
remain follow-up work. They do not implement replay or persistence of a Go
stack.

## Real demonstration

The new integration entrypoint uses the existing deliberately broken shipping
CLI fixture and real native agents:

```sh
go test -tags=integration ./internal/workflows \
  -run '^TestGoProgramLoop$' -count=1 -v -timeout=25m
```

It runs the experimental CLI, preserves an acceptance executable outside the
agent workspace, observes baseline failures, checks final standard and
expedited shipping outputs, and checks that the acceptance contract survived.
Its artifact directory is retained. This is separate from the repository's
mandatory `TestCanonicalLoop` proof for a new Tractor build.

## Results from the frozen candidate

Proved source revision: `13961c504bde31e2edd8e06f95e497ba4e110593`.
Both integration tests built byte-identical candidates with SHA-256:

```text
4f195c9429afa3e1d180e140cf67c9e3a48cbc7bb92265213a17c459855e73bd
```

| Run | Observed result | Local artifacts |
| --- | --- | --- |
| Go `sprint-execute` | `TestGoProgramLoop` exited 0 in 418 seconds; `passed: true`. Two implementation/review rounds, six command validations, six evidence judgments, one final goal evaluation. All six final CLI cases passed; acceptance files and ledger contract preserved. | [Receipt](/var/folders/lt/09rsy64x65s_0fp2b8zq3n7m0000gn/T/tractor-go-program-2503217793/result.json), [operation history](/var/folders/lt/09rsy64x65s_0fp2b8zq3n7m0000gn/T/tractor-go-program-2503217793/run/operations.jsonl) |
| Existing graph `sprint-execute` | Required `TestCanonicalLoop` exited 0 in 327 seconds; `passed: true`. Both sprints reviewed and engine-validated, final six CLI cases passed. Root inspected both review transcripts, including actual CLI commands and conclusions. | [Receipt](/var/folders/lt/09rsy64x65s_0fp2b8zq3n7m0000gn/T/tractor-canonical-go-1170671673/result.json) |
| Go `chapter-loop` | Same frozen binary on a tiny separate Go application. Initial command exited 1 and printed `shipment is pending`; final command exited 0 and printed `shipment is delivered`. Inner sprint and outer chapter checks/evaluations ran and both entries became done. Process exited 0. | [Before](/tmp/tractor-chapter-demo.rzCuBR/before.txt), [after](/tmp/tractor-chapter-demo.rzCuBR/after.txt), [operation history](/tmp/tractor-chapter-demo.rzCuBR/logs/operations.jsonl) |

The built candidate is retained at
`/var/folders/lt/09rsy64x65s_0fp2b8zq3n7m0000gn/T/tractor-go-program-2503217793/tractor`.
Artifacts are local; no release, remote push, or PR has been performed.
Source implementation and its subsequent startup fixes are checkpointed in
`983ebc6`, `40e4bca`, and `13961c5`. Later changes to this report are prose only.

The Go run's 17 observed operations comprise six commands and eleven native
agent calls. Native Codex usage is available per call. Taking only the last
cumulative usage event **per log file**, then summing across calls, gives
968,026 input tokens (716,672 cached) and 14,173 output tokens. Cached input is
a subset of input, not an additional chargeable count; these numbers are not
a price estimate. [The extracted usage summary](/var/folders/lt/09rsy64x65s_0fp2b8zq3n7m0000gn/T/tractor-go-program-2503217793/usage-summary.json)
retains each call's reported totals. Repeated validation calls share a name,
so grouping only by name would lose their separate usage. Timings and native
usage make the intended observer/eval seam concrete without building that UI.

The two shipping runs used different agent policies and model selections;
their elapsed times are observations, **not a controlled comparison**. This
POC does not answer whether the Go routine is faster or better. It demonstrates
that the authoring/runtime separation works on the existing application.

Ordinary repository tests, formatter/modernizer checks, and golangci-lint all
passed on the code candidate. The first native attempts exposed two integration
contracts that fake backends missed: fresh-session fidelity requires an empty
thread key, and native run-log segments must be allocated before calling the
harness. Those startup attempts failed; only the fresh passing runs above
support the result.

## Deliberate limits and next decisions

- `delivery-loop` is implemented and compiles, but was not live-executed in
  this slice. The live exercises cover the flat and nested execution loops.
- `Runtime.Validate` requires a command and/or inference gate; prose alone
  does not automatically pass. Direct `Loop` callers can supply any ordinary
  Go validation callback instead.
- Every item is validated before selection, including the initial broken
  state. This is simple and supports reconciliation, but costs extra command
  and judge calls. Selection/validation policy remains an obvious eval axis.
- The optional goal evaluator runs when all item checks pass. The old engine's
  exact evaluator cadence is not reproduced.
- The observer surface is a flat named-operation trace plus native events,
  including usage. There is no runtime tree/map UI or static graph extraction.
- No fan-out/fan-in, patrol supervisors, native-session continuity, or exact
  resume is implemented in the Go routines. The graph path remains available.
- This is an experimental package, not a general workflow framework or a
  finalized task-schema convention. Input schemas are kept beside their Go
  programs; model-result schemas are supplied per call.
