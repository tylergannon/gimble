# P1: One Clean PR Lands

- `HARNESS CONTROL -> environment`: install the shared fixture with `main=M` and
  open PR `A`.
- `PRODUCT -> MergeHerder`: submit PR `A` through the real, still-TBD front door.
- `GitHub -> MergeHerder PR SNAPSHOT`: through the selected, still-TBD source,
  provide open PR `A`, base `main@M`, and source head `A0`.
- `MergeHerder -> GIT`: fetch `main@M` and `A0` using installation credentials.
- `MergeHerder -> WORKER`: request one squashed contribution based on `M`; receive
  `R0=M-SA`.
- `MergeHerder -> GIT`: using an App installation token, push release ref `R`
  from absent/recorded old value to exact SHA `R0`; the fixture workflow
  explicitly matches this release branch/ref pattern.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `R`, `after=R0`.
- `WORKFLOW -> environment`: execute the real workflow against exact checkout
  `R0`; exit successfully.
- `GH WEBHOOK -> MergeHerder`: signed `workflow_run` action `requested` with
  status `queued`, then action/status `in_progress`, then action `completed`
  with conclusion `success`, for one run ID/attempt with `head_branch=R` and
  `head_sha=R0`.
- `MergeHerder -> GIT`: push `main` from `M` to `R0` without force; receive
  fast-forward success.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `main`, `before=M`, `after=R0`.
- `OBSERVE`: `main==R0`, source ref `A==A0`, `R0` passed CI, and the repository is
  no longer occupied by a current batch.
