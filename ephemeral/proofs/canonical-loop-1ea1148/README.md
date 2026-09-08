# Canonical loop proved

Proved code commit: `1ea11482c59f7a90cae91870b63b2c2d1a9ceb3d`.

Candidate SHA-256:
`2cc6ad6b25547c6a00545f701e1164094e14d4c094e1540046365c6905cd8d8f`.

The real `tractor run sprint-execute` completed in 271 seconds. The proof
command exited zero and [its receipt](result.json) records `passed: true`.
This directory retains the receipt and selected evidence; complete native
transcripts and run artifacts remain at
`/private/tmp/tractor-canonical-proof-1ea1148`.

## Observations

1. [Baseline CLI invocations](baseline-cli.json) exposed the seeded standard
   shipping boundary defect and the missing expedited mode.
2. A real Claude agent repaired standard shipping and committed only
   `quote.py`. The [Codex reviewer](standard-review.md) invoked the CLI and
   returned to `sprints`. The engine's command and evidence judge passed.
3. The goal evaluator invoked the still-missing expedited mode and returned
   `not_done`. The engine dispatched the second item.
4. A real Claude agent implemented expedited shipping and committed only
   `quote.py`. The [second Codex review](expedited-review.md) exercised both
   modes and returned to `sprints`.
5. The final engine validation ran standard and expedited commands and their
   evidence judges successfully. The evaluator returned `done`; the engine
   closed the [ledger](final-ledger.md) and emitted `PipelineCompleted`.
6. Six fresh [final CLI invocations](final-cli.json), using the original
   checker outside the agent workspace, all passed. The acceptance files,
   candidate hash and checkout-input hash were unchanged.

The Codex launcher passed `--disable memories`; its feature listing reported
false. Both complete reviewer tool-call streams were inspected: each operated
the current fixture and neither consulted the memory paths.

## Checks and review disposition

- Full Go race suite, vet and default golangci-lint passed. No Go source
  changed after those checks; the final workflow scope is sprint-execute.
- Python regression checks passed for missing/mismatched/dirty binary VCS
  metadata and exclusion of generated/untracked fixture files.
- The original checker rejected malformed output, nonzero application exits
  and wrong totals, and accepted all six correct results.
- An actual invocation with the previous binary exited nonzero before agent
  work because its VCS revision did not match the current checkout.
- The updated verifier rejected the earlier run's memory-contaminated review.
- Removing completion, removing review returns, removing revalidation,
  replacing the partial goal verdict with done, or failing the last evidence
  judge in copies of this real timeline each produced the expected rejection.
- Release-skill validation and `git diff --check` passed.

The [independent review](../../reviews/canonical-loop-round-01.md) identified
memory contamination, unchecked binary provenance and generated fixture
contamination. Those were corrected and the live proof rerun. Output paths
inside the checkout are now rejected early. Unproved edits to chapter-loop
and delivery-loop were withdrawn.

## Scope

This demonstrates the flat checklist loop through real native harnesses. The
initial negative evidence is from CLI invocations before the run; neither
engine validation in this successful run failed. Service lifecycle, nested
loops, failure/retry recovery and the other built-ins need their own proof.
This receipt does not make claims about those surfaces.

The subsequent evidence-only commit does not alter build inputs. The binary
and source revision named above identify the build actually exercised.
