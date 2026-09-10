# Proposed authoring shapes: critique circles and bake-offs

Design exploration, 2026-09-08. These snippets describe the API to aim for;
they are not implemented functions or a concurrency guarantee for today's POC.

Tyler's direction: choose readable orchestration pseudocode first, then derive
functions/types from it. Prefer ordinary Go goroutines and errgroup over new
graph nodes, edge types, or a mandatory promise/future abstraction.

**Status 2026-09-10.** The two code blocks below are the intended shape: the
tactic is written out in the workflow with `errgroup`. `RunTrial` and `Choose`
are local helpers of that one workflow, not library functions, and there is
no `workflows.BakeOff`. The section "What Tractor supplies underneath" is
withdrawn: no scoped context, no observation scopes, no workspace-isolation
primitive. A candidate worktree is `runtime.Command(ctx, "git worktree add
...")` in the workflow's own code. See AGENTS.md, "No wrappers".

## Starting recommendation

Use ordinary synchronous operations returning typed values and errors. A caller
chooses whether to run them sequentially or inside an errgroup. The group owns
joining/cancellation; results are ordinary values after Wait. This is the
structured-concurrency pattern: child work belongs to an enclosing operation,
and that operation joins its children before returning.

A Future is useful when a running computation must be passed around and awaited
individually. Neither of these initial algorithms needs that extra type. A
small generic parallel-map helper may later remove repetition; it would be a
convenience over errgroup, not an alternate workflow execution language.

## Critique circle: a shared proposal

Interpretation for this example: several independent reviewers inspect the same
proposal; its owner adjudicates the findings and revises it. Each round reviews
one unchanged version. Adjudication may reject overreach; readiness belongs to
the task's review contract, not a framework-mandated unanimity rule.

Assume input validation, initial drafting, model binding and prompt definitions
are outside the snippet. Functions below are proposed typed steps. Go errors
represent execution failure; critique findings are ordinary returned data.

```go
for range maxRounds {
    critiques := make([]Critique, len(reviewers))
    group, reviewCtx := errgroup.WithContext(ctx)

    for i, reviewer := range reviewers {
        group.Go(func() error {
            var err error
            critiques[i], err = reviewer.Critique(reviewCtx, proposal)
            return err
        })
    }
    if err := group.Wait(); err != nil {
        return proposal, err
    }

    decision, err := adjudicate(ctx, proposal, critiques)
    if err != nil {
        return proposal, err
    }
    if decision.Ready {
        return proposal, nil
    }

    proposal, err = revise(ctx, proposal, decision.Changes)
    if err != nil {
        return proposal, err
    }
}
return proposal, ErrNeedsDecision
```

The first version deliberately fails the round on a reviewer execution error;
accepting partial review coverage would be a separate policy choice. Negative
reviews do not return Go errors and do not cancel other reviewers.

If "circle" means each author critiques every other author's draft (as in the
existing competitive sprint planner), use the same phase structure: independent
drafts; parallel peer critiques of the frozen draft set, excluding each author's
own draft; owner revisions after the join; then either another round or synthesis.
That is a different argument/result shape and selection loop, not a new node.

## Bake-off: independent candidates

Each contestant gets the same starting repository snapshot and acceptance
contract. It builds in its own workspace. Compare the actual candidates after
running that acceptance; finishing first does not establish a winner.

```go
trials := make([]Trial, len(contestants))
group, trialCtx := errgroup.WithContext(ctx)

for i, contestant := range contestants {
    group.Go(func() error {
        trials[i] = RunTrial(trialCtx, contestant, baseline, acceptance)
        return trialCtx.Err()
    })
}
if err := group.Wait(); err != nil {
    return Decision{}, err
}

return Choose(ctx, trials)
```

`RunTrial` is an ordinary function: create an isolated candidate workspace from
baseline, run the contestant there, validate with the shared acceptance, and
return the candidate plus its observed outcome and usage/time. Preparation,
resource cleanup and artifact retention live in that helper. Failed attempts
are Trial data, so one failed candidate does not cancel all remaining candidates;
parent cancellation still terminates the bake-off. The illustration uses a
different error policy from the critique circle intentionally.

`Choose` compares eligible candidates according to the declared task. It can
return no winner when none meet acceptance. Selection produces a choice and
retained candidate; it does not silently merge or publish that candidate.

In both snippets, each goroutine writes its own preallocated result slot and
the enclosing program reads only after Wait. Keep the proposal immutable during
a review round. Use the original ctx after Wait: errgroup cancels its derived
context when Wait returns, including successful completion. SetLimit can bound
parallelism when needed.

## What Tractor supplies underneath

1. Scoped context inherited by each call, with branch-local additions. Typed
   step arguments such as proposal bind into that scope; the step prompt can
   remain "Critique the current proposal." Do not concurrently mutate a shared
   context map. Child scopes snapshot/inherit outer context and own their writes.
2. Native agent calls that compose with ordinary context cancellation, with
   per-invocation identity distinct from a human-readable step name.
3. Workspace isolation for competing implementations, and immutable artifacts
   or read-only inputs for simultaneous critiques.
4. Named observation scopes: circle/round/reviewer and bake-off/contestant/step.
   The caller's context propagates the parent relationship; scheduling itself
   need not be owned by Tractor for the runtime to record this tree.

The current POC is sequential and needs changes before these examples are safe:
`program/codergen.go` uses Backend.InterruptAll on one call's cancellation, and
passes the step's display name as NodeID. In `harness/backend.go`, fresh-fidelity
binding locks are keyed by NodeID+role, so equal step names can serialize calls.
These are concrete reasons to fix call ownership and identity before exposing
parallel composition. No concurrency implementation was made in this turn.

## Pattern and platform comparison

The primary Go precedent is [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup):
goroutine groups, joining, error propagation, optional cancellation and limits.
Its standard parallel example already collects independently computed results
into distinct slice entries. Concurrency does not require graph nodes or futures.

[Temporal's Go workflow model](https://docs.temporal.io/develop/go/workflows/basics)
adds deterministic replay. Workflow code uses workflow.Go, workflow.Context and
other SDK substitutes, and external I/O is delegated to Activities. Its
[multithreading documentation](https://docs.temporal.io/develop/go/best-practices/multithreading)
describes the deterministic runner and prohibits native threading in workflow
code. A server-backed execution history and durable scheduling are the substantive
platform commitment; ordinary parallel calls and typed results alone are not it.

Recommendation: stay with native Go for the currently stated recovery needs.
Reconsider Temporal if keeping execution alive across process/machine failures,
long durable waits, or distributed worker scheduling becomes part of the task's
actual contract. If it does, evaluate adopting that machinery rather than
building a durable scheduler into Tractor.

[Genkit Go flows](https://genkit.dev/docs/go/flows/) are a closer reference for
the present developer experience: typed Go functions, schema-bearing flow inputs,
and observable steps. Its RunWithContext gives a step a child context for nested
tracing; its [concurrency guide](https://genkit.dev/docs/go/concurrency/) describes
ordinary goroutines and context cancellation. This is relevant prior art for
Tractor's context/observation interface. Adopting Genkit would still require
integration with native coding-agent sessions and Git workspaces; no migration
benefit for those responsibilities was demonstrated here. A focused integration
spike can answer whether its schema/UI/tracing facilities justify that work.

The immediate next implementation should be the shared-proposal critique circle:
scoped context, concurrent typed calls, joined critiques, and owner revision.
Then add isolated candidate workspaces for the bake-off. Keep the algorithms
above readable while implementing those seams; no Future type is required first.
