Findings:

1. Chapter 1 — `chapters.md`: “An agent inside a run asks a question with `tractor ask`, the run's timeline records it, and `tractor answer` unblocks the agent with the answer.” At its end, a real run can complete an ask/answer exchange end to end. Owning node: `decompose`.

2. Chapter 2 — `chapters.md`: “A built-in planning workflow interviews the caller through `tractor ask` and ends with a brief, a loop-format checklist, and a size recommendation.” At its end, a seed can be planned into usable execution artifacts. Owning node: `decompose`.

3. Chapter 3 — `chapters.md`: “Built-in MEDIUM and LARGE execution workflows run a planning output to completion on the loop node.” At its end, a generated plan can execute end to end. Owning node: `decompose`.

4. Chapter 4 — `chapters.md`: “Every prompt, doctrine page, and skeleton the built-in workflows use is an embedded library file with a render test, an orphan walk, and a `workflow show` command; `Build` returns what it returned before.” At its end, existing workflows remain runnable while P8 is exercised through `workflow show`, rendering, and orphan detection. Owning node: `decompose`.

5. Chapter 5 — `chapters.md`: “The `plan` workflow is the v2 graph with the brief/research loop, the validation design loop, the seven-pass review loop, and four supervisors, all as library content with `models.yaml` roles; P1 to P5 and P7 demonstrated by nested runs.” At its end, the complete planner runs from intake through approval and demonstrates its assigned promises. Owning node: `decompose`.

6. Chapter 6 — `chapters.md`: “`medium` and `large` carry `replan`, `large` carries `verify` with the holdout handoff, two seeds run end to end through plan and execution, and the proof record is written; P6, P9, P10.” At its end, both planning and execution are exercisable on real seeds with verification and recorded proof. Owning node: `decompose`.

7. Chapter 4, sprint 1 — `chapters/04-library/sprints.md`: “`Build` returns the same graphs it returned before, byte for byte in every prompt and command; no prompt body remains in a Go string.” At its end, all three existing workflows remain runnable through the migrated library, demonstrating the first part of P8. Owning node: `decompose`.

8. Chapter 4, sprint 2 — `chapters/04-library/sprints.md`: “`tractor workflow show <name>` prints each node's prompt and command as `Build` materialized them for the given parameters; `--stage <dir>` diffs against a stage's prompt.md with the frame stripped.” At its end, users can exercise the new CLI against built and recorded workflow materializations, directly demonstrating P8. Owning node: `decompose`.

9. Chapter 4, sprint 3 — `chapters/04-library/sprints.md`: “The doctrine pages and artifact skeletons … exist under the library … and the migrated prompts include them so the orphan walk passes.” At its end, existing workflows render and expose the wired doctrine through `show`, exercising another P8 path. Owning node: `decompose`.

10. Chapter 4, sprint 4 — `chapters/04-library/sprints.md`: “an agent reading only them can add a doctrine page and see it in `show` output.” At its end, the documented edit-to-render workflow is exercisable by a docs-only reader, demonstrating the chapter-4 portion of P9. Owning node: `decompose`.

11. Chapter 5, sprint 1 — `chapters/05-planner/sprints.md`: “Proof: unit tests plus `workflow show` printing provider and model per node.” At its end, `models.yaml` flows through `Build` to the CLI, while unsatisfiable independence constraints fail lint; this exercises and de-risks the provider-selection path crossed by P3 and P5. Owning node: `decompose`.

12. Chapter 5, sprint 2 — `chapters/05-planner/sprints.md`: “Proof: nested run on a seed with a planted source; leaves and findings exist; the timeline shows the branches.” At its end, seed-to-research-artifacts fan-out and fan-in are exercisable, de-risking the research side of P2. Owning node: `decompose`.

13. Chapter 5, sprint 3 — `chapters/05-planner/sprints.md`: “Brief node: elicit then prune, batched question files, promise list, exclusions, open plan entries; the halt tool node and the loop. Proof: P1 and P2 scenarios.” At its end, the interview/research loop is runnable and directly demonstrates P1 and P2. Owning node: `decompose`.

14. Chapter 5, sprint 4 — `chapters/05-planner/sprints.md`: “Proof: `validate-plan` shape rules pass on both sizes from one seed each.” At its end, seeds can produce accepted MEDIUM and LARGE packages, exercising and de-risking the package-layout seam from `plan` to execution. Owning node: `decompose`.

15. Chapter 5, sprint 5 — `chapters/05-planner/sprints.md`: “Validation ledger generation; `design` with both archetypes and the holdout writer; `review` with pass and fail edges. Proof: P3 and P4 scenarios.” At its end, both successful and rejected validation designs traverse the loop end to end, directly demonstrating P3 and P4. Owning node: `decompose`.

16. Chapter 5, sprint 6 — `chapters/05-planner/sprints.md`: “Proof: a nested run in which one pass fails, its owning node runs, and the pass is re-selected and marked.” At its end, reviewer failure, owner correction, reselection, and engine marking are exercisable as one path, de-risking P5. Owning node: `decompose`.

17. Chapter 5, sprint 7 — `chapters/05-planner/sprints.md`: “`assemble` and `approve`, the human gate. Proof: P5 scenario.” At its end, the seven-pass-reviewed package can be assembled and approved through the human gate, completing the P5 scenario. Owning node: `decompose`.

18. Chapter 5, sprint 8 — `chapters/05-planner/sprints.md`: “The four supervisors in the YAML with `models.yaml` roles. Proof: P7 scenario.” At its end, a planning run can receive and react to supervisor steering, directly demonstrating P7. Owning node: `decompose`.

19. Chapter 5, sprint 9 — `chapters/05-planner/sprints.md`: “an agent reading only the docs starts a nested plan run on a one-line seed and answers its first question.” At its end, the documented plan-and-answer path is exercised by a docs-only reader, demonstrating part of P9. Owning node: `decompose`.

20. Chapter 6, sprint 1 — `chapters/06-execution/sprints.md`: “proof that it edits only open items and never the chapter ledger.” At its end, `medium` and `large` can execute a replan lap while preserving engine-owned chapter state, exercising and de-risking the checklist-item seam. Owning node: `decompose`.

21. Chapter 6, sprint 2 — `chapters/06-execution/sprints.md`: “Proof: a small package with one universal promise whose holdout sample the coder never sees.” At its end, `large` can hand a hidden sample to `verify` and route on its result, exercising the holdout seam and de-risking P6. Owning node: `decompose`.

22. Chapter 6, sprint 3 — `chapters/06-execution/sprints.md`: “`plan` then `medium` for one, `plan` then `large` for the other; the proof record … is written from these runs in the same sprint. Proof: P10.” At its end, both seeds execute from description to accepted program and proof record, directly demonstrating P10. Owning node: `decompose`.

23. Chapter 6, sprint 4 — `chapters/06-execution/sprints.md`: “the docs-only reader runs plan, ask, answer, and show. Proof: P9.” At its end, the complete documented user path is exercised without implementation context, directly demonstrating P9. Owning node: `decompose`.

ROUTE: pass
