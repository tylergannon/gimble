1. No qualifying cheap game is apparent. Hard-coding the two known examples is within P10’s deliberately narrow scope, while forging events is expressly conceded. A show/run graph substitution would require the excluded same-ID decoy-node trick.

2. Yes. The package-identity check relies on a later `workflow show` reconstruction ([design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P10/design.md:75)), but the execution run never captures its materialized graph or checklist paths: the manifest records only identity/name/workdir ([control.go](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/engine/control.go:19)), and loop records contain only item names and commands ([loop.go](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/engine/loop.go:28)). Thus the evidence does not establish that the graph shown afterward is the graph that actually ran, which is central to “runs it.”

3. No concrete false failure is evident. MEDIUM/LARGE is required by P10, either size is accepted, and the ledger provenance restrictions implement the project’s durable-chapter/open-sprint replanning contract rather than adding scope ([decisions.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/decisions.md:200)).

ROUTE: fail
