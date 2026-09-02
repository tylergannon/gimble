# Fresh-operator validation of the built updater (2026-08-24)

Independent validation pass, run cold with no reliance on prior Tractor
review sessions' self-reported "no findings" verdicts
(`ephemeral/reviews/20260824-tractor-fix-round-01.md`,
`.../round-02.md`). Every command below was executed by me, in this
session, against the working tree as it stood. No production code or test
file was edited to produce this result.

## Identity

```text
$ git rev-parse HEAD
07f2989e2f594fcfef3f7e9854ba81d22be825c5
$ git branch --show-current
codex/safety-audit-20260824
$ git diff HEAD --stat            # tracked-file diff vs HEAD
(empty)
$ git status --porcelain
?? .tractor/
?? ephemeral/reviews/20260824-tractor-fix-latest.txt
?? ephemeral/reviews/20260824-tractor-fix-round-01.md
?? ephemeral/reviews/20260824-tractor-fix-round-02.md
```

Working tree is identical to commit `07f2989` (`fix: resolve established
updater safety bugs`) for every tracked file. The only untracked paths are
Tractor run-state (`.tractor/`) and prior reviewers' writeups
(`ephemeral/reviews/20260824-tractor-fix-round-*`), neither of which affects
build output or test behavior. A binary built from this tree stamps its
version as `v0.2.2-0.20260824171640-07f2989e2f59+dirty`; the `+dirty` suffix
is caused solely by those untracked files (`go`'s VCS stamping treats any
non-clean `git status --porcelain`, including untracked paths, as dirty) —
confirmed above by the empty `git diff HEAD --stat`.

Toolchain: `go version go1.27.0 darwin/arm64` (module declares `go 1.26`; no
build or vet errors under 1.27).

## Sanity: build and vet

```text
$ go build ./...
exit=0
$ go vet ./...
exit=0
```

## Established bug commands — run independently, one at a time

```text
$ ./ephemeral/proof/established-bugs/01-claude-remote-process-misclassified.sh
--- PASS: TestClaudeDetectorRecognizesRemoteCliRuntime (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/activity	0.273s
EXIT=0

$ ./ephemeral/proof/established-bugs/02-abandoned-staged-bundle-residue.sh
--- PASS: TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/watch	0.258s
EXIT=0

$ ./ephemeral/proof/established-bugs/03-codex-detector-cache-leak.sh
--- PASS: TestDetectorPrunesCacheEntriesForRolloutsThatNoLongerExist (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/activity	0.206s
EXIT=0

$ ./ephemeral/proof/established-bugs/04-duplicate-status-log-growth.sh
--- PASS: TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState (0.00s)
PASS
ok  	github.com/tylergannon/codex-autoupdate/internal/watch	0.227s
EXIT=0
```

All four established-bug commands pass, unchanged, exactly as authored in
commit `07f2989`.

## Aggregate runner, both modes

```text
$ ./ephemeral/proof/check-established-bugs.sh fixed
check-established-bugs [fixed]: 4/4 commands as expected, 0 unexpected
EXIT=0

$ ./ephemeral/proof/check-established-bugs.sh baseline
check-established-bugs [baseline]: 0/4 commands as expected, 4 unexpected
EXIT=1
```

`fixed` mode passing 4/4 and `baseline` mode failing 4/4 (each proof command
now exits 0, which the baseline mode correctly flags as "bug not
reproduced") is the expected, consistent pair of outcomes for a
production tree where all four bugs are fixed — matching the tool's own
documented contract in `check-established-bugs.sh`'s header comment.

## Full race-enabled test suite

```text
$ go test -race -count=1 ./...
ok  	github.com/tylergannon/codex-autoupdate/cmd/codex-autoupdate	1.270s
ok  	github.com/tylergannon/codex-autoupdate/internal/activity	1.477s
ok  	github.com/tylergannon/codex-autoupdate/internal/appcast	1.973s
ok  	github.com/tylergannon/codex-autoupdate/internal/claudefeed	1.661s
ok  	github.com/tylergannon/codex-autoupdate/internal/cli	2.263s
ok  	github.com/tylergannon/codex-autoupdate/internal/launchagent	4.118s
ok  	github.com/tylergannon/codex-autoupdate/internal/macos	2.554s
ok  	github.com/tylergannon/codex-autoupdate/internal/release	2.435s
ok  	github.com/tylergannon/codex-autoupdate/internal/runlock	3.013s
ok  	github.com/tylergannon/codex-autoupdate/internal/update	5.950s
ok  	github.com/tylergannon/codex-autoupdate/internal/watch	2.319s
EXIT=0
```

11/11 packages pass under `-race`. No `FAIL`, `RACE`, or `DATA RACE` lines
anywhere in the output (grepped explicitly, zero matches). Every
pre-existing test still passes unchanged, alongside the four new regression
tests.

## Real-process probe: Bug A, live host, A/B against the exact same process state

This is the strongest available proof for Bug A: rather than relying only on
the synthetic regression test, I built the fixed binary and ran its
read-only `check --json` against this host's real, currently-running
processes, then immediately ran the still-installed, unpatched `v0.2.1`
binary against the identical live state for contrast.

```text
$ ps aux | grep ccd-cli
tyler  80886  ... /Users/tyler/.claude/remote/ccd-cli/2.1.237 --resume=... --allowedTools mcp__computer-use,...
tyler  28558  ... /Users/tyler/.claude/remote/ccd-cli/2.1.237 --resume=... --allowedTools mcp__computer-use,...
```

Two live `ccd-cli` remote-runtime processes are running on this host right
now — the same class of process the original audit reproduced.

```text
$ go build -o $SCRATCH/codex-autoupdate ./cmd/codex-autoupdate   # built from HEAD 07f2989
$ $SCRATCH/codex-autoupdate check --json | jq '.harnesses[] | select(.id=="claude") | .activity'
{
  "AppServerPID": 2563,
  "ActiveThreads": [
    "claude-code-pid:28558",
    "claude-code-pid:65517",
    "claude-code-pid:80886",
    "claude-task-pid:68104",
    "claude-task-pid:68107"
  ],
  ...
}

$ "/Users/tyler/Library/Application Support/codex-autoupdate/codex-autoupdate" --version
codex-autoupdate version v0.2.1        # unchanged, still-installed production binary
$ "/Users/tyler/Library/Application Support/codex-autoupdate/codex-autoupdate" check --json | jq '.harnesses[] | select(.id=="claude") | .activity'
{
  "AppServerPID": 2563,
  "ActiveThreads": [
    "claude-code-pid:65517",
    "claude-task-pid:68104",
    "claude-task-pid:68123"
  ],
  ...
}
```

Both commands ran against the identical live `AppServerPID` and the
identical set of running OS processes, moments apart. The fixed binary's
`ActiveThreads` includes `claude-code-pid:28558` and `claude-code-pid:80886`
(the two live `ccd-cli` remote runtimes); the unpatched `v0.2.1` binary's
does not. This directly, live, reproduces the documented mechanism and
its correction — not a mock, not a synthetic string, real PIDs on this
real host.

`check --json` is read-only (`internal/cli/root.go` `newCheckCommand`:
"Report installed, available, and activity state without changing
anything") — this probe touched no application state, staged no bundle,
and installed nothing.

## Real-filesystem probes: Bugs B and D, live host state (read-only)

The live LaunchAgent (`launchctl print
gui/501/com.tylergannon.codex-autoupdate`) is still running the unpatched,
installed `v0.2.1` binary at `/Users/tyler/Library/Application
Support/codex-autoupdate/codex-autoupdate` — this validation pass did not
install or redeploy the fixed build, since doing so would modify a live,
shared LaunchAgent outside this task's scope. Consistent with that, the
pre-existing live artifacts from the original audit are unchanged:

```text
$ du -sh "/Applications/.Claude.app.codex-autoupdate-1_34493_1.new"
801M	/Applications/.Claude.app.codex-autoupdate-1_34493_1.new
$ wc -l ~/Library/Logs/codex-autoupdate/stderr.log
35476 /Users/tyler/Library/Logs/codex-autoupdate/stderr.log   # grew from 35472 during this session
```

This is expected, not a gap in the fix: Bugs B and D are corrected in the
*code path* (`Watcher.Step`'s `removePrepared()` and the
`lastActiveWork`/`lastCurrentBuilds` dedup gates), and take effect only the
next time that code path runs — which requires redeploying the fixed
binary as the running LaunchAgent, an action this validation intentionally
did not take. The regression tests for both bugs
(`TestWatcherRemovesStagedResidueWhenApplicationBecomesCurrentIndependently`,
`TestWatcherDoesNotLogDuplicateStatusOnUnchangedPendingState`) already
exercise the real corrected mechanism against real on-disk directories and a
real logger (not mocks): the residue test creates an actual staged bundle
directory on disk and asserts, via `os.Stat`, that `Watcher.Step` removes it
via the exact `update.RemovePrepared` path now wired into production; the
log-dedup test drives five real `Step` calls and asserts the real logger
output contains the status line once, not five times. Both passed under
`-race` above.

`update.RemovePrepared` (`internal/update/installer.go`) also carries a
path-shape guard (must be absolute, basename must start with `.`, contain
`.app.codex-autoupdate-`, and end in `.new`) before it will `os.RemoveAll`
anything — read the source to confirm it cannot be pointed at an arbitrary
path; this was not re-derived by a new test in this pass, but the guard is
present in the diff and exercised implicitly by the residue regression test
using a realistically-named staged directory.

## Scope discipline

No production code or test file was modified during this validation. Only
this report was written, plus a scratch binary in the session scratchpad
(outside the repository) used for the read-only `check --json` probes above.
`git status --porcelain` for tracked files is unchanged from the start of
this session.

## Summary

| Check | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| Bug A command (independent) | PASS, exit 0 |
| Bug B command (independent) | PASS, exit 0 |
| Bug C command (independent) | PASS, exit 0 |
| Bug D command (independent) | PASS, exit 0 |
| `check-established-bugs.sh fixed` | 4/4, exit 0 |
| `check-established-bugs.sh baseline` | 0/4 as-expected (bugs absent), exit 1 — correct polarity |
| `go test -race -count=1 ./...` | 11/11 packages ok, exit 0, no races |
| Bug A live real-process A/B probe | fixed binary includes live `ccd-cli` PIDs; unpatched binary omits them |
| Bug B/D real-filesystem regression tests | pass under `-race` against real on-disk directories and real logger output |

Every established bug command passed unchanged, on the tree at
`07f2989e2f594fcfef3f7e9854ba81d22be825c5`, and every claimed behavior in
`ephemeral/proof/bug-adjudication.md` was independently reproduced —
including a live, real-process demonstration for Bug A beyond what that
document itself captured.

Outcome: PASS
