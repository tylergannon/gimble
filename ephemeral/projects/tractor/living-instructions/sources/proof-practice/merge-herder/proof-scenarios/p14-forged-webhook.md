# P14: A Forged Webhook Cannot Affect Coordinator State

- `HARNESS CONTROL -> environment`: establish a current batch and construct a
  syntactically valid authoritative `workflow_run` completed/success raw body.
- `HARNESS CONTROL -> GH WEBHOOK`: `POST /api/v1/github-webhook` with delivery ID
  `G`, the valid raw body, and missing or invalid `X-Hub-Signature-256`.
- `MergeHerder -> HARNESS CONTROL`: return `401`; do not record an accepted
  delivery, change batch state, call a worker, or call Git.
- `HARNESS CONTROL -> GH WEBHOOK`: send the exact same raw body and delivery ID
  `G` with the correct HMAC-SHA256 signature.
- `MergeHerder -> HARNESS CONTROL`: accept the valid delivery normally.
- `OBSERVE`: the forged attempt caused no durable deduplication entry that could
  suppress the later valid delivery, and only the valid delivery may produce
  its ordinary coordinator transition.
