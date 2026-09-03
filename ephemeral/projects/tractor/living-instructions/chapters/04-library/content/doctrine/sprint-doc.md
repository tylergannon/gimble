# The sprint doc

A sprint doc tells one agent, in one fresh context, what to build and
how it will know it is done. Write it for that reader: someone capable
who has never seen the repository and will not ask you a second time.

Contents, in order:

1. Title and one paragraph: the slice, and what is exercisable at its
   end.
2. What to read first: the files and lines this touches, by path. Do
   the survey once, here, so the agent does not.
3. The work: what changes, where, in the order that keeps the build
   green. Name seams the sprint must not cross.
4. Definition of done: the ledger item's `check`, `command`, and
   `infer` restated in prose, and the proof script by name and what it
   proves.
5. Not in this sprint: what a keen agent would add and must not.
6. Ask the reviewer: the decisions this doc leaves open, each with the
   options and a recommendation, so the agent asks well.

Rules:

- One turn. If the doc needs a second turn, split it.
- Commands in the doc run from the repository root as written.
- No guessing left for the agent that the planner could have settled.
- The doc is prose the engine never reads; the ledger item is what
  the engine reads. Keep them saying the same thing.

Source: sources/diffusioninc/.claude/skills/df-sprint-plan/SKILL.md; chapters/01-ask/SPRINT-01.md through SPRINT-03.md (the hand-written standard).
