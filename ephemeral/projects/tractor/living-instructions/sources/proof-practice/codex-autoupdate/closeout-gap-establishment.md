# Closeout gap establishment (2026-08-24)

Follow-up to `ephemeral/proof/bug-adjudication.md` and
`ephemeral/proof/tractor-final-validation.md`, addressing worklog decision
#10 (`ephemeral/worklog/202608241000-system-instability-audit.md`):

> Primary closeout rejected the first Tractor validation despite four green
> regressions: installing the new binary cannot remove the already-orphaned
> 801 MiB bundle because the persisted path is no longer in watcher memory,
> and active-status deduplication does not reset across an intervening
> current-state branch. Both gaps require new binary red/green tests before
> installation.

This pass adds one focused regression test per gap, co-located in
`internal/watch/watch_test.go`, plus one independent executable proof
command per gap under `ephemeral/proof/closeout-gaps/`, and a
`baseline`/`fixed` aggregate runner, `ephemeral/proof/check-closeout-gaps.sh`,
that requires both commands by name and checks each independently — matching
the established-bugs proof structure, but without directory globbing (both
commands are named explicitly, and a missing/non-executable command is a
hard failure in either mode, never a silent pass).

No production code was modified. The four existing established-bug tests
(`TestClaudeDetectorRecognizesRemoteCliRuntime`,
`TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently`,
`TestDetectorPrunesCacheEntriesForRolloutsThatNoLongerExist`,
`TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState`) were not
changed: `git diff -- internal/watch/watch_test.go` is purely additive (111
insertions, 0 deletions).

## Identity

```text
$ git rev-parse HEAD
0ef88be7d19b0c4da5f9c017ec2c0b45449f57c5
$ git branch --show-current
codex/safety-audit-20260824
$ git status --porcelain
 M internal/watch/watch_test.go
?? .tractor/
?? ephemeral/proof/check-closeout-gaps.sh
?? ephemeral/proof/closeout-gaps/
$ git diff --stat -- internal/ cmd/
 internal/watch/watch_test.go | 111 +++++++++++++++++++++++++++++++++++++++++++
 1 file changed, 111 insertions(+)
```

## Gap 1 — orphaned staged bundle survives with no in-memory bookkeeping

**Claim:** `Watcher.Step`'s "application is current" branch only calls
`removePrepared()`, which is a no-op when `w.prepared` is `nil`. The existing
established-bug regression test for Bug B
(`TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently`)
only proves cleanup when `w.prepared` still references the staged directory
in memory (it sets `watcher.prepared` directly before calling `Step`). A
truly orphaned bundle — left by a process that staged it and then exited or
was restarted — leaves a fresh process's `w.prepared` `nil` on its first
`Step` call, with only the on-disk directory as evidence it ever existed.
Nothing in `Watcher.Step` or `removePrepared()` scans the filesystem next to
`AppPath` for such a sibling.

**Regression test:** `internal/watch/watch_test.go` —
`TestWatcherRemovesPreExistingOrphanResidueWithNoInMemoryBookkeeping`. Starts
with `watcher.prepared` at its zero value (`nil`), creates a realistically
named sibling staged directory (`.Claude.app.codex-autoupdate-1.new`,
matching `Installer.stagedPath`'s naming convention) next to `AppPath` on a
real `t.TempDir()` filesystem, drives the "application is current" `Step`
branch (installed build == candidate build, `force=false`), and asserts the
orphan directory no longer exists afterward.

**Proof command:**
`ephemeral/proof/closeout-gaps/01-orphan-residue-with-no-in-memory-bookkeeping.sh`

**Baseline result (run this session):**
```
$ ./ephemeral/proof/closeout-gaps/01-orphan-residue-with-no-in-memory-bookkeeping.sh
=== RUN   TestWatcherRemovesPreExistingOrphanResidueWithNoInMemoryBookkeeping
=== PAUSE TestWatcherRemovesPreExistingOrphanResidueWithNoInMemoryBookkeeping
=== CONT  TestWatcherRemovesPreExistingOrphanResidueWithNoInMemoryBookkeeping
    watch_test.go:232: pre-existing orphaned staged bundle was not removed: /var/folders/.../.Claude.app.codex-autoupdate-1.new (stat err = <nil>)
--- FAIL: TestWatcherRemovesPreExistingOrphanResidueWithNoInMemoryBookkeeping (0.00s)
FAIL
FAIL	github.com/tylergannon/codex-autoupdate/internal/watch	0.230s
FAIL
$ echo $?
1
```
Exit status: **1** (nonzero, as required for baseline — the gap reproduces).

## Gap 2 — active-status dedup does not reset across an intervening current-version cycle

**Claim:** `Watcher.Step`'s "application is current" branch resets
`w.lastCurrentBuilds` but never touches `w.lastActiveWork`, which is
otherwise cleared only when a `Step` call observes idle (non-active) work.
The existing established-bug regression test for Bug D
(`TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState`) only proves
dedup within a single unbroken pending run against one candidate. If the same
active work is still running when a later, genuinely newer candidate
appears — after an intervening tick where the application was briefly
current against the old candidate — the stale dedup key silently swallows
the log record for the new pending period, even though it represents a
distinct blocking event against a different candidate.

**Regression test:** `internal/watch/watch_test.go` —
`TestWatcherLogsActiveWorkAgainAfterInterveningCurrentVersionCycle`. Drives
three `Step` calls: (1) active work observed against candidate build "2"
(installed "1") — logs once; (2) installed catches up to "2", a
current-version cycle (`Feed` and `Inspector` both report build "2"); (3) a
newer candidate "3" appears and the identical active work (`ActiveThreads`)
is observed again. Asserts `"waiting for active work to finish"` appears
**twice** in the captured log output — once per distinct pending period —
not once.

**Proof command:**
`ephemeral/proof/closeout-gaps/02-active-work-dedup-does-not-reset-across-current-cycle.sh`

**Baseline result (run this session):**
```
$ ./ephemeral/proof/closeout-gaps/02-active-work-dedup-does-not-reset-across-current-cycle.sh
=== RUN   TestWatcherLogsActiveWorkAgainAfterInterveningCurrentVersionCycle
=== PAUSE TestWatcherLogsActiveWorkAgainAfterInterveningCurrentVersionCycle
=== CONT  TestWatcherLogsActiveWorkAgainAfterInterveningCurrentVersionCycle
    watch_test.go:316: logged 1 active-work status records across observe -> current -> observe(newer candidate), want 2 (an intervening current-version cycle must reset active-work dedup)
--- FAIL: TestWatcherLogsActiveWorkAgainAfterInterveningCurrentVersionCycle (0.00s)
FAIL
FAIL	github.com/tylergannon/codex-autoupdate/internal/watch	0.223s
FAIL
$ echo $?
1
```
Exit status: **1** (nonzero, as required for baseline — the gap reproduces).

## Aggregate runner, both modes (run this session)

```text
$ ./ephemeral/proof/check-closeout-gaps.sh baseline
==> [baseline] 01-orphan-residue-with-no-in-memory-bookkeeping.sh
... (test output, FAIL as above) ...
==> ok: 01-orphan-residue-with-no-in-memory-bookkeeping.sh exited 1 as expected in baseline mode
==> [baseline] 02-active-work-dedup-does-not-reset-across-current-cycle.sh
... (test output, FAIL as above) ...
==> ok: 02-active-work-dedup-does-not-reset-across-current-cycle.sh exited 1 as expected in baseline mode

check-closeout-gaps [baseline]: 2/2 commands as expected, 0 unexpected
$ echo $?
0

$ ./ephemeral/proof/check-closeout-gaps.sh fixed
==> [fixed] 01-orphan-residue-with-no-in-memory-bookkeeping.sh
... (test output, FAIL as above) ...
==> FAIL: 01-orphan-residue-with-no-in-memory-bookkeeping.sh exited 1 (gap still present) in fixed mode
==> [fixed] 02-active-work-dedup-does-not-reset-across-current-cycle.sh
... (test output, FAIL as above) ...
==> FAIL: 02-active-work-dedup-does-not-reset-across-current-cycle.sh exited 1 (gap still present) in fixed mode

check-closeout-gaps [fixed]: 0/2 commands as expected, 2 unexpected
$ echo $?
1
```

`baseline` mode correctly passes (both gaps reproduce, 2/2 as expected, exit
0); `fixed` mode correctly fails (0/2 as expected, exit 1), since neither gap
has been corrected in production code — none was touched in this session, as
instructed. This is the same polarity relationship documented for
`check-established-bugs.sh` in `ephemeral/proof/bug-adjudication.md`.

I also verified the "required command missing" guard independently, outside
this repository: pointing a copy of `check-closeout-gaps.sh` at a
`closeout-gaps/` directory missing both scripts, in a scratch directory. Both
`baseline` and `fixed` modes exited 1 with `check-closeout-gaps: required
command missing or not executable: .../01-orphan-residue-with-no-in-memory-bookkeeping.sh`,
so a missing proof command can never be mistaken for either "both gaps
reproduced" or "both gaps fixed."

## Regression scope check — established tests and full suite (run this session)

```text
$ ./ephemeral/proof/check-established-bugs.sh fixed
==> ok: 01-claude-remote-process-misclassified.sh exited 0 as expected in fixed mode
==> ok: 02-abandoned-staged-bundle-residue.sh exited 0 as expected in fixed mode
==> ok: 03-codex-detector-cache-leak.sh exited 0 as expected in fixed mode
==> ok: 04-duplicate-status-log-growth.sh exited 0 as expected in fixed mode
check-established-bugs [fixed]: 4/4 commands as expected, 0 unexpected
$ echo $?
0
```

All four established-bug commands still pass unchanged — the two new
closeout-gap tests were added without touching the four existing tests.

```text
$ go build ./...
exit=0
$ go vet ./...
exit=0
$ go test -count=1 ./...
ok  	.../cmd/codex-autoupdate
ok  	.../internal/activity
ok  	.../internal/appcast
ok  	.../internal/claudefeed
ok  	.../internal/cli
ok  	.../internal/launchagent
ok  	.../internal/macos
ok  	.../internal/release
ok  	.../internal/runlock
ok  	.../internal/update
FAIL	.../internal/watch      (2 new regression tests: closeout gap 1, gap 2)
```

Every pre-existing test in the repository still passes unchanged; only the
two new closeout-gap regression tests fail, exactly as required for a
baseline (red) proof against unmodified production code.

## Scope discipline

No production code (`internal/`, `cmd/`) was modified during this session —
`git diff --stat -- internal/ cmd/` shows only the additive test-file change
above. Only `internal/watch/watch_test.go` (new tests + a `sequenceFeed` test
helper), the two new proof scripts under `ephemeral/proof/closeout-gaps/`,
`ephemeral/proof/check-closeout-gaps.sh`, and this report were written.
