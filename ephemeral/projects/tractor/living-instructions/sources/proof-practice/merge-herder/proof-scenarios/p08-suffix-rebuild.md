# P8: A Changed Middle Source Rebuilds Only The Suffix

- `HARNESS CONTROL -> environment`: establish paused
  `R1=M-SA-SB-SC-SD-SE-H` with recorded source heads `A0...E0`.
- `human/agent -> GIT`: push source ref `C` from `C0` to `C1`.
- `GH WEBHOOK -> MergeHerder`: signed `push` for source ref `C`,
  `before=C0`, `after=C1`.
- `PRODUCT -> MergeHerder`: invoke Unpause.
- `GitHub -> MergeHerder PR SNAPSHOT`: through the selected snapshot source,
  revalidate all five members and observe only PR `C` has a new head SHA.
- `MergeHerder -> GIT`: fetch current `main`, `C1`, prior release history, and
  any other objects needed for reconstruction.
- `MergeHerder -> WORKER`: retain `SA-SB`, rebuild contributions for `C,D,E`, and
  replay repair range `H` above the new contribution tip.
- `MergeHerder -> GIT`: push `R2` using explicit
  `--force-with-lease=refs/heads/R:R1`; normal receive-pack processing rejects
  the update if the server ref no longer matches `R1`.
- `GH WEBHOOK -> MergeHerder`: signed `push` for `R`, `before=R1`, `after=R2`.
- `WORKFLOW -> environment`: execute the real workflow at exact `R2`.
- `GH WEBHOOK -> MergeHerder`: deliver `workflow_run` actions `requested`
  (status `queued`), `in_progress`, and `completed` for a new run with
  `head_sha=R2`.
- `OBSERVE`: SHA identities for `SA` and `SB` are unchanged; old runs cannot land;
  an unresolved replay conflict stops before CI and emits the still-TBD
  human-attention interaction.
