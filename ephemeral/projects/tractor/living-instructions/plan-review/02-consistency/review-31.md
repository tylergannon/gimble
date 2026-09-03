1. `chapters/04-library/SPRINT-01.md` permits either seam: “a package-level `func parseDefinition...` ... or an exported `BuildFrom...`. Pick the smaller.” But `chapters/04-library/SPRINT-03.md` assumes one choice: “the `BuildFrom(fs.FS)` seam from sprint 1.” Sprint 3 can therefore reference a seam sprint 1 explicitly allowed the implementer not to create. Change `SPRINT-03.md` to reference the chosen filesystem-injection seam, or make `BuildFrom` mandatory in `SPRINT-01.md`. Owner: `decompose`.

2. `decisions.md` decision 47 says every design lap “fills the `command` and `infer` of the sprint item that will demonstrate the promise.” `planning-workflow.md` instead allows that item not to exist yet: “for a chapter whose sprint ledger is still prose (`items: []`), [the design’s `## Verifier` section] is their only home until the execution `plan` node turns the prose into items and copies them in.” Change decision 47 to include the deferred-copy case. Owner: `design`.

3. `planning-workflow.md` calls the mapped sprint item “one per promise; when the promise is split across items, each item carries the leg’s `check`.” Its SIMPLE rule says the opposite: “each lap appending its `command` and `infer` to the same item (the SIMPLE item aggregates; nothing else does).” The first statement also contradicts itself by allowing a promise to split across several items. Change the validation-loop cardinality statement. Owner: `decompose`.

4. `planning-workflow.md` names the design section “The design’s `## Verifier` section,” but its project-layout contract calls the same design component “`design.md (story, evidence, validator, not proven)`.” The actual contract has two names, `Verifier` and `validator`. Change the layout entry to `verifier`. Owner: `design`.

5. `declaration.md` states that `large` has “`verify` at chapter exit reading a holdout outside the workdir,” while `planning-workflow.md` limits that read to “a universal promise with a holdout” and explicitly says “a chapter whose promises are all scenarios has none.” Change the declaration to make the holdout read conditional. Owner: `brief`.

6. `chapters/06-execution/CHAPTER.md` likewise says unconditionally that `verify` runs “reading the holdout from outside the workdir,” then narrows it in its own L1 contract to “for a universal promise with a holdout.” Change the chapter introduction to use the conditional rule. Owner: `decompose`.

7. `declaration.md`’s chapter-4 summary requires “every doctrine page [to be] referenced by a rendered prompt.” P8 in the same file instead permits reference by “an agent-facing library file (a prompt body, supervisor brief, or pass).” A doctrine page referenced only by a supervisor brief or pass satisfies P8 but violates the chapter summary. Change the summary to say “rendered agent-facing library file.” Owner: `brief`.

ROUTE: fail
