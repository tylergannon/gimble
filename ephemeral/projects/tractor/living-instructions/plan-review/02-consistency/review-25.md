1. `ceiling`/`halted` are required nodes but disappear from the implementation inventories.

   - `planning-workflow.md`: “`brief`'s `max_visits` is the ceiling; its exhaustion edge leads to `ceiling` … and routes back to `brief` or to `halted`.”
   - `declaration.md`: “`plan` v2: intake, the brief/research loop with its tool-node halt, decompose, the validation design loop, assemble, the plan review loop, approve; four supervisors.”
   - `chapters/05-planner/CHAPTER.md` repeats the latter inventory, while `sprints.md` specifies only “The `index_gate` tool node and the `halt` tool node with the loop's routes.”

   The declared and decomposed graph omits the mandatory exhaustion branch. Change `declaration.md` (`brief`) and chapter 05’s `CHAPTER.md`/`sprints.md` (`decompose`).

2. SIMPLE decomposition is required in the design but excluded from the chapter-5 work.

   - `decisions.md` decision 51: “SIMPLE: the same graph, with `decompose` writing a one-item ledger.”
   - `planning-workflow.md`: “SIMPLE | the full graph; `decompose` writes a one-item sprint ledger.”
   - `declaration.md`: “decompose for MEDIUM and LARGE.”
   - `chapters/05-planner/sprints.md`: “`decompose` for MEDIUM and LARGE … Proof: `validate-plan` shape rules pass on both sizes.”

   The implementation and proof cover two sizes where the governing rule requires three. Change `declaration.md` (`brief`) and chapter 05’s `sprints.md` (`decompose`).

3. Chapter 5 gives two incompatible rules for sprint proof.

   - `chapters/05-planner/CHAPTER.md`: “Each sprint's proof is a nested run on a small seed with the scripted observer, inspected from the timeline and the ledgers.”
   - `chapters/05-planner/sprints.md`, item 1: “Proof: unit tests plus `show` output.”

   Item 1 has no nested run, timeline, ledger, or scripted observer. Change chapter 05’s `CHAPTER.md` to narrow the review-posture rule, or change item 1’s proof; owner: `decompose`.

4. The `verify`-failure rule alternates between conditional escalation and unconditional repair insertion.

   - `planning-workflow.md`: “On the `verify` fail edge it instead appends one open item from the verifier's findings (or asks the human when the findings say the chapter is wrong).”
   - `chapters/06-execution/CHAPTER.md`: “fail to `replan`, which appends a repair sprint or asks the human.”
   - `chapters/06-execution/sprints.md`: “on the `verify` fail edge it appends exactly one open item.”

   The sprint requires an append even in the case where the governing rule requires a human question. Change chapter 06’s `sprints.md`; owner: `decompose`.

ROUTE: fail


