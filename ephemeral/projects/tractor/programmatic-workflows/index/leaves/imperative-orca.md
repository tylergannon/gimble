# Orca: imperative flow DSL with optional durable stage semantics

Purpose: VirtusLab Orca is the closest evidence for authoring an agent workflow as ordinary typed code while deliberately adding a runtime boundary around durable side effects. Snapshot: Apache-2.0, commit `76e46122add1fd781432f8e6281728a55fc938f8`; GitHub API reported it non-archived on 2026-09-08, latest release `v0.1.6` published 2026-08-28. Corpus and hashes: `imperative/manifest.json`.

## Findings for the programmatic-workflow decision

- The author-facing flow is ordinary Scala control flow: a `stage("Plan")`, a `for task <- plan.tasks`, and a final bounded review loop. The important observation is that the loop is source code, while `stage` supplies a separate durability/effect boundary; the loop itself is not serialized as a graph. `imperative/orca/README.md:100-139` ([pinned source](https://github.com/VirtusLab/orca/blob/76e46122add1fd781432f8e6281728a55fc938f8/README.md#L100-L139)).
- `stage[T: JsonData]` first computes an ID, tries to decode a saved `T`, otherwise executes its body with capability evidence, records the result, and commits. This is a *stronger, costly* contract than Tractor's proposed fresh-revalidate restart: it couples saved typed intermediate values and source mutation into a checkpoint. `imperative/orca/flow/Flow.scala:40-50`, `imperative/orca/flow/Flow.scala:52-80`, `imperative/orca/flow/Flow.scala:116-142` ([pinned source](https://github.com/VirtusLab/orca/blob/76e46122add1fd781432f8e6281728a55fc938f8/flow/src/main/scala/orca/Flow.scala#L40-L142)).
- The design explains the hidden machinery behind safe skip/replay: hierarchical stable stage IDs, JSON decoding to the current call-site type, and no concurrent stages because commits share a Git index. It also records a stage baseline commit for review, and notes that nested checkpoints sweep existing working-tree edits. `imperative/orca/adr/0018-stage-bound-flow-runtime.md:61-86`, `imperative/orca/adr/0018-stage-bound-flow-runtime.md:101-166` ([pinned ADR](https://github.com/VirtusLab/orca/blob/76e46122add1fd781432f8e6281728a55fc938f8/adr/0018-stage-bound-flow-runtime.md#L61-L166)).
- Its reusable per-call typed-output mechanism is separate from stages. `resultAs[O]` requires a type carrying a schema and strict JSON codec; the autonomous path sends schema to the backend, parses the reply, and corrective-retries parse failures. Build schema construction fails before a destructive stage can start. `imperative/orca/tools/JsonData.scala:17-53`, `imperative/orca/tools/Agent.scala:95-103`, `imperative/orca/tools/AgentCall.scala:109-147`, `imperative/orca/tools/AgentCall.scala:189-295` ([pinned source](https://github.com/VirtusLab/orca/blob/76e46122add1fd781432f8e6281728a55fc938f8/tools/src/main/scala/orca/agents/AgentCall.scala#L109-L295)).
- Orca's capability split is its other durable lesson: a parallel fork may get read/LLM authority but not stage-start or workspace-write authority, structurally preventing concurrent Git checkpoints. The stated types are language-specific, but the separation of read-only/review calls from one exclusive mutation/checkpoint owner is portable. `imperative/orca/adr/0018-stage-bound-flow-runtime.md:173-218` ([pinned ADR](https://github.com/VirtusLab/orca/blob/76e46122add1fd781432f8e6281728a55fc938f8/adr/0018-stage-bound-flow-runtime.md#L173-L218)).

## Mechanisms to copy, not merely cite

1. A narrow typed agent-call boundary: `Call[In, Out]` generates/receives an explicit schema, validates/decodes locally, and returns `Out` or a classified error. Keep retry policy owned by the call layer, not hidden in routing prose.
2. A named effect boundary for agent mutation and validation. In the proposed weaker recovery mode it can emit structured run evidence and establish a fresh validation baseline without saving `Out` or committing every boundary.
3. If checkpoint skipping is later needed, add it as an explicit mode with stable names, persisted versioned data, decode-before-skip, and a concurrency/working-tree policy. Do not accidentally imply it from ordinary Go functions.

## Do not copy yet / counterevidence

- The complete stage contract is disproportionate if restart intentionally re-derives plans and reruns validation: it adds progress-log format, atomic writes, branch/prompt binding, commit policy, session rehydration, ID stability, and idempotence obligations. In particular, Orca's own ADR says a nested checkpoint may include earlier outer edits, so a replay must tolerate leftovers. `imperative/orca/adr/0018-stage-bound-flow-runtime.md:148-171`.
- Orca is an early Scala application/CLI, not evidence that its runtime extracts into an embeddable Go library with Tractor's existing adapters. Its README claims resumability; this corpus inspected code/design but did not execute it. `imperative/orca/README.md:152-159`.

## Retrieval recipes

- For typed agent output and corrective retries: open `imperative/orca/tools/AgentCall.scala:109`.
- For exact checkpoint and recovery cost: open `imperative/orca/adr/0018-stage-bound-flow-runtime.md:61` then `imperative/orca/flow/Flow.scala:40`.
- For an actual imperative planning/task/review loop: open `imperative/orca/README.md:100`.

Unknowns: backend-specific schema enforcement differs (the README documents Gemini as prompt-enforced), and source inspection does not establish failure behavior with Tractor's native agents or current worktree model. `imperative/orca/README.md:179-205`.
