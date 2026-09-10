# Adversarial Review: Go workflow context consensus (round 01)

Date: 2026-09-09 (local)
Reviewer: Claude (adversarial-review skill, performed directly, read-only except this file)
Prior round: `ephemeral/reviews/202609092010-adversarial-review-round-01.md`

## Review target

- Repository: `/Users/tyler/.codex/worktrees/a5b1/gimble`, branch `codex/go-workflow-sketches`
- Range: `074b302..HEAD` (`de7b770`, `9c06425`, `370f014`, `20f281c`, `cf28d8a`) plus the
  uncommitted working tree (29 modified files under `docs/semantic-index/`,
  `ephemeral/projects/`, `ephemeral/worklog/`, and `examples/go-workflows/`; `ephemeral/reviews/`
  untracked).
- Subject: the Go workflow design and executable examples (`examples/go-workflows/`), the stub
  primitive package (`internal/program`), the unadopted `contextcheck` analyzer, the design
  documents that describe the context/ownership contract, and the semantic-index routing that
  points at them.

Narrowing instructions in the launch prompt: none. The read-only boundary, the artifact path,
and the five-finding cap are valid operating constraints and were honored.

## Authoritative intent (reconstructed)

From the caller's restated conversation context, `AGENTS.md`, the worklog
(`ephemeral/worklog/202609091800-go-workflow-sketches.md`), `CONTEXT-OWNERSHIP.md`,
`CONTEXT-SCOPES.md`, `CONTEXT-FILES.md`, and the designer rules
(`ephemeral/projects/gimble/workflow-designer/rules.md`, "least engineering first"):

1. Simple, compiling Go workflow sketches with stubbed agent and backend operations.
2. Ordinary Go concurrency; workflow-owned roles and argument shapes.
3. Filesystem context with index projection into prompts; automatic scopes for loop items plus
   arbitrary user scopes.
4. Declared value ownership independent of assignment order; inherited values refreshed at each
   agent entry into one fixed view for that call.
5. Any adopted ownership analyzer must traverse the call graph; the local-only prototype stays
   preserved but unadopted. Schema emission deferred.
6. No speculative complexity, no new requirements, no production-backend expansion, no planning
   machinery.

## Evidence inspected

Documents: `AGENTS.md`; `docs/semantic-index/programmatic-workflows/README.md`, `recipes.md`,
`sources/compiling-examples.md`, `sources/context-files.md`, `sources/context-scopes.md`
(anchors spot-checked against current line ranges; all resolved); `.semantic-index/evals.jsonl`
diff; `ephemeral/projects/gimble/programmatic-workflows/{CONTEXT-OWNERSHIP,CONTEXT-SCOPES,CONTEXT-FILES}.md`;
the worklog; in-range diffs to `docs/direction.md`, `docs/workflows-as-programs.md`,
`docs/README.md`, `examples/README.md`; the round-1 review.

Code (read in full): `examples/go-workflows/internal/program/{context,scope,view,loop,runtime,main}.go`
and all seven test files (1014 lines); examples `scopes/`, `context/`, `sprints/`, `bakeoff/`,
`critique/`, each with README and fixtures; `contextcheck/main.go`, `contextcheck/README.md`,
`contextcheck/testdata/conflict/conflict.go`; `internal/contextcheck/analyzer.go`,
`analyzer_test.go`, and its `testdata/src` stub package; `go.mod` diff; `examples/go-workflows/README.md`.

Verification performed:

- `go build ./examples/go-workflows/...`, `go vet ./examples/go-workflows/...`, `gofmt -l`: clean.
- `go test -race -count=1 ./examples/go-workflows/...`: all packages pass.
- Fresh runs of all five examples succeeded. The scopes, context, and sprints outputs match the
  supplied artifacts stage for stage (scope names, revisions, `chapter_focus` values), and the
  retained directories under `/private/tmp/gimble-scopes-*` contain the symlinked `values/` and
  `index.json` files the reports describe.
- `go mod tidy -diff`: reports a pending change (see finding 4).
- Behavioral probes in a scratchpad copy of the package (no project files touched):
  qualification of effective keys under nested same-name declarations; a loop body redeclaring
  the loop's metadata key; `Sprints` without an attached store; index/view churn under a
  snapshotting validator.

## Round-1 status

- F1 (first-writer ownership): resolved. `DeclareContext` returns a scope-bound `Key`;
  `SetContext` rejects a key owned by another scope (`context.go:129-135`); tests cover
  descendant-before-owner and ancestor-writes-child cases.
- F2 (frozen inheritance): resolved. `Codergen` snapshots at entry (`runtime.go:37-70`); the
  artifact and a fresh run show sprint 2 receiving the chapter's post-sprint-1 update at
  revision 16; a running call keeps its fixed view (`runtime_test.go`).
- F3 (schema emission): explicitly deferred by Tyler. Not re-reported.
- F4 (over-engineering density): not addressed. See finding 2.
- F5 (temp-dir names / per-sweep metadata rewrite): dir naming fixed (`scope.go`, `scopePrefix`);
  the unconditional rewrite remains. See finding 3.

## Findings

### 1. Issue: one same-name declaration anywhere in the chain renames every effective key, contradicting the documented contract and breaking documented file paths

Evidence: `examples/go-workflows/internal/program/context.go:168-188`. The collision scan sets a
single `qualify` flag for the whole chain (`:173-176`) and, when set, prefixes every name with
`scope-<id>.` (`:181-183`), not only the colliding ones.

Reproduction (scratchpad probe against a copy of the package): root declares `goal`; a chapter
scope declares `sprint`; a nested sprint scope declares `sprint`. The inner snapshot's view
contains `values/scope-0.goal.json`, `values/scope-1.sprint.json`, `values/scope-2.sprint.json`,
and its prompt renders `scope-0.goal = "deliver"`. Sibling scopes without a nested collision keep
plain `values/goal.json`.

Contradicted statements:

- `CONTEXT-OWNERSHIP.md:26` "a name collision qualifies effective index keys with scope IDs";
  `:34` "Cross-scope name collisions are qualified in the effective view".
- `CONTEXT-SCOPES.md:26` "effective indices qualify collisions with scope IDs so neither value
  hides the other"; `:123` "Native tools can read `values/goal.json`".
- `examples/go-workflows/README.md:72` "cross-scope name collisions are qualified with scope IDs".
- `examples/go-workflows/scopes/README.md:38` "scope-qualified index keys when necessary".

Impact: the design advertises nested same-kind loops (`Sprints` inside `Sprints`, both keyed
`sprint`) and native-tool access by stable path. Under that advertised shape, every inherited,
non-colliding value changes name and path for the inner scope only, so an agent instruction or
tool that reads `values/goal.json` works in one scope and fails in its child. The scope-id
prefix is an opaque allocation counter (`scope.go:33`), so the renamed path is not predictable
from the workflow source either. Either the code should qualify only the colliding names (what
the four documents say) or the documents should stop promising stable `values/<key>.json` paths.
The nearest test (`loop_scope_test.go`, nested Sprints) asserts only that both `sprint` values
are visible, so the rename of unrelated keys is untested and undocumented.

### 2. Issue: the stub primitive package is still built to production defensiveness, against "least engineering first" and the explicit rejection of speculative complexity

Evidence, measured on the working tree:

| Unit | Lines |
| --- | --- |
| `internal/program` non-test (`context` 348, `loop` 143, `runtime` 125, `scope` 78, `main` 57, `view` 47) | 798 |
| `internal/program` tests | 1014 |
| Workflow algorithm files across the five examples | 388 |

Cancellation checks (`ctx.Err()`): `context.go` 10 (`:66,87,109,140,149,156,199,204,236,250`),
`runtime.go` 6 (`:41,61,74,84,103,112`), `loop.go` 3, `scope.go` 3, `view.go` 3. Ten explicit
rollback branches remove partially written files or directories (`context.go` publication
rollback `:236-258`, `scope.go:44-70`, `view.go`). Several checks bracket a synchronous stub
callback that already received `ctx` (`runtime.go:61` then `:74`; `:103` then `:112`).

Round 1 reported this as an issue; the ownership fix grew `context.go` from 320 to 348 lines and
added a second store-identity check and a second sanitization path rather than removing any.
Impact: the primitives now outweigh the workflows they exist to sketch by two to one, and
reading the "simple compiling sketches" requires understanding an atomic-publication,
tree-wide-revision cache, and rollback protocol that Tyler did not ask for and said not to
build. This is unrequested infrastructure, not a correctness defect; it is material because the
brief's stated purpose is legibility of the workflow shape.

### 3. Nitpick: every validation sweep rewrites unchanged loop metadata and bumps the tree-wide revision, invalidating every scope's snapshot cache

Evidence: `loop.go:82-88` calls `SetContext` for every item on every sweep regardless of change;
`SetContext` increments `tree.revision` (`context.go:161`); `SnapshotContext` caches on
`store.snapshot.Revision == store.tree.revision` (`:207`), so any write in any scope forces a
re-index and a new `view-NNNNNN-*` directory for every other scope on its next call.

Reproduction (probe): three items, three loop bodies, a validator that snapshots once per sweep
produced 12 `index.json` files and 12 view directories where 3 to 4 would carry the same
information. In the scopes artifact the two parallel reviewers received revisions 10 and 11 for
identical inherited content because each reviewer's own `review_focus` write invalidated the
sibling's cache.

Impact: harmless in the stubs, but `Call.Context.Revision` no longer identifies a scope's
effective content, which is the property the docs lean on for "one fixed view per call". Skip
the write when `Item`/`Attempt` are unchanged (round-1 F5, still open).

### 4. Nitpick: `go.mod` is untidy because the unadopted analyzer imports `golang.org/x/tools` directly

Evidence: `go.mod:53` marks `golang.org/x/tools v0.49.0 // indirect`, but
`examples/go-workflows/contextcheck/main.go:10` and `internal/contextcheck/analyzer.go` import
`golang.org/x/tools/go/analysis/...` directly; `go mod tidy -diff` shows the promotion to a direct
requirement. The range diff promoted `x/sync` to direct but not `x/tools`.

Impact: the next `go mod tidy` will change `go.mod` as a side effect of preserving code Tyler
said is unadopted. Its CLI fixture (`contextcheck/testdata/conflict/conflict.go`) and analyzer
stub package (`internal/contextcheck/testdata/src/.../program.go`) still use the superseded
string-key API, so the prototype can no longer fire against the current `Key` API
(`analyzer.go` requires an `*ssa.Const` string at `Args[1]`); its README acknowledges this.
Either tidy `go.mod` now or move the prototype behind a build tag so it stops shaping the main
module's requirements.

### 5. Nitpick: automatic item scopes silently require `NewContext`, and a loop body can redeclare the loop's own metadata key

Evidence: `loop.go:65-79` always calls `Scope` for `Chapters`/`Sprints`; without a store the
first yield is the error "no context store is attached" (`scope.go:21-24`), while plain `Loop`
works without one. `examples/go-workflows/README.md:88` says the two helpers "create item
scopes automatically" without stating the precondition. Separately, a body may call
`DeclareContext(it.Context, "sprint")` and overwrite the iterator's `itemState`; the iterator
then silently overwrites it back on the next sweep (`loop.go:84`). `loop.go:17` calls the
binding "loop-owned metadata", but nothing owns it against the body.

Impact: both are legibility gaps rather than defects. One sentence in the README for the
precondition, and either a reserved-name check in `DeclareContext` or a comment retracting
"loop-owned", would close them.

## Requirement coverage (no finding)

Compiling stubs, `errgroup` fan-out, workflow-owned roles and input shapes, filesystem context
with prompt projection and external spill, automatic and arbitrary scopes, declared ownership
independent of order, refresh at agent entry with a fixed per-call view, unadopted analyzer with
the call-graph requirement recorded, and schema emission deferred are all present and verified
by tests, fresh runs, and the supplied artifacts. No race or crash was found under `-race`; lock
discipline in `context.go`/`scope.go` never runs a callback under the tree mutex.

## Outcome

material findings remain
