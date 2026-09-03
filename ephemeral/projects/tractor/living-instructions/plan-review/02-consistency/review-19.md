Findings:

1. `declaration.md:64` makes a check the proof for P10:

   > “the end-to-end run plus the seed's own examples, run by the check, are the proof.”

   `decisions.md:179-184` forbids that:

   > “Checks never prove a promise… Every promise needs a judgment over recorded evidence by a model that is not the coder's.”

   `declaration.md` should add the independent judgment. Owning node: `brief`.

2. P8 changes scope between its two statements. `declaration.md:62` requires:

   > “Every prompt body, doctrine page, supervisor brief, pass, and skeleton is a library file and Go supplies templates only data values… `workflow show` prints every node…”

   After items 1–7 restate P1–P7 in order, `planning-workflow.md:357-360` reduces item 8 to:

   > “`show --stage` reports no diff… and the orphan walk fails when a doctrine page is added…”

   This drops most of P8’s contract and its independent judgment. `planning-workflow.md` should restate the whole promise or label this as only a partial claim. Owning node: `brief`.

3. P8 ownership contradicts the chapter decomposition. `chapters.md:25-30` assigns full P8 to chapter 4:

   > “Every prompt body, doctrine page, supervisor brief, pass, and skeleton… (P8)”

   But `chapters/05-planner/sprints.md:47-50` defers:

   > “`p8-passes-are-files.sh` (P8's pass leg).”

   And `chapters/06-execution/sprints.md:24-28` defers:

   > “`p8-show-stage.sh` (P8's real-stage leg…)”

   Meanwhile the chapter 5 and 6 promise lists omit P8. `chapters.md` and the chapter 5/6 `CHAPTER.md` promise mappings should represent the split explicitly. Owning node: `decompose`.

4. Payload ownership has two incompatible rules. `chapters/04-library/SPRINT-01.md:50-56` says:

   > “`Build` materializes each node's payload with `Render` and adds nothing to it”

   `chapters/04-library/SPRINT-03.md:44-47` says:

   > “`Build` adds nothing to a prompt; tool commands and checklist paths are `Build`'s own”

   The first statement should say “each prompt-bearing node’s prompt.” Document to change: `SPRINT-01.md`. Owning node: `decompose`.

5. The chapter-item command is specified both with and without count guards. `planning-workflow.md:137-142` says:

   > “A chapter item carries as `command` the required checks plus the engine-count guards `chapters.md` uses…”

   `planning-workflow.md:223-225` says:

   > “its item carries the required checks as `command` and nothing else.”

   The execution-side statement should explicitly retain the count guards and say only that there is no `infer`. Owning node: `design`.

6. The holdout pointer is called both build-scoped shared state and run-private state. `declaration.md:80` says:

   > “its path is recorded… at `$XDG_STATE_HOME/tractor/holdouts/<build>.path` by the design lap and read by the execution run”

   `chapters/06-execution/CHAPTER.md:18-21` says:

   > “the path the run's private workflow state carries”

   Those are different storage and handoff contracts. `CHAPTER.md` should name the build-scoped XDG pointer. Owning node: `decompose`.

7. Review passes are extensible, but approval is fixed to seven outcomes. `planning-workflow.md:152-154` says:

   > “`plan-review/ledger.md` from the built-in pass list plus any passes the project adds.”

   `planning-workflow.md:184-186` says:

   > “Shows the human… the seven pass outcomes.”

   Approval should show every ledger outcome, including project-added passes. Document to change: `planning-workflow.md`. Owning node: `brief`.

8. Chapter 6 promises per-lap replanning before its own first sprint supplies `replan`. `declaration.md:88-91` says:

   > “later chapters… are re-planned each lap once `replan` exists (chapter 6; until then the plan node edits the backlog when it runs)”

   `chapters/06-execution/sprints.md:8-13` says:

   > “written when the chapter is entered, re-planned each lap”  
   > “1. `replan` in `medium` and `large`…”

   Since the execution graph is materialized before the run, chapter 6’s current run cannot acquire `replan` after sprint 1. `sprints.md` needs an explicit bootstrap/new-run boundary. Owning node: `decompose`.

ROUTE: fail


