# Sprint 1: prompts become library files

Move the three graphs and the four prompts out of `workflow/workflow.go`
into `workflow/library/`, embedded, rendered with `text/template`. The
outside of the package does not change: `Build`, `Parameters`, `List`,
and every existing test keep passing, and every prompt and command
`Build` returns is byte-identical to what it returned before this sprint.

Read `research/workflow-package-inventory/migration-inventory.md` first;
it names every line this touches.

## Layout

```
workflow/library/
  README.md                the template contract (below); how to test an edit
  workflows/plan.yaml      moved from workflow/
  workflows/medium.yaml
  workflows/large.yaml
  prompts/plan/planner.md
  prompts/medium/implement.md
  prompts/large/plan.md
  prompts/large/implement.md
```

One `//go:embed library` into an `embed.FS` in a new `workflow/library.go`.
The three `[]byte` definition vars go away; the tests that reassign them
to prove parse failure (inventory §3) need a new seam: a package-level
`func parseDefinition(name string) (*graph.Graph, error)` that the tests
can call with a corrupt `fs.FS`, or an exported `BuildFrom(fsys fs.FS,
…)` used by `Build`. Pick the smaller; keep the test intent.

## Template contract

- Engine: `text/template`, delimiters `<<` and `>>` (or another
  non-default pair). The README states them on one line of the form
  `Delimiters: << >>`; the sprint 2 proof script reads that line.
  Doctrine text may contain `{{`, which is then literal.
- Data: a struct built from `Parameters` plus the derived values the old
  functions computed (`projectDir`, `briefPath`, `checklistPath`,
  `interviewDir`, `questionCommand`). Two functions in the func map:
  `quote` (today's `strconv.Quote`, for display) and `shell` (today's
  `shellQuote`, for executable words). They are different jobs; the
  inventory's gotcha on `'/tmp/tractor'\''s binary'` is a test.
- `include "name"` and `doctrine "name"` actions exist but are not yet
  used by any prompt (sprint 4 does that). Wire the func map now so
  sprint 4 is content only.
- An empty render is an error, never a fallback to the node label
  (inventory gotcha on `PromptValue`).
- Export `Render(name string, params Parameters) (string, error)`: the
  standalone rendering of one library file (`prompts/...`,
  `doctrine/...`, `templates/...`) with the same data and func map
  `Build` uses. The render test and the chapter's proof scripts call
  it; `Build` materializes each prompt-bearing node's prompt with `Render`
  and adds nothing to it (tool commands and checklist paths are
  `Build`'s own); graph construction, path resolution, and validation stay in
  `Build`.

## Byte equality

Before touching anything, write `TestBuildMatchesSnapshot`: it calls
`Build` for each workflow with fixed parameters and compares, byte for
byte, the prompt of every codergen node, the `tool_command` of every
tool node, and the `checklist` of every loop node against
`workflow/testdata/<workflow>/<node>.txt`. Capture the snapshots from
the code as it is now, commit them, then do the move; the same test
must pass against the templates. Keep the test; it is the tripwire for
every later content edit that is meant to be a refactor. (Sprint 2's
`show` is checked against a program that calls `Build` directly, not
against these files.)

Fixed parameters, recorded in `workflow/testdata/README.md`: project
`demo`, workdir `/tmp/demo`, executable `/opt/tractor/bin/tractor`,
seed `/tmp/demo/seed.md` (plan only). Sprint 2's proof script runs
`show` with real paths and substitutes these back, so the values must
not appear anywhere else in a prompt.

## Not in this sprint

`show` (sprint 2), the render-all and orphan tests (sprint 3), new
content (sprint 4), docs (sprint 5), `models.yaml` (chapter 5).

## Ask the reviewer

Ask through `tractor ask`, one file per independent group:

- Which delimiter pair, if you have a reason to prefer one; the
  recommendation is `<<` `>>`.
- Whether the definition-parse-failure tests keep a `BuildFrom(fs.FS)`
  seam or a corrupt-FS injection; recommend `BuildFrom`.
