# Go workflow API examples

Start with these workflow functions; their dependencies are small stubs in
[internal/program](internal/program). No agents, commands, context storage,
indexing, or Git operations are implemented here.

| Example | What to inspect |
| --- | --- |
| [BakeOff](bakeoff/main.go) | Parallel candidates with ordinary `errgroup`, a judge, and integration |
| [CritiqueCircle](critique/main.go) | Parallel proposals, then peer critiques |
| [SprintExecute](sprints/workflow.go) | Sprint iteration, review, and repair in ordinary Go |
| [ContextWalkthrough](context/main.go) | Context declarations and updates between agent calls |
| [NestedScopes](scopes/main.go) | Chapter/sprint scopes and arbitrary parallel review scopes |

Each directory has its argument type in `input.go` and canned responses in
`demo.go`. Roles such as `sswe`, `eng-mgr`, and `tester` belong to the workflow.

```sh
go build ./examples/go-workflows/...
go run ./examples/go-workflows/bakeoff
# Replace bakeoff with critique, sprints, context, or scopes.
go run ./examples/go-workflows/bakeoff --example
go run ./examples/go-workflows/bakeoff -input argument.json
```

`--example` prints the argument JSON; `-input -` reads a complete argument from
stdin. The intended calling convention is any JSON shape supported by Polytype;
these examples use structs. Schema generation is deferred. A future builtin
catalog should describe when to use each named workflow, expose its argument
schema, and validate that argument before calling it. Stable argument contracts
allow the implementation behind a named workflow to improve independently.

## Context contract being illustrated

A declared key belongs to one scope. Only that scope may change its value.
Chapters and sprints should create scopes automatically; `Scope` supplies an
arbitrary one. Each agent entry should resolve current ancestor values and wait
for its index, then receive a fixed view for that call. Later calls can see
parent updates. No ownership analyzer is included.

These are intended semantics. The context functions are **no-ops**. Loop stubs
visit each supplied item once; they do not validate, retry, or set `Done`.
Worktree names are symbolic and agent/command responses are canned.

## Illustrative prompts

These are hand-written pictures of the intended context projection, not output
produced by the stubs. Paths are illustrative. Storage and indexing are undecided.

Small context, before the planning call:

```text
Scope: goal
goal: Implement the quote CLI.
constraints: Use integer cents; free shipping starts at 50.00.
Task: Outline the implementation and its proof.
```

After adding large research:

```text
Scope: goal
goal: Implement the quote CLI.
research: Detailed observations at context/research.json.
Task: Implement the quote command.
```

After many small notes exceed the shared context budget:

```text
Scope: goal / chapter: Quote command / sprint: Price boundaries
goal: Implement the quote CLI.
chapter_focus: Build on the completed amount parser.
Context index: context/index.md — research, constraints, boundary cases.
Task: Exercise the boundary cases and judge the recorded evidence.
```
