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

## Tyler's follow-up

decision: Tyler accepted the recommendation and requested an optional `version`
key on model objects. Use the same model shape across all node types and loop
roles. Third-party model routers are outside the current scope.

correction: Tyler rejected the need for an authored provider selection because
the model name identifies its provider. This supersedes the earlier proposal
to require provider for unrecognized names. The authored object has `name`,
optional `version`, and optional `effort`; provider and harness are derived
resolution results. Reject names whose provider cannot be resolved rather than
guessing or adding a provider override.

recommendation: Make version a string, not a number. Omission selects Tractor's
maintained default release for that name; an explicit version pins that release
without silent fallback. A fully versioned native name remains a supported
escape hatch and does not also take a separate version. Reject a selection that
cannot be resolved. These version semantics are agent proposals implementing
Tyler's requested optional key, not additional decisions made by Tyler.

example: `{model: {name: fable, version: "5.1", effort: high}}`. This is proposed
syntax; no runtime or schema implementation has been changed in this review.

## Executable issue

decision: Tyler requested updating the original GitHub issue into an executable
task and an assessment of whether GPT-5.6 Sol can implement it. Rewrote issue 53's
title/body from `ephemeral/issue-53-executable.md`, distinguishing accepted design
from agent execution clarifications, and read back the published body to verify
it matches. The original loop-only scope and alternative-shape questions are
superseded by the common model contract and all-consumer migration.

assessment: Moderate-to-large migration with settled design; principal risk is
cross-surface consistency and native effort behavior. Recommend GPT-5.6 Sol at
high effort for implementation. GPT-6 is optional for a focused final semantic
review or genuinely new architectural questions, not required for the routine
migration. This is an engineering judgment, not a measured comparison on this
task. Current official model documentation was checked for Sol's capabilities
and supported effort settings.

## Authorized implementation and independent proof

decision: Tyler authorized implementation by a subagent, parent final review,
and PR creation plus squash merge once satisfactory. Existing work was already
isolated in the b683 worktree; preserved its recommendation commits on branch
`codex/unified-model-selection`. GPT-5.6 Sol owns implementation; parent owns
the proof fixture and final review. The main checkout was fast-forwarded without
removing its unrelated untracked work.

proof_design: Parent created a disposable shipping-quote application with a
real inclusive-threshold bug and missing expedited mode. The first worker turn
captures the fault; a judge must reject it; the worker repairs; the evaluator
must add remaining work before completion. Captures contain actual CLI outputs,
and the final oracle lives outside the worker workspace. Baseline execution
confirmed the intended faults and independent oracle rejection.

proof_preparation: Transparent native CLI observers preserve traffic unchanged
and record only model/effort/session metadata, excluding prompts and credentials.
An early candidate passed 29 independent CLI preflight cases, including 23
invalid inputs with no native launch or run-log creation, and cross-consumer
atomic-resolution inspection. These are interim checks, not final acceptance
or live convergence proof.

review_findings: The first native proof uncovered shared fidelity-none binding
identity across a loop's distinct roles, preventing a cross-harness evaluator.
Sol isolated binding keys by role and the repeated real run passed. Source review
also found preflight/per-turn system default divergence; Sol moved normalization
into their shared resolver and cross-path regression checks passed.

proof_result: Native Fable/Flash/Sol run completed after judge fail, repair/pass,
evaluator not_done, appended expedited work, both items passing, and evaluator
done. Independent external oracle passed all six shipping cases. Browser
save/reload preserved a version pin and demonstrated name-only atomic replacement.
Final CLI checks passed 29 preflight cases (23 invalid, zero harness launches/logs)
and cross-consumer resolution checks. Seven sanitized proof artifacts uploaded;
all URLs returned HTTP 200 with curl. Python urllib HEAD returned 403 in this
environment, so link checks used curl. Full report records exact tested revisions
and the limited intervening change. No unresolved material review findings.
