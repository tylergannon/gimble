# Issue #131 fixed runtime proof

## Identity

- Model: `gpt-5.6-luna` (Codex), as requested for attestation.
- The tested repository head is recorded in `head.sha`.
- The existing proof test's blob identity is recorded in `issue131_proof_test.sha`; `worktree.status` records the complete worktree state at capture time.
- No production code or tests were changed by this proof agent. New evidence files are confined to this `fixed/` directory.

## Commands

`status.txt` records the exit status for each command. Raw combined stdout and stderr are retained per command:

- `raw-issue131.txt`: focused Issue #131 tests.
- `raw-baseline.txt`: run-log reader and #127 supervisor cancellation/join tests.
- `raw-race.txt`: focused Issue #131 tests under the race detector.
- `raw-vet.txt`: `go vet ./...`.
- `raw-build.txt`: `just build`, including generated/web prerequisites.
- `raw-all.txt`: `go test ./...`.

All six commands exited 0.

## Claims demonstrated

- `TestIssue131RunClosesOwnedSessionLogs` creates four sessions through the fake adapter, runs turns through the normal `NewSession.Generate` path, and after `Run` returns checks every retained writer's `*os.File` with `Stat`; each descriptor is closed. This directly exercises closing every retained session writer before return.
- `TestIssue131RunSurfacesRecordingFailures` independently injects project-log open failure, session-log open failure, session-log write failure, and session-log close failure. Each error reaches the returned `Run` error. The combined case returns a workflow sentinel and a run-log persistence failure; `errors.Is` preserves the workflow sentinel while the returned error also contains the recording failure.
- The project-open case reads the completed run log with `internal/runlog.Read` and checks `Complete.RecordingError`, demonstrating that a non-run-log persistence failure is encoded in a readable run record as an explicit incomplete marker.
- `TestRunLogCanBeRead` remains green, preserving normal complete-log readability. `TestSuperviseCancelsAndJoinsLook` remains green for both worker success and worker failure, including the supervisor cancellation and join boundary from #127.
- The focused Issue #131 tests also pass under `go test -race`; repository-wide vet, web prerequisite/build, and all-package tests pass.

## Material findings

No material defect or unproven requested claim was found in the tested change. The proof uses the repository's local fake adapter and filesystem failure injection, so no native or paid agent harness was invoked.
