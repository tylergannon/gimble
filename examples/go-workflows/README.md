# Go workflow examples with executable stubs

Start with the algorithms:

- [Bake-off](bakeoff/main.go): three concurrent builders, join, judge, integrate,
  then check the integrated result. Translates [bake-off.yaml](../loops/bake-off.yaml).
- [Critique circle](critique/main.go): three independent proposals, join, then
  each author critiques the other two. Translates [critique-circle.yaml](../loops/critique-circle.yaml).
- [Sprint execution](sprints/workflow.go): range over a checklist iterator,
  implement, review, repair material defects, then let validation decide done.
  Follows the [recovered Go POC](../../ephemeral/projects/gimble/programmatic-workflows/POC-WORKFLOWS.md).

These are real Go programs. Every external effect is a stub: agent responses and
command results are canned; workspaces are symbolic names; integration only
logs an action. Running them demonstrates Go control flow, not successful agent
work, command execution, or Git isolation. The existing Gimble runtime is untouched.

From the repository root:

```sh
go run ./examples/go-workflows/bakeoff
go run ./examples/go-workflows/critique
go run ./examples/go-workflows/sprints
```

Each prints a stub trace to stderr and a JSON result to stdout. The sprint demo
deliberately needs a repair and another implementation attempt before its
scripted check passes; both checklist entries must finish. The bake-off always
chooses the second candidate. Custom inputs change the task passed to each
stub, not these scripted outcomes.

Every workflow owns its argument type in `input.go`. There are no model arguments:
roles such as `sswe`, `eng-mgr`, and `tester` are declared by the workflow. Show
the calling convention or supply an entire replacement JSON argument:

```sh
go run ./examples/go-workflows/bakeoff -example
printf '%s\n' '{"topic":"How should we explain Gimble?"}' |
  go run ./examples/go-workflows/critique -input -
go run ./examples/go-workflows/sprints -input /path/to/sprints.json
```

The input types are ordinary workflow-specific structs, including nested objects
and lists. They are the intended roots for Polytype argument schemas. This demo
uses Go JSON decoding and small semantic checks; it does not implement schema
generation, schema-complete validation, or the proposed builtin catalog.

## The small primitives

[The stub package](internal/program/runtime.go) exposes synchronous `Codergen[T]`,
`Command`, `Worktree`, and `Integrate`. [Loop](internal/program/loop.go) returns an
`iter.Seq2[Iteration, error]` and owns checklist validation and done reconciliation.
It rechecks every item on each arrival, including after the final permitted
implementation attempt. The in-memory ledger has no file reload or persistence.

Parallelism is directly visible in the workflows: `errgroup.WithContext`,
`group.Go`, and `group.Wait`. Each goroutine writes its own result slot. Later
phases use the parent context because the group's context is canceled after
`Wait`, even on success. Operational errors stop a parallel phase; negative
reviews and failed acceptance checks remain typed observations for the caller
to interpret. The callbacks run concurrently, outside the trace writer's lock.

`prompts.go` contains wording; `demo.go` contains canned responses and the
executable entrypoint. Workspace context is inherited by calls, but automatic
agent context, real role-to-provider bindings, timeouts, retries, supervisors,
and persistence remain outside this example. Separate observation and steering
support will be needed to retain those arts while simplifying authoring.

```sh
go build ./examples/go-workflows/...
go test -race ./examples/go-workflows/...
```

The [semantic index](../../docs/semantic-index/programmatic-workflows/README.md)
routes to the original POC, earlier concurrency designs, and the broader
role/input/catalog direction. These stubs explore that API; they do not port
the native POC into Gimble or establish runtime parity with YAML.
