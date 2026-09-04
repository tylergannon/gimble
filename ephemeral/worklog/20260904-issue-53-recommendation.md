# Issue 53 recommendation

Agent recommendation, not approved by Tyler. Source inspection at `492bd6b`;
no implementation or live harness experiment performed.

correction: Tyler explicitly requested independent judgment about the workflow
language and Attractor's goals; issue 53 is agent-authored input, not an accepted
design contract.

finding: `graph/parse.go:193-222` inherits model, provider, and effort separately
for ordinary agent turns as well as loop evaluators. A loop-only resolver would
leave two incompatible authoring rules in the same language.

finding: `engine/loop.go:355-485` already separates item evidence judgment from
goal evaluation and replanning, with fresh turns and engine-owned bookkeeping.
Named role configuration can expose this distinction without requiring authors
to reproduce the protocol as ordinary graph nodes.

recommendation: Keep the checklist loop and its arbitrary body subgraph. Give
its roles explicit names (`item_judge`, `goal_evaluator`) and use one model
selection object everywhere a turn is configured. Prefer whole-selection
replacement over fieldwise inheritance. Preserve the deliberate independent
judge default and the evaluator's normal pipeline fallback.

recommendation: A present model object requires a name; provider may come from
the alias or a recognized native name, otherwise it is required. Effort belongs
to that selection and cannot leak from a replaced selection. Reject provider-only
and effort-only replacement objects. Document the small cost: changing effort
requires repeating the model name.

recommendation: Resolve and validate all turn selections through a shared path
before execution, preserve authored provenance, and show effective role, provider,
model, effort, and harness. Translate authored effort through adapter-supported
native forms; never assume stripping a native model suffix produces a valid ID.
For explicitly pinned native IDs with embedded effort, reject contradictory
authored effort.

doc_bug: README.md's claim that done is an exit code conflicts with the current
loop evaluator's goal judgment and the workflow rules' insistence on material
proof. Update the explanation alongside any accepted language migration.

scope: Recommend a coordinated model-selection contract change with issue 11,
without redesigning fan-out topology, checklist semantics, or adding arbitrary
role subgraphs, profile registries, or model benchmarks. Retain strict migration
and discriminating live proof of item rejection, repair, goal-not-met replanning,
and eventual goal satisfaction. That proof establishes the tested behavior, not
general judge quality.
