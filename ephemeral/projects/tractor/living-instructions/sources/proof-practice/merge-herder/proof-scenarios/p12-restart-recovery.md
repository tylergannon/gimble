# P12: Restart Recovery Is Idempotent

Run the following trace once for each failpoint.

- `HARNESS CONTROL -> MergeHerder`: start a real process with durable Postgres
  and begin a batch through the real product front door.
- `HARNESS CONTROL -> MergeHerder`: terminate the process at one boundary:
  after claim/before assembly; after `R` push/before its transition is recorded;
  after CI success is recorded/before landing; or after `main` moves/before the
  batch becomes terminal.
- `HARNESS CONTROL -> MergeHerder`: restart against the same database and Git
  repository.
- `MergeHerder -> GIT`: fetch/advertise current `main`, `R`, and member refs as
  required to reconcile durable intent with actual refs.
- `MergeHerder -> GH REST`: if recovery cannot be derived from durable webhook
  receipt and Git state, call a **TBD workflow-run inspection endpoint**; do not
  silently assume one is unnecessary.
- `GH WEBHOOK -> MergeHerder`: redeliver any webhook whose processing boundary
  is under test, preserving its delivery ID.
- `MergeHerder -> GIT/WORKER`: resume or repeat only idempotent work; guarded ref
  updates must reject unexpected movement.
- `OBSERVE`: exactly one current batch, no duplicated contribution/hotfix,
  `main` moves at most once, and final product-visible state agrees with Git.

Contract hole: each failpoint needs an accepted recovery algorithm before the
required workflow-run read API, ref inspection, or internal replay interaction
can be frozen.
