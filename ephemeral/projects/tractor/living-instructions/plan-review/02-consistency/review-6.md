1. `planning-workflow.md` gives the same `research` node two incompatible successors. The graph says:

   > `intake ─▶ research ─▶ brief ─▶ research ─▶ halt?`

   But its node contract says:

   > “A tool node, `index_gate`, after the fan-in checks the index against the quality bar and routes to `halt`.”

   Thus the initial research pass can reach `halt`, whose empty state routes directly to `decompose`, skipping the required first interview. Document to change: `planning-workflow.md`. Owning node: `brief`.

2. `planning-workflow.md` promises a `max_visits` human escalation that neither its graph nor Tractor supplies:

   > “The enclosing `loop` node's `max_visits` is the ceiling, and reaching it is a question to the human, not an exit.”

   Tractor’s `docs/spec.md` states:

   > “If exclusion leaves a node with an empty offered set, the run fails with an explicit reason. Authors who want a softer landing draw an edge to an escalation node…”

   No such escalation node appears in the proposed graph. This also conflicts with the supposedly exhaustive list:

   > “Human gates… the promise interview in `brief`, proof-mechanism questions inside the validation design loop, and `approve`.”

   Document to change: `planning-workflow.md`, with the corresponding count clarified in `decisions.md` decision 50. Owning node: `brief`.

3. P4’s declared evidence permits the rejected design to bypass its second review. `declaration.md` says:

   > “the timeline shows `review` routing to `design`, then to the loop; the item ends `done: true`.”

   `planning-workflow.md` instead says:

   > “It routes pass to the loop node and fail to `design`.”

   and:

   > “only the pass edge returns to the loop.”

   The declaration must require `review → design → review → loop`, with the final review routing pass. Document to change: `declaration.md`. Owning node: `brief`.

4. P6 is restated with a weaker execution requirement. `declaration.md` requires:

   > “`tractor workflow run large` on a v2 package reaches `COMPLETED`, and every chapter's `done: true` is preceded… by a `verify` turn…”

   `planning-workflow.md` claim 6 allows:

   > “`tractor workflow run medium` (or `large`)…”

   A successful `medium` run has no chapter verification and therefore cannot demonstrate P6. Document to change: `planning-workflow.md`. Owning node: `brief`.

5. The holdout path is assigned two incompatible storage rules inside `planning-workflow.md`. The design-loop section says:

   > “the holdout sample under the XDG state root in a random-token directory whose path is recorded only in the verifier prompt…”

   The execution section says:

   > “the holdout path the run's private workflow state carries (written there by the design lap and rendered only into this prompt…)”

   `declaration.md` agrees with the latter:

   > “recorded in the run's private workflow state, rendered only into `verify`.”

   “Recorded only in the prompt” must become “stored in private workflow state and disclosed only to the verifier prompt.” Document to change: `planning-workflow.md`. Owning node: `design`.

6. Supervisor output from `workflow show` is simultaneously mandatory and undecided. `planning-workflow.md` says:

   > “`tractor workflow show <name>…` prints each node's prompt and supervisor brief exactly as `Build` materialized them…”

   `chapters/04-library/SPRINT-02.md` says:

   > “Whether `show` should also print supervisor briefs once supervisors exist… Recommend yes… nothing to do now.”

   A fixed requirement cannot remain a reviewer choice, and chapter 5’s backlog does not explicitly schedule the deferred work. Document to change: `chapters/04-library/SPRINT-02.md` and the chapter 5 backlog if implementation is not already generic. Owning node: `decompose`.

7. Doctrine authorship is ordered both before and after delimiter selection. `planning-workflow.md` says:

   > “The template delimiters are non-default, chosen by the coder before any page is written…”

   `chapters/04-library/SPRINT-04.md` says the pages are:

   > “Written by Claude, not the coder… before the chapter's run starts,”

   and only afterward:

   > “The coder's turn copies them into `workflow/library/doctrine/`…”

   The governing rule should say delimiters are selected before pages are installed/rendered, or the chapter order must change. Document to change: `planning-workflow.md`. Owning node: `decompose`.

8. Chapter 5 claims per-lap replanning without the node that performs it. `chapters/05-planner/sprints.md` says:

   > “the `replan` step (or, until it exists, the plan node on re-entry) edits open items after every lap…”

   But `chapters/05-planner/CHAPTER.md` makes:

   > “`replan` and `verify` (chapter 6)”

   non-goals, while decision 44 specifically requires:

   > “a `replan` node after every implement lap.”

   The current `large` graph routes `implement` back to `sprints`; it does not re-enter `plan` after each implementation lap. Document to change: `chapters/05-planner/sprints.md`. Owning node: `decompose`.

ROUTE: fail
