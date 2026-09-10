# Source leaf: Go workflow programs with stubbed effects

## Purpose

Actual Go source for reading, compiling, and running the proposed authoring
shape. Start with [the example guide](../../../../examples/go-workflows/README.md).
These are new isolated examples, not the recovered native-agent POC or a
published runtime API. Agent/command replies are canned, workspaces and
integration are symbolic, and the checklist is in memory. There is no builtin
catalog, Polytype schema generation, automatic context engine, or runtime-parity
claim.

## Key concepts

- Bake-off declares engineer/tester roles and typed candidate/judgment results;
  ordinary `errgroup` starts candidates, then Go joins, judges, integrates, and
  checks the winner: `examples/go-workflows/bakeoff/main.go:11-75`.
- Critique circle has two ordinary `errgroup` phases: proposals, then peer
  critiques of the unchanged proposal set. It returns both collections:
  `examples/go-workflows/critique/main.go:32-79`.
- Sprint execution uses `range program.Loop`, implementation, typed review,
  and bounded repair; a separate callback combines a command observation and
  judgment: `examples/go-workflows/sprints/workflow.go:15-62`.
- Each program owns its JSON argument type, separate from prompts and control
  flow: `examples/go-workflows/bakeoff/input.go:8-22`,
  `examples/go-workflows/critique/input.go:8-17`, and
  `examples/go-workflows/sprints/input.go:10-38`.
- Effects are synchronous stub callbacks; orchestration belongs to callers:
  `examples/go-workflows/internal/program/runtime.go:1-34`.
- The example iterator updates `Done` only through its validation callback and
  returns `iter.Seq2[Iteration, error]`; it has no persistent ledger:
  `examples/go-workflows/internal/program/loop.go:28-81`.

## Retrieval recipes

- For "show me real compiling programs", open the example guide, then the
  workflow source above. From the repository root, the demo entrypoints are
  `go run ./examples/go-workflows/bakeoff`,
  `go run ./examples/go-workflows/critique`, and
  `go run ./examples/go-workflows/sprints`.
- For calling conventions, use each entrypoint's `--example` and `--input`
  options and inspect its `input.go`; these do not generate JSON Schema.
- For prior native agents, persisted checklist handling, and source
  provenance, follow [POC recovery](poc-recovery.md). Stub output does not
  reverify the historical native execution or demonstrate application success.

## Themes

Ordinary Go control flow; explicit errgroup concurrency; iterator validation;
workflow-owned roles and inputs; prompts separate from algorithms; stubbed
effects versus native execution.
