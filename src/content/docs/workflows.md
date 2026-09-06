---
title: Built-in workflows
description: Five whole workflows ship inside the binary — sprint execution, a nested chapter loop, competitive planning, a delivery loop, and a promise loop. Run one by name; nothing to copy.
eyebrow: Reference
order: 2
sourceLabel: Read the pipelines
sourceUrl: https://github.com/tylergannon/tractor/tree/main/internal/workflows
---

The [loop examples](/docs/loops/) are single shapes you copy and edit. The
built-in workflows are whole pipelines that ship inside the binary and run by
name, with nothing to copy first:

```sh
tractor workflows
tractor run sprint-execute --workdir . --logs .tractor/run
```

`tractor workflows` lists what this binary carries and the situation each one
is for. `tractor workflows show <name>` prints the pipeline itself — redirect
it into a file when you want to adapt one, and the copy is an ordinary
pipeline with no tie back to the binary.

A name resolves to a built-in only when no file of that name exists, so a
pipeline on disk is never shadowed by one that ships.

## What a workflow works on

`--goal` replaces the pipeline's goal, and every workflow carries that into its
working prompts, so a goal always steers the run:

```sh
tractor run delivery-loop --goal "Build what docs/SPEC.md describes" \
  --workdir . --logs .tractor/run

tractor run sprint-plan --goal "Replace polling with server-sent events" \
  --workdir . --logs .tractor/run
```

What differs is whether a goal is the assignment or a steer. `sprint-plan` has
nothing to plan and `delivery-loop` has nothing to build without one, so both
refuse to start rather than running against the generic goal their file
carries, and `tractor workflows` marks them. `sprint-execute` and
`chapter-loop` iterate a ledger that already names every sprint, and
`promise-loop` reads `PROMISE_ID` from the environment; a goal frames those
without replacing the item in hand, and a workdir alone is enough to start.

## When to run which

| The situation | Workflow | What it works on |
| --- | --- | --- |
| A sprint ledger is planned and you want it worked to done | `sprint-execute` | `docs/sprints/ledger.md` |
| A chapter's worth of work should run as one long run | `chapter-loop` | `docs/chapters/ledger.md` |
| The next sprint needs planning properly, not off the cuff | `sprint-plan` | `--goal`, the seed |
| A specification exists and you want software from it while you are away | `delivery-loop` | `--goal`, naming the spec |
| One repository promise needs advancing and a verdict recording | `promise-loop` | `$PROMISE_ID` |

## What they have in common

**The ledger is the checklist.** `sprint-execute`, `chapter-loop`, and
`delivery-loop` all iterate a checklist file with a `loop` node. The engine
re-reads it every arrival, validates the item the previous lap worked on, and
marks it done itself. No agent writes `done`, so no protocol is needed to stop
one from claiming work it did not finish.

**Validation runs on the other provider.** Whoever reviews is never whoever
wrote the code, and the reviewer cannot route to success — leaving a loop is
the loop node's decision against the checklist's definition of done.

**Supervisors coach, they never route.** Each workflow carries supervisors
named by the question they ask: whether the work in flight is the work that
was asked for, whether proof is being demonstrated or merely constructed, and
whether a plan says what done looks like instead of listing commands to run.
They patrol on their own clock and steer one named target at a time.

## Adapting one

Print it, save it, edit it:

```sh
tractor workflows show delivery-loop > delivery.yaml
tractor edit delivery.yaml
```

The workflows carry their reasoning in comments, so the file you get is also
the explanation of why it is shaped that way.
