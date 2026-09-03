# Sprint 1: prompts become library files

`workflow/workflow.go` embeds `plan.yaml`, `medium.yaml`, and `large.yaml`
and fills their prompts from Go functions (`plannerPrompt`, `mediumPrompt`,
`largePlanPrompt`, `largeImplementPrompt`, `validatorCommand`). Move every
prompt body into `workflow/library/<workflow>/<node>.md`, embed the
directory, and have `Build` render the files (Go `text/template`; keep the
shell quoting `validatorCommand` needs). `Build` must return byte for byte
what it returns today; write the test that proves it before you move
anything.

Add `tractor workflow show <name>` in `cmd/tractor/workflow.go` that prints
each node's rendered prompt or tool command, so a person can read what an
agent will be sent.

Existing tests in `workflow_test.go` reassign the embedded YAML; keep them
passing. No new dependencies.
