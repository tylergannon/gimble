1. The doctrine-page size limit has three drifting names and thresholds:

   - `declaration.md`: “each page is **under a screen**”
   - `chapters/04-library/sprints.md`: “each **a page or less**”
   - `chapters/04-library/SPRINT-04.md`: “Each page: **under about sixty lines**”

   A screen, a page, and roughly sixty lines are not equivalent acceptance rules. Reconcile `declaration.md` (owning node: `brief`) and the chapter 4 sprint documents (owning node: `decompose`) around one measurable limit.

2. The orphan rule disagrees about what may reference doctrine:

   - `declaration.md`: “every doctrine page is referenced by at least one **agent-facing library file (a prompt body, supervisor brief, or pass)** that some node renders”
   - `planning-workflow.md`: “fail on a doctrine page no rendered **prompt** references”
   - `chapters/04-library/sprints.md`: “fails … on a doctrine page no rendered **prompt** references”

   A doctrine page referenced only by a supervisor brief or review pass satisfies the declaration but fails the narrower “prompt” rule. `planning-workflow.md` and `chapters/04-library/sprints.md` should adopt the declaration’s canonical “agent-facing library file” rule. Owning node: `decompose`.

ROUTE: fail
