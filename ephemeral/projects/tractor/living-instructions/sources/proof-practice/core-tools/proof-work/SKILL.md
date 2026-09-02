---
name: proof-work
description: >
  Own implementation, proof, PR closeout, and merge-readiness work from current
  origin/main through verified merge or explicit blocker. Use when a repo asks
  for proof-work, implementation proof, PR closeout, auto-merge verification,
  exact-head CI checks, or when an agent must keep a branch mergeable after
  opening a pull request.
---

# Proof Work

Use this skill when the job is more than choosing proof gates. The agent owns
the whole path: current base, isolated worktree, implementation, evidence,
pull request, exact-head checks, auto-merge, and final merge verification.

## Workflow

1. Start from current `origin/main`. Fetch first; if the branch is stale, rebase
   or restart from current `origin/main` without treating that as user work.
2. Work in a task branch/worktree, using the target repo's convention or
   `.worktrees/<short-task>` with a `codex/` branch.
3. Create or update the session worklog before material edits. Record branch,
   worktree, goal, proof plan, commands, failures, and PR state.
4. Read the target repo's `AGENTS.md`, proof policy, and relevant skills.
5. Implement the smallest coherent change.
6. Use `repo-proof-policy` to choose proof gates. Static checks support proof
   but do not replace behavior evidence when behavior changed.
7. Commit, push, and open or update the PR unless the user explicitly opts out.
8. Put exact proof, important artifacts, and known inherited debt in the PR body
   or PR comment.
9. Verify the PR head SHA before trusting CI, preview, or browser-test results.
   Skipped or stale checks are not evidence.
10. Enable auto-merge with the repo's standard merge method once local proof
   passes and required checks are pending or green.
11. If the PR becomes dirty or behind, rebase/update, rerun impacted proof, and
   keep auto-merge configured.
12. Follow through until the PR merges, or escalate only when the current diff
   exposes a real product bug, incompatible dependency breakage, missing secret,
   or external service outage that cannot be fixed in scope.

## Failure Triage

- Distinguish current-diff failures from inherited repo-wide debt.
- Record inherited debt with commands and logs, but do not let it block a PR
  when targeted proof for the changed surfaces passed and repo policy allows
  merge.
- For flaky or environment-only failures, inspect logs, rerun when justified,
  and keep the branch on the merge path unless the failure is caused by the
  current diff.
- Do not stop at PR creation, first CI failure, dirty branch state, or a
  preview/build readiness signal.

## Closeout

Report:

- branch, worktree, PR, and head SHA,
- proof commands and artifact paths,
- exact-head CI/preview status and skipped-check notes,
- auto-merge state or merge SHA,
- blockers that are genuinely caused by the current diff or external state.

source: pagerguild/core-tools
