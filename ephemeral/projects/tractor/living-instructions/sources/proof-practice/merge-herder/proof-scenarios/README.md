# Proof Scenario Interaction Contracts

These files refine the product-level scenarios in
[`MERGE_HERDER_PROOF_SCENARIOS.md`](../../MERGE_HERDER_PROOF_SCENARIOS.md) into
ordered boundary interactions. They exist to determine what the independent
proof harness must provide, keep real, or control.

Each bullet is one externally observable interaction. Internal function calls
and database statements are intentionally absent.

A request and its immediate response form one interaction bullet. Unless a
scenario overrides it:

- a valid App-token request returns `201` and an opaque scoped token;
- a valid targeted PR read returns `200` and the named snapshot;
- an accepted signed webhook returns a `2xx` response after durable receipt;
- an accepted cancellation returns `202` without promising terminal state;
- cancellation of an already terminal run returns `409`; and
- Git smart HTTP returns the real protocol result, including ref-update
  rejection rather than a fabricated success.

## Vocabulary

- **`PRODUCT`**: an operator or agent invokes MergeHerder behavior. If its
  transport is unsettled, the interaction is named but no URL is invented.
- **`GH REST`**: a GitHub-compatible REST request. This is a controlled external
  dependency in deterministic proof and a real call in live conformance.
- **`PR SNAPSHOT`**: the semantic PR identity/head/base/open-state observation.
  Its source remains unsettled: targeted REST read, signed `pull_request`
  webhook, or both. A scenario does not select that transport implicitly.
- **`GH WEBHOOK`**: `POST /api/v1/github-webhook` with
  `X-GitHub-Delivery`, `X-GitHub-Event`, and `X-Hub-Signature-256` over the exact
  raw body.
- **`GIT`**: real Git smart-HTTP and real objects. Fetch uses
  `GET .../info/refs?service=git-upload-pack` plus
  `POST .../git-upload-pack`; push uses
  `GET .../info/refs?service=git-receive-pack` plus
  `POST .../git-receive-pack`.
- **`WORKFLOW`**: the exact candidate is checked out and the repository's real
  workflow is executed using `act` or an equivalent Actions-compatible local
  executor.
- **`WORKER`**: the integration/coding-agent boundary. Deterministic scenarios
  use a controlled worker; agent-quality evaluations use the real agent.
- **`OBSERVE`**: a black-box assertion against Git refs/trees, workflow evidence,
  notifications, or product-visible state.
- **`HARNESS CONTROL`**: an operation needed only to arrange or disorder the
  proof world. Its eventual transport belongs to the proof repository and is
  not a GitHub or MergeHerder product API.

`POST /app/installations/{installation_id}/access_tokens` appears before an
authenticated GitHub REST or Git operation when MergeHerder has no usable
installation token. A valid cached token may remove the repeated call without
changing scenario semantics.

## Scenario Index

- [P1: one clean PR](./p01-one-clean-pr.md)
- [P2: three-PR batch](./p02-three-pr-batch.md)
- [P3: one current batch](./p03-single-current-batch.md)
- [P4: append-only hotfix](./p04-hotfix.md)
- [P5: Pause race](./p05-pause-race.md)
- [P6: unchanged Unpause](./p06-unpause-unchanged.md)
- [P7: paused hotfix](./p07-paused-hotfix.md)
- [P8: suffix rebuild](./p08-suffix-rebuild.md)
- [P9: Cancel](./p09-cancel.md)
- [P10: stale base](./p10-stale-base.md)
- [P11: event disorder](./p11-event-disorder.md)
- [P12: restart recovery](./p12-restart-recovery.md)
- [P13: agent conflict](./p13-agent-conflict.md)
- [P14: forged webhook](./p14-forged-webhook.md)

## Shared Fixture Interactions

Scenarios may refer to this fixture rather than repeat its setup transport:

- `HARNESS CONTROL -> GIT`: create real bare repository with `main=M` and real
  source refs `A`, `B`, `C`, `D`, and `E` as required.
- `HARNESS CONTROL -> GH REST model`: create repository, installation, and open
  PR snapshots whose head SHAs resolve to those real refs.
- `GH WEBHOOK -> MergeHerder`: deliver signed `installation` and
  `installation_repositories` events so the repository exists in MergeHerder's
  normal local state.
- `OBSERVE`: record base and source SHAs before MergeHerder acts.

Fixture creation is a harness capability, not an endpoint MergeHerder consumes.
The proof repository must still test the GitHub-shaped setup adapter it chooses.

## Initial Controlled-Surface Inventory

The scenario traces currently require the controlled GitHub boundary to provide:

- `POST /app/installations/{installation_id}/access_tokens`;
- PR snapshots, implemented by the still-unsettled choice of
  `GET /repos/{owner}/{repo}/pulls/{pull_number}`, signed `pull_request`
  webhooks, or both;
- `POST /repos/{owner}/{repo}/actions/runs/{run_id}/cancel`;
- one unresolved same-SHA rerun or dispatch operation for P6;
- possibly one unresolved run-inspection operation for P12;
- Git smart-HTTP upload-pack and receive-pack over real repositories;
- signed `installation`, `installation_repositories`, `push`, and `workflow_run`
  webhook deliveries;
- missing/invalid-signature webhook rejection before durable deduplication; and
- harness-only control over PR snapshots, run timing, delivery IDs, delivery
  order, redelivery, terminal conclusions, and non-authoritative workflow
  identity at an existing SHA.

The proof harness must **not** mock Git object semantics, candidate trees,
workflow commands, process death, or the MergeHerder coordinator itself.

## Known Contract Holes

The interaction traces intentionally expose, rather than fill, these holes:

- transport for Submit, Pause, Unpause, Cancel, and state inspection;
- batch formation trigger and ordering source;
- GitHub-valid mechanism for fresh CI on unchanged `R` after Unpause;
- treatment of CI automatically triggered by a hotfix pushed while paused;
- restart reconciliation when a webhook or post-push database transition is
  missing; and
- notification interaction when automation cannot continue safely.
