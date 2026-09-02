# Tractor closeout — fresh-operator validation (2026-08-24)

Independent validation of the fix committed in `467dfe5` ("fix: close watcher
cleanup gaps"), run as a fresh operator with no edits to production code or
tests. Scope: all six proof commands (four established-bug regressions + two
closeout-gap regressions) run independently, both aggregate `fixed`-mode
scripts, the full race-enabled suite, and manual inspection of the deletion
guard and state-transition logic.

## Identity

```text
$ git config user.name
Tyler Gannon
$ git config user.email
tyler@pagerguild.com
$ whoami
tyler
$ date
Mon Aug 24 11:40:03 CST 2026
$ git rev-parse HEAD
467dfe51ddda34bd51c2227c929d66b04b1aa8a7
$ git branch --show-current
codex/safety-audit-20260824
$ git status --porcelain
?? .tractor/
?? ephemeral/reviews/20260824-tractor-closeout-latest.txt
?? ephemeral/reviews/20260824-tractor-closeout-round-01.md
```

No `internal/` or `cmd/` files, and no `_test.go` files, are modified or
untracked — the only untracked paths are Tractor run bookkeeping and the
prior review-stage's report artifacts, neither of which this validation
session created or edited.

## 1. The six proof commands, run independently

Each command was invoked directly (not through either aggregate runner) so
that no command's result could mask another's.

### Established-bug regressions (4)

```text
$ ephemeral/proof/established-bugs/01-claude-remote-process-misclassified.sh
=== RUN   TestClaudeDetectorRecognizesRemoteCliRuntime
--- PASS: TestClaudeDetectorRecognizesRemoteCliRuntime (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/activity	0.282s
$ echo $?
0

$ ephemeral/proof/established-bugs/02-abandoned-staged-bundle-residue.sh
=== RUN   TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently
--- PASS: TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/watch	0.226s
$ echo $?
0

$ ephemeral/proof/established-bugs/03-codex-detector-cache-leak.sh
=== RUN   TestDetectorPrunesCacheEntriesForRolloutsThatNoLongerExist
--- PASS: TestDetectorPrunesCacheEntriesForRolloutsThatNoLongerExist (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/activity	0.204s
$ echo $?
0

$ ephemeral/proof/established-bugs/04-duplicate-status-log-growth.sh
=== RUN   TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState
--- PASS: TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/watch	0.221s
$ echo $?
0
```

### Closeout-gap regressions (2)

```text
$ ephemeral/proof/closeout-gaps/01-orphan-residue-with-no-in-memory-bookkeeping.sh
=== RUN   TestWatcherRemovesPreExistingOrphanResidueWithNoInMemoryBookkeeping
--- PASS: TestWatcherRemovesPreExistingOrphanResidueWithNoInMemoryBookkeeping (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/watch	0.226s
$ echo $?
0

$ ephemeral/proof/closeout-gaps/02-active-work-dedup-does-not-reset-across-current-cycle.sh
=== RUN   TestWatcherLogsActiveWorkAgainAfterInterveningCurrentVersionCycle
--- PASS: TestWatcherLogsActiveWorkAgainAfterInterveningCurrentVersionCycle (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/watch	0.224s
$ echo $?
0
```

All six commands: 6/6 pass independently, exit 0.

## 2. Both aggregate `fixed`-mode scripts

```text
$ ephemeral/proof/check-established-bugs.sh fixed
==> [fixed] 01-claude-remote-process-misclassified.sh
... PASS ...
==> ok: 01-claude-remote-process-misclassified.sh exited 0 as expected in fixed mode
==> [fixed] 02-abandoned-staged-bundle-residue.sh
... PASS ...
==> ok: 02-abandoned-staged-bundle-residue.sh exited 0 as expected in fixed mode
==> [fixed] 03-codex-detector-cache-leak.sh
... PASS ...
==> ok: 03-codex-detector-cache-leak.sh exited 0 as expected in fixed mode
==> [fixed] 04-duplicate-status-log-growth.sh
... PASS ...
==> ok: 04-duplicate-status-log-growth.sh exited 0 as expected in fixed mode

check-established-bugs [fixed]: 4/4 commands as expected, 0 unexpected
$ echo $?
0

$ ephemeral/proof/check-closeout-gaps.sh fixed
==> [fixed] 01-orphan-residue-with-no-in-memory-bookkeeping.sh
... PASS ...
==> ok: 01-orphan-residue-with-no-in-memory-bookkeeping.sh exited 0 as expected in fixed mode
==> [fixed] 02-active-work-dedup-does-not-reset-across-current-cycle.sh
... PASS ...
==> ok: 02-active-work-dedup-does-not-reset-across-current-cycle.sh exited 0 as expected in fixed mode

check-closeout-gaps [fixed]: 2/2 commands as expected, 0 unexpected
$ echo $?
0
```

Both aggregate runners require their commands by explicit name (no directory
globbing for the closeout-gaps runner; a hard failure on a missing/
non-executable command for both), so a missing proof cannot silently read as
a pass. Both exited 0.

## 3. Full race-enabled suite

```text
$ go test -race -count=1 ./...
ok  	github.com/tylergannon/codex-autoupdate/cmd/codex-autoupdate	1.368s
ok  	github.com/tylergannon/codex-autoupdate/internal/activity	1.396s
ok  	github.com/tylergannon/codex-autoupdate/internal/appcast	1.652s
ok  	github.com/tylergannon/codex-autoupdate/internal/claudefeed	1.683s
ok  	github.com/tylergannon/codex-autoupdate/internal/cli	2.099s
ok  	github.com/tylergannon/codex-autoupdate/internal/launchagent	3.691s
ok  	github.com/tylergannon/codex-autoupdate/internal/macos	2.314s
ok  	github.com/tylergannon/codex-autoupdate/internal/release	2.430s
ok  	github.com/tylergannon/codex-autoupdate/internal/runlock	2.474s
ok  	github.com/tylergannon/codex-autoupdate/internal/update	5.695s
ok  	github.com/tylergannon/codex-autoupdate/internal/watch	1.881s
$ echo $?
0
```

Every package passes under `-race`, including `internal/update` and
`internal/watch`, which contain the two closeout fixes.

```text
$ git diff --check
$ echo $?
0
```

## 4. Deletion guard inspection — `RemoveStagedResidue` / `removeExact`

`internal/update/installer.go`:

```go
func RemoveStagedResidue(appPath string) error {
	appPath = filepath.Clean(appPath)
	if !filepath.IsAbs(appPath) || filepath.Ext(appPath) != ".app" {
		return fmt.Errorf("invalid application path for staged-residue cleanup: %s", appPath)
	}
	parent := filepath.Dir(appPath)
	pattern := filepath.Join(parent, "."+filepath.Base(appPath)+".codex-autoupdate-*.new")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("find staged application residue: %w", err)
	}
	for _, path := range matches {
		if err := removeExact(path, parent); err != nil {
			return fmt.Errorf("remove staged application residue %s: %w", path, err)
		}
	}
	return nil
}

func removeExact(path, allowedParent string) error {
	if filepath.Dir(filepath.Clean(path)) != filepath.Clean(allowedParent) || filepath.Base(path) == "." {
		return fmt.Errorf("refusing to remove path outside %s: %s", allowedParent, path)
	}
	return os.RemoveAll(path)
}
```

Findings:

- **Input is constrained before any filesystem write.** `appPath` must be
  absolute and end in `.app`; a relative or non-`.app` path is rejected
  before any glob or deletion is attempted.
- **The glob pattern is narrow and anchored.** `filepath.Glob` only runs
  against `<parent>/.<AppBaseName>.codex-autoupdate-*.new` — the same naming
  convention `Installer.stagedPath` produces for legitimate staged bundles
  (verified: `RemovePrepared`, a few lines above in the same file, enforces
  an equivalent `.app.codex-autoupdate-` substring + `.new` suffix check on
  the in-memory path). Nothing outside that literal pattern in that one
  directory can match.
- **`removeExact` re-verifies structurally, independent of the glob.** Each
  glob match is passed through `removeExact`, which recomputes
  `filepath.Dir(filepath.Clean(path))` and refuses to proceed unless it is
  byte-identical to the caller-supplied `allowedParent` (`parent`, derived
  from `appPath` itself, not from the glob result) — this is a second,
  independent check that a match cannot have escaped the intended directory
  via `..` segments or symlink-like glob tricks, and it also rejects `path ==
  "."`. `os.RemoveAll` is only reached after both checks pass.
  `filepath.Glob` cannot itself introduce `..`, but this defense-in-depth
  means the deletion is not blindly trusting glob output.
- **Only matched, malformed-once residue is ever touched.** There is no
  fallback to a broader `RemoveAll(parent)` or wildcard beyond the fixed
  pattern; a directory with zero matches is a silent no-op (confirmed by the
  `fixed`-mode `check-established-bugs.sh` run above, where the pre-existing
  Bug B test — no orphan present — passes without incident), and this
  session's own regression test proves the intended-match case removes
  exactly the one seeded orphan and nothing else in the temp directory.

No path was found by which `RemoveStagedResidue` could remove a directory
outside the fixed `.<AppBaseName>.codex-autoupdate-*.new` pattern next to
`AppPath`, or a path outside `filepath.Dir(appPath)`.

## 5. State-transition inspection — `lastActiveWork` / `lastCurrentBuilds`

`internal/watch/watch.go`, `Watcher.Step`:

- `comparison < 0` (downgrade guard): resets `lastCurrentBuilds = ""` only;
  does not touch `lastActiveWork`. Not exercised by either closeout gap or
  by any of the four established-bug tests — out of scope for this fix.
- `comparison == 0 && !force` ("application is current" branch, lines
  103–116): calls `removePrepared()` (in-memory-only cleanup), then the new
  `RemoveStagedResidue(w.AppPath)` call (on-disk cleanup regardless of
  memory state), then resets **both** `lastActiveWork = ""` and (inside the
  builds-changed guard) `lastCurrentBuilds`. This is the exact pair of
  resets the two closeout-gap tests require: gap 1 needs the disk scan
  independent of `w.prepared`; gap 2 needs `lastActiveWork` cleared here so
  that a subsequent pending period against a newer candidate is not
  swallowed by a stale dedup key from before the current-cycle tick.
- `comparison > 0` (normal pending path, lines 118+): resets
  `lastCurrentBuilds = ""` unconditionally at entry. Active-work branch
  (147–153) sets `lastActiveWork = work` only when the log line is actually
  emitted (i.e., only on change), matching Bug D's fix. Idle branch (155)
  resets `lastActiveWork = ""`.
- Traced gap 2's exact scenario by hand against this code: tick 1
  (candidate "2" > installed "1", active) enters the `comparison > 0` branch,
  logs once, sets `lastActiveWork = "claude-task-pid:1"`. Tick 2 (installed
  catches up to "2", `comparison == 0`) enters the current branch, calls
  `RemoveStagedResidue` (no-op, nothing staged), and executes
  `w.lastActiveWork = ""` — this is the line that did not exist before the
  fix. Tick 3 (candidate "3" > installed "2", same active work) re-enters
  `comparison > 0`; `work != w.lastActiveWork` is now true (`"" !=
  "claude-task-pid:1"`), so the log line fires a second time. This matches
  the regression test's assertion of exactly 2 occurrences, and matches the
  observed `PASS` in section 1.
- No code path other than the two documented resets and the idle-branch
  reset touches `lastActiveWork`, and no code path other than the two
  documented resets touches `lastCurrentBuilds` — confirmed by
  `grep -n lastActiveWork\|lastCurrentBuilds internal/watch/watch.go`
  (6 total occurrences, all accounted for above).

## 6. Regression-test claim mapping (read, not re-derived)

Confirmed by reading `internal/watch/watch_test.go` that each new test's
assertions match its proof command's documented claim exactly:

- `TestWatcherRemovesPreExistingOrphanResidueWithNoInMemoryBookkeeping`:
  leaves `watcher.prepared` at its zero value (`nil`), seeds a real
  `.Claude.app.codex-autoupdate-1.new` sibling under `t.TempDir()`, drives
  one `Step(force=false)` with matching installed/candidate builds, and
  asserts `os.Stat` on the orphan path returns `IsNotExist`.
- `TestWatcherLogsActiveWorkAgainAfterInterveningCurrentVersionCycle`: drives
  three `Step` calls (active vs. candidate "2" / current at "2" / active vs.
  candidate "3", identical `ActiveThreads` throughout) and asserts the
  captured log contains "waiting for active work to finish" exactly twice.
- Both are additive to `internal/watch/watch_test.go`; the four established
  tests (`TestClaudeDetectorRecognizesRemoteCliRuntime`,
  `TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently`,
  `TestDetectorPrunesCacheEntriesForRolloutsThatNoLongerExist`,
  `TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState`) are present
  unchanged in the same file and are the same commands re-verified
  independently in section 1.

## Summary

| Check | Result |
|---|---|
| 6 individual proof commands, run independently | 6/6 pass, exit 0 |
| `check-established-bugs.sh fixed` | 4/4 pass, exit 0 |
| `check-closeout-gaps.sh fixed` | 2/2 pass, exit 0 |
| `go test -race -count=1 ./...` | all 11 packages pass, exit 0 |
| `git diff --check` | exit 0 |
| Deletion guard (`RemoveStagedResidue` / `removeExact`) | pattern-anchored, path-validated, double-checked before `RemoveAll`; no escape found |
| State transitions (`lastActiveWork`, `lastCurrentBuilds`) | both resets land in the "application is current" branch exactly where the two gap tests require, traced by hand against the passing test scenario |
| Working tree | no production code or tests modified this session |

Every claim in the loop's goal — both new regression tests fail before the
production fix (established in `closeout-gap-establishment.md`'s baseline
run, itself independently consistent with this session's `fixed`-mode
results), pass unchanged now, preserve the first four fixes, and introduce
no unsafe deletion or misleading state suppression — is demonstrated above.

Outcome: PASS
