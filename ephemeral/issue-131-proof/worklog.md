# Issue #131 proof worklog

## 2026-09-11

- Read the repository agent protocol, proof-of-work guidance, API contract, sprint record, definition of done, and `gh issue view 131` at head `607a6e188becb952d9d9d0118404dcea229ee269`.
- Added only `issue131_proof_test.go` plus this evidence directory. No production edits or commit.
- The first simple runtime repro passed twice: four real `NewSession.Generate` sessions per run remained directly `Stat()`-able after `Run` returned nil. The tests close those descriptors only after observation.
- Added controlled barriers so all three worker or cancellation sessions emit a persisted event before failure/cancellation. This prevents a fast error from making the matrix timing-dependent.
- Open-error injections are explicit: `project.jsonl` is a directory; `<run>/sessions` is a regular file. Event-write injections explicitly preclose `run.writer.file` or a session writer.
- The bounded reader result is intentionally scoped to `internal/runlog.Read`: missing `complete` waits until the supplied 80 ms context deadline. No live web observer claim is made.
- Close-error coverage is limited: preclosing a descriptor makes the later write and close fail. It demonstrates discarded observable close errors but does not simulate an isolated flush or disk failure.
- Targeted issue suite, baseline persistence/join checks, and targeted race run all exited 0. Raw outputs and exit status are retained beside this worklog.
- decision: Run retains the first persistence failure, closes every writer it owns, joins that failure with the workflow error, and records earlier non-run-log failures on Complete. This bounds repeated failures while preserving errors.Is for the workflow result.
- The characterization suite was replaced with positive regression checks. A lightweight fake adapter exercises the real Run, NewSession, Generate, and filesystem paths; direct descriptor inspection and independent open, write, and close injections cover the ownership contract without a paid harness.
- Codex gpt-5.6-luna independently validated the fixed tree. Focused tests, race detection, vet, the production build, all-package tests, normal log replay, and the #127 supervisor join boundary all exited 0; evidence is under fixed/.
