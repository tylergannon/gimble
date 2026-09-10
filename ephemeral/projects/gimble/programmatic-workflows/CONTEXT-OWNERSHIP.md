# Context values owned by one scope

Tyler's direction, 2026-09-09: a value can be edited only from its owning scope.
A descendant that tries to write an ancestor's value gets an error. Catch this
statically with a Go analyzer where possible.

This replaces the agent's proposed rule that parents must freeze while child
work runs. Ownership restricts where writes occur; it does not require a new
scheduler, filesystem, or agent-role permission system.

## Runtime rule in the examples

The first successful `SetContext` establishes the binding's owner. Subsequent
writes from that scope replace its value. Ordinary Go contexts derived with
cancellation or deadlines still refer to the same scope.

- Ancestors and descendants cannot own the same key. The check applies in
  either write order, including when the child was created before the ancestor
  first wrote the key. A stale inherited snapshot cannot bypass ownership.
- Siblings may each own a local binding with the same name. Those bindings are
  different values; neither sibling can edit the other's binding through its
  own context.
- Rejected writes do not change files, revisions, or the last published view.
- This is a scope rule, not a goroutine rule. Goroutines sharing the same scope
  can still write that scope's values; the library serializes writes, but the
  workflow must choose meaningful ordering if they target the same value.

The ancestry rule in both write orders and independent sibling bindings are
implementation choices consistent with Tyler's restriction. Keys differing
only in case are also checked to preserve portable filesystem views.

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

The parent may continue updating its own values. The examples still freeze
inherited data at scope creation and publish immutable prompt/index revisions
before agent calls. Ownership and freshness are separate policies; this change
does not implement live inheritance or automatic result promotion.

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
