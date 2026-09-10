# Context values owned by one scope

Tyler's direction, 2026-09-09: a value can be edited only from its owning scope.
A descendant that tries to write an ancestor's value gets an error. Catch this
statically with a Go analyzer where possible.

This replaces the agent's proposed rule that parents must freeze while child
work runs. Ownership restricts where writes occur; it does not require a new
scheduler, filesystem, or agent-role permission system.

## Runtime rule in the examples

Tyler's "Fix" accepted declared ownership and refreshed inheritance after
review exposed first-writer ownership and frozen child scopes as mistakes.
The earlier first-writer rule is superseded.

`DeclareContext(ctx, name)` returns an opaque `Key` bound to that scope before
it has a value. `SetContext(ctx, key, value)` accepts writes only from the
declaring scope; it never acquires ownership. Ordinary Go contexts derived
with cancellation or deadlines still refer to the same scope.

- A descendant cannot write an ancestor's key, even before its first value is
  assigned. An ancestor likewise cannot write a child's key.
- Repeating a declaration in one scope returns the same key. The same display
  name in another scope denotes a different binding, not an override. Both
  remain visible; any name collision qualifies ALL effective keys with scope
  IDs, avoiding generated-name aliases. Nested loops can both be named `sprint`.
- Rejected writes do not change files, revisions, or the last published view.
- This is a scope rule, not a goroutine rule. Goroutines sharing the same scope
  can still write that scope's values; the library serializes writes, but the
  workflow must choose meaningful ordering if they target the same value.

Names differing only in case are rejected within one scope. A cross-scope
collision qualifies every effective key; use the current index for filenames.

The filesystem remains managed through `SetContext`. Directly editing a
symlink target bypasses the API, so agent tools continue to treat context files
as read-only and keep working files separately. This is trusted routing, not
an operating-system security boundary.

## Workflow shape

The root owns `goal` and `focus`; each chapter owns its `chapter` metadata and
`chapter_focus`; each sprint owns its `sprint` metadata. Parallel reviewer
scopes independently own `review_focus`. They add information without hiding
inherited values. The loop's attempt number lives inside its chapter or sprint
JSON object, so nested loops do not shadow a shared `attempt` key.

The parent may continue updating its own values. Scopes retain their parent
relationship; each agent entry resolves current inherited values and publishes
one fixed prompt/index view. The received `Call.Context` records that exact
snapshot. Already-running calls keep their earlier view. The scopes example
updates chapter context after sprint 1, and sprint 2 receives the update even
though its scope was created during initial validation.

There is no automatic result promotion or invalidation of derived findings.
Those remain ordinary decisions in the workflow, not new backend machinery.

## Static checking

Tyler rejected the agent's intraprocedural analyzer as insufficient and liable
to produce false confidence. Any adopted analyzer must traverse the call graph;
that work is deferred. The immediate task is compiling workflows with stubbed
agent/backend operations and the small context ownership rule above.

The agent had already written a local-only `contextcheck` prototype when Tyler
corrected the scope. At his instruction to preserve written software, its code
and tests remain under `examples/go-workflows/`. It is unadopted reference work,
not an ownership validator or an enabled workflow check. Its successful exit
must not be interpreted as evidence that a workflow obeys scope ownership.
It describes the superseded string-key API, not the current declared-key API.
