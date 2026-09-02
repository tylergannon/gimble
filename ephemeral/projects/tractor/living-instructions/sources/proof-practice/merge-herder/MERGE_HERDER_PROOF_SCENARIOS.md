# MergeHerder Proof Scenarios

## Purpose

This draft defines what must be proven about MergeHerder before choosing the
test world that will prove it. It does not assume a GitHub twin, `act`, a mock
server, or live GitHub as the implementation. Scenario semantics remain
proposals until accepted in the product-design conversation.

The concise ordered boundary interactions for each scenario live in
[`docs/proof-scenarios/`](./docs/proof-scenarios/README.md). Those interaction
contracts determine the precise external surfaces the proof harness must keep
real or control.

The primary proof suite should live outside the MergeHerder repository. It
should exercise MergeHerder as a black-box system using product operations and
externally observable Git and CI effects. This repository may contain smaller
unit and integration tests, but it must not be the sole author of its own
acceptance standard.

## Common Repository Fixture

Unless a scenario says otherwise, begin with a real Git repository whose
default branch is at commit `M` and three open PR source branches:

- `A = M-a1-a2`
- `B = M-b1-b2`
- `C = M-c1-c2`

The changes are chosen so their intended effects can be asserted from the
resulting files and trees, not merely from commit messages. The batch order is
`A`, `B`, `C`.

Successful assembly produces:

```text
M -- SA -- SB -- SC
                  ^
                  R0
```

`SA`, `SB`, and `SC` are newly created single-commit contributions. Each has
the complete intended effect of its source PR after integration with its
predecessors. The source branches themselves remain unchanged.

The proof driver records all original source SHAs, the base SHA, every derived
contribution SHA, every release-head SHA submitted to CI, and the final
default-branch SHA. These identifiers are observations and correlation keys,
not assertions about MergeHerder's database schema.

## MVP Release-Gating Scenarios

### P1. One clean PR lands

Purpose: prove the smallest complete coordinator path before batching adds more
variables.

Setup:

- `main` points to `M`.
- PR branch `A` is submitted through the real MergeHerder front door.
- The repository has no current batch.

Actions:

1. Allow MergeHerder to claim `A` and assemble release head `R0 = M-SA`.
2. Report authoritative CI success for exact SHA `R0`.

Pass conditions:

- `SA` contains the complete effect of `A` as one commit whose parent is `M`.
- The source branch `A` is not rewritten.
- Exactly one candidate is treated as current.
- `main` advances by fast-forward from `M` to exact SHA `R0`.
- The SHA that passed CI and the SHA now at `main` are identical.
- The batch becomes terminal and the repository can accept later work.

### P2. Three clean PRs share one CI run and land in order

Purpose: prove the economic premise of the application.

Setup:

- `A`, `B`, and `C` are waiting in the accepted deterministic order.
- The repository is eligible to form a batch.

Actions:

1. Let MergeHerder claim all three PRs into one current batch.
2. Let it assemble `R0 = M-SA-SB-SC`.
3. Check out exact SHA `R0` and execute the proof repository's real
   `.github/workflows` CI definition with `act` or an equivalent
   Actions-compatible local executor.
4. Feed the resulting run identity, exact head SHA, status, logs, and conclusion
   through the same CI-observation seam MergeHerder uses in production.

Pass conditions:

- No member remains simultaneously in the waiting queue.
- `R0` has exactly three ordered contribution commits above `M`.
- Each commit contains the intended integrated effect of its corresponding PR.
- The three source branches are unchanged.
- Only the whole release head is submitted to authoritative CI; MergeHerder
  does not spend separate authoritative CI runs on `SA` and `SB`.
- The successful conclusion comes from executing the actual workflow commands
  against `R0`, not from a test directly declaring the candidate successful.
- `main` fast-forwards to exact SHA `R0`, preserving the three-commit chain.

The front-door operation, batch-formation trigger, and ordering rule must be
settled before this scenario can be executable. The proof must use the selected
product behavior rather than a private database fixture to bypass it.

### P3. Waiting work cannot overlap the current batch

Purpose: prove the single integration position and the queue/batch distinction.

Setup:

- Batch `X` is current and its CI is running.
- `A`, `B`, and `C` arrive while `X` remains current.

Actions:

1. Observe the repository while `X` is running.
2. Complete `X` successfully.
3. Allow normal processing to continue.

Pass conditions:

- `A`, `B`, and `C` remain waiting while `X` is current.
- No second release candidate enters CI concurrently with `X`.
- After `X` becomes terminal, one subsequent batch claims the waiting work.
- At no instant is a PR both waiting and a member of a current batch.
- At no instant are two batches current for the same repository.

### P4. A failed batch is repaired by appending a hotfix

Purpose: prove the first supported failure-remediation path without unpacking a
batch that has already been assembled.

Setup:

- `R0 = M-SA-SB-SC` is current.
- The proof workflow contains a real integration assertion that fails when
  executed against exact SHA `R0`; each source contribution can still be
  individually well-formed.

Actions:

1. Append repair commit `H` directly to the release branch, producing
   `R1 = R0-H`.
2. Execute the same real workflow against exact SHA `R1` and observe success.
3. After landing, redeliver `run-R0`'s actual completed/failure event once with
   its original delivery ID and once as a semantic duplicate with a new ID.

Pass conditions:

- `SA`, `SB`, and `SC` are not rebuilt or rewritten merely to add `H`.
- `H` remains visible as a batch-level repair commit above the contribution
  chain.
- The failed `R0` is never eligible to land.
- `main` fast-forwards to exact SHA `R1`.
- The late/duplicate `R0` failure changes neither Git nor batch state. P11
  separately proves that a real superseded run's late success cannot land.

### P5. Pause wins a race with successful CI

Purpose: prove that operator intent invalidates an underway run immediately,
even when GitHub cancellation is asynchronous.

Setup:

- `R0` is current and its authoritative CI run is in progress.
- Other PRs are waiting.

Actions:

1. In one run, invoke Pause, accept cancellation with `202`, and then deliver a
   racing successful completion.
2. In a second run, make CI terminal/success while withholding its webhook,
   invoke Pause, return `409` from cancellation, and then deliver the withheld
   success.

Pass conditions:

- Pause immediately makes that run unable to authorize landing.
- `main` does not move when the racing success arrives.
- The batch remains current and paused.
- Waiting PRs remain waiting; no later batch begins.
- The batch record and release Git artifacts remain available without a
  separate preservation operation.

### P6. Unpause with no Git changes requires fresh CI

Purpose: prove that a run invalidated by Pause cannot later be reused.

Status: required by the accepted Pause/Unpause behavior but blocked until the
product chooses GitHub rerun or workflow dispatch as the fresh-execution
mechanism. The proof tool must not manufacture a no-op-push substitute.

Setup:

- The batch from P5 remains paused at `R0`.
- `main`, every source branch, and `R` are unchanged.

Actions:

1. Invoke Unpause.
2. Allow the selected product mechanism to request fresh authoritative CI for
   exact SHA `R0`.
3. Report success for the new run.

Pass conditions:

- The old run remains permanently ineligible even though its SHA is still
  `R0`.
- A fresh authoritative execution is correlated by
  `(run_id, run_attempt, head_sha)`. A GitHub rerun may increment `run_attempt`
  while retaining the original `run_id`; a dispatch may create a new run ID.
- Only the new run may authorize fast-forwarding `main` to `R0`.

This scenario forces a product decision: a no-op push cannot be assumed to
create a new push-triggered workflow run. MergeHerder needs an explicit,
GitHub-valid rerun or dispatch mechanism.

### P7. A hotfix added while paused is preserved

Purpose: prove the fastest intended human repair path.

Setup:

- A failed or running batch is paused at `R0 = M-SA-SB-SC`.

Actions:

1. A human or agent appends repair commit `H` directly to `R`, producing `R1`.
2. Invoke Unpause.
3. Report CI success for exact SHA `R1`.

Pass conditions:

- MergeHerder recognizes `H` as work above the recorded contribution-chain
  tip; it does not discard or overwrite it.
- It does not rebuild `SA`, `SB`, or `SC` when their inputs are unchanged.
- It requires fresh CI for `R1` and lands exact SHA `R1`.

### P8. A changed middle source rebuilds only the invalid suffix

Purpose: prove incremental reconstruction and repair replay.

Setup:

- The batch is paused at `R1 = M-SA-SB-SC-SD-SE-H`.
- Source branches `A`, `B`, `D`, and `E` are unchanged.
- Source branch `C` advances from recorded head `C0` to `C1`.

Actions:

1. Invoke Unpause.
2. Let MergeHerder compare current source and base heads with those recorded by
   the batch.
3. Let the integration worker rebuild from `C` and replay `H`.

Pass conditions:

- Existing commits `SA` and `SB` remain byte-for-byte unchanged.
- New contributions are produced for `C`, `D`, and `E` in the original order.
- Repair `H` is replayed above the new contribution-chain tip, or the batch
  stops for human attention if that replay cannot be resolved safely.
- The old release head and its CI results cannot authorize landing.
- Fresh CI runs only for the newly produced release head.

### P9. Cancel terminates the batch and releases the repository

Purpose: distinguish Cancel from Pause and from resubmission.

The following is the current proposed Cancel behavior and still requires
explicit ratification.

Setup:

- Batch `X` is current, with `A`, `B`, and `C` as members.
- `D` and `E` are waiting.

Actions:

1. Invoke Cancel through the real MergeHerder product operation.
2. Allow processing to continue normally.

Pass conditions:

- `X` becomes terminal and can never land, including after late CI success.
- `A`, `B`, and `C` do not silently re-enter the waiting queue.
- Their GitHub PRs are not converted to Draft merely because the combination
  was cancelled.
- The repository integration position is released.
- `D` and `E` may form the next batch without waiting for work on `X`.
- There is no Resume operation for `X`; repaired members must use the normal
  front door again.

### P10. A moved default branch prevents stale landing

Purpose: prove MergeHerder never overwrites work or lands an obsolete candidate.

Setup:

- `R0`, built from base `M`, has passed authoritative CI.
- Before landing, another actor advances `main` from `M` to `M1`.

Actions:

1. Allow MergeHerder's landing attempt to proceed.

Pass conditions:

- The guarded fast-forward to `R0` fails.
- `main` remains at `M1`; it is never force-updated to `R0`.
- The green result for `R0` is not recorded as a successful landing.
- Subsequent recovery cannot land until a candidate rebuilt from the new base
  passes fresh authoritative CI.

### P11. Duplicate, late, and reordered events cannot cause a second action

Purpose: prove idempotency and exact-run correlation independently of happy
delivery order.

Setup:

- Two candidates, `R0` and newer `R1`, have distinct CI runs.

Actions:

1. Deliver run events out of order.
2. Redeliver one event with the same delivery identifier.
3. Deliver a semantic duplicate with a different delivery identifier.
4. Deliver a late success for `R0` after `R1` is current.
5. Deliver success for a non-authoritative workflow at current SHA `R1` before
   the configured authoritative workflow succeeds.

Pass conditions:

- A delivery is durably accepted at most once by identifier.
- Repeated semantic information does not repeat a state transition or Git
  operation.
- Only the configured authoritative workflow and current
  `(run_id, run_attempt, head_sha)` can authorize landing.
- `main` moves at most once and never to `R0` after `R1` became current.

### P12. Process death does not create a second batch or a second landing

Purpose: prove durable orchestration rather than in-memory happy-path behavior.

Run the same batch repeatedly with the MergeHerder process forcibly terminated
at these boundaries:

1. after members are claimed but before assembly starts;
2. after `R` is pushed but before the push/run observation is recorded;
3. after CI success is recorded but before landing;
4. after `main` moves but before the batch is marked terminal.

After each restart, pass conditions are:

- Processing converges without manual database repair.
- No second current batch is created.
- No PR is claimed by two live batches.
- No duplicate contribution or repair commit is appended.
- `main` moves zero or one times as appropriate, never to an untested SHA.
- The final durable state agrees with the externally observable Git state.

## Assembly And Agent Proof

### P13. A real rebase conflict is resolved in the assembled chain

Setup `A` and `B` so they edit the same lines and cannot be applied in sequence
without a semantic decision. The expected final behavior is stated as fixture
tests, not as an exact textual patch the agent can copy.

The deterministic coordinator regression uses a controlled integration worker
that returns a known resolution. It proves that MergeHerder supplies the right
repository state, accepts exactly one resolved squash contribution, records it,
and continues to CI.

A separate nondeterministic agent evaluation invokes the real configured coding
agent on multiple conflict fixtures. It measures whether the agent produces a
valid resolution whose resulting repository passes the fixture tests. Agent
quality is reported statistically; it is not allowed to make the coordinator's
deterministic regression flaky.

### P14. A forged webhook cannot affect coordinator state

Given a current batch and a syntactically valid workflow-run body, send the body
once with a missing or invalid `X-Hub-Signature-256`. MergeHerder must return an
authentication failure, store no accepted delivery, perform no transition, and
make no Git call. Sending the exact body afterward with a valid signature and
the same delivery ID must be accepted normally; the forged attempt must not
poison delivery deduplication.

## Scenarios That Must Force Decisions Before MVP Is Declared Complete

These are not specified by the current product design. The proof suite should
not invent answers for them:

- a source PR advances while its batch is running but not paused;
- a source PR is closed or its branch is deleted while waiting or current;
- the release ref `R` is moved or force-pushed by an unexpected actor;
- CI never completes;
- the first CI failure may be retried unchanged before requesting a hotfix;
- a hotfix or suffix replay conflicts and the integration agent cannot resolve
  it;
- successful landing cleanup for the originating GitHub PRs;
- the exact batch-formation trigger, ordering rule, and maximum batch size.

Each requires an explicit product decision followed by a scenario in the same
setup/action/pass-condition form.

## External-Service Contract Proof

The core behavioral scenarios above should be capable of running against more
than one environment. Separately, a deliberately small live-GitHub suite proves
only the external facts MergeHerder relies upon:

- App installation authentication can fetch and push the required refs;
- a release-ref update triggers the configured authoritative workflow for the
  pushed exact SHA;
- the workflow emits the run identity, statuses, conclusion, branch, and exact
  head SHA MergeHerder uses for correlation;
- cancellation is asynchronous and may race completion;
- a guarded non-fast-forward update of `main` is rejected; and
- webhook signatures, identifiers, redelivery, and payload shapes match the
  consumer contract.

The live suite does not need to reproduce every MergeHerder scenario. Its job is
to catch drift in the external assumptions used by the reusable behavioral
suite.

## What The Scenarios Require From A Test World

Only after defining the scenarios can we infer these capabilities:

- real Git commits, trees, refs, fetches, pushes, and guarded updates;
- real MergeHerder process restarts and a durable database;
- controlled PR identities and mutable source refs;
- observable CI runs tied to exact SHAs;
- controllable CI success, failure, non-completion, and cancellation races;
- deliberate duplicate, late, reordered, and redelivered events;
- execution of a small real workflow where workflow fidelity is the subject of
  the scenario; and
- product-level submit, Pause, Unpause, Cancel, state-inspection, and
  notification operations as those seams are finalized.

This list is a set of proof capabilities, not a commitment that one monolithic
"DTU GitHub" service must implement all of them.

The local Actions executor is required as soon as P2 is implemented. Its role is
to answer whether the assembled three-PR candidate actually satisfies the real
workflow. It is not a substitute for the controllable run/event mechanism
needed to produce cancellation races, delayed completions, duplicates, and
reordering.
