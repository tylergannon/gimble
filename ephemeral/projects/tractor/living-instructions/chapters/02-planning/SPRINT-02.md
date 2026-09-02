# Sprint 2: `tractor workflow`

Expose the embedded library through the dedicated surface selected in interview
0006:

```text
tractor workflow list
tractor workflow run plan --project <build> --seed <path> --workdir <dir> --logs <dir>
```

`workflow list` prints available built-ins in stable order with enough help text
to discover `workflow run`. Do not claim `medium` or `large` is available until
chapter 3 registers it.

`workflow run plan` requires both `--project` and `--seed`. The project is an
explicit safe directory name, never derived from the seed. Resolve the seed
against `--workdir`, require a regular readable file, and create
`ephemeral/projects/<project>/` beneath that workdir. It takes the same required
`--logs` and optional `--workdir` semantics as `tractor run`; share the existing
runner path rather than implementing another engine lifecycle. This is one run,
not a child run.

Before the agent turn starts, set its interview directory to
`ephemeral/projects/<project>/interview` so plain `tractor ask <question>` uses
the right project. Preserve the current run-directory propagation and timeline
behavior. Add a hidden implementation subcommand, or an equally narrow bridge,
that lets the plan graph's tool node invoke sprint 1's Go artifact validator
through the same executable.

After the graph reaches `COMPLETED`, read the validated recommendation and print
the paths of `brief.md`, `checklist.md`, and `recommendation.md`, followed by its
`Size` and `Next` values. A caller should not have to guess the output root or
open the file to learn the handoff. Unknown workflows and unsafe/missing inputs
fail before a harness starts and name the bad argument.

Keep command execution injectable enough that the focused Cobra tests do not
launch a real model. Tests named `TestWorkflowList`, `TestWorkflowRun`,
`TestWorkflowRejects`, and `TestWorkflowHandoff` cover listing, required and
unsafe arguments, shared workdir/logs behavior, graph materialization, interview
environment, and the exact success handoff. Add `check-sprint-02.sh` beside this
document; it runs those tests with `-count=1 -v`, proves each named test ran and
passed, and also exercises `workflow list` and both workflow help paths through
the built binary. Run all required BUILD.md gates before committing.
