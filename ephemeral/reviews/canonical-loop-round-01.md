# Adversarial review: canonical loop build proof, round 01

## Target

Branch `codex/canonical-loop-proof` at d52ea23, reviewed as the diff from
d645991. The requirement, taken from AGENTS.md, the worklog entry
`ephemeral/worklog/202609072000-canonical-loop-proof.md`, and the caller's
restatement: establish an example workflow as the canonical loop case, require
it whenever a new build is proved, and make it work through a real run.

The caller's constraints (read-only apart from this file, artifact path) were
honored. No caller instruction narrowed the subject matter, so nothing was
ignored.

## Evidence inspected

- AGENTS.md, README.md, examples/loops/README.md, the release-tractor skill,
  and the rules file at ephemeral/projects/tractor/workflow-designer/rules.md.
- The full diff d645991..d52ea23, including the reviewer edge rewrites in
  internal/workflows/{sprint-execute,chapter-loop,delivery-loop}.yaml.
- scripts/prove-build.py and every file under examples/loops/canonical.
- engine/loop.go (event shapes, validation set, evaluator gating),
  checklist/checklist.go (done rewrite, YAML dump options, infer.files
  parsing), cmd/tractor/root.go (run flags), harness/codex/adapter.go
  (thread start parameters).
- A probe that ran checklist.MarkDone on the fixture ledger and fed the
  result to the script's ledger_contract comparison: equal, two done lines.
- `go test ./cmd/tractor/... ./internal/... ./lint/... ./checklist/... ./engine/...`
  passed (exit 0).
- The live run at /tmp/tractor-canonical-proof-d52ea23: result.json,
  baseline-cli.json, binary-build.txt, run/timeline.jsonl, the stage prompt
  for the first implement lap, the workspace git log, and the first reviewer
  transcript run/events/000002-review.jsonl.

## State of the live run at review time

The prove script (pid 43304) was still running when this review was
written, so the requirement "make it work through a real run" is not yet
demonstrated by a completed receipt. What the timeline showed so far:

- Baseline CLI captured the expected failures (standard 50.00 wrong,
  expedited unsupported).
- Lap 1: implement 1m01s, one commit touching only quote.py; review 1m21s
  routed to sprints; engine validation of Standard shipping passed with
  exit 0 and judge verdict pass; goal evaluator returned not_done with a
  correct explanation.
- Lap 2: Expedited shipping selected; implement 34s; review in progress.

No finding below depends on how the run ends. If it ends with `passed: true`
the findings still stand; if it fails, finding 3 and the timeout note are the
first places to look.

## Findings

### 1. Issue: the "independent" reviewer is not fresh-context; it read the operator's Codex memories about this very fixture

Evidence: run/events/000002-review.jsonl. The gpt reviewer's first three
tool calls read `~/.codex/memories/MEMORY.md` lines 184-205 and
`~/.codex/memories/rollout_summaries/2026-09-04T22-10-37-l69a-shipping_quote_checklist_verified.md`,
a summary of an earlier session that evaluated a shipping-quote checklist
ledger and recorded it as verified. Only after that did it run quote.py and
check.py.

The fixture README says "Read the reviewer transcripts when assessing the
proof: a passing route alone cannot establish that the review was sound", and
sprint-execute.yaml says the reviewer "runs on another provider with fresh
context". harness/codex/adapter.go:88-92 starts the thread with only
approvalPolicy and sandbox; nothing disables memories, and the proof script
does nothing about it either. On this machine the reviewer therefore arrives
already primed with a prior verdict on the same problem. That is exactly the
contamination the repo's own memory warns about (codex memories derail
reviewers). Impact: the proof's review leg does not establish what the
documentation claims, and the receipt cannot tell a primed pass from a real
one. At minimum the fixture README should say this is a known weakness and
how to inspect for it; the proper fix is to start reviewer threads with
memories disabled or otherwise isolated.

### 2. Issue: with `--binary`, the receipt claims a source binding it never checks

Worklog decision: "bind its receipt to the candidate binary and source".
scripts/prove-build.py:129-137 writes `source_revision` and `source_sha256`
from the checkout, hashes the binary, and saves `go version -m` output, but
the only link it verifies between the two is that `workflows show
sprint-execute` matches the checkout's YAML (line 142). A binary built from
any other commit whose embedded sprint-execute happens to match (the file is
unchanged across most commits) passes, and the receipt then records a source
revision the binary was never built from. The saved binary-build.txt already
contains `vcs.revision=d52ea23…` and `vcs.modified=false`; the script reads
it but does not compare it to `source_revision`, nor fail on
`vcs.modified=true`. Impact: the mandatory gate can attest the wrong source
for a release binary, which is the case the `--binary` flag exists for.

### 3. Issue (verifiable bug): ignored files in the fixture directory make the mandatory proof fail on a good build

examples/loops/canonical/.gitignore ignores `__pycache__/` and
`evidence-*.json`, so both are expected to appear there whenever anyone runs
`python3 check.py standard` inside the fixture, which the README invites by
describing check.py as the public checker. Two paths then break:

- `__pycache__`: copytree (line 147) skips it in the workspace, but
  verify_run (line 89) iterates `FIXTURE.rglob("*")` without skipping it,
  so `sha(workspace / ...)` raises FileNotFoundError and the run is reported
  NOT PROVED after all agent quota was spent.
- `evidence-standard.json`: copied into the seed and committed, then
  overwritten by the engine's item command, so line 91 reports
  "acceptance file changed: evidence-standard.json".

Reproduction (not executed here, read-only):

```
cd examples/loops/canonical && python3 check.py standard; cd ../../..
python3 scripts/prove-build.py
```

Impact: a false negative from the gate AGENTS.md makes mandatory, costing a
full agent run each time, with an error message that points at the agents
rather than at the checkout. The fix is to derive the file set from
`git ls-files` under the fixture (as source_hash already does for the repo)
or to apply the same ignore list in both places.

### 4. Nitpick: `--output` inside the checkout self-invalidates the proof

source_hash (lines 36-44) includes untracked, non-ignored files. An output
directory inside the repo but outside `ephemeral/` (the README only says
"an absolute new path") gets result.json rewritten mid-run, so the closing
check at line 190 fails with "source changed during proof". Document that
the output must be outside the checkout or under ephemeral/, or exclude the
output root from the hash.

### 5. Nitpick: the reviewer edge rewrite in three shipped workflows is undocumented

The diff changes the review node's edge conditions in sprint-execute,
chapter-loop and delivery-loop so the reviewer no longer decides "done" and
only routes back on a named material defect. This is a reasonable
correction, the tests pass, and it serves the live run, but neither the
worklog entry nor any doc records it as a decision or explains why the
chapter and delivery loops changed too, and only sprint-execute is proved by
the fixture. Record it so the next reader knows it was deliberate.

## Things checked that hold up

- Event names and fields used by verify_run match engine/loop.go emitters
  (LoopItemSelected.item, LoopValidated.validations[].item/passed/exit_code/
  infer.verdict, LoopEvaluated.verdict, StageCompleted.next).
- The engine's ledger rewrite preserves the authored contract line for line
  (probe above), so the ledger comparison and the done-count regex are
  sound.
- The evaluator is only consulted after a passing validation or when no item
  is open, so requiring verdicts[0] == not_done is correct for this fixture.
- The seed baseline expectation ([True, False, True] standard, all expedited
  failing) matches quote.py's `> 5000` bug and missing mode.
- `tractor run <name> --workdir --logs` and `tractor workflows show` exist
  with the flags the script uses.
- The instruction surface (AGENTS.md, README, release skill, examples index)
  consistently names the proof as mandatory for every new build.

## Outcome

material findings remain
