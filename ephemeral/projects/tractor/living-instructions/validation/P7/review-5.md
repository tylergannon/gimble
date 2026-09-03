1. Yes. The cheapest game exploits the missing at-steer snapshot. One pre-steer helper call can write the bait and delete it within that same call; a post-steer “remove” call can then succeed as a no-op. There is no “later pre-S call” to reject, and the final package lacks the bait, yet the bait was already absent when steering arrived. The design incorrectly says event order alone proves persistence ([design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P7/design.md:37)).

2. Yes. Final absence is trivially true if the bait never persisted, while neither the segment nor observer captures filesystem state at S. “Successful” is also not reliably captured: the neutral event schema provides only `tool_result { call_id, output }` ([spec.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/docs/spec.md:2916)), and Codex command projection discards command status, retaining only aggregated output ([events.go](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/harness/codex/events.go:159)). Thus the claimed successful write and persistence depend on evidence the design does not record.

3. Yes. The promise permits any steered-turn output change, but the validator accepts only specific bait in a ledger `name:`/`check:` line or `promises.md` row ([design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P7/design.md:38)). A correct run could, for example, move dashboard scope from a `requirements.md` paragraph into exclusions after the delivered steer; P7 would be true, but this validator would remain inconclusive.

ROUTE: fail


