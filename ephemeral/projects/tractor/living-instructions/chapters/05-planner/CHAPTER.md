# Chapter 5: the planner

Status: planned

The `plan` workflow becomes the graph in `planning-workflow.md` §2: intake,
the brief/research loop with its tool-node halt, decompose, the
validation design loop, assemble, the plan review loop, approve; four
supervisors; `models.yaml`. Every node is content in the library from
chapter 4. Promises P1 to P5, P7, P8's pass leg (passes are library files, sprint 11), and the chapter 5 portion of P9.

## Pyramid index

- L0: The planner interviews about promises, researches, decomposes into
  slices, designs a proof per promise, reviews the plan seven ways, and
  hands the human a package, with supervisors steering throughout.
- L1:
  - `models.yaml` assigns provider, model, and effort per role with an
    explicit provider and a `not` constraint, applied in `Build`.
  - Intake and the brief/research loop: elicit then prune, batched
    question files, research branches writing leaves and findings, a halt
    tool node.
  - Decompose writes chapters or sprints as vertical slices.
  - The validation design loop: `design` chooses an archetype and fills
    the item; `review` routes pass or fail.
  - The plan review loop: seven passes, each a fresh reviewer routing
    pass or fail to the owning node.
  - Four supervisors named by their question; assemble and approve.
- L2: sprints in `sprints.md` (backlog; edited when the plan node runs
  until chapter 6 adds `replan`); node
  contracts in `planning-workflow.md` §3; models in §6b; passes in §3.

## Vector

Decisions 37 to 52, 54, 57 to 59. Research R2 (reasoning before routing;
rubric-free seventh pass), R4 (every field exists; explicit providers).
No engine change: every node type is already in the spec.

## Review posture

Each sprint's proof is a nested run on a small seed with the scripted
observer, inspected from the timeline and the ledgers, in the manner of
chapter 2 sprint 4. The implement prompt (library content since chapter
4) forbids the coder to start a child Tractor run; sprint 1 of this
chapter adds the exception: a validator named in the sprint doc may
start one, the coder runs it as the doc says, and the engine runs it
again after the turn. Reviewer and verifier nodes run on a provider other
than the node they judge, and the check reads the provider from the
run's events, not from the YAML.

## Non-goals

`replan` and `verify` (chapter 6). A holdout over this project's own promises (scratch packages under test may carry one).
Multi-level supervision. Fan-out drafts of chapter docs.
