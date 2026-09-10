# Filesystem-backed workflow context

Design note, 2026-09-09. The generation trigger and budget behavior below
summarize Tyler's direction. The timing, API spelling, and first selection
policy are agent proposals explored in the [Go examples](../../../../examples/go-workflows/README.md),
not an accepted production contract.

## Tyler's direction

- `SetContext` supplies workflow context backed by ordinary files. Small
  values can appear inline; larger values are available through an index.
- Consider the combined prompt budget: many individually small values can
  require moving some material out of the prompt and reorganizing the index.
- Updating context triggers the work needed to produce a coherent view.
  An agent must wait for that view instead of receiving a half-updated index.
- Explore existing open-source context systems, while letting broader
  research wait until the immediate authoring examples are concrete.

The purpose remains the [one-context direction](../../../../docs/workflows-as-programs.md#one-context-organized-around-success):
give an agent an understandable assignment and useful paths to detail.
An inline threshold alone does not decide what the agent needs to succeed.

## Proposed timing: publish at the agent boundary

The simplest first rule is that each agent receives one coherent snapshot
when its call begins. `SetContext` preserves the complete value and marks the
view dirty. Several updates can accumulate while ordinary Go work continues.
The next agent call prepares the view before invoking the agent; callers can
also request an explicit `SnapshotContext`.

This is deferred, coalesced indexing. The example setter does not start a
background worker. The automatic call boundary satisfies the proposed wait
without pausing unrelated workflow work. An indexing error prevents that
agent call. Once acquired, its snapshot stays stable; later updates belong
to a later snapshot. Ordinary child contexts inherit the same store, so they
do not create private branches of context.

A later optimization could prepare the view asynchronously, preserving
the same entry rule. Its scheduling and cancellation are separate design
choices; neither a global workflow pause nor live mutation of an active
agent's prompt follows from this proposal.

## What the current example explores

`NewContext(ctx, dir, ContextLimits{ValueBytes, PromptBytes})` attaches a new
store. `SetContext(ctx, key, value)` writes immutable JSON values.
`SnapshotContext(ctx)` creates a versioned index and returns the prompt,
revision, index path, and inline/external key lists. `Codergen` acquires that
snapshot before calling its agent callback.

Both thresholds use bytes, as a temporary proxy for token accounting.
`ValueBytes` applies to each encoded value; `PromptBytes` bounds the whole
context projection, including its routes. It is not a bound on the complete
model request including step instructions, history, or other harness content.
Values over the individual limit are external. If the combined projection
is too large, the selector spills the largest remaining values first, with
keys breaking ties. If routes themselves exceed the budget, the prompt keeps
only the index entrypoint.

The JSON index contains keys, placeholder descriptions, byte sizes, and paths
to exact values. It is plain-text navigation, not an LLM-generated semantic
index. Old indices continue pointing to immutable value files. The context
demo stages small inline data, oversized research, then enough small values
to require aggregate spilling; it exposes the composed snapshots and files.

Filesystem writes and context projection are real here. Agent replies,
commands, workspaces, and integration remain stubs. There is no native-agent
integration, Polytype schema generation, or complete runtime parity. No
external context dependency has been adopted. Pinning goals or rules,
choosing by relevance, and preserving useful attention cues need further
judgment; size limits do not establish that an agent received effective context.

## Two leads for later inspection

- **OpenViking** stores context in a virtual filesystem with abstract,
  overview, and full-content tiers. Its asynchronous semantic/embedding work
  has an explicit `wait_processed` barrier; generated summary writes use
  version checks to discard stale results. Inspect its processing pipeline
  and completion API for index generation and publication ideas.
  [Architecture](https://github.com/volcengine/OpenViking/blob/main/docs/en/concepts/01-architecture.md),
  [wait API](https://github.com/volcengine/OpenViking/blob/main/docs/en/api/07-system.md#wait_processed),
  [write consistency](https://github.com/volcengine/OpenViking/blob/main/docs/en/concepts/09-transaction.md).
  Its current open-source license is [AGPL-3.0](https://github.com/volcengine/OpenViking/blob/main/LICENSE).
- **Letta Code MemFS** projects a Git-backed memory filesystem into context:
  `system/` files are in the prompt, with other files reached through the file
  tree. Committed updates take effect on later recompilation; background
  memory workers can use worktrees without blocking the main agent. Its
  context-doctor measures total/per-file tokens and guides restructuring.
  Inspect the projection and recompile boundary, then its budgeting guidance.
  [MemFS](https://docs.letta.com/concepts/memfs),
  [update timing](https://github.com/letta-ai/letta-code/blob/main/src/agent/prompts/letta.md),
  [context-doctor](https://github.com/letta-ai/letta-code/blob/main/src/skills/builtin/context-doctor/SKILL.md).
  Its current open-source license is [Apache-2.0](https://github.com/letta-ai/letta-code/blob/main/LICENSE).

This was a bounded source inspection, not a comparative runtime evaluation.
Neither inspected system establishes the complete automatic `SetContext`
contract: aggregate prompt rebalance plus coherent index publication and an
agent-entry wait. Those remain our design questions.
