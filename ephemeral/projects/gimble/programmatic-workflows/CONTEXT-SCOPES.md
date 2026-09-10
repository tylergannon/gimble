# Context scopes and physical filesystem layers

Design/example note, 2026-09-09. This extends
[filesystem-backed context](CONTEXT-FILES.md). Tyler's requested direction is
separated below from the prototype's chosen semantics. These examples do not
establish a production API or native-agent runtime parity.

## Tyler's direction

Chapter and sprint iteration should supply the appropriate context
automatically, and ordinary Go should also be able to create arbitrary named
scopes. The filesystem should reflect those nested and parallel branches.
Agents can be trusted to follow the supplied routes and work in their assigned
layers. Reading another visible layer is not itself a danger to design around;
the immediate concern is clear context and write ownership.

## Prototype contract

The root starts with `NewContext` and workflow-supplied goal data.
`Scope(ctx, name)` copies the parent's effective values and revision under
the parent lock, reuses their immutable file paths, and creates a private
mutable map and a new directory physically inside its parent directory.
The name is recorded in the snapshot/index scope path; generated directory
names identify the physical layers.

`SetContext(child, key, value)` writes a new file in the child's directory.
Its effective index resolves that key to the local value, shadowing the
inherited reference. Other inherited keys still point to ancestor files.
Later parent updates do not change an existing child; child writes do not
change siblings or automatically promote results to a parent. Files are
immutable by API convention, not protected from direct filesystem edits.

`Chapters` and `Sprints` wrap the same Go iterator. Each yielded
`Iteration.Context` belongs to that item; its validation callbacks receive
the same context. The workflow passes `chapter.Context` into `Sprints`, then
`sprint.Context` to its agent calls. Item and attempt metadata are refreshed,
while explicitly stored feedback survives retries within that iterator
activation. A typed agent reply enters later context only when the workflow
explicitly calls `SetContext` with it.

The initial validation sweep creates every item's scope, including siblings
not yet selected for work. This freezes their inherited values at that time.
Updating the parent afterward will not reach those already-created siblings.
Starting a new iterator activation creates new item scopes; retry reuse does
not extend across separate iterator invocations.

Arbitrary scopes use the same `Scope` function and ordinary `errgroup` for
parallel work. Cancellation follows the supplied Go context. Returning to a
parent means using the original context again; there is no global stack or
pop operation. Scope exit does not cancel escaped goroutines or remove files.
The caller joins child work, and retained snapshots remain readable.

Each agent entry still obtains one coherent snapshot before its callback.
The byte budgets and deterministic index remain the
[earlier prototype](CONTEXT-FILES.md#what-the-current-example-explores).
Scope isolation does not decide which goals, rules, or evidence deserve
attention, and does not establish that the resulting context is effective.

## A manifest overlay, visible on disk

The scope directory tree follows the workflow's ancestry. Each layer owns
its new values and indices; its index provides the effective view of local
and inherited data. `ContextSnapshot.View` materializes that view as an ordinary
directory of symlinks, described below. Native file tools can list and read
its meaningful filenames without interpreting the manifest first. The view
requires no mounted union filesystem and is not a security boundary.

The [scopes example](../../../../examples/go-workflows/scopes/README.md) combines
nested chapter/sprint loops with parallel arbitrary reviewer scopes. It
captures exact delivered prompts, scope paths, index/view paths, and decoded
effective values, including the parent view after child work. Context files
are real and retained; agent replies and command outcomes are canned.

## Three filesystem implementations to inspect

- **Afero `CopyOnWriteFs`:** a small Go union filesystem whose reads prefer
  the writable upper and whose inherited writes copy upward. Its `Fs`
  interfaces compose into branches, but an unmodified base read remains live;
  this is not a snapshot over a mutable parent. There are no whiteouts:
  removing an override reveals the base again, and base-only delete/rename
  is unsupported. Native shell tools do not use the Afero view. Its CI covers
  macOS, Linux, and Windows. [Source](https://github.com/spf13/afero/blob/master/copyOnWriteFs.go),
  [platform CI](https://github.com/spf13/afero/blob/master/.github/workflows/ci.yaml).
- **Dagger `Directory`:** derive child snapshots from an immutable parent,
  then branch again. File replacement/removal changes the child result;
  the implementation commits new snapshots from parent references. Host
  inputs are lazy, so the intended base must be captured before claiming an
  exact fork point. Dagger supports macOS but requires an engine/container
  runtime. [Directory API](https://docs.dagger.io/api/reference/directory/),
  [snapshot implementation](https://github.com/dagger/dagger/blob/main/core/directory.go),
  [installation](https://docs.dagger.io/cli/install/).
- **Linux OverlayFS:** transparent merged directories, upper-name precedence,
  copy-up writes, whiteout deletions, and stacked lower layers. Each branch
  needs its own writable upper/work layer. It requires Linux mounts and
  supporting filesystem features, rather than native macOS file operations.
  Changing underlying layers while mounted is explicitly unsupported; the
  overlay itself does not freeze a changing parent.
  [Kernel documentation](https://kernel.org/doc/html/latest/filesystems/overlayfs.html).

**Agent recommendation:** keep manifest layers for this prototype. They
preserve physical ancestry, explicit write ownership, and stable inherited
references without a new runtime dependency. Transparent mounting and
deletion/tombstone semantics can be designed if a later workflow needs them.
No package has been adopted from this bounded source inspection.

## FUSE versus an ordinary revision directory

Tyler raised a custom FUSE filesystem over memory or opaque storage, with
views depending on the requestor, and asked whether a simpler option exists.
He then chose ordinary directories of symlinks, with changed files replaced,
and concluded that FUSE is unnecessary for this direction.

The examples now implement a directory per scope revision. `ContextSnapshot.View`
names `view-<revision>-<id>/`, containing `values/<key>.json` symlinks for every
effective key and an `index.json` symlink to the authoritative `ContextSnapshot.Index`.
Inherited and local values are both visible through meaningful filenames;
the view adds directories and links without copying the original value bytes.
Native tools can read `values/goal.json` or list `values/` directly.

`SnapshotContext` prepares every link before publishing the snapshot, and
`Codergen` waits for that snapshot before invoking the agent callback. The
context prompt points at `View/index.json`. Explicit revision paths keep scope
identity stable across tools and subprocesses, while earlier views remain
readable. Parent snapshots and private scope writes retain their existing
semantics; this adds a filesystem view to the manifest rather than a mount.

Context is read-only by convention. Writes go through `SetContext`, which
writes a new immutable value file; the next snapshot creates a new view whose
link resolves to that replacement. Editing through an existing symlink would
modify its target and can change ancestor data: symlinks do not provide
automatic copy-on-write. Direct tool edits would require separate writable
files and a publication contract, neither of which this example implements.
Agent replies and command effects remain canned; filesystem views do not
establish native-agent runtime parity.

The following FUSE research explains the earlier comparison; no driver or
filesystem dependency has been adopted.

FUSE can implement a memory-backed view, but selecting that view by PID adds
complexity. Linux libfuse exposes caller UID/GID and thread PID, with exceptions
for writepage operations. Name, attribute, and data caches can also bypass a
fresh userspace lookup/read. The inference is to prefer explicit scope paths,
distinct inode identities, or separate mounts over different answers for the
same path based solely on PID.
[Request context](https://libfuse.github.io/doxygen/structfuse__context.html),
[cache configuration](https://libfuse.github.io/doxygen/structfuse__config.html),
[FUSE I/O modes](https://kernel.org/doc/html/latest/filesystems/fuse/fuse-io.html).

On macOS, macFUSE documents a kernel-extension VFS backend and an FSKit
backend; the latter lacks `fuse_context_t` and requires `/Volumes` mount points.
Caller-dependent behavior is therefore also backend-dependent.
[macFUSE backends](https://github.com/macfuse/macfuse/wiki/FUSE-Backends).

A restricted read-only FUSE experiment is manageable. A dependable writable
filesystem is a separate subsystem: file-handle lifetime, concurrent writes,
rename/unlink semantics, caching, and mount lifecycle all need attention.
Opaque storage does not remove those obligations. Revisit FUSE if lazy access
or transparent filesystem behavior becomes worth that runtime commitment.

## Parent changes after a child override

Tyler raised what should happen when a parent changes after a child has
overridden some of its context. The current prototype freezes inheritance
when `Scope` is created. Later parent writes do not reach that child, even for
keys it has never overridden. A child override remains local, and each
published revision view remains stable under writes made through the API.

**Agent proposal, not accepted or implemented:** live lexical inheritance
could resolve each key from the nearest scope that defines it. A local value
wins; otherwise the next agent call resolves the latest applicable parent
value. That call would still receive one frozen effective snapshot, rather
than having its context change during execution. This would change when
inheritance is resolved; it must not be introduced silently into the current
scope API.

Precedence does not settle freshness. An intentional override should survive
an unrelated parent edit. A value derived from an older parent may instead
need to be invalidated, recomputed, or rebased; conflicting intent may require
adjudication. Those choices need an explicit workflow rule or dependency
relationship. Neither the symlink view nor "local wins" can infer them, and
the example implements no automatic invalidation, rebase, or adjudication.

## BranchFS: pinned source assessment

Tyler supplied [BranchFS](https://github.com/multikernel/branchfs/tree/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9).
This assessment pins `a4b6592d31aacb4d2f94d0c80f43d47d38063fa9`; it is source
inspection, not a FUSE runtime test. BranchFS is a Rust FUSE filesystem under
the [MIT license](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/LICENSE#L1-L9).
It exposes `@branch` paths, nested branches, frozen inherited snapshots, and
leaf commits into the immediate parent. This is snapshot isolation, not the
proposed live lexical inheritance above.
[Snapshot/commit model](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/README.md#L17-L19),
[@branch paths](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/README.md#L135-L168).

- **Parent-write precedence:** the source captures the parent's commit counter
  at fork and compares it at commit. Only child commits advance that counter.
  Thus parent value A → child override B → direct parent write C → child commit
  can replace C with B. If C instead arrives through a sibling commit, the later
  child commit returns `Conflict`, even when the sibling changed unrelated paths:
  the check is branch-wide. This is a source trace, not executed proof.
  [Fork counter](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/src/branch.rs#L479-L482),
  [conflict check](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/src/branch.rs#L803-L814),
  [root-parent increment](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/src/branch.rs#L857-L859),
  [nested-parent increment](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/src/branch.rs#L919-L923).
- **Copy cost:** snapshot creation walks the visible tree and uses `fs::copy`
  for regular files. There is no metadata-only guarantee; an OS may optimize
  the physical copy. [Tree walk](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/src/branch.rs#L723-L755),
  [copy implementation](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/src/storage.rs#L22-L32).
- **Publication:** staged files are renamed sequentially; rollback removes
  remaining temporary copies and restores staged deletions, without undoing
  earlier published replacements. The headline "atomic" does not establish a
  durable all-or-nothing transaction. [Publication and rollback](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/src/branch.rs#L185-L209).
- **macOS cost:** it requires macFUSE. CI integration runs on Ubuntu; release
  automation also builds macOS ARM. Those configurations do not demonstrate
  the macOS runtime behavior. [macOS setup](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/README.md#L65-L73),
  [integration job](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/.github/workflows/ci.yml#L32-L54),
  [release targets](https://github.com/multikernel/branchfs/blob/a4b6592d31aacb4d2f94d0c80f43d47d38063fa9/.github/workflows/release.yml#L17-L42).

BranchFS is a concrete comparison for writable filesystem branches and
explicit commits. It does not decide when a workflow should promote results,
invalidate derived context, or resolve conflicting intent. No dependency was
adopted, and this research changes neither the symlink-view example nor its
frozen inheritance contract.
