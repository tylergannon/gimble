# P6: Unpause With Unchanged Git Requires Fresh CI

- `HARNESS CONTROL -> environment`: retain paused batch at `R0`; keep `main`, all
  source refs, and `R` unchanged.
- `PRODUCT -> MergeHerder`: invoke Unpause through the still-TBD product
  transport.
- `GitHub -> MergeHerder PR SNAPSHOT`: revalidate members through the selected
  snapshot source if Unpause performs source validation.
- `MergeHerder -> GIT`: fetch/compare `main`, member heads, and `R`; observe no
  change.
- `MergeHerder -> GitHub`: request a distinct authoritative run for exact `R0`
  using a **TBD GitHub-valid rerun or dispatch interaction**. A no-op push is not
  part of this contract.
- `WORKFLOW -> environment`: execute the real workflow again at exact `R0` and
  succeed.
- `GH WEBHOOK -> MergeHerder`: deliver the selected mechanism's real event
  sequence for a fresh execution. A rerun may retain `run-R0` and increment
  `run_attempt` without emitting action `requested`; a dispatch may create a
  new run ID.
- `MergeHerder -> GIT`: fast-forward `main` to `R0` only after the fresh
  `(run_id, run_attempt, head_sha=R0)` execution succeeds.
- `OBSERVE`: the pre-Pause run never becomes eligible again.

Contract hole: choose the exact GitHub REST call and permission required for the
fresh same-SHA execution before implementing this mock.
