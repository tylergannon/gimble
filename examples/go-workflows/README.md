# Go workflow examples with executable stubs

Start with the algorithms:

- [Bake-off](bakeoff/main.go): three concurrent builders, join, judge, integrate,
  then check the integrated result. Translates [bake-off.yaml](../loops/bake-off.yaml).
- [Critique circle](critique/main.go): three independent proposals, join, then
  each author critiques the other two. Translates [critique-circle.yaml](../loops/critique-circle.yaml).
- [Sprint execution](sprints/workflow.go): range over a checklist iterator,
  implement, review, repair material defects, then let validation decide done.
  Follows the [recovered Go POC](../../ephemeral/projects/gimble/programmatic-workflows/POC-WORKFLOWS.md).
- [Context projection](context/main.go): set context values, then call agents
  with short instructions; inspect prompts as large values and aggregate
  pressure move data behind a filesystem index.
- [Nested scopes](scopes/main.go): automatic chapter/sprint contexts, arbitrary
  parallel review scopes, and parent context restored after each nested call.

These are real Go programs. Agent responses and command results are canned;
workspaces are symbolic names; integration only logs an action. Context-enabled
examples write real local JSON values and indexes. Running these demonstrates
Go control flow and context projection, not successful agent work, command
execution, or Git isolation. The existing Gimble runtime is untouched.

From the repository root:

```sh
go run ./examples/go-workflows/bakeoff
go run ./examples/go-workflows/critique
go run ./examples/go-workflows/sprints
go run ./examples/go-workflows/context
go run ./examples/go-workflows/scopes
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

## Filesystem context and the agent boundary

The [context example](context/main.go) adds a small filesystem-backed prototype
to the stubs. `SetContext(ctx, key, value)` writes the entire JSON value and marks
the prompt view stale. The next `Codergen` call automatically materializes a
coherent index and prompt before entering the agent callback. Multiple setters
can be combined; ordinary Go work between setting data and calling the agent
does not wait for indexing. There is no background indexing worker in this sketch.

Small values appear inline. A per-value limit or a total context limit moves
values behind short key-based routes to an index containing their full paths.
Values and indexes use immutable files, so an older prompt's links remain valid
after replacement. Each snapshot also publishes a `View` directory with
`index.json` and `values/<key>.json` symlinks for ordinary file tools. The prompt
points to that view's index. Keys cannot differ only by letter case, so the view
works on case-insensitive filesystems. A failed index/view update prevents the
agent call. The limits count bytes in the context contribution, excluding the separate task instruction;
token accounting and the real semantic placement policy remain to be designed.

The program prints the exact prompt received by each stub agent at three stages:
small context, oversized research, and aggregate pressure from smaller notes.
It also reports each stage's index, inline keys, and external keys. The temporary
directory is retained for inspection and printed on stderr. No raw prompts or
run artifacts are committed. This deterministic index is a placeholder for the
semantic index; it makes no judgment about which information matters most.

See [the context design note](../../ephemeral/projects/gimble/programmatic-workflows/CONTEXT-FILES.md)
for the proposed waiting contract and open-source research leads.

## Context scopes and physical layers

`Chapters` and `Sprints` create item scopes automatically and yield their
`Context`; validation receives that same context. An arbitrary
`Scope(ctx, "research")` uses the same primitive. Each child inherits a snapshot
of the parent's effective values, writes overrides in its own nested directory,
and references the original ancestor files for inherited values. An effective
index tells the agent what applies in that scope. Its published `View/values`
directory lists all effective keys as symlinks, including inherited ones.
Each view creates directory entries and links rather than copying unchanged
payloads. `SetContext` writes a fresh value file; the next view links to it.
Directly editing a symlink would edit its target, so context updates stay behind
`SetContext`. There is no FUSE mount or identity-dependent filesystem view.

Sibling writes stay separate. Passing the parent context again restores its
view; returning a result does not implicitly add it to parent context. Agents
are trusted to follow the supplied index and write to their assigned layer;
these paths do not enforce read access restrictions. Ordinary `errgroup`
joining still owns the lifetime of parallel work.

An item scope is retained across retries in one loop activation. Because the
iterator checks every item before choosing work, those first checks create the
scopes; later parent updates do not change already-created sibling views.
See [the scoping design and filesystem comparison](../../ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md).

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
executable entrypoint. Ordinary Go child contexts share their attached store;
`Scope` creates a separate writable layer. Real role-to-provider bindings,
timeouts, retries of native calls, supervisors, and production persistence
remain outside these examples. Separate observation and steering support will
be needed to retain those arts while simplifying authoring.

```sh
go build ./examples/go-workflows/...
go test -race ./examples/go-workflows/...
```

The [semantic index](../../docs/semantic-index/programmatic-workflows/README.md)
routes to the original POC, earlier concurrency designs, and the broader
role/input/catalog direction. These stubs explore that API; they do not port
the native POC into Gimble or establish runtime parity with YAML.
