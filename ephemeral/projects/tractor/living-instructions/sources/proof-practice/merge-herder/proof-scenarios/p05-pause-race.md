# P5: Pause Wins A CI Completion Race

- `HARNESS CONTROL -> environment`: establish current `R0` with `run-R0` in
  progress and other PRs waiting.
- `PRODUCT -> MergeHerder`: invoke Pause through the still-TBD product transport.
- `MergeHerder -> GH REST`:
  `POST /repos/{owner}/{repo}/actions/runs/{run-R0}/cancel`.
- `GH REST -> MergeHerder`: return `202 Accepted` without making the run terminal.
- `HARNESS CONTROL -> GH WEBHOOK`: race a signed completed/success
  `workflow_run` for `run-R0`, `head_sha=R0`, after cancellation acceptance.
- `OBSERVE`: MergeHerder performs no push to `main`, batch remains current and
  paused, and waiting work does not start.
- `OBSERVE`: cancellation acceptance and the racing success remain recorded as
  distinct facts; success does not undo Pause.

Run a second ordering against a fresh batch:

- `HARNESS CONTROL -> GitHub`: make `run-R0` terminal/success but withhold its
  completed webhook.
- `PRODUCT -> MergeHerder`: invoke Pause and durably make the batch ineligible.
- `MergeHerder -> GH REST`:
  `POST /repos/{owner}/{repo}/actions/runs/{run-R0}/cancel`.
- `GH REST -> MergeHerder`: return `409 Conflict` because the run is terminal.
- `HARNESS CONTROL -> GH WEBHOOK`: deliver the withheld completed/success event.
- `OBSERVE`: the `409` does not undo Pause and no `main` push occurs.
