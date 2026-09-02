# Sprint 4: workflow CLI and fresh run directories

Expose all three registered definitions through the existing surface:

```text
tractor workflow list
tractor workflow run plan --project <build> --seed <path>
tractor workflow run medium --project <build>
tractor workflow run large --project <build>
```

`--project` is the only workflow-specific parameter shared by execution
workflows. `--seed` remains required for `plan` and is not required for
`medium` or `large`; unknown names and invalid or missing inputs fail before a
harness begins. All workflows retain the common `--workdir` and `--logs`
controls and use the shared foreground runner. Configure the project's
interview directory for every workflow so a planning or implementation turn can
call plain `tractor ask`.

Make `--logs` optional for all three workflows. With no override, allocate a
fresh, collision-safe run directory beneath the same XDG or user state root
used by the MCP run store, never beneath the committed project. Print the
absolute chosen logs path before the run begins so it is observable even when
the run blocks for an interview. An explicit `--logs` preserves current
relative-to-workdir resolution and fresh-directory safety. Factor state-root
selection rather than cloning subtly different environment logic.

After a completed plan, preserve the artifact/size/next handoff. After a
completed execution workflow, print the project, workflow name, and completed
logs path. The fixed `Next:` commands from chapter 2 must run as printed; they
must not be mere prefixes that require an undisclosed flag.

## Proof

Extend the Cobra tests and add named `TestWorkflowExecutionRun` and
`TestWorkflowDefaultLogs` cases. Cover stable `plan`, `medium`, `large` listing;
per-workflow arguments; materialized graph selection; shared workdir and runner;
interview environment; XDG and home fallbacks; unique fresh allocations;
printed-before-run logs discovery; explicit override; and success output.

Add `check-sprint-04.sh` beside this document. It proves the existing workflow
tests and both new names ran, builds the real CLI, verifies all three list/help
surfaces, and checks that the exact `Next:` commands pass argument validation
without `--logs` or `--seed` by using an injected/test boundary rather than
launching a model. Run every BUILD.md gate before committing.
