1. Yes. The cheapest game is to leave the design substantively unchanged, perform a logged read of the notes, and replace `command: true` with another no-op such as `command: ":"`. The validator never compares the two designs or proves the notes affected the redesign ([design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P4/design.md:28)).

2. Yes. “Command other than `true`” is only a string inequality; `:`, `exit 0`, or `true # revised` passes trivially. The engine treats exit zero as success ([loop.go](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/engine/loop.go:186)). Likewise, merely containing a read event does not establish a successful read of the reviewer-produced notes.

3. Yes. The promise does not require a command, and the workflow expressly says validation-ledger items have no command because the reviewer’s pass edge causes the loop to mark them done ([planning-workflow.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/planning-workflow.md:119)). It also permits notes supplied in the frame, while the validator requires a separate logged read. Either correct implementation could therefore fail this design.

ROUTE: fail
