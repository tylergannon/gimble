Findings:

1. [declaration.md:7](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:7) understates its governing decisions:

   > “the rulings are decisions 37–53.”

   But [chapters/04-library/CHAPTER.md:29](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/CHAPTER.md:29) says:

   > “Decisions 53, 54, 59.”

   Decisions 54–59 materially govern models, chapter division, holdouts, verdicts, review passes, and P8. `declaration.md` should say decisions 37–59. Owner: `brief`.

2. The “Go supplies only data” rule contradicts the required template function map. [declaration.md:58](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:58) says:

   > “Go supplies only data values to templates”

   [chapters/04-library/SPRINT-01.md:39](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-01.md:39) requires:

   > “Two functions in the func map: `quote` … and `shell` …”

   and [SPRINT-01.md:45](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-01.md:45) adds:

   > “`include "name"` and `doctrine "name"` actions exist”

   `declaration.md` and Decision 59 should distinguish content-bearing values from generic rendering functions. Owner: `brief`.

3. Chapter 4 alternates between four and three existing prompts. [chapters/04-library/SPRINT-01.md:3](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-01.md:3) says:

   > “the three graphs and the four prompts”

   [chapters/04-library/SPRINT-03.md:8](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-03.md:8) says:

   > “adds the `doctrine` includes to the three migrated prompts”

   Its own table names planner, large-plan, and both implement prompts—four. The same wrong count appears in `declaration.md` and `chapters/04-library/sprints.md`. Those “three” references should change to four. Owners: `brief` for `declaration.md`; `decompose` for the chapter and sprint documents.

4. The orphan walk has two incompatible scopes. [chapters/04-library/sprints.md:9](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/sprints.md:9) requires:

   > “fails on a doctrine or template file no prompt references”

   [chapters/04-library/SPRINT-02.md:55](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-02.md:55) defines:

   > “every file under `doctrine/` must be named by a `doctrine` action”

   and then explicitly says:

   > “Skeletons under `templates/` are not in the walk”

   `chapters/04-library/sprints.md` should remove templates from the orphan-walk requirement. Owner: `decompose`.

5. The documentation sprint both mandates restatement and forbids it. [chapters/04-library/SPRINT-04.md:6](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-04.md:6) requires the spec to state:

   > “the layout, the template contract (delimiters, `quote` and `shell`, the `include` and `doctrine` actions, data values only from Go), the render and orphan tests”

   [SPRINT-04.md:16](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/SPRINT-04.md:16) says:

   > “`workflow/library/README.md` is the authoritative contract; the docs point at it rather than restating it.”

   `SPRINT-04.md` must choose summary-plus-reference or full duplication. Owner: `decompose`.

6. Universal promises have contradictory holdout rules. [planning-workflow.md:144](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:144) defines proof quality as:

   > “holdouts present for every universal promise”

   [declaration.md:53](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:53) classifies P3 as:

   > “Universal over promises … No holdout: the set is small and checked exhaustively.”

   [declaration.md:62](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:62) further says:

   > “No promise here carries a holdout.”

   Decision 56 likewise says:

   > “No holdout in this project's proof.”

   `planning-workflow.md` and Decision 42 must encode the later explicit exhaustive-set exception, or the declaration’s archetypes must change. Owner: `design`.

7. Reviewer notes are assigned two different transports. [planning-workflow.md:123](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:123) says:

   > “The reviewer's notes are ordinary files under `validation/<promise>/`, read by `design` on the next lap.”

   [planning-workflow.md:306](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:306) instead requires re-entry with:

   > “the reviewer's notes in the frame”

   Engine frames contain the selected item, validation result, and item `doc`; they do not automatically include arbitrary review-note files. Section 8 should say the notes are read from their ordinary files. Owner: `design`.

8. The declaration says verified items have no command, while the execution design requires a chapter command. [declaration.md:155](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:155) says:

   > “Review and verification are codergen nodes with pass and fail edges; routing is the verdict, the item has no command”

   [planning-workflow.md:190](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:190) says:

   > “its item carries the required checks as `command` and nothing else.”

   The no-command rule applies to validation and plan-review ledger items, not verified chapter items. `declaration.md` and Decision 57 should be qualified. Owner: `brief`.

9. P9 has two different proof mechanisms. [declaration.md:59](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:59) specifies:

   > “`infer` judge over the docs”

   [chapters/06-execution/sprints.md:22](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/06-execution/sprints.md:22) specifies:

   > “the docs-only reader runs plan, ask, answer, and show. Proof: P9.”

   Chapter 5 likewise requires the reader’s actual run directory. `declaration.md` should identify the executable reader run as the verifier, or the sprint documents must consistently specify only inference. Owner: `brief`.

ROUTE: fail
