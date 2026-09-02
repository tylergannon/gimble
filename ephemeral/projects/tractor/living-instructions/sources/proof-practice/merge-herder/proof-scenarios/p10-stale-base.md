# P10: A Moved Default Branch Prevents Stale Landing

- `HARNESS CONTROL -> environment`: establish green current candidate `R0` built
  from `main=M`; delay the corresponding `main` push webhook.
- `external actor -> GIT`: fast-forward `main` from `M` to unrelated `M1`.
- `GH WEBHOOK -> MergeHerder`: deliver completed/success `workflow_run` for
  current run at exact `head_sha=R0` before delivering the `main` push event.
- `MergeHerder -> GIT`: attempt a non-forced push of `R0` to `main`; receive-pack
  sees current head `M1` and rejects `R0` by the fast-forward ancestry check.
- `GH WEBHOOK -> MergeHerder`: deliver the delayed signed `push` for `main`,
  `before=M`, `after=M1`.
- `OBSERVE`: `main==M1`, `R0` is not recorded as landed, and no forced update is
  attempted.
- `MergeHerder -> GH REST/GIT/WORKER`: if automatic recovery is later accepted,
  revalidate members, fetch `M1`, rebuild a new candidate, and require fresh CI.
