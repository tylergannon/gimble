# Chapter 2: The planning workflow

Status: active

A workflow baked into the binary that turns a seed into a plan. It
interviews the caller with `tractor ask` until intent is clear, then writes
a brief, a checklist in the loop node's format, and a size recommendation,
and hands back to the caller. "As important as having a fantastic workflow
engine."

## Pyramid index

- L0: Tractor ships its own planning workflow, and running it is how a
  caller learns how to use Tractor for the job at hand.
- L1:
  - A library of named workflows embedded in the binary, listable and
    runnable by name without a pipeline file.
  - The planning workflow: interview node using `ask`, then a writing node
    that produces `brief.md`, `checklist.md`, and a recommendation under
    `ephemeral/projects/<build>/`.
  - The recommendation names a size (SIMPLE, MEDIUM, LARGE) and, for
    MEDIUM and LARGE, the execution workflow from chapter 3 to run next.
    SIMPLE means "execute the plan yourself."
  - Docs and skill teach the caller to start here.
- L2: the sprint ledger is empty; the `plan` node writes it after
  interviewing the reviewer.

## Vector

Decisions 1, 2, 5, 7, 8, 12, 14, 20, 21 in `decisions.md`, and the brief's
"Definition" dimension: the interview is proportional to the size, keeps
asking only while an answer would change the contract, and stops after
two rounds that surface only derivable detail. The checklist it writes is
the loop node's format from `loop-node.md` §2, with `check` fixed per item
and `command`/`infer` filled where they can be known.

## Planning this chapter

The planner must ask the reviewer at least about:

- How embedded workflows are invoked (a subcommand, a flag on `run`, or a
  name where a pipeline path goes) and where their outputs land.
- The shape of the interview node: one codergen node with a long timeout
  that calls `ask` in a loop, or something else.
- What the recommendation file looks like and how a caller reads it.

Sprints must each fit one agent turn and each carry a real `command`. The
last sprint should be a live proof: the planning workflow run against a
small seed, with the reviewer answering the interview.

## Non-goals

Executing the plan (chapter 3). Multiple planning agents drafting in
parallel. A web client.
