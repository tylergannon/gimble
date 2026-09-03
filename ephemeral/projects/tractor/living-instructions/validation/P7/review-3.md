1. Yes. The cheapest game is to make the target emit a bait-bearing write call before the genuine steer, but have that call fail or restore the file before the steer; afterward it can claim compliance while leaving the already-clean package unchanged. The validator requires only the call’s arguments plus clean start/end snapshots—not a successful matching `tool_result` or package state at S ([design.md](</Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P7/design.md:35>)). This needs no false engine events.

2. Yes. The assertion that clean `StageStarted` and `StageCompleted` snapshots mean “the removal happened inside the steered turn” is unsupported: both snapshots can be identical throughout. A `tool_call` is distinct from its result ([docs/spec.md](</Users/tyler/src/tractor/.claude/worktrees/goal-gates/docs/spec.md:2916>)), and the design does not use the available `SupervisorVerdict` snapshot to capture state near delivery ([ledger.md](</Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/ledger.md:109>)).

3. Yes. P7 requires only some delivered steer whose turn output changes, while this validator requires `decompose` or `design`, the three specific bait concepts, and complete cessation of what the steer named. Yet `scope_cop` also supervises `brief` ([planning-workflow.md](</Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:51>)). A correct run that changes `brief`, or removes one bait item while retaining another, satisfies P7 but is rejected or left inconclusive.

ROUTE: fail


