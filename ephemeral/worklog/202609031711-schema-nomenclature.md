decision: ISSUE-41 definition of done supersedes its stale Open section; old authored node types and routing fields must fail with no compatibility aliases.
decision: This task is a schema vocabulary rename only; fan_out branches remain top-level, branch_edges retains the former fan-out template-edge behavior, and no other authored fields move.
friction: A broad multiline example rewrite briefly discarded three fan-out branch destinations; use explicit or schema-aware edits for routing-shape migrations and validate every example immediately afterward.
