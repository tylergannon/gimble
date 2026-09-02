# Bug adjudication: system-instability safety audit follow-up

## Scope and method

This is an independent adjudication pass over the five system-instability
safety-audit reviews (`ephemeral/reviews/202608241000-system-instability-{gpt54,gpt55,luna,sol,terra}.md`)
and the worklog decisions recorded in
`ephemeral/worklog/202608241000-system-instability-audit.md`, against:

- current source at `HEAD` (branch `codex/safety-audit-20260824`, unchanged
  from `6031e9f4b22039cc6a8497542c3399d03e05335a` / `v0.2.1`),
- the existing test suite,
- live read-only host state, re-verified independently below rather than
  taken on the reviewers' word.

No production code was modified. Every accepted bug below is proven by a new,
focused regression test co-located with the package it regresses, plus one
independent, executable proof command under
`ephemeral/proof/established-bugs/` whose exit status is nonzero against the
current implementation and will become zero only when that specific behavior
is corrected. `ephemeral/proof/check-established-bugs.sh` runs all of them in
`baseline` or `fixed` mode; both were executed and their output is captured
below.

Admissibility standard applied (per worklog decisions #1–#2, #9): a finding is
accepted only if it is directly reproducible via code + a deterministic test,
or grounded in observed live evidence I re-verified myself, with a
non-arbitrary definition of "corrected" behavior. Speculative races,
compound/edge-case failure chains not observed live, and claims already
rejected in the worklog are excluded.

## Live host evidence I re-verified independently (2026-08-24, this session)

```text
$ du -sh "/Applications/.Claude.app.codex-autoupdate-1_34493_1.new"
801M    /Applications/.Claude.app.codex-autoupdate-1_34493_1.new
(Info.plist CFBundleVersion 1.34493.1, matching installed /Applications/Claude.app)

$ wc -l ~/Library/Logs/codex-autoupdate/stderr.log
   35472 /Users/tyler/Library/Logs/codex-autoupdate/stderr.log
$ ls -la ~/Library/Logs/codex-autoupdate/stderr.log
-rw-r--r--  1 tyler  staff  5120148 ... stderr.log

$ launchctl print gui/501/com.tylergannon.codex-autoupdate
    state = running   (single instance; matches all five reviewers' observation
                        of exactly one running LaunchAgent PID, no restart loop)

$ codex-autoupdate --version
codex-autoupdate version v0.2.1
```

The log grew from the audit's earlier reading (35,470 lines / 5,119,873 bytes)
to 35,472 lines / 5,120,148 bytes over the course of this session with the
watcher otherwise idle — direct, freshly observed confirmation of ongoing,
unbounded growth (see Bug D). The 801 MiB abandoned bundle is unchanged and
still present (see Bug B).

I agree with all five reviewers and the worklog decisions that there is **no
evidence of updater process/child leakage or a launchd restart loop**: one
running PID, `runs = 1`, no zombie children attributable to the updater. That
class of instability is not present in this proof set.

---

## Accepted bugs

### Bug A — Live Claude Code remote-runtime processes are misclassified as idle

**Claim:** `internal/activity/claude.go`'s `isClaudeCodeProcess` and the
stdio-based fallback both fail to recognize the Claude Code remote CLI runtime
at `~/.claude/remote/ccd-cli/<version>`, so a currently executing remote task
— including one explicitly granted `mcp__computer-use` — is reported idle by
`codex-autoupdate check --json`, and the watcher (`internal/watch/watch.go`)
has no other gate blocking replacement. This directly contradicts the
documented contract (`README.md:45-50`, `llms.txt:95-100`): "currently
executing Codex, Claude Code, or Claude Cowork work does [block replacement]."

**Evidence:**
- Two reviewers (`gpt54`, `sol`) independently reproduced this live on the
  host: PIDs 28558/80886 running
  `/Users/tyler/.claude/remote/ccd-cli/2.1.237 --resume=... --allowedTools mcp__computer-use,...`,
  confirmed executing (accumulating CPU time across a 2-second observation),
  while `codex-autoupdate check --json` simultaneously reported
  `ActiveThreads: null` and a 3-day-stale `LastLifecycle`. This is recorded as
  established in worklog decision #5 ("Two reviewers independently
  reproduced...").
- Source confirms the mechanism: `isClaudeCodeProcess`
  (`internal/activity/claude.go:131-155`) only recognizes basename
  `claude`/`claude-code`, a path containing `/.local/share/claude/versions/`,
  or a later argument containing `/@anthropic-ai/claude-code/`. The observed
  executable's basename is a bare version string (`2.1.237`) under
  `~/.claude/remote/ccd-cli/`, matching none of those. The stdio fallback
  (`internal/activity/claude.go:69-80`, `internal/macos/system.go:247-274`)
  requires fd 0/1/2 to be files under `/private/tmp/claude-<uid>`; live
  `lsof` showed anonymous pipes for this process, so that route also misses
  it.

**Regression test:** `internal/activity/claude_test.go` —
`TestClaudeDetectorRecognizesRemoteCliRuntime`. Feeds `ClaudeDetector.Detect`
the exact observed command string with no `TaskRoot` open files (matching the
live anonymous-pipe evidence) and asserts `report.Active()`.

**Proof command:** `ephemeral/proof/established-bugs/01-claude-remote-process-misclassified.sh`

**Baseline result (run this session):**
```
=== RUN   TestClaudeDetectorRecognizesRemoteCliRuntime
    claude_test.go:158: live remote Claude Code runtime was not reported active: {... ActiveThreads:[] ...}
--- FAIL: TestClaudeDetectorRecognizesRemoteCliRuntime (0.00s)
FAIL
```
Exit status: **1** (nonzero, as required for baseline).

---

### Bug B — Independently updated applications leave a full staged bundle abandoned forever

**Claim:** When Claude or ChatGPT updates itself to the build `codex-autoupdate`
had already staged, `Watcher.Step`'s "application is current" branch
(`internal/watch/watch.go:97-100`) and its "installed application changed
while staged" branch (`:102-109`) both call only `w.clearPrepared()`, which
nils in-memory bookkeeping. Neither calls, nor has any way to call, cleanup of
the on-disk staged bundle at `w.prepared.StagedPath` — `cleanupResidue` is
reachable only from a later `Installer.Prepare` for a *different, non-current*
candidate (`internal/update/installer.go:101-113`), and the `watch.Installer`
interface (`internal/watch/watch.go:28-31`) exposes only `Prepare`/`Apply`,
with no cleanup/abandon method at all.

**Evidence:**
- Two reviewers (`gpt55`, `terra`) independently traced this mechanism and
  matched it to a real, currently-present artifact on the host:
  `/Applications/.Claude.app.codex-autoupdate-1_34493_1.new`, 801 MiB,
  `CFBundleVersion 1.34493.1` — identical to the installed `Claude.app`
  build. `gpt55` additionally quoted the log sequence proving how it was
  created and then never cleaned up (staged at `21:36:53`, current at
  `22:31:23`, no cleanup log line between).
- I re-verified this artifact still exists on the host in this session (see
  live evidence above) — it has not self-healed since the original audit.
- This is one of the three mechanisms the worklog (`decision #4`) already
  recorded as established from direct source and live evidence ("an
  abandoned 801 MiB staged bundle").

**Regression test:** `internal/watch/watch_test.go` —
`TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently`.
Sets `watcher.prepared`/`preparedInstalledBuild` directly (bypassing
`Prepare`, since the defect is in `Step`'s bookkeeping, not in `Prepare`
itself) pointing at a real on-disk staged directory, drives `Step` through the
"application is current" branch, and asserts the staged directory no longer
exists afterward.

**Proof command:** `ephemeral/proof/established-bugs/02-abandoned-staged-bundle-residue.sh`

**Baseline result (run this session):**
```
=== RUN   TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently
    watch_test.go:171: abandoned staged bundle residue was not removed: .../.Claude.app.codex-autoupdate-2.new (stat err = <nil>)
--- FAIL: TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently (0.00s)
FAIL
```
Exit status: **1** (nonzero, as required for baseline).

---

### Bug C — Codex activity detector permanently retains one cache entry per rollout for the app-server's lifetime

**Claim:** `Detector.cache` (`internal/activity/detector.go:39-53`) is a map
keyed by rollout file path. It is reset only when the tracked app-server PID
changes or disappears (`Detect`, lines 66-74); no code path ever deletes an
individual entry — not on session completion, archival, or file deletion. For
a long-lived Codex Desktop app-server process, every distinct rollout file
ever observed accumulates a permanent map entry, and `Detect` keeps walking
both session trees on every check regardless.

**Evidence:**
- `terra` traced the exact mechanism from code with a step-by-step causal
  argument (allocate → insert → never delete except full PID-change reset).
- This is the third mechanism the worklog (`decision #4`) already recorded as
  established ("monotonic rollout-cache retention"), distinguished there from
  the (unproven) UI symptom.

**Regression test:** `internal/activity/detector_test.go` —
`TestDetectorPrunesCacheEntriesForRolloutsThatNoLongerExist`. Creates one
rollout, lets `Detect` cache it, deletes the file from disk, creates a second
rollout under the same unchanged server PID, calls `Detect` again, and asserts
`len(detector.cache) == 1` (only the currently-existing rollout).

**Proof command:** `ephemeral/proof/established-bugs/03-codex-detector-cache-leak.sh`

**Baseline result (run this session):**
```
=== RUN   TestDetectorPrunesCacheEntriesForRolloutsThatNoLongerExist
    detector_test.go:127: cache retains 2 entries after one rollout was deleted from disk, want 1
--- FAIL: TestDetectorPrunesCacheEntriesForRolloutsThatNoLongerExist (0.00s)
FAIL
```
Exit status: **1** (nonzero, as required for baseline).

---

### Bug D — Identical status is re-logged on every poll tick with no change gate, growing the log without bound

**Claim:** `Watcher.Step` (`internal/watch/watch.go:97-100,129`) logs
`"application is current"` or `"waiting for active work to finish"`
unconditionally on every single call, with no check for whether the reported
state changed since the previous tick. Combined with the coordinator's
5-second `ActivityPollInterval` during any pending period
(`internal/watch/watch.go:271-320`), a long pending or steady-state window
grows the log file linearly and without bound purely from repeating identical
records. The plist wires this stream directly to a file with no rotation
(`internal/launchagent/launchagent.go:341-344`), and uninstall explicitly
retains it (`llms.txt:86-87`, `internal/launchagent/launchagent_test.go`).

**Evidence:**
- `terra` labeled this "Critical" and quantified it: 14,245
  `waiting for active work to finish` records and 19,795 `application is
  current` records in the live log, with consecutive active-work records
  ~5.36 seconds apart.
- `gpt55` independently reproduced the same mechanism from a different angle,
  quoting five consecutive near-identical log lines from the live host at
  5-6 second cadence.
- I independently re-measured the live log this session: it grew from
  35,470/5,119,873 bytes (original audit) to 35,472/5,120,148 bytes over the
  course of this otherwise-idle session — freshly observed, ongoing,
  unbounded growth, exactly matching the claimed mechanism.
- This is the first mechanism the worklog (`decision #4`) already recorded as
  established ("unbounded diagnostic-log growth").

I deliberately proved this as "identical status is unconditionally re-logged
every tick" rather than "the log has no size cap," because the codebase and
docs make no promise about a specific size bound — inventing a byte threshold
would be an arbitrary pass/fail line, not an established defect. Suppressing
duplicate, no-op status logging is the precise, non-arbitrary corrected
behavior that the live evidence (long runs of byte-identical lines) directly
supports.

**Regression test:** `internal/watch/watch_test.go` —
`TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState`. Drives
`Watcher.Step` five times with an `Activity` source that always returns the
same active report, captures the logger's output, and asserts
`"waiting for active work to finish"` appears once rather than once per tick.

**Proof command:** `ephemeral/proof/established-bugs/04-duplicate-status-log-growth.sh`

**Baseline result (run this session):**
```
=== RUN   TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState
    watch_test.go:205: logged 5 identical status records across 5 ticks with unchanged state, want 1
--- FAIL: TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState (0.00s)
FAIL
```
Exit status: **1** (nonzero, as required for baseline).

---

## Rejected candidates

### R1 — ChatGPT bundle quiescence ignores the computer-use companion service (`gpt55`, finding 1)

**Reason for rejection:** this is the same class of claim the worklog already
adjudicated and explicitly rejected in `decision #7`: "although command-prefix
matching omits the helper, the latest real update proves the old helper
exited, replayd stopped and deallocated its session, and a single replacement
helper started; survival across replacement was not reproduced." `gpt55`'s
own evidence (PID 89652 `SkyComputerUseService` holding `ChatGPT.app`
resource files open) is real — `BundleProcesses`
(`internal/macos/system.go:236-245`) is indeed a pure command-path-prefix
match against the bundle path, and a companion process launched from a
different bundle (`Codex Computer Use.app`) will never match it. But two other
independent reviewers (`sol`, `terra`) traced the *same* live replacement
end-to-end via unified logs and showed the old `SkyComputerUseService`
cleanly exited, `replayd` deallocated its session, and the replacement service
was accepted without incident. No reviewer, and no live evidence I could
find, reproduced the service surviving across a bundle replacement. Per
decision #2's admissibility standard, an un-reproduced "if it survives" claim
is excluded as speculative.

### R2 — Pending state drives 5-second feed re-fetch and full-harness re-scan traffic (`gpt55`, finding 3, network/scan half)

**Reason for rejection:** the log-growth half of this finding is the same
mechanism as accepted **Bug D** and is covered there. The remaining claim —
that `Watcher.Step` re-fetches `Feed.Latest` and re-inspects all harnesses on
every activity tick — is accurate as code (`internal/watch/watch.go:80-118`
runs unconditionally on every `Step`), but no README/llms.txt text promises a
bound on feed-check frequency, and "5-second polling while an update is
pending" is stated, working-as-designed behavior
(`--activity-poll-interval`, `llms.txt`). Without a documented contract this
violates, I cannot define a non-arbitrary "corrected" behavior distinct from
Bug D's, so this is folded into Bug D rather than treated as a separate
accepted bug.

### R3 — Failed-release archive ZIPs and post-rollback backup bundles have no eventual garbage collection (`terra`, finding 3)

**Reason for rejection:** `terra`'s code trace is directionally correct —
`download` (`internal/update/installer.go:215-279`) only removes its archive
after a full successful `Prepare` (`:158-160`), and `cleanupResidue`
(`:574-595`) only globs `.new`/`.failed-*` patterns, never cache ZIPs or
`.codex-autoupdate-backup-*` directories left behind when
`TestRollbackDoesNotMoveBundlesWhenFailedReplacementCannotStop`'s scenario
occurs live. But reproducing either half requires a compound failure chain
not observed on this host: a malformed archive that passes the length check
but fails `findExtractedApp`, or a failed replacement whose old bundle
survives a stop attempt. No reviewer supplied live evidence of either
occurring, and no second reviewer independently corroborated this finding.
Per decision #2's exclusion of "speculative races and edge-case guesses," and
given the compound precondition, I am not treating this as established for
this proof round.

### R4 — Root-cause attribution to the reported reboot-only Computer Use outage

**Reason for rejection:** already explicitly rejected in worklog `decision
#8`: "No reviewer proved that this program caused the machine-wide,
reboot-only loss of Computer Use." All five reviewers independently reached
the same conclusion — `sol` and `luna` found no admissible instability
finding at all, and `gpt54`/`gpt55`/`terra`'s accepted findings above are
scoped to their proven mechanisms only, not to system-wide causation. Nothing
in this adjudication pass changes that; Bugs A–D are reported as themselves,
not upgraded into a root-cause claim.

### R5 — Two generations of a resumed Claude remote session under separate long-lived remote servers

**Reason for rejection:** recorded in worklog `decision #6` as observed but
explicitly *not* established as caused by `codex-autoupdate` or as a cause of
the reported outage. No reviewer added evidence closing that gap in this
round, so it remains unestablished and out of scope for a regression test.

---

## Commands executed this session (all four proof commands + the check script)

```text
$ ./ephemeral/proof/established-bugs/01-claude-remote-process-misclassified.sh ; echo $?
FAIL ... exit=1
$ ./ephemeral/proof/established-bugs/02-abandoned-staged-bundle-residue.sh ; echo $?
FAIL ... exit=1
$ ./ephemeral/proof/established-bugs/03-codex-detector-cache-leak.sh ; echo $?
FAIL ... exit=1
$ ./ephemeral/proof/established-bugs/04-duplicate-status-log-growth.sh ; echo $?
FAIL ... exit=1

$ ./ephemeral/proof/check-established-bugs.sh baseline
==> ok: 01-claude-remote-process-misclassified.sh exited 1 as expected in baseline mode
==> ok: 02-abandoned-staged-bundle-residue.sh exited 1 as expected in baseline mode
==> ok: 03-codex-detector-cache-leak.sh exited 1 as expected in baseline mode
==> ok: 04-duplicate-status-log-growth.sh exited 1 as expected in baseline mode
check-established-bugs [baseline]: 4/4 commands as expected, 0 unexpected
$ echo $?
0

$ ./ephemeral/proof/check-established-bugs.sh fixed
==> FAIL: 01-... exited 1 (bug still present) in fixed mode
==> FAIL: 02-... exited 1 (bug still present) in fixed mode
==> FAIL: 03-... exited 1 (bug still present) in fixed mode
==> FAIL: 04-... exited 1 (bug still present) in fixed mode
check-established-bugs [fixed]: 0/4 commands as expected, 4 unexpected
$ echo $?
1
```

`fixed` mode correctly fails today, since none of the four bugs have been
corrected in production code (none was touched in this session). I verified
the empty-directory guard separately by pointing a copy of the script at an
empty `established-bugs/` directory: both `baseline` and `fixed` modes exited
1 with "no bug commands found," so an empty proof set can never be
mistaken for either "all bugs reproduced" or "all bugs fixed."

I did not fabricate a "fixed" run against a patched implementation, because
doing so would require editing production code, which this task explicitly
forbids. `fixed` mode's correctness was instead verified structurally: it is
the same command set and the same exit-code comparison as `baseline` with the
pass/fail condition inverted, exercised above and read from
`ephemeral/proof/check-established-bugs.sh`.

## Full test suite status

`go build ./...` and `go vet ./...` succeed. `go test ./...` fails only on the
four new regression tests above; every pre-existing test in the repository
still passes unchanged:

```text
ok      .../cmd/codex-autoupdate
FAIL    .../internal/activity   (2 new regression tests: Bug A, Bug C)
ok      .../internal/appcast
ok      .../internal/claudefeed
ok      .../internal/cli
ok      .../internal/launchagent
ok      .../internal/macos
ok      .../internal/release
ok      .../internal/runlock
ok      .../internal/update
FAIL    .../internal/watch      (2 new regression tests: Bug B, Bug D)
```
