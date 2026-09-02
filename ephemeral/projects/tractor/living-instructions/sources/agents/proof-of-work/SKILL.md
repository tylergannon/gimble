---
name: proof-of-work
description: >
  Define and assess proof for application changes. Use while authoring coding
  tasks, reviewing a PR, or reassessing whether changed behavior was actually
  demonstrated. Do not use for docs-only changes.
---

# Proof Of Work

Prove changed behavior through the running application and produce an artifact
that demonstrates the result. If the repo provides `/repo-proof`, load it for
the project-specific protocol.

Unit tests, linters, and compile checks are required checks, not proof, unless
they exercise the whole application behavior being claimed.

For docs, policy, scripts, or tooling, use focused checks instead of full-stack
application proof unless the change affects runtime behavior.

## Proof Claims

In a task or issue, describe the observable claims the implementation must
demonstrate. Do not provide a checklist of commands or list routine checks such
as unit tests and compilation.

In nlspec-governed work, claims arrive classified *cell-local* (demonstrated
through one cell's contract) or *integration* (demonstrated end-to-end); see
the spec-writing standard. Integration claims are demonstrated only through
the running application; a cell-local claim may be demonstrated against a
mock authenticated per slice-design.

## PR Evidence

Lead with each claim and what demonstrated it. Put complete, concrete evidence
in the PR body and upload artifacts instead of citing local paths. Use the
repo's proof-artifact uploader when one exists and verify its links.

If evidence cannot be uploaded, identify what remains local and why. Do not
claim complete proof for evidence reviewers cannot inspect.

At closeout, identify the proved head SHA, satisfied and unmet claims, artifact
links, relevant checks, and any proof blocker.
