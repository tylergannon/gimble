1. `planning-workflow.md` overstates the library boundary:

   > “Everything an agent is told lives as a file under `workflow/library/`”

   but later says:

   > “Frames and `$goal` are engine additions”

   `chapters/04-library/CHAPTER.md` repeats the overstatement:

   > “Everything an agent is told by a built-in workflow becomes a file under `workflow/library/`”

   Runtime frames are agent-visible but remain engine content. Change the two “everything” statements to the narrower P8 inventory. Owning nodes: `brief` for `planning-workflow.md`; `decompose` for `CHAPTER.md`.

2. `$goal` handling by `workflow show` is contradictory. `chapters/04-library/SPRINT-02.md` requires:

   > “expand `$goal` in the node's prompt when `--goal <text>` is given”

   while `chapters/04-library/SPRINT-05.md` requires documentation to:

   > “State plainly that `show` never reproduces frames or `$goal`.”

   `planning-workflow.md` repeats the latter rule:

   > “Frames and `$goal` are engine additions `show` never reproduces”

   The detailed `--stage` contract distinguishes raw output from stage comparison; the summaries erase that distinction. Change `SPRINT-05.md` and the summary in `planning-workflow.md` to say raw output does not expand `$goal`, while `--stage --goal` does. Owning node: `decompose`.

3. The validation-loop cardinality drifts. `planning-workflow.md` defines:

   > “validation design loop (one lap per promise”

   and:

   > “loop over `validation/ledger.md`, one item per promise”

   but its SIMPLE rule says:

   > “the validation loop runs one lap”

   No rule limits SIMPLE packages to one promise. Change the SIMPLE rule—also duplicated in decision 51—to preserve one lap per promise, or explicitly add a one-promise restriction. Owning node: `design`.

4. SIMPLE review outcomes contradict the approval contract. `planning-workflow.md` says:

   > “Shows the human the package summary and the seven pass outcomes.”

   but the SIMPLE rule says:

   > “the review loop runs the holistic pass only”

   `declaration.md` P5 is likewise unqualified:

   > “All seven review passes end `done: true` in `plan-review/ledger.md`”

   Change the approval text to show the applicable outcomes and qualify P5 to runs using the full pass set—or remove the SIMPLE exception. Owning node: `brief`.

ROUTE: fail
