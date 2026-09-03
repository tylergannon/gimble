1. `planning-workflow.md` contradicts itself about where the holdout path is recorded:

   > “a random-token directory whose path is recorded only in the verifier prompt the workflow will later materialize”

   Later:

   > “Reads the validation design and any holdout path the design recorded”

   `chapters/06-execution/CHAPTER.md` repeats the latter rule:

   > “reads the validation design and any holdout path it recorded”

   This conflicts with `declaration.md`:

   > “recorded in the run's private workflow state, rendered only into `verify`.”

   Change `planning-workflow.md` §5 and `chapters/06-execution/CHAPTER.md` so the design records no path. Owning nodes: `design` for the handoff contract; `decompose` for the chapter copy.

2. The template API excludes and requires the same functions. `chapters/04-library/CHAPTER.md` says:

   > “Go supplies data values and the rendering functions `quote` and `shell` to templates and nothing else”

   But `chapters/04-library/SPRINT-01.md` requires:

   > “`include "name"` and `doctrine "name"` actions exist”

   and:

   > “Wire the func map now”

   `declaration.md` also explicitly promises all four:

   > “the two value functions `quote` and `shell`, and the two composition actions `include` and `doctrine`”

   Decision 59 repeats the restrictive two-function wording. Change `chapters/04-library/CHAPTER.md` and decision 59 to include the composition actions. Owning node: `decompose`.

3. `chapters/04-library/SPRINT-01.md` assigns the content-only follow-up to two different sprints in the same rule:

   > “not yet used by any prompt (sprint 4 does that). Wire the func map now so sprint 3 is content only.”

   `SPRINT-03.md` instead says:

   > “Doctrine pages (sprint 4).”

   Change “sprint 3” to “sprint 4” in `SPRINT-01.md`. Owning node: `decompose`.

4. The exhaustive-universal holdout exception is simultaneously pending and adopted. Decision 42 says:

   > “a universal promise over a set the verifier checks exhaustively needs none (amendment proposed in interview 0014, question 4, pending Tyler's answer).”

   But `declaration.md` treats it as settled:

   > “No promise here carries a holdout… every other set is checked exhaustively.”

   Decision 56 likewise states:

   > “No holdout in this project's proof.”

   Change decision 42 to record the final ruling rather than “pending.” Owning node: `design`.

5. The validation artifact has two incompatible names/shapes in `planning-workflow.md`. The project layout specifies one combined artifact:

   > “`<promise>/ design.md (story, evidence, validator, not proven)`”

   The library skeleton list instead specifies:

   > “`CHAPTER.md SPRINT.md story.md evidence.md`”

   Chapter 5 reinforces the single-artifact form:

   > “a nested run leaves one design per promise”

   Change the skeleton list to `design.md`, or explicitly define `story.md` and `evidence.md` as partials composed into it. Owning node: `design`.

ROUTE: fail
