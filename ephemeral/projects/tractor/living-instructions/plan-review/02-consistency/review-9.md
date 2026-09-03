1. `planning-workflow.md` broadens P1’s input contract.

   - `declaration.md`: “Given a seed that names three features and an empty repository...”
   - `planning-workflow.md`: “Given a one-paragraph seed and an empty repository...”

   A one-paragraph seed need not name three features, so these claims quantify different inputs. Change `planning-workflow.md` §8 to match P1. Owning node: `brief`.

2. P8 requires a real recorded stage, but Sprint 2 proves `--stage` against a synthetic one.

   - `planning-workflow.md`: “`show --stage` against a recorded stage of a real run reports no diff”
   - `chapters/04-library/SPRINT-02.md`: “`--stage` exits 0 on a stage built from the frame preamble (`fixtures/frame-preamble.txt`) plus that program’s output”

   Change `SPRINT-02.md` and its ledger proof to use a stage produced by an actual run. Owning node: `decompose`.

3. “Prompt” has two incompatible meanings in the orphan rule.

   - `declaration.md` distinguishes the types: “Every prompt body, doctrine page, supervisor brief, pass, and skeleton is a library file” and then requires “every doctrine page [to be] referenced by at least one prompt.”
   - `chapters/04-library/SPRINT-03.md` instead permits citations from “`prompts/`, `supervisors/`, or `passes/`” and declares: “All three directories are prompts in P8’s sense.”

   Under the sprint definition, a doctrine page cited only by a supervisor brief or pass satisfies the test despite the declaration distinguishing those from prompts. Change `declaration.md` to name the intended umbrella—such as agent-facing library content—or narrow `SPRINT-03.md`; the declaration should own the terminology. Owning node: `brief`.

4. The validation archetype belongs to promises, but `planning-workflow.md` assigns it to chapters.

   - `decisions.md` decision 42: “Scenario (a user story): no holdout”
   - `planning-workflow.md`: “a scenario chapter has none”

   No “scenario chapter” concept is defined, and a chapter may contain promises of different archetypes. Change `planning-workflow.md` to say scenario promise or scenario validation design. Owning node: `design`.

5. Chapter 4 both claims and excludes decision 54.

   - `chapters/04-library/CHAPTER.md`: “Decisions 53, 54, 59.”
   - `decisions.md` decision 54: “`workflow/library/models.yaml` maps roles to provider, model, and effort”
   - The same `CHAPTER.md` lists as a non-goal: “`models.yaml` (chapter 5)”

   Remove decision 54 from Chapter 4’s vector or move `models.yaml` into Chapter 4. The existing chapter split indicates the former. Owning node: `decompose`.

6. Chapter 4’s sprint ledger says every item uses both validation gates, but its first item does not.

   - `chapters/04-library/sprints.md`: “Each item’s `command` and `infer` together are its definition of done”
   - Its first item contains a `command` but no `infer`.
   - The governing format in `loop-node.md` §2 makes `command` optional and `infer` conditional rather than universally paired.

   Change the blanket sentence or add the missing inference gate. Owning node: `decompose`.

ROUTE: fail
