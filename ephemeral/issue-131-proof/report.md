# Issue #131 runtime proof

## Proven head and execution identity

- Repository head: `607a6e188becb952d9d9d0118404dcea229ee269` (recorded in `head.txt`).
- The proof was run by the requested low-cost Codex model, `gpt-5.6-luna`.
- The test adapter is deliberately local and labeled `fake`; no paid or native harness call was made. It emits `UserMessage` through `NewSession.Generate`, so the normal session transcript persistence path is exercised.
- The characterization tests pass when the issue is observable. A production repair that closes the files or surfaces the errors should make the relevant tests fail until their expected behavior is updated.

## Commands and exits

The reproducible runner is `run.sh`.

| Command | Exit |
| --- | ---: |
| `go test -count=1 -run '^TestIssue131' -v .` | 0 |
| `go test -count=1 -run '^(TestRunLogCanBeRead|TestSuperviseCancelsAndJoinsLook)$' -v .` | 0 |
| `go test -race -count=1 -run '^TestIssue131' .` | 0 |

Raw stdout/stderr is retained in `raw-issue131.txt`, `raw-baseline.txt`, and `raw-race.txt`; `status.txt` records the three exits.

## Observed claims

- `TestIssue131CharacterizationSessionWritersStayOpenAfterSuccessfulRepeatedRuns` ran two runs, each with four sessions. After `Run` returned nil, `Stat()` on every retained `*os.File` succeeded before any GC. The output lists all four session IDs on both repeats.
- `TestIssue131CharacterizationSessionWritersStayOpenAfterWorkerFailure` synchronized three worker turns before allowing one to fail. `Run` retained the worker error while all three session descriptors were directly observed open.
- `TestIssue131CharacterizationSessionWritersStayOpenAfterCancellation` synchronized three active turns, then canceled their parent. `Run` returned `context canceled`; all three session descriptors were directly observed open.
- `TestIssue131CharacterizationIgnoresProjectAndSessionOpenErrors` made `project.jsonl` a directory before `Run`, and made the run's `sessions` path a regular file before the first session event. Both runs returned nil. The session case also left `r.sessions` empty.
- `TestIssue131CharacterizationIgnoresEventWriteErrors` preclosed the run writer and, separately, a session writer after its first persisted event. Both workflows returned nil. The run log had no `complete`; the session transcript retained one line after a second generated event.
- `TestIssue131CharacterizationKeepsWorkflowErrorWithoutRecordingFailure` preclosed the run writer, then returned an original workflow sentinel. `Run` returned the sentinel, and the observed closed-file write error was absent from the returned error.
- `TestIssue131CharacterizationReaderWaitsForMissingComplete` verified the run file lacked `complete`, then bounded `internal/runlog.Read` with an 80 ms context. It returned `context deadline exceeded`.

## Close-error coverage and boundaries

`TestIssue131CharacterizationCloseErrorIsDiscarded` shows an observable second `Close` error after the descriptor was deliberately preclosed, while `Run` still returned nil. This is a combined preclosed-descriptor injection that also causes write errors; it does not isolate a filesystem flush or disk-only close failure. The proof therefore establishes that the current close path can discard an observable close error, with that limitation.

The missing-`complete` result is explicitly about the internal file reader. This proof makes no claim about a live web observer or #130 snapshot/live-event behavior; that surface is outside the current proof.

No production files were changed and no commit was made. Owned files are `issue131_proof_test.go` and this `ephemeral/issue-131-proof/` directory.
