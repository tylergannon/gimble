# Promise: All skills validate

At the recorded repository subject fingerprint, every skill bundle under .claude/skills and .agents/skills passes scripts/validate-skills.py and every bundled Python helper compiles.

## Scope

Included:

- `.claude/skills/**`
- `.agents/skills/**`
- `scripts/validate-skills.py`

Excluded:

- None beyond Promise state.

## Fulfillment gates

- `inventory` — Both skill mirrors contain the same non-empty set of df-* skill bundles.
- `validator` — scripts/validate-skills.py exits zero for the full mirrored skill inventory.
- `python-compile` — The validator and every bundled Python helper compile without syntax errors.

Human sign-off is not required. A fulfilled run is pinned to its reviewed commit,
subject fingerprint, contract and tool fingerprint, verifier, evidence, and
resource metrics.

## Operating plan

1. Inventory skill bundles in both mirrors and verify the sets match.
2. Run the repository validator and capture its exact result.
3. Compile the validator and every bundled Python helper.
4. Record one test result per skill bundle in each mirror for durable coverage.
5. Record the validation report as passing evidence for every gate and finish every run with its truthful verdict; fulfillment requires every gate to pass.

Cadence: **manual**

Runner: **Human-invoked local Codex task**

Default authority: **Read-only validation and promise-state updates; skill remediation, commits, scheduling, and dirty review require explicit user authority.**

## Economics and evidence

Estimated full-run cost: **lt-10** (USD band)

Estimate basis: Recurring local deterministic validation and compact evidence recording; one-time Promise design is excluded.

Per-run approval: **not required**. Maximum run cost: **$9.99**.

Staging plan: Not required for this cost band.

Evidence mode: **internal**. Every fulfillment gate needs a
passing event linked to durable evidence.

Verifier: **tools/run-validation.py**

## Promise-local tools

- `tools/run-validation.py` — Run the full deterministic skill validation suite and write a compact JSON report.

## State

Current machine-readable state is in `state.json`. Compact run summaries and
append-only events live under `runs/`; latest per-subject coverage is in
`coverage.json`. Large or sensitive raw artifacts belong in ignored `.work/`,
with hashes or durable external references recorded as evidence.
