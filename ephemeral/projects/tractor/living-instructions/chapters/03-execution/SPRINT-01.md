# Sprint 1: execution-ready planning artifacts

Make chapter 2's planning output a valid direct input to the execution workflow
it recommends. Preserve the three fixed top-level artifacts; a LARGE plan may
also contain the supporting chapter files required by its nested loop.

## Size-specific contract

Update the embedded planner prompt and mechanical artifact validation together:

- SIMPLE has at most one flat checklist item.
- MEDIUM has more than one flat checklist item. Every item is one implementation
  sprint and must not carry a nested `checklist`.
- LARGE has more than one chapter item in top-level `checklist.md`. Every item
  carries a unique `doc` path to a non-empty chapter document and a unique
  `checklist` path to that chapter's initially empty sprint ledger. The planner
  writes both supporting files beneath the project directory. A chapter is not
  constrained to one agent turn; the sprints later written beneath it are.

All generated paths remain relative to the repository workdir, resolve beneath
`ephemeral/projects/<project>/`, and reject absolute paths or traversal. Empty
sprint ledgers are valid checklist files with an `items: []` frontmatter list.
The planner never writes `done` at either level.

Parse the recommendation early enough for `ValidatePlanArtifacts` to apply the
matching shape. Keep the existing recommendation field and `Next:` contract
unchanged. Do not invent a conversion step in `large`: its printed handoff must
be directly runnable against these files.

## Proof

Extend the workflow package tests, including a named
`TestPlanExecutionShapes`, to cover valid SIMPLE, MEDIUM, and LARGE output plus
wrong item counts, nested MEDIUM items, missing/out-of-root/duplicate LARGE
paths, non-empty initial sprint ledgers, and engine-owned `done` at every level.
`TestBuiltInPlan` must assert the materialized prompt teaches the same contract.

Add `check-sprint-01.sh` beside this document. It runs the focused workflow
tests with `-count=1 -v` and fails unless `TestBuiltInPlan`,
`TestPlanArtifacts`, `TestPlanExecutionShapes`, and `TestRecommendation` all
actually ran and passed. Run every BUILD.md gate before committing.
