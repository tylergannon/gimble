# P4: A Failed Batch Is Repaired By Appending A Hotfix

- `HARNESS CONTROL -> environment`: establish current release
  `R0=M-SA-SB-SC` whose real integration assertion fails.
- `WORKFLOW -> environment`: execute the real workflow at exact `R0`; capture a
  failing exit and logs.
- `GH WEBHOOK -> MergeHerder`: signed `workflow_run` actions `requested`
  (status `queued`), `in_progress`, and `completed` with conclusion `failure`
  for `run-R0`, attempt 1, `head_sha=R0`.
- `PRODUCT -> MergeHerder`: request fix-forward repair through a still-TBD
  operation, or designate the current release branch for a human repair.
- `MergeHerder/human -> WORKER`: supply `R0` and failure evidence; receive repair
  commit `H` producing `R1=R0-H`.
- `MergeHerder/human -> GIT`: push release ref `R` from `R0` to exact `R1`.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `R`, `before=R0`, `after=R1`.
- `WORKFLOW -> environment`: execute the same real workflow at exact `R1`;
  capture successful exit and logs.
- `GH WEBHOOK -> MergeHerder`: signed `workflow_run` actions `requested`
  (status `queued`), `in_progress`, and `completed` with conclusion `success`
  for distinct `run-R1`, attempt 1, `head_sha=R1`.
- `MergeHerder -> GIT`: fast-forward `main` from `M` to exact `R1`.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `main`, `after=R1`.
- `HARNESS CONTROL -> GH WEBHOOK`: redeliver `run-R0`'s immutable
  completed/failure body with its original delivery ID, then send a semantic
  duplicate of that failure with a new delivery ID.
- `OBSERVE`: `SA-SB-SC` were not rebuilt, `H` is retained, `main==R1`, and the
  late/duplicate `R0` failure causes no interaction with Git. P11 owns the
  separate stale-success proof.
