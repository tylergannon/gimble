# Source leaf: context scopes and physical layers

## Purpose

Route to the [scope design note](../../../../ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md)
and [nested/parallel scope example](../../../../examples/go-workflows/scopes/README.md).
Find Tyler's scope direction and the current declared-key and per-call snapshot
rules, with earlier filesystem research kept separate.

## Key concepts

- Current [scope ownership rule](../../../../ephemeral/projects/gimble/programmatic-workflows/CONTEXT-OWNERSHIP.md):
  `DeclareContext` fixes ownership before assignment; only that scope may
  edit the returned key. Any name collision qualifies ALL effective keys with scope IDs.
  The local-only analyzer was rejected and is retained unadopted at Tyler's
  request. Call-graph analysis is deferred; agent/backend effects remain stubs.
- Automatic chapter/sprint scopes, arbitrary named branches, trusted read
  visibility, and explicit write ownership:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:8-15`.
- A child retains its parent, resolves current ancestor values at each agent
  entry, and owns a named, nested directory. Each call receives a fixed view;
  writes use declared keys and do not automatically promote results:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:19-35`.
- Item contexts and validators agree; retries reuse a scope within one iterator
  activation. Initial validation creates every sibling scope, but later calls
  refresh inherited data; typed replies require explicit context writes:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:37-50`.
- Ordinary Go context selection restores the parent; `errgroup` owns joins and
  cancellation. Scope exit does not cancel goroutines or delete files:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:52-62`.
- The effective index routes to local and ancestor files, and each revision
  has a symlink view for native tools. Files are real and agent effects canned:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:64-77`.
- Afero's live-base Go overlay, Dagger's engine-backed snapshots, and Linux
  OverlayFS's mounted layers; the agent recommends manifest layers for this
  prototype. No dependency is adopted:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:79-109`.
- Tyler's chosen symlink directories without FUSE, implemented `ContextSnapshot.View`,
  coherent publication, and replacement writes through `SetContext`:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:111-139`.
- Earlier FUSE comparison: memory/opaque storage, caller identity, caches,
  macOS backends, and filesystem lifecycle obligations:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:141-163`.
- Earlier parent/child override discussion, superseded by declared ownership
  and current ancestor reads at each call. Automatic invalidation of derived
  values remains unimplemented:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:165-183`.
- BranchFS at a pinned revision: frozen snapshots, explicit leaf commits,
  direct-parent-write versus sibling-commit conflict behavior, file-copy cost,
  sequential publication, and macFUSE requirements. Source inspection only:
  `ephemeral/projects/gimble/programmatic-workflows/CONTEXT-SCOPES.md:185-225`.

## Implementation entrypoints

- Parent-linked scope, readable physical directory, and immutable-by-API convention:
  `examples/go-workflows/internal/program/scope.go:12-77`.
- Declaration-bound ownership, write rejection, and effective inherited values:
  `examples/go-workflows/internal/program/context.go:97-188`.
- Item scope creation, metadata refresh, validation, and retries:
  `examples/go-workflows/internal/program/loop.go:39-131`.
- Automatic chapter/sprint wrappers:
  `examples/go-workflows/internal/program/loop.go:133-143`.
- Nested orchestration and arbitrary `errgroup` reviewer branches:
  `examples/go-workflows/scopes/main.go:16-109`.
- Complete revision directory with effective-value and authoritative-index links:
  `examples/go-workflows/internal/program/view.go:10-47`.
- For snapshot publication, prompt budgets, and the agent-entry boundary,
  continue to [filesystem context](context-files.md).

## Retrieval recipes

- For "where does this sprint's context come from?", read the iterator lifetime
  above, then follow `chapter.Context` and `sprint.Context` through the example.
- For "what can an agent see on disk?", inspect the revision-view decision
  and run `go run ./examples/go-workflows/scopes` from the repository root.
  The demo retains files and prints exact prompts and decoded effective values.
- For transparent filesystem branching, compare the three primary-source
  leads; a mutable lower layer does not by itself provide snapshot consistency.
- For FUSE or a different view per caller, read the revision-directory section.
  Tyler chose ordinary symlink views. Updates use `SetContext`; direct writes
  through these links are not copy-on-write.
- For "does a parent change overwrite my child's value?", read the precedence
  and ownership notes. A descendant cannot write an ancestor-owned value.
  The next call sees current inherited values; already-published views stay fixed.
- For static checking, read the ownership note's status section. No adopted
  analyzer validates ownership; retained local-only code does not meet Tyler's
  requirement to traverse the call graph.
- For BranchFS, read the pinned source assessment before inferring behavior
  from its branching or atomic-commit description. Its commit counter tracks
  sibling commits, not direct parent edits; no package or driver is adopted.

## Themes

Automatic item context; explicit arbitrary scopes; declared keys; current ancestor reads; retry
lifetime; trusted visibility; physical ancestry; local write ownership;
manifest overlays; materialized symlink views; ordinary Go cancellation and concurrency.
