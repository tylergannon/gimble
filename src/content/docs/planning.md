---
title: Start with a plan
description: Turn a seed into an interviewed brief, a runnable checklist, and the right-sized Tractor handoff.
eyebrow: Start here
order: 1
sourceLabel: Read the built-in workflow contract
sourceUrl: https://github.com/tylergannon/tractor/blob/main/docs/spec.md#312-built-in-workflows
---

If the job is not already a small, settled change, start with Tractor's built-in
planning workflow. It keeps the planning conversation in one agent turn, asks
only questions that can change the contract, and leaves a plan you can inspect
before execution begins.

## Run the planner

Write the initial request in a seed file in the target repository. Then list
the workflows in your installed binary and run `plan`:

```sh
tractor workflow list
tractor workflow run plan \
  --project my-build \
  --seed seed.md
```

`--project` is a safe directory name, not a path. The seed path is relative to
`--workdir`, which defaults to the current directory; planning artifacts land
under that workdir. Tractor first prints `Logs: <absolute-path>`. By default it
allocates a fresh directory under `$XDG_STATE_HOME/tractor/workflow-runs/`, or
`~/.local/state/tractor/workflow-runs/`; pass `--logs <empty-directory>` to
choose one yourself.

## Answer the interview

Watch `<printed-logs>/timeline.jsonl`. When a `QuestionAsked` event
appears, open the numbered file in its `question` field and answer it:

```sh
tractor answer ephemeral/projects/my-build/interview/0001.md \
  "Keep the first release local-only."
```

The planner continues in the same turn. It covers intent, scope and non-goals,
constraints, and an observable Definition of success, stopping when more
answers would only repeat derivable detail. See [Interviews](/docs/interviews/)
for the file protocol, stdin answers, and resuming an interrupted wait.

## Read the handoff

When planning finishes, the command prints the paths to:

- `ephemeral/projects/my-build/brief.md` — the agreed contract.
- `ephemeral/projects/my-build/checklist.md` — one-turn items with real checks.
- `ephemeral/projects/my-build/recommendation.md` — size, rationale, and next
  action.

SIMPLE is at most one sprint: execute the checklist yourself. MEDIUM is more
than one sprint but fewer than two chapters. LARGE is multiple chapters. For a
MEDIUM or LARGE plan, Tractor prints the executable `Next:` command:

```sh
tractor workflow run medium --project my-build
# or
tractor workflow run large --project my-build
```

Run it from the same repository, or add the same `--workdir` used for
planning. The command prints a new absolute `Logs:` path before execution so
you can follow its timeline. One run covers the whole plan. The
[Loops guide](/docs/loops/) explains the MEDIUM and LARGE execution shapes,
reviewer questions, validation evidence, and engine-owned marking.
