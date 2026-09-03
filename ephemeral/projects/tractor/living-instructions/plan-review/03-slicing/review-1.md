1. `chapters/01-ask/CHAPTER.md:34`: “a real agent in a real run asks and gets answered, end to end.” Exercisable: `tractor ask` blocks, `tractor answer` releases it, and the timeline records the exchange. **Pass.** Owner: decompose.

2. `chapters/02-planning/CHAPTER.md:5-9`: “turns a seed into a plan… writes a brief, a checklist… and a size recommendation.” Exercisable: run the built-in planner against a small seed and obtain its planning package. **Pass.** Owner: decompose.

3. `chapters/03-execution/CHAPTER.md:5-9`: “Two more embedded workflows that run a plan from chapter 2 to completion.” Exercisable: execute a planner-produced checklist through MEDIUM, with engine validation and marking. **Pass.** Owner: decompose.

4. `chapters/04-library/CHAPTER.md:12-14`: “editable embedded content with a render test, an orphan walk, and a `show` command.” Exercisable: inspect materialized workflow content through the CLI and exercise P8’s library guarantees. **Pass.** Owner: decompose.

5. `chapters/05-planner/CHAPTER.md:13-15`: “interviews about promises, researches, decomposes into slices, designs a proof per promise, reviews the plan seven ways, and hands the human a package.” Exercisable: a nested planning run produces an approved package and demonstrates P1–P5 and P7. **Pass.** Owner: decompose.

6. `chapters/06-execution/CHAPTER.md:5-9`: “Two seeds go end to end through `plan` and then `medium` or `large`.” Exercisable: plan and execute real seed projects, including verification and proof capture, demonstrating P6, P9, and P10. **Pass.** Owner: decompose.

7. `chapters/04-library/sprints.md:4-7`: “`Build` returns the same graphs it returned before… no prompt body remains in a Go string.” Exercisable: build workflows from embedded files through the existing `Build` surface; this directly demonstrates part of P8 and de-risks the library-template seam. **Pass.** Owner: decompose.

8. `chapters/04-library/sprints.md:8-11`: “`tractor workflow show <name>` prints each node’s prompt… `--stage` diffs against a stage’s prompt.md.” Exercisable: run `show`, compare it to `Build`, and detect both stage drift and synthetic orphans. **Pass.** Owner: decompose.

9. `chapters/04-library/sprints.md:12-20`: “the migrated prompts include them so the orphan walk passes.” Exercisable: render real prompts containing the new doctrine, with orphan and semantic checks; this demonstrates P8’s content/reference path. **Pass.** Owner: decompose.

10. `chapters/04-library/sprints.md:21-32`: “an agent reading only them can add a doctrine page and see it in `show` output.” Exercisable: follow the documented edit-to-render workflow through `workflow show`, de-risking P9’s documentation seam. **Pass.** Owner: decompose.

11. `chapters/05-planner/sprints.md:14-17`: “`models.yaml` and per-role resolution in `Build`… Proof: unit tests plus `workflow show` printing provider and model per node.” Exercisable: resolve content-defined roles through `Build` and inspect the result through the CLI, de-risking the independence path used by P3, P5, and P7. **Pass.** Owner: decompose.

12. `chapters/05-planner/sprints.md:18-21`: “nested run on a seed with a planted source; leaves and findings exist; the timeline shows the branches.” Exercisable: run research from intake through parallel branches, fan-in, persisted findings, and quality halt, de-risking the P2 path. **Pass.** Owner: decompose.

13. `chapters/05-planner/sprints.md:22-24`: “the halt tool node and the loop. Proof: P1 and P2 scenarios.” Exercisable: a scripted planning run elicits and excludes a promise, then halts research correctly. **Pass.** Owner: decompose.

14. `chapters/05-planner/sprints.md:25-27`: “`validate-plan` shape rules pass on both sizes from one seed each.” Exercisable: decompose real seeds into MEDIUM and LARGE packages accepted by the consumer-facing validator, de-risking the package-layout seam crossed by P6 and P10. **Pass.** Owner: decompose.

15. `chapters/05-planner/sprints.md:28-30`: “`design` with both archetypes… `review` with pass and fail edges. Proof: P3 and P4 scenarios.” Exercisable: generate, reject, revise, independently approve, and engine-mark validation designs. **Pass.** Owner: decompose.

16. `chapters/05-planner/sprints.md:31-32`: “generated review ledger, `reviewer` with routes to the owning node; assemble; approve. Proof: P5 scenario.” Exercisable: run all seven review passes through correction routing to an approved package. **Pass.** Owner: decompose.

17. `chapters/05-planner/sprints.md:33-34`: “The four supervisors… Proof: P7 scenario.” Exercisable: run a supervised planning turn and observe a delivered steer changing its output. **Pass.** Owner: decompose.

18. `chapters/05-planner/sprints.md:35`: “Docs and skill.” This names neither an exercised workflow nor a seam-risk proof. By contrast, P9 requires an agent to run the documented commands: `validation/P9/design.md:22-31` says “Run the commands, do not describe them.” As written, this is a horizontal documentation task with nothing exercisable at completion. **Fail.** Owner: decompose.

19. `chapters/06-execution/sprints.md:12-13`: “proof that it edits only open items and never the chapter ledger.” Exercisable: execute `replan` after an implementation lap and inspect its bounded ledger mutation, de-risking the ledger/package seam crossed by P6 and P10. **Pass.** Owner: decompose.

20. `chapters/06-execution/sprints.md:14-17`: “a small package with one universal promise whose holdout sample the coder never sees.” Exercisable: execute `verify` with the external holdout and observe pass/fail routing, directly de-risking P6’s verification seam. **Pass.** Owner: decompose.

21. `chapters/06-execution/sprints.md:18-20`: “`plan` then `medium` for one, `plan` then `large` for the other. Proof: P10.” Exercisable: two complete planning-to-running product paths satisfying their seed examples. **Pass.** Owner: decompose.

22. `chapters/06-execution/sprints.md:21-22`: “Docs and skill (P9).” The explicit P9 reference supplies an end-to-end exercise: `validation/P9/design.md:54-72` requires the docs-only reader to complete `plan`, use `ask`/`answer`, and inspect materialized prompts with `show`. **Pass.** Owner: decompose.

23. `chapters/06-execution/sprints.md:23`: “Proof record and closeout.” This merely packages evidence produced by earlier slices. It names no new executable behavior, promise demonstration, or seam-risk exercise, so it is a horizontal closeout task and must be folded into the slice whose proof it records. **Fail.** Owner: decompose.

ROUTE: fail
