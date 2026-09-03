1. `decisions.md` defines a different promise schema from `declaration.md`.

   - `decisions.md:157`: “A promise has a statement, what it must not imply, a scope, a verifier, and evidence.”
   - `declaration.md:45`: “Each promise: the statement, what it must not imply, the archetype … and the verifier.” Its table likewise has no distinct scope or evidence fields.

   File to change: `declaration.md`. Owning node: brief.

2. SIMPLE follows two incompatible paths.

   - `decisions.md:245`: “SIMPLE: brief and one validation lap, recommend self-execution.”
   - `planning-workflow.md:164`: “SIMPLE | the full graph; `decompose` writes a one-item sprint ledger, the validation loop runs one lap, the review loop runs the holistic pass only”.

   File to change: `planning-workflow.md`, unless decision 51 is formally amended. Owning node: decompose.

3. The universal-holdout rule both requires and exempts exhaustive universal promises.

   - `planning-workflow.md:107`: “for a universal promise, the holdout sample”.
   - `planning-workflow.md:145`: “holdouts present for every universal promise”.
   - `planning-workflow.md:303`: “a universal promise over a set the verifier checks exhaustively needs none”.
   - `decisions.md:191`: the same exception is still marked “pending Tyler’s answer,” while `declaration.md:62` declares: “No promise here carries a holdout.”

   Files to change: `planning-workflow.md` and the stale amendment status in `decisions.md`. Owning node: design.

4. Sprint-ledger timing is eager in the governing rule but lazy in chapters 5 and 6.

   - `decisions.md:199`: “Sprint ledgers get an upfront backlog”.
   - `planning-workflow.md:102`: “a sprint ledger per chapter; sprint ledgers start with a backlog.”
   - `chapters/05-planner/sprints.md:3,8`: “items: []” and “Empty ledger; the plan node writes the backlog when the chapter is entered”.
   - `chapters/06-execution/sprints.md:3,8`: “items: []” and “Empty ledger; written when the chapter is entered”.

   Files to change: both chapter sprint ledgers, or the governing rule must state this transitional exception. Owning node: decompose.

5. The proof doctrine says checks never prove a promise, but P8 names only checks as its verifier.

   - `decisions.md:179`: “Checks never prove a promise. … Every promise needs a judgment over recorded evidence by a model that is not the coder’s”.
   - `declaration.md:58`, P8 verifier: “Render test; orphan walk over the embedded tree; `show --stage` against a recorded run reports no diff.”
   - `chapters.md:29-30` gives Chapter 4 only a shell `command`, with no independent judgment covering P8.

   Files to change: `declaration.md` and Chapter 4’s proof plan. Owning node: design.

6. The template contract promises only two functions while also requiring two additional callable actions.

   - `declaration.md:58`: “Go supplies templates only data values and the two rendering functions `quote` and `shell`”.
   - `declaration.md:77`: “the `include "name"` and `doctrine "name"` actions”.
   - `SPRINT-01.md:41-47`: “Two functions in the func map” followed by “`include "name"` and `doctrine "name"` actions exist … Wire the func map now”.

   File to change: `declaration.md` should distinguish value-transforming functions from composition actions, or the sprint contract must remove the extra functions. Owning node: brief.

7. Doctrine inclusion is assigned to two different sprints.

   - `SPRINT-01.md:45`: “`include "name"` and `doctrine "name"` actions exist but are not yet used by any prompt (sprint 3 does that).”
   - `SPRINT-04.md:6`: “The coder’s turn … adds the `doctrine` includes to the four migrated prompts”.

   File to change: `chapters/04-library/SPRINT-01.md`. Owning node: decompose.

8. The canonical doctrine inventory omits a page required by Chapter 4.

   - `planning-workflow.md:237-242` enumerates the doctrine files but does not include `ledger.md`.
   - `SPRINT-04.md:29`: “`doctrine/ledger.md` | planner, both implement prompts”.

   File to change: `planning-workflow.md`. Owning node: decompose.

ROUTE: fail
