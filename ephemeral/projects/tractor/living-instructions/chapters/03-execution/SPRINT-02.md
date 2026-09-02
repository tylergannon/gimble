# Sprint 2: the embedded `medium` workflow

Add `medium` to the workflow library as a runnable embedded definition. It
receives one workflow parameter, the safe project name, and resolves
`<workdir>/ephemeral/projects/<project>` while materializing the graph. Write
the resolved `checklist.md`, `brief.md`, and interview paths into node fields or
prompts. Do not add graph variables or overload `$goal`.

The graph is one `loop` node over the plan's flat checklist and one codergen
body node returning to that loop. The loop owns validation and marking, routes
to success only after every item is done, and uses `max_visits` as its only
ceiling. The body prompt scopes the agent to the current frame, the brief, and
any item doc; it tells the agent to complete the whole current sprint, run its
validator before returning, never write `done`, and commit its work. It may use
blocking `tractor ask` only when a missing validator or repeated failure raises
a material question. There is no child run or failure-escalation node.

Separate plan-only parameters such as the seed from the common workflow
parameters so building `medium` does not require dummy input. Validate the
embedded YAML through the same graph and lint paths as every other definition.

## Proof

Add tests named `TestBuiltInMedium` and
`TestMediumRunsPlanningChecklist`. The first fixes the exact graph shape,
resolved project paths, prompt contract, registry ordering, and rejection of a
bad project. The second runs the materialized graph through the real engine
with a deterministic fake codergen over a two-item planning checklist; require
two selected items, both commands actually executed, both `done` fields written
by the engine, unchanged Markdown body, validation records, and terminal
completion.

Add `check-sprint-02.sh` beside this document. It runs those named tests with
`-count=1 -v` and refuses an empty or renamed suite. Run every BUILD.md gate
before committing.
