---
name: request-adversarial-review
description: >
  Construct and launch a minimal prompt for an independent agent to load
  `/adversarial-review` and review a design, implementation, or proof. Use from
  the calling agent for a one-pass review or to start a `/consensus` loop.
---

# Request Adversarial Review

This skill is for the calling agent. The spawned reviewer owns the review
method by loading `/adversarial-review`; do not reproduce that skill's rubric in
the launch prompt.

Implementation scope and review scope are different. An authoritative
exclusion may constrain the solution, but the caller must not turn it into a
limit on what the reviewer may consider.

The `agent` CLI selects the reviewer automatically. Never pass a model or agent
name as a positional argument; `<workdir>` must immediately follow `agent`.

1. Identify the worktree and authoritative issue, specification, plan, or proof
   artifact the reviewer should read. Choose a new review path such as
   `ephemeral/reviews/YYYYMMDDHHMM-<task>-round-01.md`.
2. Include only context needed to locate the target plus operating and output
   constraints. If a requirement exists only in conversation, state it
   concisely as authoritative context without turning it into a checklist or
   expected answer. Point to authoritative exclusions; do not restate them as
   reviewer instructions.
3. Remove positive or negative limits on the defects, files, or subject matter
   the reviewer may consider, plus expected findings or conclusions, declared
   safe areas, or desired verdicts. Then launch it with a short prompt.

Bad:
```text
agent <workdir> "/adversarial-review Review these artifacts. Do not broaden
into routing. Re-check the expected failures below. Write to <review-file>."
```

Good:
```text
agent <workdir> "/adversarial-review Review the current implementation against
the issue and repository instructions. Inspect the working tree and relevant
surrounding code. Write the complete review to <review-file>. Do not edit any
other files."
```

Retain the reported session ID and review-file path. Read the artifact before
acting on findings. For a one-pass review, summarize and link the artifact. For
repeated review and adjudication, continue with `/consensus`. An
`invalid review request` outcome is not a review result; discard that session
and relaunch with a clean prompt.
