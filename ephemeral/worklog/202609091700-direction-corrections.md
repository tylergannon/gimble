# Direction corrections fold

## Input ledger and disposition

- Tyler's pseudocode correction: folded into `docs/direction.md` as an API-design ideal; exact wording retained in `DIRECTION-CORRECTIONS-VERBATIM.md`.
- Tyler's telemetry correction: folded into `docs/direction.md` with bad runs as the priority; exact wording retained.
- Tyler's recovery correction: folded into `docs/direction.md` as an open capability with supervisor steering as the only concrete current mechanism; exact wording retained.
- Tyler's Go rationale: folded into `docs/workflows-as-programs.md`; exact wording retained.
- Tyler's README boundary: folded into `docs/workflows-as-programs.md`; exact wording retained.
- Rename and repository-history corrections from the same session: already represented by the history reboot and the mandatory repository-size policy in `AGENTS.md`; excluded from the architecture docs.
- Existing programmatic-workflow source statements, workflow-design rules, and rendered diagrams remain active source material and were retained unchanged.

## Checker debt

- The docs-fold source-handle checker interprets the literal runtime filename `steering.jsonl` in `docs/spec.md` as a missing citation. It is not a source handle, so the normative specification was left unchanged.

## Proof

- `check-doc-indexes.py` passes with the two direction documents required from the new docs index.
- `git diff --check` passes, and every newly added Markdown link resolves in the checkout.
