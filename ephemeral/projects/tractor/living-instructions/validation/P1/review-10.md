1. No qualifying cheap game. An unrelated or pre-existing exclusion can satisfy the structural heading check, but the infer judge compares the seed, question, pre-answer brief, and final brief and requires the same previously unmentioned capability to be asked, declined, and newly excluded ([design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P1/design.md:62)). Special-casing the committed feature pool is explicitly conceded under “Not proven.”

2. No substantive check is trivially true or dependent on uncaptured evidence. The observer-generated answer-file check and nonempty-heading check are weak alone, but redundant: the nonce binds the blocking call to its answer, while infer performs the semantic comparison ([design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P1/design.md:44)). Run-log events are timestamped and paired by `call_id`, as the shared rules require.

3. Yes. Requiring the paired `tool_result` to contain the nonce is stricter than P1 ([design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P1/design.md:48)). A correct planner may run `answer="$(tractor ask question.md)"`, save or subsequently read the answer, and update `brief.md`; the real call blocks and the decline is honored, but command substitution prevents the nonce from appearing in that tool result. The seam requires `tractor ask`/`answer`, and the spec only requires `ask` to print the answer to stdout—not that the enclosing harness result echo it ([docs/spec.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/docs/spec.md:514)).

ROUTE: fail


