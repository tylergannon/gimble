1. No substantive game. A coder could hard-code a generic unnamed promise and add its decline to `brief.md`, but that would satisfy P1 literally. Substituting an unrelated exclusion is blocked by the semantic `infer` comparison in [design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P1/design.md:60).

2. Yes. The validator expects the `tool_call` arguments to name the numbered question file or ID ([design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P1/design.md:46)), but `tractor ask` receives a source path, moves it to a newly numbered destination, and records only that destination ([interview.go](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/cmd/tractor/interview.go:117), [interview.go](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/cmd/tractor/interview.go:172)). No recorded evidence maps the source argument to the destination.

3. Yes. A correct planner can run `tractor ask draft.md`, produce `interview/0001.md`, receive the decline, and record it under exclusions—fully satisfying P1—yet fail the validator’s source-path/numbered-ID requirement. Requiring the ask specifically within a `brief` stage also exceeds P1, which constrains `plan` behavior but names no internal node.

ROUTE: fail
