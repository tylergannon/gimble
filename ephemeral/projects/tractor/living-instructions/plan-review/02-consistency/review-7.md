1. Validation-design directories use two names.

   - [declaration.md:45](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:45): “live in its validation design under `validation/<id>/design.md`”
   - [planning-workflow.md:109](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:109): “writes, under `validation/<promise>/`”
   
   The persisted package actually distinguishes a promise from its ID; the existing directories use IDs such as `P1`. Change `planning-workflow.md` to `<id>`. Owning node: `design`.

2. Holdout custody has mutually exclusive descriptions.

   - [planning-workflow.md:114](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:114): “a random-token directory whose path is recorded only in the verifier prompt”
   - [planning-workflow.md:190](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:190): “the holdout path the run’s private workflow state carries (written there by the design lap and rendered only into this prompt)”
   
   “Recorded only in the prompt” contradicts “written into private workflow state.” The declaration also specifies private state. Change the validation-loop description in `planning-workflow.md`. Owning node: `design`.

3. The first research pass has two successors.

   - [planning-workflow.md:29](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:29): “`intake ─▶ research ─▶ brief`”
   - [planning-workflow.md:69](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:69): “A tool node, `index_gate`, after the fan-in checks the index against the quality bar and routes to `halt`.”
   
   The same research node cannot unconditionally route both to `brief` and to `halt`; an empty first research pass could therefore skip the promised interview. Change `planning-workflow.md` to state the first-pass exception or show the actual routing. Owning node: `brief`.

4. The convergence rule has drifted.

   - [decisions.md:163](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/decisions.md:163): “Stop when a full pass changes nothing.”
   - [planning-workflow.md:80](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:80): “Exits 0 when `research/findings.md` is empty and `research/plan.md` has no open entry; routes to `decompose`.”
   
   Empty research files do not establish that the preceding brief pass changed nothing. The workflow can halt immediately after accepting a promise change. Change `planning-workflow.md` so the halt predicate implements decision 38. Owning node: `brief`.

5. The LARGE-plan promise gates have no consistent owning item.

   - [decisions.md:157](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/decisions.md:157): “The checklist item’s `check` is the promise; `command` and `infer` are its gates.”
   - [planning-workflow.md:97](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:97): “LARGE writes `checklist.md` as the chapter ledger”
   - [planning-workflow.md:118](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:118): “fills the sprint item that will demonstrate the promise, in the chapter or sprint ledger, with `command` … and `infer`”
   - [planning-workflow.md:198](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:198): “its [chapter] item carries the required checks as `command` and nothing else.”
   
   In LARGE, the checklist item is a chapter item, but the design rule calls it a sprint item and assigns it `infer`, while the execution rule prohibits that `infer`. Change `planning-workflow.md` to identify one gate-bearing item and one rule for its fields. Owning node: `design`.

6. Chapter 4 gives two definitions of done.

   - [decisions.md:157](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/decisions.md:157): “`command` and `infer` are its gates; `done: true` is the attestation.”
   - [chapters/04-library/sprints.md:52](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/sprints.md:52): “Commands are the definition of done”
   
   Several Chapter 4 items have an `infer`, so their command alone is not their definition of done. Change `chapters/04-library/sprints.md`. Owning node: `decompose`.

7. The package both excludes and uses a holdout in its own proof.

   - [decisions.md:276](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/decisions.md:276): “No holdout in this project’s proof.”
   - [chapters/05-planner/CHAPTER.md:49](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/05-planner/CHAPTER.md:49): “A holdout in this project’s proof” is a non-goal.
   - [chapters/05-planner/sprints.md:38](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/05-planner/sprints.md:38): “Proof: a nested run leaves … for a universal promise, a holdout under the state root.”
   - [chapters/06-execution/sprints.md:14](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/06-execution/sprints.md:14): “Proof: a small scratch package with one universal promise whose holdout sample the coder never sees.”
   
   Narrow `decisions.md` and the Chapter 5 non-goal to the declaration’s precise rule—no declared P1–P10 promise carries a holdout—while retaining fixture-based proof of holdout machinery. Owning nodes: `brief` for the ruling and `decompose` for the chapter wording.

ROUTE: fail
