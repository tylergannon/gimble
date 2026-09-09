# Durable execution, recovery, and visualization — bounded findings

This is research for the programming-workflows decision, not a product decision or implementation authorization. Source snapshots, immutable revisions, licenses, SHA-256 hashes, and GitHub API maintenance evidence are in `corpus/durable/MANIFEST.md` and `corpus/durable/status.json`; detailed retrieval leaves are `index/leaves/durable-*.md`. No fetched code was executed.

## What the source mechanisms actually show

| Project | Code model | Recovery mechanism demonstrated by source | Effect boundary | Visualization evidence |
|---|---|---|---|---|
| Temporal Go | workflow code can branch, but uses deterministic workflow APIs to schedule activities | event-history replay; `SideEffect` records a value returned on replay | an Activity is separate from the workflow decision; source here does not make a remote HTTP/PR/deploy atomic | runtime UI/CLI summaries are mentioned; no static compiler found |
| Restate Go | normal handler flow plus `Context` actions | journal replays completed `Run` results | `Run` callback can perform nondeterministic work; source does not prove atomic commit with an arbitrary remote API | docs advertise execution trace; no static Go extractor found |
| DBOS Transact Go | typed ordinary Go workflow and step callbacks | Postgres checkpoint ledger returns prior step outcome | `RunAsStep` is explicitly at-least-once; source orders callback before checkpoint; `runAsTxn` joins a database effect with checkpoint | recorded step names/statuses after execution; no static extractor found |
| Prefect | Python flow code | outside scope for durable recovery | outside scope | `visualize()` runs non-task code; mock values select a dynamic route, so output is a controlled path trace |

The strongest source result is DBOS's actual ordering: the callback is executed and then its result is recorded (`corpus/durable/dbos-go/workflow.go.txt:2623-2689`; [pinned source](https://github.com/dbos-inc/dbos-transact-golang/blob/ab56911fdd78552e1e7fe648cff7c831a1e760c8/dbos/workflow.go#L2623-L2689)). Its own contract calls this at-least-once (`corpus/durable/dbos-go/workflow.go.txt:2527-2531`; [pinned source](https://github.com/dbos-inc/dbos-transact-golang/blob/ab56911fdd78552e1e7fe648cff7c831a1e760c8/dbos/workflow.go#L2527-L2531)). This is the concrete crash-after-effect-before-record ambiguity, rather than an abstract distributed-systems caveat.

Temporal takes the heavier route: `SideEffect` records output and returns it during replay (`corpus/durable/temporal-go/workflow.go.txt:416-456`; [pinned source](https://github.com/temporalio/sdk-go/blob/9aae5eec926f359f1fb986b7d3125bd968d1067a/workflow/workflow.go#L416-L456)), and evolves workflow code with recorded version markers (`corpus/durable/temporal-go/workflow.go.txt:526-536`; [pinned source](https://github.com/temporalio/sdk-go/blob/9aae5eec926f359f1fb986b7d3125bd968d1067a/workflow/workflow.go#L526-L536)). Restate provides the related journaled-step model: `Run` stores final result/error and replay yields the same value (`corpus/durable/restate-go/run.go.txt:12-25`; [pinned source](https://github.com/restatedev/sdk-go/blob/3f7852328e1fc1f5b0716d8a1e556a614d83dc97/run.go#L12-L25)). These are viable when exact continuation is a requirement; they do not establish that it is one here.

## Material counterexamples for a fresh-validation restart

The requested recovery contract can deliberately restart an interrupted coding run by inspecting the repository/artifact state, rechecking the goal, and rerunning validation. Do not add durable replay merely because an agent loop is long-running. The cases where a small persisted record or named stage becomes material are narrower:

1. **Human answer changes authority.** A reviewer or user grants/rejects a scope, deployment, or destructive action. Repository state cannot derive that decision. Persist the answer and its context, or stop and ask again; replay cannot manufacture it.
2. **External effect cannot be discovered or deduplicated.** A PR/deploy/API call may succeed before the local record is written. Use the provider's idempotency key, store/lookup an external identifier, or a compensating action. DBOS's shared-database transaction avoids this only for that database (`corpus/durable/dbos-go/workflow.go.txt:2692-2694`; [pinned source](https://github.com/dbos-inc/dbos-transact-golang/blob/ab56911fdd78552e1e7fe648cff7c831a1e760c8/dbos/workflow.go#L2692-L2694)), not a Git provider or deployment API.
3. **A validation result is being reused as if still current.** A green validator from before a repository update is stale evidence. The correct low-durability recovery is to rerun the named validator against current state. Persist only the evidence that must remain attributable or auditable, not its obsolete pass/fail as a scheduling shortcut.

These counterexamples argue for explicit effects/authority/freshness boundaries, not for serializing Go stack frames or adopting a general workflow server.

## Pre-execution visualization: what is honest

Ordinary code can support a useful **abstract possible-route outline**: AST/source inspection can show both syntactic branches, known direct calls, loops, and opaque dynamic calls. It cannot determine the concrete future branch or number of nodes when a tool/agent result decides them. Temporal's exclusive-choice sample makes that dependency visible (`corpus/durable/temporal-go/choice-exclusive-workflow.go.txt:24-46`; [pinned source](https://github.com/temporalio/samples-go/blob/a3d54e384db1842459c57df7d9d3db66b204d9db/choice-exclusive/workflow.go#L24-L46)); Restate's input-sized fan-out does too (`corpus/durable/restate-go/parallelizework-main.go.txt:24-53`; [pinned source](https://github.com/restatedev/sdk-go/blob/3f7852328e1fc1f5b0716d8a1e556a614d83dc97/examples/parallelizework/main.go#L24-L53)).

Prefect gives a particularly clean contrast: its documented visualization executes code outside tasks (`corpus/durable/prefect/visualize-workflow-structure.mdx:7-15`; [pinned source](https://github.com/PrefectHQ/prefect/blob/34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef/docs/v3/how-to-guides/workflows/visualize-workflow-structure.mdx#L7-L15)) and needs mock return values to select loop/conditional paths (`corpus/durable/prefect/visualize-workflow-structure.mdx:101-127`; [pinned source](https://github.com/PrefectHQ/prefect/blob/34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef/docs/v3/how-to-guides/workflows/visualize-workflow-structure.mdx#L101-L127)). The implementation builds edges from encountered tracked tasks, confirming this is a path trace (`corpus/durable/prefect/visualization.py:72-205`; [pinned source](https://github.com/PrefectHQ/prefect/blob/34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef/src/prefect/utilities/visualization.py#L72-L205)).

Therefore, a separate declared graph is optional product metadata, not a necessary duplicate of code. It can be justified if it declares goals, validators, authority boundaries, or effect classes that source structure omits. If the current graph already supplies those, retain it as that contract; if source-derived possible routes are enough, show an outline with opaque/dynamic markers. Neither representation should pretend to predict the agent's actual outcome.

## Proposed POC boundary and falsifiers

Proposal only: test a Go-authored loop with typed agent output, a goal/validator interface, ordinary branches and bounded loops, a source-derived possible-route outline, and restart that re-reads repository/artifact state then revalidates. Keep one explicit hook for persisting authority answers or external effect IDs when present; do not implement workflow replay.

The POC is falsified if any of these occur:

- recovery repeatedly cannot decide whether an external PR/deploy occurred because the provider lacks an idempotency/readback path;
- authority-bearing human answers cannot be made available on restart without unsafe repetition;
- recovery must preserve an intermediate result whose reproduction is too costly or impossible and repository/provider state cannot recover it;
- a source-derived outline cannot give users enough preflight understanding even with opaque/dynamic markers; or
- dynamic behavior needs long-lived waits/signals where exact continuation, rather than fresh validation, becomes a stated product requirement.

Current status evidence: all four repositories were active and non-archived in API checks on 2026-09-08. Temporal Go `main` was at `9aae5eec926f359f1fb986b7d3125bd968d1067a` (latest release `v1.48.0`, 2026-08-18); Restate Go `3f7852328e1fc1f5b0716d8a1e556a614d83dc97` (`v1.0.4`, 2026-08-21); DBOS Go `ab56911fdd78552e1e7fe648cff7c831a1e760c8` (`v1.3.0`, 2026-09-02); Prefect `34ee7056b2b83cb0eaaa80378c5202d6c0b5eeef` (`3.8.5`, 2026-09-03). See `corpus/durable/status.json` for dates and licenses.
