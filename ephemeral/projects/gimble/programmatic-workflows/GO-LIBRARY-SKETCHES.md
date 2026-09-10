# Additive Go workflows and a dependable builtin catalog

Working proposal, 2026-09-09. Requested by Tyler; the roadmap and API names
below are agent proposals, not accepted contracts or implemented APIs. These
are hypothetical Go programs, with prompt construction and registration
boilerplate separated from their control flow.

## Direction

Tyler proposes adding the Go library while retaining the existing workflow
language, taking arbitrary Polytype-supported argument shapes, referring to
workflow-defined roles rather than passing model settings at each call, and
concentrating investment on a small curated library of improving workflows.
Agents with the binary should discover each workflow's purpose and calling
convention, validate an argument, and call a stable name from scheduled work.

I agree. The maintained workflow and its invocation contract are the main
product. The public Go package makes those workflows easier for us to author
and improve; adoption need not depend on users becoming workflow authors.

One source-level correction: `graph` defines the DAG's representation;
`engine` and `harness` implement its capabilities. Project those capabilities
into a public `workflow` package. Do not make ordinary Go execution depend on
building a `graph.Graph` first. The old graph interpreter can continue to use
the same execution machinery.

## Short roadmap

- **Translate immediately, starting from readable Go sketches.** Add
  `workflow`; translate `fix-until-green`, `bake-off`, and `sprint-execute`
  first. Extract shared execution operations only as these examples require
  them. Keep YAML parsing, schemas, and execution available throughout.
- **Give each workflow a typed entrypoint.** Accept `Args` chosen by that
  workflow; generate its JSON Schema, validation, and any required JSON
  codecs with Polytype. Go owns the algorithm; the JSON describes its input.
- **Declare roles beside the workflow.** Separate responsibility, step name,
  and staffing. Bind models through curated role profiles; retain explicit
  independence, context freshness, and session-continuity policies.
- **Make the binary its own calling manual.** Extend the existing catalog
  with purpose, argument schema, examples, prerequisites, observable success,
  outputs, and validation. Validate before allocating a run or worktrees.
- **Improve methods behind stable contracts.** Keep names and compatible
  argument semantics dependable; version breaking contracts separately from
  implementation revisions. Translate the remaining builtins, compare real
  outcomes, and improve the maintained methods over time. Deprecation is a
  separate decision after the Go path has earned it.

## Library shape

Prefer ordinary functions with a conventional signature:

```go
func BakeOff(ctx context.Context, run *workflow.Run, in BakeOffArgs) error
```

`Run` owns execution resources and exposes scoped agent, command, worktree,
supervision, and human-wait operations. Ordinary Go expresses branches,
loops, calls, and concurrency. A small parallel-worktree helper is useful
because its lifecycle and artifact handoff are more than goroutine syntax.
Avoid a fluent node/edge builder and a universal task envelope.

Use a typed result when an agent supplies a decision:

```go
review, err := workflow.Ask[Review](ctx, run, "review", tester, reviewTask(sprint))
```

This is a generic package function; it is not a generic method on `Run`.
The response schema describes a `Review`, not a set of graph edge IDs. Go
then decides what to do with that result. Prompt factories such as
`reviewTask` assemble instructions and relevant context; they perform no
hidden retries, branching, or validation.

Every operation gets an invocation identity for logs and steering, separate
from the role. Concurrent invocations of `sswe` cannot be addressed safely
by the role name alone. A child scope carries its own contextual contribution
while sharing its parent's run and artifacts. Context includes both the
assembled request and routes to ordinary local information/index files.

### Declared arguments

These nested structs are specific to bake-off. Other workflows need not
have `Goal`, `Knowledge`, or any of these fields.

```go
type BakeOffArgs struct {
    // Desired behavior every candidate must implement.
    Goal string `json:"goal"`
    Acceptance Acceptance `json:"acceptance"`
    Knowledge Knowledge `json:"knowledge"`
}

type Acceptance struct {
    // Shell command run against the integrated winner; zero means success.
    Command string `json:"command"`
    // Candidate files the judge may bring into the main workspace.
    Deliverables []string `json:"deliverables"`
}

type Knowledge struct {
    // Workspace-relative entry point explaining the task's available knowledge.
    Index string `json:"index"`
    // Additional local source paths the candidates should know about.
    Sources []string `json:"sources"`
}
```

Generate the schema and validator from these types at build time. Each
workflow can use the nested structs, slices, enums, optional fields, and
registered unions that its pinned Polytype version supports. There is no
promise that every possible Go type is JSON-projectable. Reject unsupported
types at generation time rather than widening them to untyped objects.
Use generated owner codecs where the chosen shape needs them, including
sealed unions. Argument schema validation comes before decoding; semantic
validation then checks things such as readable input paths and a nonempty
deliverable list. Neither step starts agents.

### Roles

The workflow declares these names; their staffing is maintained with it:

```go
const (
    sswe       workflow.Role = "sswe"
    alternate  workflow.Role = "sswe-alternative"
    challenger workflow.Role = "sswe-challenger"
    tester     workflow.Role = "tester"
    engMgr     workflow.Role = "eng-mgr"
)
```

A role definition holds its responsibility and a named staffing profile.
Profiles resolve to provider/model/settings outside the algorithm. The
bake-off definition assigns its three engineer seats distinct providers,
matching the original example. The tester receives fresh context and must
operate candidates itself; it does not inherit their conversations. In
`sprint-execute`, require the tester to use another provider from the engineer.
Resolve and validate these relationships before launching. Role names alone
do not establish independence. `delivery-loop` separately specifies that its
engineer's native session continues across coding laps.

## Example: bake-off

Generalizes [the shipped example](../../../../examples/loops/bake-off.yaml):
three isolated candidates, reports, a judge that operates each one and
integrates the winner, then a command against the integrated result.

```go
func BakeOff(ctx context.Context, run *workflow.Run, in BakeOffArgs) error {
    candidates, err := run.ParallelWorktrees(ctx, "attempts",
        []workflow.Role{sswe, alternate, challenger},
        buildAndDemonstrate(in, "REPORT.md"),
    )
    if err != nil {
        return err
    }

    _, err = run.Agent(ctx, "judge", tester,
        judgeAndIntegrate(in, candidates, "WINNER.md"))
    if err != nil {
        return err
    }

    return run.RequireCommand(ctx, "verify", in.Acceptance.Command)
}
```

`ParallelWorktrees` runs the calls concurrently in three isolated worktrees
from the same initial workspace snapshot and returns candidate locations,
outcomes, and collected reports. The run keeps these available through
judging. It owns cancellation and resource cleanup. This initial spelling
requires all candidates to finish successfully; keeping partial candidates
would be an explicit later policy, not an accidental consequence of errors.

`judgeAndIntegrate` tells the tester to run every candidate, judge demonstrated
behavior, copy the winning allowed deliverables into the main workspace, and
write the rationale to `WINNER.md`. Reports alone are not evidence of working
behavior. `RequireCommand` returns an error for a failing command; success
therefore cannot follow just from the tester's positive report.

The exact FizzBuzz example supplies `deliverables: ["fizzbuzz.sh"]` and the
original shell check, including the check for `WINNER.md`. Arbitrary paths
and a task argument are proposed generalizations of that fixed example.

## Example: fix until green

This translation retains the original five implementation attempts. Command
failure is a result to act on; cancellation or failure to launch is an error.

```go
type FixArgs struct {
    Goal string `json:"goal"`
    Check string `json:"check"`
}

func FixUntilGreen(ctx context.Context, run *workflow.Run, in FixArgs) error {
    var feedback string
    for attempt := 0; attempt < 5; attempt++ {
        _, err := run.Agent(ctx, "implement", sswe, fixTask(in.Goal, feedback))
        if err != nil {
            return err
        }
        check, err := run.Command(ctx, "check", in.Check)
        if err != nil {
            return err
        }
        if check.ExitCode == 0 {
            return nil
        }
        feedback = check.Output
    }
    return workflow.ErrAttemptsExhausted
}
```

Passing the previous command output to the next attempt is an explicit
context improvement in this sketch. The existing example carries the goal
through its loop; the Go shape makes this feedback choice easy to see.

## Example: execute a sprint ledger

The cursor is justified by a domain invariant: the runtime owns attestation.
It is not just an iterator over unchecked boxes.

```go
type SprintArgs struct {
    Ledger string `json:"ledger"`
}

var sprintLedgerPolicy = workflow.ChecklistPolicy{
    MaxVisits: 40,
    Evaluator: "evidence-judge",
    ItemStepLimits: map[string]int{"implement": 12, "review": 12},
}

func SprintExecute(ctx context.Context, run *workflow.Run, in SprintArgs) error {
    return run.Supervised(ctx, sprintCoaches, func(scope *workflow.Run) error {
        sprints := scope.Checklist(in.Ledger, sprintLedgerPolicy)
        for sprints.ValidateAndSelect(ctx) {
            if err := implementAndReview(ctx, sprints.Scope()); err != nil {
                return err
            }
        }
        return sprints.Err()
    })
}

type Review struct {
    MaterialDefect bool `json:"material_defect"`
    Findings string `json:"findings"`
}

func implementAndReview(ctx context.Context, sprint *workflow.Run) error {
    for {
        if _, err := sprint.Agent(ctx, "implement", sswe, implementCurrentSprint()); err != nil {
            return err
        }
        review, err := workflow.Ask[Review](ctx, sprint, "review", tester, reviewCurrentSprint())
        if err != nil {
            return err
        }
        if !review.MaterialDefect {
            return nil // Ready for the runtime's check; not an attestation.
        }
        sprint = sprint.WithFeedback(review.Findings)
    }
}
```

`ValidateAndSelect` rereads the ledger on each arrival, runs the selected
item's and already-done items' command/inference gates, updates `done` in both
directions, and selects work still needing demonstration. The ledger's own
goal evaluator controls final exit. A false return means either completion
or error, distinguished by `Err`. A handwritten `done` field never skips
validation. Its returned scope supplies the current item, relevant goal,
documents, and validation feedback to subsequent tasks.

`sprintLedgerPolicy` names the existing 40-arrival limit, goal-evaluator
role, and separate 12-visit implementation/review ceilings. Its shared item
scope counts each named step across helper calls, preserving counts when
validation fails and resetting them when a new item is selected. `Agent`
and `Ask` return a limit error when those ceilings are reached. `WithFeedback`
preserves this accounting while changing the agent-facing context. The Go
loop therefore cannot accidentally restart its allowance after a failed
runtime check. A limit produces an error/needs-action outcome, never success.

`sprintCoaches` declares both existing advisory relationships:
`serves-the-requirement` watches implementation and review;
`proof-is-evidence` watches review. Their patrol/turn intervals remain 180
seconds. `Supervised` starts them, scopes their observation, propagates startup
errors, and stops/drains them on exit. Coaches can steer the work; they cannot
choose its next branch or mark anything done.

The remaining translations follow the same vocabulary:

- `chapter-loop`: an outer chapter cursor calls the sprint function with
  the selected chapter's scoped run and ledger. Its own command/inference
  gates chapter completion. One run, with limits per activation.
- `critique-circle`: isolated proposals, unchanged collection, isolated
  critiques of the other proposals, unchanged collection. No synthesis or
  winner is silently added.
- `sprint-plan`: orientation, three drafts, mutual critique, human-only
  questions, a blocking interview, then merge. Failed interviews regenerate
  questions within the existing three-visit limit.
- `delivery-loop`: plan, critique, revise, then the checked item loop;
  retain the engineer session across laps and use fresh independent review.

## A self-describing catalog

Use one typed registration to connect executable code, documentation, and
input validation. `Definition` is ordinary Go metadata, not a workflow DSL:

```go
var BakeOffWorkflow = workflow.Definition[BakeOffArgs]{
    Name: "bake-off",
    Contract: 1,
    When: "Try independent implementations and keep the best demonstrated result.",
    Schema: BakeOffArgs{}.Schema(),
    ValidateJSON: BakeOffArgs{}.ValidateJSON,
    Validate: validateBakeOff,
    Roles: bakeOffRoles,
    Examples: bakeOffExamples,
    Run: BakeOff,
}
```

Schema and JSON validation methods above stand for generated Polytype
outputs. The registry can erase the type internally for listing/dispatch;
the entrypoint still receives `BakeOffArgs`. Help and execution consume this
same registration. Prerequisites, allowed effects, produced files, success
meaning, and failure/needs-human outcomes also belong in its calling manual.

Proposed CLI, with existing YAML commands preserved:

```sh
gimble workflows --json
gimble workflows describe bake-off --json
gimble workflows schema bake-off --contract 1
gimble workflows validate bake-off --contract 1 --args @task.json
gimble run bake-off --contract 1 --args @task.json
```

`describe` returns when to use it, prerequisites, examples, contract versions,
outputs and completion meanings, plus the input schema. `schema` emits only
the schema. `validate` checks that same contract without running the method;
`run` repeats admission validation. The existing `workflows show` remains
source-oriented. Keep runtime options such as workdir and delivery outside
the workflow's arbitrary argument object.

A scheduled task pins the argument contract, for example `bake-off` contract
1, while installed releases can improve the implementation and staffing.
Compatible improvement preserves field meaning, defaults, promised outcomes,
and allowed effects, not just whether old JSON still parses. Breaking changes
get a new contract version; retain the old implementation or an honest adapter
for old calls. Unsupported versions fail clearly before work starts.

Record the resolved contract, binary/implementation revision, and role
bindings with each run. This identifies what executed without demanding exact
replay. Scheduled runs benefit from improvements when their installed binary
is upgraded; a stable workflow name is not a remote self-update mechanism.

## The trade and the first acceptance bar

Go improves authoring and composition; it costs the current graph engine's
easy static topology and node-addressed continuation. Keep runtime events,
context inspection, and steering attached to scoped invocations so freedom
of authorship does not erase visibility. Deriving a helpful diagram can
follow later. Existing graph checkpoints cannot restore an arbitrary Go
stack; retain YAML resume and specify Go restart behavior separately.

The first implementation should demonstrate the translated workflows with
the real software: worktree isolation and winner verification for bake-off,
failed-check retry and exhaustion for fix-until-green, and command/inference
ownership plus independent review for sprint-execute. Compare meaningful
behavior against the YAML originals. Any new-build proof must additionally
run the repository's canonical native-agent loop and inspect its required
artifacts; these sketches themselves make no build or runtime claim.

## Current source anchors

- `graph/graph.go` and `graph/schema.go`: DAG representation and definition schema.
- `engine/runner.go`, `engine/loop.go`, `engine/parallel.go`: execution lifecycle,
  runtime attestation, and isolated branches; several seams remain runner-coupled.
- `harness/backend.go`: arbitrary schema-validated agent results already exist
  below graph-specific routing; useful foundation for `Ask[T]`.
- `internal/workflows/workflows.go` and `cmd/gimble/workflows.go`: existing
  builtin purpose/needs catalog and listing/source commands to extend.
- `docs/direction.md`, `docs/workflows-as-programs.md`, and
  `docs/five-arts.md`: the direction and the authoring/visibility/steering trade.
