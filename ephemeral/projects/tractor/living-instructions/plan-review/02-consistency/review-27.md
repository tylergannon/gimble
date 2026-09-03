1. Required-check rule drifts for chapters 1–3.

   [declaration.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:50):

   > “Required checks (`go build`, `go vet`, `go test`, `golangci-lint`), with the count guards `chapters.md` carries, apply to every chapter”

   [chapters.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters.md:8) gives chapters 1–3 only `go build`, `go test`, and a `done` count; neither `go vet` nor `golangci-lint` appears:

   > `go build ./... && go test ... && test "$(grep -c 'done: true' ...)" -ge ...`

   Change `chapters.md` items 1–3 to carry the declared checks, or explicitly narrow the declaration. Owning node: `decompose`.

2. Plan-review retry semantics contradict each other.

   [planning-workflow.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:166):

   > “It routes pass to the loop node, which marks the pass done, or fail to the owning node … whose edge returns to `reviewer` for the same pass; only a passing reviewer returns to the loop”

   [chapters/05-planner/sprints.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/05-planner/sprints.md:51):

   > “one pass fails, its owning node runs, and the pass is re-selected and marked”

   Under the stated graph, failure never returns to the loop, so the item is not re-selected; the reviewer reruns inside the same lap and only then returns for marking. Change `chapters/05-planner/sprints.md`. Owning node: `decompose`.

3. P7’s repeated proof contract is weaker in `planning-workflow.md`.

   [declaration.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:62) requires:

   > “an `infer` judge over the steer text and the target's segment split at the steer, deciding that the change is what the steer asked”

   [planning-workflow.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:377) restates the demonstration as only:

   > “`scope_cop` delivers at least one steer during the run, recorded in the timeline, and the steered turn's output changes.”

   Mere change does not establish that the steer caused the requested change. Restore the causal judgment in `planning-workflow.md`. Owning node: `design`.

ROUTE: fail


