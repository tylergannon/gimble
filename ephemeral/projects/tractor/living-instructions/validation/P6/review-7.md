1. No substantive game. Hand-marking skips validation ([engine/loop.go](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/engine/loop.go:75)), while the validator requires matching `LoopValidated` and `validation.json`; relabelling a node is defeated by inspection of its prompt, response, and segment. A passing validation event directly precedes the engine’s `done` write ([engine/loop.go](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/engine/loop.go:90)).

2. No. Completion and ordering come from recorded timeline events, codergen identity is evidenced by stage artifacts and segments, and the shared observer captures package state at stage boundaries. These checks can distinguish missing verification, hand-marking, stale verification, and incomplete execution.

3. Yes. Decision 44 permits the chapter ledger to be edited with a reason “by a human or the planner” ([decisions.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/decisions.md:197)), but the validator accepts replacement only through a human question and answer. A planner could legitimately replace an obsolete chapter, record its reason, then have every remaining chapter verified and engine-marked before completion; P6 would hold, yet this validator would reject it.

ROUTE: fail
