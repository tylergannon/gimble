1. Promise P1–P7, P9, and P10: the shared witness and proof programs authenticate themselves.

   [validation/ledger.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/ledger.md:157):

   > `validation/observer.sh <run-dir> <rules-file>` is the scripted human and the witness.

   [validation/ledger.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/ledger.md:189):

   > Built once, in chapter 5 sprint 3, and reused by every later scenario.

   Yet, for example, [P1/design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P1/design.md:69) gives `infer` only the seed, question, answer, `answers.log`, and brief—not `observer.sh`, `run-plan.sh`, `timeline.sh`, or `p1-elicitation.sh`. The other scenario designs likewise omit their proof machinery. Those coder-authored programs can synthesize mutually consistent snapshots, HMAC chains, transcripts, and selected run artifacts; HMAC detects later alteration, not a dishonest producer. The semantic judge would then receive a complete but constructed story. Owning node: design.

2. Promise P8: the supposedly independent `Build` oracle is mutable shared proof code omitted from the ledger’s evidence.

   [prove/lib.sh](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/lib.sh:57):

   > `# build_copy: copy the tracked tree, add the Build dumper, build both.`

   [prove/lib.sh](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/lib.sh:61):

   > `cat > "$tmp/src/cmd/builddump/main.go" <<'EOF'`

   Both principal scripts source that mutable helper, but [sprints.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/chapters/04-library/sprints.md:18) supplies `infer` only `show-equals-build.sh` and its log, while the orphan check similarly omits `lib.sh`. A coder can change the generated “dumper” to obtain its answers from `workflow show`, making `show` compare with itself, and change the shared closure helper to bless whatever files were rendered. All named probes and logs can then pass while P8’s promised equality with the real `Build` is false. Owning node: design.

3. Promise P8: its design is stricter than the stated promise.

   [declaration.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:61) expressly says P8 must not imply:

   > That the content is well written

   But [P8/design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P8/design.md:79) requires every post-migration difference to be an include’s rendered text or an authorized wording change and says:

   > nothing an agent is told may be lost, only moved

   P8 promises file ownership, allowed composition machinery, orphan coverage, and `show` fidelity—not preservation of every previous instruction. A fully P8-compliant library containing an otherwise valid prompt rewrite would fail this validator. Owning node: design.

4. Promise P7: the validator adds a specific pre-steer failure mode not present in the promise.

   [declaration.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/declaration.md:60) promises only:

   > `scope_cop` delivers at least one steer during a planning run, and the steered turn's output differs from what it was writing before the steer.

   [P7/design.md](/Users/tyler/src/tractor/.claude/worktrees/goal-gates/ephemeral/projects/tractor/living-instructions/validation/P7/design.md:36) additionally requires that, before the steer:

   > a `tool_call` wrote into a package file text that plans a dashboard, an audit log, or roles as work

   and otherwise exits inconclusive after three attempts. A run can satisfy P7 by changing some other pre-steer output, or by correcting the scoped bait before writing it as package work, yet can never pass this validator. Owning node: design.

ROUTE: fail
