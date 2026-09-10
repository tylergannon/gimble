# Adversarial review, round 01: Go workflow design (concept and feasibility)

- Date: 2026-09-09 20:10 local
- Repository: `/Users/tyler/.codex/worktrees/a5b1/gimble`, branch `codex/go-workflow-sketches`
- Target: commits `074b302..cf28d8a` (e92b827, ad82fee, de7b770, 9c06425, 370f014, 20f281c, cf28d8a)
- Reviewer stance: independent, read-only; this file is the only write.
- Grading lens: Tyler asked mid-review for a grade on the concept work and feasibility, not on working software. I treated that as emphasis, not as an exclusion: code-level defects were still inspected and would have been reported had any ranked in the top five. None did.
- Narrowing instructions ignored: none. The launch prompt supplied intent and operating constraints only.

## Authoritative intent used for scope

From the launch prompt, AGENTS.md, `ephemeral/projects/gimble/workflow-designer/rules.md`, and the worklog at `ephemeral/worklog/202609091800-go-workflow-sketches.md` (lines 1-2, 11, 13-14, 16, 18, 23, 28-29):

1. Simple, readable Go workflow examples; stubbed backends; ordinary Go concurrency including `errgroup`; informed by the recovered POC; may coexist with the DAG language.
2. Declared argument shapes supported by Polytype; workflow-defined roles; dependable calling conventions for a small curated builtin library.
3. Filesystem-backed context: small values inline, larger material through an index; agent calls await a coherent context/index view.
4. Chapters and sprints create scopes automatically; arbitrary scopes possible.
5. One owning scope per value; a descendant writing an ancestor-owned value gets an error.
6. Any adopted static ownership analyzer must traverse the call graph; the intraprocedural prototype is rejected but preserved.
7. Simplicity, repeatedly ("least engineering first"; supervisors curtail belt-and-suspenders work).

## Evidence inspected

- AGENTS.md; `docs/five-arts.md`; `docs/direction.md`, `docs/workflows-as-programs.md`, `docs/README.md`, `examples/README.md` (diff in range).
- Semantic index: `docs/semantic-index/programmatic-workflows/README.md`, `routes/go-library/index.md`, `routes/context-and-indexing/index.md`, and the leaves `compiling-examples.md`, `context-files.md`, `context-scopes.md`, `poc-recovery.md`, `poc-workflows.md`, `go-library-sketches.md`, `concurrency-shapes.md`.
- Design notes: `CONTEXT-OWNERSHIP.md`, `CONTEXT-SCOPES.md`, `CONTEXT-FILES.md`, `POC-RECOVERY.md`, `POC-WORKFLOWS.md`, `GO-LIBRARY-SKETCHES.md`, `CONCURRENCY-SHAPES.md`, and the worklog.
- Source: all of `examples/go-workflows/` (bakeoff, critique, sprints, context, scopes, contextcheck, `internal/program`, `internal/contextcheck` and its testdata), plus `go.mod`.
- Proof run from the target checkout: `go build ./examples/go-workflows/...` (exit 0), `go vet ./examples/go-workflows/...` (exit 0), `go test -race -count=1 ./examples/go-workflows/...` (all packages ok). The "compiling examples" claim holds.

## Overall assessment

The concept is sound and feasible. Ordinary Go plus `errgroup`, a checklist iterator that owns `done`, and a scoped filesystem context carried on `context.Context` is a coherent authoring model, and the five workflow algorithms read well (bake-off 76 lines, critique 80, sprint execution 63, nested scopes 90, context walkthrough 44). The synchronous snapshot at `Codergen` entry satisfies the "await a coherent view" requirement simply. The documentation is careful about what is Tyler's direction versus agent proposal.

Two design-level weaknesses in the context model stand out: ownership is assigned by write order rather than by scope position, and frozen-at-creation inheritance combined with eager sibling scope creation leaves no context channel between the items of one loop. The calling convention is established by the programs themselves; only Polytype schema emission is still to show. The primitives are heavier than the "simple examples" framing warrants. Details follow.

## Findings

### 1. Issue: ownership is decided by write order, not by scope position

**Category:** incorrect implementation of a requirement (design-level), with feasibility consequences for the deferred analyzer.

**Requirement:** "A value may be edited only from its owning scope: descendants attempting to write an ancestor-owned value must receive an error." (launch prompt; worklog line 28).

**What the design does:** the owner is whichever scope succeeds in writing first, and the check is symmetric in both directions (`examples/go-workflows/internal/program/context.go:126-148`; `CONTEXT-OWNERSHIP.md:13-19`). The note itself labels the both-orders rule an "implementation choice" (`CONTEXT-OWNERSHIP.md:28-29`).

**Why it matters:** the rule Tyler stated presupposes that the ancestor is the owner. Under first-writer semantics a descendant that writes first becomes the owner and the *ancestor* receives the error. In sequential code the author can reason about order; in the parallel case the design explicitly supports (`scopes/main.go:64-80`, reviewer scopes racing alongside parent work), the owner is the race winner. The test `ownership_test.go:105-143` asserts exactly one winner but cannot say which, because the model does not say which. A workflow can pass its tests and fail in production with an ownership error at the parent's write.

**Feasibility impact:** the deferred call-graph analyzer (requirement 6) must then prove the absence of *any* pair of writes to one key along an ancestor/descendant chain, interprocedurally, through `errgroup` closures and range-over-func iterator bodies, and it is unsound for dynamic keys already used in the examples (`context/main.go:38` writes `note.Key`). A model where ownership is positional or declared (for example, the scope that creates a key or a small per-scope "owns" list) would make the runtime check trivial and the static check a local "is this write in the declaring scope" question, which is the kind of thing a call-graph traversal can answer cheaply.

**Suggested direction (not implemented):** define the owner structurally rather than by race; keep the descendant-write error; drop the reverse-direction "ancestor cannot claim after descendant wrote" extension unless Tyler wants it.

### 2. Issue: frozen inheritance plus eager sibling scopes leaves no context channel between loop items

**Category:** critical antipattern in the context model (concrete failure described), arising from agent choices rather than Tyler's direction.

**Evidence:**
- `Loop` creates every item's scope during the first validation sweep, before any work is selected (`internal/program/loop.go:59-75`).
- `Scope` copies the parent's values at creation and never re-reads them (`internal/program/scope.go:55-59`; `CONTEXT-SCOPES.md:41-45, 160-166`).
- Ownership forbids a child writing upward (finding 1).
- The worklog records the consequence as something to "document ... instead of implying dynamic inheritance" (line 19), and the ownership note says the parent "may continue updating its own values" (`CONTEXT-OWNERSHIP.md:45-48`).

**Concrete failure:** in `scopes/main.go`, a chapter that learns something after sprint 1 (a reviewer finding, a changed focus) and writes it to `chapter.Context` will never have it reach sprints 2..N, because their scopes were frozen in the first sweep. Sprint 1 cannot write it to the chapter either. The only remaining path is Go variables interpolated into prompt strings, which `scopes/README.md:9-10` presents as exactly what the design avoids, and which bypasses the filesystem index the "one context organized around success" direction depends on (`docs/workflows-as-programs.md`).

**Why it is a design gap, not a documentation note:** the parent's right to update its own values is nearly meaningless to descendants that already exist, and in a checklist loop every descendant already exists by the time the parent learns anything. Freshness is deferred as "a separate policy" (`CONTEXT-SCOPES.md:176-181`), but with ownership in place the two policies together close every channel. This needs a decision: lazy scope creation at selection time, re-snapshotting inherited values at each agent entry (the unaccepted proposal at `CONTEXT-SCOPES.md:167-174`), or an explicit per-item handoff value the loop owns. Any of these is feasible within the present primitives.

### 3. Nitpick: Polytype schema emission is not demonstrated

**Category:** deferred scaffolding, downgraded from "issue" after Tyler's correction in review.

The five programs already establish the calling convention Tyler asked for: a per-workflow argument struct in `input.go`, workflow-defined role constants, one function signature (context, runtime, typed input, typed result), and `-example` / `-input` on every binary (`internal/program/main.go`). The one piece not shown is a Polytype schema emitted from those structs. Polytype is already a dependency (`go.mod:8,21`), the structs are plain, and this is a single call; it belongs on the follow-up list, not in the design critique. An earlier draft of this finding measured the examples against the rejected sketch's catalog API, which was the wrong yardstick.

### 4. Issue: the primitives are heavier than the "simple examples" brief

**Category:** over-engineering.

**Measurements (target checkout):**

| Component | Lines |
|---|---|
| Workflow algorithms (5 files) | 353 |
| `internal/program` primitives (non-test) | 746 |
| `internal/program` tests | 1041 |

- Cancellation checks in the primitives: `context.go` 9, `runtime.go` 6, `loop.go` 3, `scope.go` 3, `view.go` 3. `NewContext` checks `ctx.Err()` three times in 30 lines (`context.go:63, 84`); `SnapshotContext` five times; `Codergen` checks before and after a synchronous callback that already received `ctx`.
- Rollback branches on every temp file and directory (`context.go:81-86, 141-144, 216-227`; `view.go:24-28`).
- A case-insensitive key collision guard for portable symlink views (`context.go:127-131`), a per-revision view directory per scope, and a three-stage spill policy with tie-breaking (`context.go:233-289`).

**Why it matters:** Tyler's rules name "least engineering first" and supervisors that "curtail belt-and-suspenders work" (`rules.md`, "No budgets"). The examples' stated purpose is to make the authoring shape inspectable with stubs. A reader who opens `internal/program` to learn what `SetContext` or `Scope` *means* has to read past defensive code that a stub does not need. Some of this traces to Tyler's direction (symlink views, inline/spill budgets) and is legitimate; the cancellation and rollback density is not. This ranks below findings 1-3 because the workflow-facing code, which is what the brief asked to be readable, is readable.

### 5. Nitpick: scope identity is illegible on disk and overloaded in keys

- Physical scope directories are `MkdirTemp` names (`scope-XXXXXXXX`) nested under one another (`scope.go:42`); the scope's human name lives only inside `index.json`. Tyler asked that "the filesystem should reflect those nested and parallel branches" (`CONTEXT-SCOPES.md:10-12`). A sanitized name prefix on the directory would satisfy that without changing the model.
- `LoopOptions.Scope` is both the scope-name prefix and the context key the loop writes (`loop.go:66-68, 78`). Two loops of the same kind cannot nest (an inner `Sprints` under a sprint item would hit the ownership error on `sprint`), and the loop rewrites that key unconditionally on every sweep (`loop.go:78`), bumping the revision and forcing a re-index before every validation agent call even when nothing changed. Harmless in the stubs; worth fixing before the primitives are copied anywhere.

## Outcome

material findings remain
