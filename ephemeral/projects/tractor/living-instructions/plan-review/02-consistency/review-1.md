Findings:

1. The existing prompt count is both three and four.

   - `declaration.md:30-32`: “`workflow/library/` with graphs, prompts…” and `declaration.md:93-96`: “Move the three graphs and three prompts…”
   - `chapters/04-library/SPRINT-01.md:3-7`: “Move the three graphs and the four prompts…” followed by four prompt paths.
   - `chapters/04-library/SPRINT-03.md:9-12` again says “three existing prompts.”

   Change `declaration.md` (owner: `brief`) and the Sprint 3 references (owner: `decompose`) to four.

2. The library composition action is called both `template` and `include`, and the declaration uses delimiters that are supposed to be literal.

   - `declaration.md:77`: “`{{doctrine "name"}}`, `{{template "name"}}`”
   - `chapters/04-library/SPRINT-01.md:35-46`: “delimiters `<<` and `>>`” and “`include "name"` and `doctrine "name"` actions”
   - `chapters/04-library/SPRINT-04.md:6-9` also names “the `include` and `doctrine` actions.”

   Change `declaration.md` (owner: `brief`) to the actual delimiters and `include` name.

3. Template files are simultaneously covered and excluded by the orphan rule.

   - `planning-workflow.md:252-254`: tests “fail on an unreferenced doctrine or template file.”
   - `chapters/04-library/sprints.md:8-10`: the walk “fails on a doctrine or template file no prompt references.”
   - `chapters/04-library/SPRINT-02.md:59-60`: “Skeletons under `templates/` are not in the walk; P8 does not promise that every skeleton is used.”

   Decision 53 mentions only unreferenced doctrine. Change `planning-workflow.md` and `chapters/04-library/sprints.md` (owner: `decompose`) or explicitly extend the governing promise.

4. Chapter 4’s content boundary disagrees with its declared §6a completeness check.

   - `declaration.md:100-104`: doctrine and skeletons have the check “page list matches `planning-workflow.md` §6a.”
   - `planning-workflow.md:234-242` includes validation-specific pages and `story.md`/`evidence.md`.
   - `chapters/04-library/SPRINT-03.md:9-12`: validation-archetype and reviewer-independence pages “come with their prompts in chapter 5,” while lines 36-40 omit `story.md` and `evidence.md`.

   Change `declaration.md` (owner: `brief`) so Chapter 4 checks only the current-prompt subset.

5. Later chapters are described both as code-free and as requiring workflow-package Go.

   - `declaration.md:86-87`: “chapters 5 and 6 are content plus workflow-package Go.”
   - `declaration.md:91`: “Everything later is content.”
   - `chapters/04-library/CHAPTER.md:5-8`: “After this chapter, every later chapter adds content, not code.”
   - `chapters/05-planner/sprints.md:14-17` requires “per-role resolution in `Build`.”

   Change `declaration.md` (owner: `brief`) and `chapters/04-library/CHAPTER.md` (owner: `decompose`) to say that later work includes workflow-package Go but no engine changes.

6. Universal promises both require and categorically omit holdouts.

   - `decisions.md:187-190`: “Universal… withhold a sample as the holdout. Scenario… no holdout.”
   - `planning-workflow.md:142-144`: “holdouts present for every universal promise.”
   - `declaration.md:53-58` classifies P3, P5, and P8 as universal, while `declaration.md:62-64` says: “No promise here carries a holdout… every other set is checked exhaustively.”

   Change `declaration.md` (owner: `brief`) to comply with the two-archetype rule, or amend the governing rule with an explicit exhaustive-finite-set exception.

7. The validation ledger item is required both to contain and not contain `command`/`infer`.

   - `planning-workflow.md:105-114`: the loop is over `validation/ledger.md`, and design “fills the promise’s checklist item with `command`… and `infer`.”
   - `planning-workflow.md:119-121`: “The ledger item has no `command`.”
   - `decisions.md:216-224` repeats both rules in the same decision.

   Change `planning-workflow.md` and clarify Decision 47 (owner: `design`) by naming the distinct downstream checklist if `command`/`infer` belong somewhere other than `validation/ledger.md`.

8. The initial model table violates its own independence constraints and misclassifies nodes.

   - `planning-workflow.md:51-55`: every supervisor uses “a provider other than the node it watches.”
   - `planning-workflow.md:115-116`: validation review uses an “other provider.”
   - `planning-workflow.md:268-280`: planner nodes including `design` are Claude; validation and plan reviewers are also Claude with a `not` constraint; supervisors are Claude although most watch Claude planner nodes.
   - `planning-workflow.md:63-83` calls `halt` a tool, but line 281 assigns `halt` an LLM model.
   - `planning-workflow.md:124` calls `assemble` cheap, but lines 276-277 assign it the planner’s high-effort role.

   Change the role table in `planning-workflow.md` (owner: `decompose`).

9. The review loop has both six and seven built-in passes.

   - `decisions.md:225-231`, Decision 48: “Passes, in order: traceability, consistency, slicing, proof quality, scope, executability.”
   - `decisions.md:270-272`, Decision 58: “Seven review passes… the seventh is holistic.”
   - `planning-workflow.md:135-152` uses seven.

   Change `decisions.md` to mark Decision 48 as amended by Decision 58 (owner: `brief`).

10. Sprint backlogs are mandatory in one rule and optional in another.

   - `decisions.md:197-202`: “Sprint ledgers get an upfront backlog.”
   - `planning-workflow.md:100-103`: “sprint ledgers may start with a backlog.”

   Change `planning-workflow.md` (owner: `decompose`) from “may” to the decided requirement.

11. The revised build order assigns the v2 planner and execution nodes to the wrong chapter.

   - `decisions.md:288-293`: “the v2 workflow is `planning-workflow.md` (chapter 4…)” and LARGE gains `replan`/`verify` “with chapter 4.”
   - `declaration.md:89-123`: Chapter 4 is the library, Chapter 5 is the planner, and Chapter 6 adds `replan` and `verify`.
   - The corresponding `CHAPTER.md` files use that 4/5/6 split.

   Change the revised build order in `decisions.md` (owner: `decompose`).

12. `chapters.md` says it contains three chapters although its ledger contains six.

   - `chapters.md:4-42` lists Ask and answer, Planning workflow, Execution workflows, The library, The planner, and Execution and the live proof.
   - `chapters.md:47-54`: “Three chapters, in order… Chapter 1… Chapters 2 and 3…”

   Change the prose in `chapters.md` (owner: `decompose`).

13. Chapter completion is attributed both to sprint-loop exit and to verifier passage.

   - `chapters.md:47-50`: “The engine marks a chapter done when its sprint loop exits and the command above holds.”
   - `planning-workflow.md:183-190`: verify runs after the sprint loop, and “A chapter is proven only through this leg.”
   - `chapters/06-execution/CHAPTER.md:36-39`: chapter completion without a preceding verify pass fails P6.

   Change `chapters.md` to state the staged/current behavior or the final verifier-gated behavior explicitly (owner: `decompose`).

14. The anti-overfitting seed-freeze deadline moves by a full chapter.

   - `declaration.md:60`: the two descriptions are “written before chapter 5 starts.”
   - `chapters/06-execution/sprints.md:18-20`: the seeds are written “before sprint 1 of this chapter starts.”

   Change `chapters/06-execution/sprints.md` to require creation before Chapter 5 (owner: `decompose`).

15. Required chapter checks are declared but absent from all new chapter commands.

   - `declaration.md:45-47`: “Required checks (`go build`, `go vet`, `go test`, `golangci-lint`) apply to every chapter.”
   - `chapters.md:29-42`: Chapters 4–6 run `go build` and `go test`, but omit `go vet` and `golangci-lint`.

   Change the Chapter 4–6 commands in `chapters.md` (owner: `decompose`).

16. Doctrine provenance is required to point exclusively under `sources/` in one place but may point to `decisions.md` in another.

   - `chapters/04-library/sprints.md:12-20`: each page has “a provenance pointer into `sources/`,” and the infer prompt fails if it “lacks a pointer to its source under… `sources/`.”
   - `chapters/04-library/SPRINT-03.md:27-32`: each page has a `Source:` line “pointing into `sources/` or `decisions.md`.”

   Change `chapters/04-library/sprints.md` and its infer prompt to allow both sources (owner: `decompose`).

ROUTE: fail
