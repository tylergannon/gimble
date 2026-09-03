1. `research/INDEX.md` has two names. `planning-workflow.md` says “a research directory with a **routing index**,” but later calls `research/INDEX.md` “the **routing tree**.” Change `planning-workflow.md` to use one term. Owner: `brief`.

2. The index-checking node has three names. `declaration.md` says “**quality tool node**”; `planning-workflow.md` says “the **index quality gate**”; `chapters/05-planner/sprints.md` says “The **index quality tool node**.” Establish one node ID and use it throughout. Documents/owners: `declaration.md` → `brief`; `chapters/05-planner/sprints.md` → `decompose`; `planning-workflow.md` → `brief`.

3. SIMPLE’s route contradicts the graph. `planning-workflow.md`’s graph routes `halt` unconditionally through “**decompose**” before validation, while its size table says SIMPLE is “**intake, one research lap, brief, one validation lap, review, approve**,” omitting both `halt` and `decompose`. Change `planning-workflow.md` to show the SIMPLE branch explicitly. Owner: `brief`.

4. P3 is restated with opposite holdout requirements. `declaration.md` says: “**No holdout: the set is small and checked exhaustively**” and “**No promise here carries a holdout**.” `planning-workflow.md` claim 3 repeats the ledger/review claim but adds: “**at least one universal promise has a holdout outside the workdir and outside the run directory**.” This also conflicts with its own exclusion: “**A holdout in this project's own proof (none of its sets is large enough)**.” Change `planning-workflow.md` claim 3. Owner: `design`.

5. Decision 42 uses “universal” incompatibly in one rule. It first says: “**Universal (‘for every case in a set’): withhold a sample as the holdout**,” then concludes: “**Holdouts are always considered, never universal**.” As written, the second statement negates or ambiguously reuses the archetype name established by the first. Clarify `decisions.md`. Owner: `design`.

6. Holdouts are optional by archetype but mandatory in the verifier descriptions. `decisions.md` says: “**Scenario (a user story): no holdout**.” Conversely, `planning-workflow.md` says `verify` “**Reads the validation design and the holdout path**,” and `chapters/06-execution/CHAPTER.md` says `verify` “**reads the validation design and the holdout path**.” Since `verify` runs once per chapter, scenario-only chapters have no such path. Change both descriptions to read any applicable holdout path. Owner: `decompose`.

7. Chapter 4 has incompatible sprint maps. `chapters/04-library/sprints.md` says “**Five sprints**,” while `chapters/04-library/CHAPTER.md` says “**sprints 1 to 4**.” `declaration.md` assigns doctrine to “**chapter 4 sprint 3**,” but the actual document is `SPRINT-04.md`. `SPRINT-01.md` retains the old map: “**render-all and orphan tests (sprint 2), new content (sprint 3), docs (sprint 4)**,” whereas those are now sprints 3, 4, and 5. Documents/owners: `declaration.md` → `brief`; `CHAPTER.md` and `SPRINT-01.md` → `decompose`.

8. Chapter 4 both forbids and requires changing prompt text. `chapters.md` promises “**Build returns what it returned before**,” and `chapters/04-library/CHAPTER.md` says “**a sprint that changes what a prompt says has failed, not refactored**.” But `SPRINT-04.md` requires: “**This changes prompt text, so the sprint 1 byte-equality snapshot is updated in the same commit**.” Change the chapter-level statements to limit byte identity to the migration sprint and require reviewed changes afterward. Owner: `decompose`.

9. The template input boundary drifts. `declaration.md` and decision 59 allow “**data values and the two rendering functions `quote` and `shell`**.” `chapters/04-library/CHAPTER.md` instead says: “**Go supplies data values to templates and nothing else**.” Change `CHAPTER.md` to name the two permitted functions. Owner: `decompose`.

10. Template-reference requirements conflict. `declaration.md` requires a “**render test that every template renders and every doctrine and template file is referenced**.” `SPRINT-03.md` says: “**Skeletons under `templates/` are not in the walk; P8 does not promise that every skeleton is used**.” Change `declaration.md` so only doctrine files must be referenced. Owner: `brief`.

11. The chapter status vocabulary conflicts with the chapters it templates. `chapters/04-library/content/templates/CHAPTER.md` permits only “**Status: `<active | done>`**,” while both `chapters/05-planner/CHAPTER.md` and `chapters/06-execution/CHAPTER.md` say “**Status: planned**.” Add `planned` to the template or change the chapter statuses. Owner: `decompose`.

ROUTE: fail
