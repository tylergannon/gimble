# P3: Waiting Work Does Not Overlap The Current Batch

- `HARNESS CONTROL -> environment`: establish current batch `X` with workflow run
  `run-X` in progress.
- `PRODUCT -> MergeHerder`: submit PRs `A`, `B`, and `C` while `run-X` is active.
- `GitHub -> MergeHerder PR SNAPSHOT`: provide snapshots for the three submitted
  PRs if the selected submission design requires them at this point.
- `OBSERVE`: `A`, `B`, and `C` are waiting; MergeHerder makes no release-ref push
  and starts no workflow for them while `X` is current.
- `GH WEBHOOK -> MergeHerder`: signed completed/success `workflow_run` for
  `run-X` and its exact current SHA.
- `MergeHerder -> GIT`: fast-forward `main` to the tested head of `X`.
- `GH WEBHOOK -> MergeHerder`: signed `push` recording that landing.
- `GitHub -> MergeHerder PR SNAPSHOT`: revalidate `A`, `B`, and `C` through the
  selected snapshot source when the next batch begins.
- `MergeHerder -> GIT`: fetch the new `main` base and recorded member heads.
- `MergeHerder -> WORKER`: assemble the next candidate from `A,B,C`.
- `OBSERVE`: no time interval contains two current batches, and no member is both
  waiting and current.
