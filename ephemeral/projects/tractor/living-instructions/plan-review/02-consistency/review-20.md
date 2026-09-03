Findings:

1. Chapter-to-promise ownership still drifts.

   - `declaration.md:99`: “Promise P8.”
   - `declaration.md:122`: “Promises P1 to P5, P7.”
   - `declaration.md:134`: “Promises P6, P9, P10.”
   - `chapters/05-planner/CHAPTER.md:9` assigns chapter 5 “P8's pass leg” and part of P9.
   - `chapters/06-execution/CHAPTER.md:9-10` assigns chapter 6 P8’s real-stage leg.

   `declaration.md` should represent the P8 and P9 splits. Owning node: `brief`.

2. Chapter 4 both claims and defers P8’s real-stage proof.

   - `chapters.md:26`: “`--stage` diffing a real stage (P8, except its … real-stage leg, proven in chapter 6).”
   - `chapters/04-library/CHAPTER.md:22-24`: “`--stage` diffs a real stage minus its frame.”
   - `chapters/04-library/SPRINT-02.md:53-55` instead constructs the stage from `fixtures/frame-preamble.txt` and dumper output.
   - `chapters/06-execution/sprints.md:31-32` assigns the recorded-real-run comparison to chapter 6.

   `chapters.md` and chapter 4’s `CHAPTER.md` should call chapter 4’s input a frame-shaped fixture and reserve real stages for chapter 6. Owning node: `decompose`.

3. Chapter 4’s final check still requires an immutable pre-migration output while a later sprint intentionally changes it.

   - `chapters.md:26`: “`Build` returns what it returned before the migration.”
   - `chapters/04-library/SPRINT-04.md:53-56`: doctrine inclusion “changes prompt text” and updates the snapshot.
   - `chapters/04-library/CHAPTER.md:41-43` limits byte identity to sprint 1 and permits reviewed content changes from sprint 4 onward.

   `chapters.md` should state that sprint 1 preserves the baseline and later reviewed edits update it. Owning node: `decompose`.

4. P10’s sprint sketch still calls checks the proof after the declaration added the required independent judgment.

   - `decisions.md:179-184`: “Checks never prove a promise. Every promise needs a judgment over recorded evidence by a model that is not the coder's.”
   - `declaration.md:64` now requires “an `infer` judge over the proof tooling's logs, the run directories, and the built program's transcripts.”
   - `chapters/06-execution/sprints.md:25-31` says the handoff and acceptance examples are the “Proof” for each half of P10, without the judge.

   Chapter 6’s `sprints.md` should include P10’s independent judgment in both seed items. Owning node: `decompose`.

5. The bootstrap exception covers `replan` but not the later-added `verify`.

   - `chapters.md:50-52`: “from chapter 6 on, the `large` workflow puts a `verify` turn between the sprint loop's exit and the mark.”
   - `chapters/06-execution/sprints.md:8-12` acknowledges that chapter 6’s current run was materialized before item 1 adds `replan`.
   - `chapters/06-execution/sprints.md:18` does not add `verify` until item 2.

   The already-materialized chapter 6 run cannot acquire `verify` either. `chapters.md` and chapter 6’s `sprints.md` should say verifier-gated marking begins with runs started after item 2, including the seed runs, not with chapter 6’s own mark. Owning node: `decompose`.

6. `planning-workflow.md`’s duplicate promise list still stops at P8.

   - `planning-workflow.md:333` introduces “Claims to demonstrate before this replaces `plan`,” but the file ends at item 8/P8 on line 367.
   - `declaration.md:63-64` defines P9 and P10.
   - `chapters/06-execution/CHAPTER.md:8-10` assigns both to the implementation plan.

   The cutoff does not explain the omission because the list already includes chapter-6 obligations from P6 and P8. `planning-workflow.md` should add P9 and P10 or explicitly define a different subset. Owning node: `brief`.

ROUTE: fail


