# Findings from the 2026-09-03 issue-32-37 session

Everything established in the session that is not already captured in a
commit or an issue. Recorded so it survives the conversation.

## Landed

| Issue | Commit |
|---|---|
| #36 evidence globs `**` and `./` | `371f882` |
| #37 judge model independent of defaults | `e44b056` |
| #32 validate every done item every lap | `7eda6d6` |

## Open, with rulings

### A. Judge and gate stage files collide (new bug)

`engine/codergen.go:58` writes `prompt.md` into `scope.StageDir`, and
`writeResponse` writes `response.md` beside it. Both paths are fixed.

Before #32 one judge turn ran per arrival, so nothing collided. #32 made the
validation set every done item plus the framed item, and `validate` calls
`judge` for each item that has an `infer` block, all with the same
`scope.StageDir`. Two `infer` items in one set therefore overwrite each
other's `prompt.md` and `response.md`; only the last survives. #33's exit
gate adds a third writer to the same two paths.

#32 gave each item its own `validation-NNN.log` and stopped there.

Fix: give each judge turn and the gate turn its own prompt/response paths in
the loop stage directory, the way `validation-NNN.log` already works.

### B. The exit gate must not use the judge's model slot (amends #33)

A loop node's `llm_model`, `llm_provider` and `reasoning_effort` have only
ever configured the infer judge. #37 stopped them inheriting pipeline
`defaults` and gave the judge its own default, the `flash` alias resolving to
`gemini-3.8-flash-medium`. Flash is for reading screenshots.

#33's text says the exit gate evaluator runs "on the loop node's model
fields, the same slot the infer judge uses." #33 predates the #37 ruling. As
written it puts the turn that decides the whole run is finished, and that
authors the remaining checklist items, on the screenshot judge's model.

**Ruling: the gate does not borrow the judge's slot.** It gets its own
selection and runs on the pipeline's default model, the same one the working
agent uses. Sol's uncommitted draft routes the gate through `evaluate`, the
judge's shared slot, and needs changing.

### C. Nomenclature (see the naming issue)

Recorded in full in its own GitHub issue. Summary of what was established:

- `body` on a loop node collides with the checklist file's markdown body,
  which is the definition of done. Same word, two meanings, one feature.
- `body` is one edge to one node, not a region. No region is declared
  anywhere in the authored schema; `lint.loopBodyNodes` derives the body
  node-set by walking from `body` until it reaches the loop node again.
- Three fields mean "the nodes under this one" and share no root:
  `branches` (parallel), `body` (loop), `supervises` (supervisor).
- `tool` runs a shell command through `/bin/sh` and has nothing to do with
  tools in the MCP sense; its own field is `tool_command`.
- `codergen` is jargon for an agent turn.
- `parallel` and `parallel.fan_in` are asymmetric. `docs/spec.md:191` already
  defines the type as "Parallel fan-out" and the prose calls it a fan-out
  throughout.
- A `parallel` node carries the full agent field set (`prompt`, `fidelity`,
  `thread_id`, `max_retries`, the three model fields) and its handler never
  runs a turn. Those fields are the template each branch is built from, and
  a branch's `codergen` block overrides them field by field
  (`graph/parse.go:125-143`). So `prompt` on a `parallel` node means "the
  default prompt for every branch", which is not what `prompt` means on any
  other node type.

Routing itself is not an accident and was left alone: `edges` with conditions
is what an agent chooses among, `on_*` fields are what the engine decides.

## Process notes

- Sol built #32 in a detached worktree per the repo's agent protocol, so the
  pipeline's check node ran in the unchanged workdir and passed on nothing.
  The briefs now forbid a separate worktree. The check node still cannot tell
  a no-op run from a real one, because `go build && go test` passes on a
  clean tree.
