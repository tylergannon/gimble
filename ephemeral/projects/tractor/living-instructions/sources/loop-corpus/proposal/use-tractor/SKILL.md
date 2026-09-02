---
name: use-tractor
description: >
  Choose, author, validate, start, monitor, steer, and resume small Tractor
  pipelines. Use when a coding or research task benefits from bounded repeated
  attempts, a mechanical verification back-edge, a fresh planning or review
  stage, parallel independent work, or a durable long-running run. Do not use
  merely because a task has several ordinary steps that one Codex session can
  complete and verify coherently.
---

# Use Tractor

Turn a task into the smallest Tractor graph that materially improves its odds
of finishing correctly. Prefer ordinary two- and three-node loops over elaborate
software-factory graphs.

## Choose the shape

Ask what must happen outside a single worker's context:

- **A command decides done:** use `worker -> check`, with check failure routed
  back to the worker.
- **The approach deserves a durable handoff:** use
  `plan -> implement -> check`; replan only when the approach is wrong.
- **A bug needs a falsifiable theory:** use `diagnose -> fix -> reproduce`.
- **Quality needs independent judgment:** use `draft -> review -> revise`, with
  fresh reviewer context.
- **Work is genuinely independent:** use parallel branches and one fan-in.

Stay in the current Codex task when none of those boundaries is useful. Do not
make separate nodes for private cognitive steps such as “think,” “code,” and
“double-check.”

Read [patterns.md](references/patterns.md) when authoring a new pipeline or
choosing among these shapes.

## Author the pipeline

1. Read repository instructions and identify the workspace the run may change.
2. State the outcome, the evidence that admits success, the correction path,
   and the maximum useful number of laps.
3. Call `get_pipeline_schema` before relying on remembered syntax. Tractor's
   current schema is authoritative.
4. Start from the closest reference pattern and replace its goal, prompts,
   artifact names, and verification command. Keep the graph strict JSON or
   YAML accepted by the current schema.
5. Keep one coherent responsibility per node. Add a node only for a durable
   artifact, independent check, authority boundary, fresh perspective, or
   parallel isolation.

For every back-edge, set `max_visits` on a node that bounds the whole cycle.
Include a deliberate `failure` route when an agent can determine the task is
impossible as stated.

## Write useful handoffs

Use the shared workspace as the default data channel:

- planning writes a short plan file that implementation reads;
- checks write their latest output to a known file while preserving the real
  exit status;
- review writes concrete findings that revision reads;
- progress or research that must survive a resume lives in a file, not only in
  prose output.

Let a revisited worker use Tractor's default compacted session. Set
`fidelity: "none"` on an independent reviewer. Share `thread_id` across nodes
only when conversation continuity is intentionally more important than role
separation and an auditable file handoff.

## Make verification external

Use a `tool` node for deterministic evidence such as tests, builds, linters,
schema checks, file assertions, or a real entry-point probe. Route exit code 0
with `on_success` and nonzero with `on_error`.

Use a `codergen` reviewer only for criteria that require judgment. Name the
question precisely—requirements coverage, security, usability, minimality—not
just “review.” Give multiple edges concrete prose conditions; the chooser reads
those conditions, while the engine does not parse them.

Match the final check to the claim. Unit tests do not prove a browser flow, and
an LLM verdict does not prove the program builds.

## Validate and run

1. Save the pipeline in the task worktree or another user-approved workspace.
2. Call `validate_pipeline` with its path and intended work directory. Fix every
   structural or lint error before execution.
3. Call `start_run` and retain the returned `run_id`, process ID, logs path, and
   work directory.
4. Use `get_run_status` until the run succeeds, fails, or needs intervention.
   Inspect stage artifacts when a route or result is surprising.
5. Use `steer_run` only for new authoritative information, a clear scope drift,
   or an active turn stuck on a bad assumption. A rejected steer with no active
   turn is normal; do not spam retries.
6. Use `stop_run` when the user asks to stop or continuing would be unsafe or
   wasteful. Resume from the same logs only when the pipeline and workspace are
   still compatible with that checkpoint.

## Close out

Report the pipeline path and shape, validation result, run ID, terminal status,
the evidence that established the outcome, and any exhausted budget or remaining
human decision. Do not call a passing parser or an agent's self-report proof of
functionality.
