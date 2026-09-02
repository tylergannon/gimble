# P11: Duplicate, Late, And Reordered Events Are Harmless

- `HARNESS CONTROL -> environment`: establish old candidate/run `R0/run-0` and
  current candidate/run `R1/run-1` with real Git objects.
- `HARNESS CONTROL -> GH WEBHOOK`: send completed/success for `run-1` before its
  requested and in-progress events.
- `HARNESS CONTROL -> GH WEBHOOK`: redeliver that exact raw body with the same
  `X-GitHub-Delivery` and valid signature.
- `HARNESS CONTROL -> GH WEBHOOK`: send semantically identical completed/success
  content with a new delivery ID and valid signature.
- `HARNESS CONTROL -> GH WEBHOOK`: send late completed/success for
  `run-0`, `head_sha=R0`.
- `HARNESS CONTROL -> GH WEBHOOK`: deliver delayed requested/in-progress events
  for both runs in an adversarial order.
- `MergeHerder -> GIT`: at most one fast-forward attempt may occur, and its
  target may only be exact current SHA `R1`.
- `OBSERVE`: each delivery ID is accepted durably at most once, semantic repeats
  do not repeat transitions, and no event for `R0` moves a ref.

Then exercise workflow identity independently:

- `HARNESS CONTROL -> GH WEBHOOK`: send completed/success for a different,
  non-authoritative workflow with current `head_sha=R1`.
- `OBSERVE`: no `main` push occurs despite the matching SHA.
- `HARNESS CONTROL -> GH WEBHOOK`: send completed/success for the configured
  authoritative workflow and the same `head_sha=R1`.
- `OBSERVE`: only the authoritative workflow may enable the single landing.
