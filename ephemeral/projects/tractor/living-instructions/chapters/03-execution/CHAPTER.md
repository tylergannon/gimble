# Chapter 3: The execution workflows

Status: active

Two more embedded workflows that run a plan from chapter 2 to completion on
the loop node. MEDIUM is one sprint loop over the plan's checklist. LARGE
is the nested shape this very build is using: a chapter loop whose lap is
a plan node and a sprint loop. Each may ask the reviewer through `ask` when
a validator is missing or a sprint fails repeatedly.

## Pyramid index

- L0: Tractor ships the workflows its own planner recommends.
- L1:
  - MEDIUM: loop over `checklist.md`, body implements one item, engine
    validates and marks.
  - LARGE: `chapters.md` outer loop with a `plan` body, sprint loop inside,
    exactly `pipeline.yaml` of this build made generic.
  - Both take the build directory as their single parameter.
  - Docs and skill: the planner's recommendation names one of these and
    the caller runs it by name.
- L2: the sprint ledger is empty; the `plan` node writes it after
  interviewing the reviewer.

## Vector

Decisions 8, 10, 11, 15, 16, 17, 18, 19 in `decisions.md`; `loop-node.md`
throughout. One run for the whole job. The agent never marks items. A
failed item is re-entered; `max_visits` is the only ceiling.

## Planning this chapter

The planner must ask the reviewer at least about:

- How a workflow takes the build directory as a parameter given that
  `$goal` is the only substitution (decision 25).
- Whether the LARGE workflow's `plan` node should be allowed to ask the
  reviewer, or plan silently from the chapter doc.

The last sprint should be a live proof: MEDIUM run against a plan chapter
2's workflow produced.

## Non-goals

Escalation on repeated failure. Parallel sprints. Anything that spawns a
child run.
