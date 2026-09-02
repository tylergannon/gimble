# P9: Cancel Terminates The Batch And Releases The Repository

- `HARNESS CONTROL -> environment`: establish current batch `X` containing
  `A,B,C`, optionally with an active run, while `D,E` wait.
- `PRODUCT -> MergeHerder`: invoke Cancel through the still-TBD product transport.
- `MergeHerder -> GH REST`: if a run is active, call
  `POST /repos/{owner}/{repo}/actions/runs/{run-X}/cancel`.
- `GH REST -> MergeHerder`: return cancellation acceptance or the real terminal
  race result.
- `OBSERVE`: `X` becomes terminal and the repository position becomes available;
  there is no Resume interaction for `X`.
- `OBSERVE`: no GitHub REST call changes PR Draft/open state and no interaction
  silently resubmits `A,B,C`.
- `GitHub -> MergeHerder PR SNAPSHOT`: revalidate waiting `D,E` through the
  selected snapshot source when forming the next batch.
- `MergeHerder -> GIT/WORKER`: fetch and assemble the next batch from `D,E`.
- `HARNESS CONTROL -> GH WEBHOOK`: deliver a late completed/success event for
  `run-X`.
- `OBSERVE`: the late event causes no `main` push and does not restore `X`.

The no-Draft and no-automatic-resubmission observations remain proposed Cancel
semantics pending explicit ratification.
