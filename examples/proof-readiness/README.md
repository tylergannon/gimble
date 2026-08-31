# Delivery and discovery modes

[`required-happy-path.yaml`](required-happy-path.yaml) is a generic loop with
one primary-success assertion and one boundary assertion. Both are declared
in `proof_contract`, so the empty-input boundary cannot complete the run by
itself. The contract also labels the intended architecture edges and the
workspace artifacts Tractor snapshots before and after each assertion.

From the repository root:

```sh
tractor validate examples/proof-readiness/required-happy-path.yaml
tractor run examples/proof-readiness/required-happy-path.yaml \
  --workdir /path/to/a/git/repository \
  --logs /path/to/empty/run-directory
```

Observable success means independently executed tool nodes have shown both
that `hello` becomes `HELLO` and that empty input remains empty in one run.
The resulting checkpoint ties each case to the run, successful route,
execution reference, and fingerprinted input/output snapshots.

[`discovery.yaml`](discovery.yaml) is intentionally contract-light. It
requires a non-empty learning artifact, but its `success` route is reported as
`LEARNING_COMPLETED`, never `COMPLETED`, because it declares `mode: discovery`.

```sh
tractor validate examples/proof-readiness/discovery.yaml
tractor run examples/proof-readiness/discovery.yaml \
  --workdir /path/to/a/git/repository \
  --logs /path/to/empty/run-directory
```
