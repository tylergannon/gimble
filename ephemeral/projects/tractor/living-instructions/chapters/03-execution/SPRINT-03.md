# Sprint 3: the embedded `large` workflow

Add `large` as the generic nested shape used by this living-instructions build.
It takes the same single project parameter and materializes resolved project
paths without extending graph substitution.

The graph is exactly this lifecycle:

```text
chapters loop -> plan codergen -> sprints loop -> implement codergen
     ^                                  |                 |
     |                                  +-- on_done ------+
     +------------- chapter done after inner completion --+
```

More precisely, the outer loop iterates the plan's `checklist.md`, enters the
chapter `plan` body, and reaches success only when all chapters finish. `plan`
returns to an inner loop with no static checklist; it resolves the outer item's
`checklist` field. The implementer returns to that inner loop, whose `on_done`
returns to the outer loop.

The chapter planner reads the project brief and the outer frame's chapter doc.
If the sprint ledger already has open, documented sprints covering the chapter,
it leaves it alone. Otherwise it asks the reviewer one blocking question at a
time only when an answer changes sprint scope or validation, then writes
one-turn sprint docs and appends items with name, observable check, real command
or infer, and doc. It never writes `done` and commits the plan. When the chapter
doc settles the work, it plans silently. The implementer follows the same
current-item, validation, reviewer-question, engine-owned marking, and commit
contract as MEDIUM. No node spawns another run and no repeated-failure
escalation route is added.

## Proof

Add tests named `TestBuiltInLarge` and
`TestLargeRunsNestedPlanningChecklist`. Fix the exact lint-clean graph and
materialized paths in the first. In the second, run the real engine with a
deterministic fake planner and implementer against at least two chapter items
whose distinct sprint ledgers begin empty. Require the planner to populate each
ledger, every sprint command to execute, both nested frame levels to reach the
implementer prompt, inner items to be engine-marked before their chapter, outer
items to be engine-marked in order, and terminal completion. Also exercise the
planner's leave-existing-open-sprints path.

Add `check-sprint-03.sh` beside this document. It runs and proves both named
tests rather than trusting an aggregate green result. Run every BUILD.md gate
before committing.
