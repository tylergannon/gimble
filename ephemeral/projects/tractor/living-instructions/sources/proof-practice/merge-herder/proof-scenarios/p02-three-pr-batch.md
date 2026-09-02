# P2: Three PRs Share One Workflow Run

- `HARNESS CONTROL -> environment`: install the shared fixture with `main=M` and
  open PRs `A`, `B`, and `C`.
- `PRODUCT -> MergeHerder`: submit `A`, `B`, and `C` through the real front door.
- `PRODUCT/HARNESS -> MergeHerder`: allow the still-TBD batch trigger to select
  the accepted order `A,B,C`.
- `GitHub -> MergeHerder PR SNAPSHOT`: through the selected, still-TBD source,
  provide open state, `base=main@M`, and heads `A0`, `B0`, `C0` for the three
  members.
- `MergeHerder -> GIT`: fetch `main@M` plus source heads `A0`, `B0`, and `C0`.
- `MergeHerder -> WORKER`: provide the repository, base, ordered members, and
  recorded heads; receive `R0=M-SA-SB-SC`.
- `MergeHerder -> GIT`: using an App installation token, push release ref `R` to
  exact SHA `R0`; the fixture workflow explicitly matches this release
  branch/ref pattern.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `R`, `after=R0`.
- `WORKFLOW -> environment`: check out exact `R0` and run the real workflow with
  `act` or equivalent; capture command logs and successful exit.
- `GH WEBHOOK -> MergeHerder`: signed `workflow_run` actions `requested`,
  `in_progress`, and `completed` for one run ID/attempt and exact `head_sha=R0`;
  the requested event has status `queued` and completed has conclusion `success`.
- `MergeHerder -> GIT`: fast-forward `main` from `M` to exact `R0`.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `main`, `before=M`, `after=R0`.
- `OBSERVE`: one batch claimed all members, one authoritative workflow executed
  the complete candidate, source refs are unchanged, and `main==R0`.
