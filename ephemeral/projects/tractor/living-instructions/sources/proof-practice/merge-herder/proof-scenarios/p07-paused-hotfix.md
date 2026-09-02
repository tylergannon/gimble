# P7: A Hotfix Added While Paused Is Preserved

- `HARNESS CONTROL -> environment`: establish a paused batch at
  `R0=M-SA-SB-SC`.
- `PRODUCT -> human/agent`: authorize work on the current release branch.
- `human/agent -> GIT`: fetch `R0`, append repair commit `H`, and push `R` from
  `R0` to `R1=R0-H`.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `R`, `before=R0`, `after=R1`.
- `GitHub -> WORKFLOW`: the configured push trigger may immediately create and
  execute a workflow for `R1`, even though the batch is paused.
- `PRODUCT -> MergeHerder`: invoke Unpause.
- `MergeHerder -> GIT`: fetch/compare the contribution-chain tip and current
  `R`; recognize `H` rather than overwriting it.
- `MergeHerder -> GitHub`: ensure one post-Pause authoritative run for exact
  `R1` exists using the **TBD paused-run policy**.
- `GH WEBHOOK -> MergeHerder`: deliver the eligible run's real `workflow_run`
  action sequence, ending in `completed`/`success` with `head_sha=R1`.
- `MergeHerder -> GIT`: fast-forward `main` to exact `R1`.
- `OBSERVE`: `SA-SB-SC` remain unchanged and `H` survives.

Contract hole: decide whether an automatically triggered `R1` run is cancelled
while paused, allowed to finish but invalidated, or retained for eligibility
after Unpause. The proof harness must not choose this policy implicitly.
