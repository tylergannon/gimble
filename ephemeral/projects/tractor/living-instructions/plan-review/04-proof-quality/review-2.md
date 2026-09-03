1. `chapters.md:25-30` validates P8 with only:

   > `test -d workflow/library && test "$(grep -c 'done: true' .../sprints.md)" -ge 5`

   But `loop-node.md:62-71` says:

   > “A hand-edited `done: true` is honored without validation.”

   A coding agent can hand-mark all five sprints, create `workflow/library/`, and bypass every sprint `command` and `infer`. The chapter validator requires no `LoopValidated` evidence, so P8 can be false while the chapter passes. Owning node: design. Promise: P8.

2. `validation/P8/design.md:29-32,71-73` defers essential proof:

   > “Passes and the plan-review ledger arrive in chapter 5; the pass leg … runs then.”

   > “then at chapter 5 `prove/p8-passes-are-files.sh`, then at chapter 6 `prove/p8-show-stage.sh`.”

   Neither later chapter ledger contains those commands; chapter 5 and 6 currently have empty item lists, and their backlog sketches never name either script. Chapter 4’s actual ledger also has no aggregate P8 `infer` matching `validation/P8/design.md:155-182`. Thus P8 can be marked while supervisor/pass files and real-stage comparison—the deferred portions of the promise—remain unvalidated. Owning node: design. Promise: P8.

3. The chapter-4 P8 validator is unsatisfiable as written. `chapters/04-library/sprints.md:26` invokes `orphan-walk-and-render.sh`; its sourced `lib.sh:6` enables:

   > `set -eu`

   Yet `orphan-walk-and-render.sh:24-41` assigns only `delim_open` and then uses:

   > `"$delim_close"`

   Later, `orphan-walk-and-render.sh:51-53` uses:

   > `case "$kind" in`

   Neither variable is assigned anywhere. With no doctrine page it dies on `delim_close`; with doctrine pages it dies on `kind`. A conforming implementation cannot satisfy this check, making the design strictly stronger than—and in practice unrelated to—the promise. Owning node: design. Promise: P8.

4. The doctrine/skeleton validator can pass while its ledger check is false. `chapters/04-library/sprints.md:34` promises content:

   > “artifact skeletons … copied from chapters/04-library/content/ … the pages say what decisions 37 to 59 say, in the library’s voice.”

   But `prove/doctrine-pages.sh:18-20` checks skeletons only with:

   > `test -f "$lib/templates/$t"`  
   > `grep -rq "$t" "$lib/prompts"`

   And the ledger’s `infer` prompt at `sprints.md:41` judges only doctrine pages, despite templates being listed as files. Blank or incorrect skeletons merely named by a prompt therefore pass. Owning node: design. Promise: P8.

5. P7’s design is stricter than P7. `declaration.md:60` promises only:

   > “the steered turn’s output differs from what it was writing before the steer.”

   But `validation/P7/design.md:36-43,50-54` additionally requires the edit to concern the bait and be:

   > “in the direction of the steer”

   and:

   > “a response to it”

   A delivered steer followed by a genuine output change satisfies P7 as written but can fail this validator. Owning node: design. Promise: P7.

6. The shared observer’s authenticity rests on an unprotected random pathname. `validation/ledger.md:142-147` claims:

   > “its own tree lives outside the run directory, at a path the check draws at random … so … no agent can find it to edit it.”

   Randomness is not filesystem isolation. The observer path is beside the run directory and passed to a live process; a coder-controlled binary or agent shell can enumerate sibling paths or process arguments and then alter the observer’s snapshots, `answers.log`, or mirror. P2 explicitly trusts those copies to prove the finding was not applied (`validation/P2/design.md:76-85`), so a planner can apply it, rewrite the observer evidence, and pass. The same unauthenticated mechanism is reused across the other scenario designs. Owning node: design. Promise: P2.

ROUTE: fail
