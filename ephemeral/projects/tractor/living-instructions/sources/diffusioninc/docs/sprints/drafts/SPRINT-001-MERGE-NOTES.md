# Sprint 001 Merge Notes

## Claude Draft Strengths

- Strongest architecture for durable enforcement: move command-form checks into `scripts/validate-skills.py`, not into ad hoc review steps.
- Best proof shape: refactor content checks into `check_skill_content(...)` and add an inline `--selftest` mode with both failing legacy and passing migrated fixtures.
- Most complete zsh coverage: includes `df-sprint-plan` and the matching `df-chapter-create` sprint-discovery probe.
- Clear risk model around false positives, mirror drift, `find` ordering, and self-test drift.

## Codex Draft Strengths

- Strongest baseline audit: start with targeted `rg` scans to prove the current migration state and identify all remaining `ls docs/sprints/SPRINT-*.md` snippets.
- Good validation scope: reject legacy command forms and zsh-fragile examples without banning valid `Gemini` lane prose or `GEMINI` filenames.
- Strong environment-risk note: `py_compile` can hit `__pycache__` write constraints in some sandboxes, so the sprint should record exact failures rather than weakening the SOP gate.

## Gemini Draft Strengths

- Concise explanation of why zsh `nomatch` happens before `2>/dev/null` can suppress `ls` output.
- Correctly keeps `/df-sprint-execute agy` aliasing out of scope and preserves `gemini` as the user-facing selector.
- Identifies real Codex preflight drift between `df-sprint-plan` and `df-sprint-execute`.

## Consensus Critiques

- The runnable `agy` migration is already largely done in the current working tree; Sprint 0001 should make it durable through validation, shell-portable examples, and proof.
- The validator must reject command-shaped legacy forms, not every occurrence of `gemini`.
- The zsh sprint-discovery bug should be fixed with `find ... | sort | tail -N`.
- The final sprint should include `df-chapter-create` because it has the same sprint glob pattern.
- Full `AGENTS.md` gates and the Claude critique gate remain required.
- The final sprint document must state the compatibility boundary: `Gemini` is the lane name and `agy` is the CLI implementation.

## Valid Critiques Accepted

- Use Claude's inline `--selftest` design instead of a new test tree.
- Keep `df-chapter-create` in scope for the zsh discovery fix.
- Add validator coverage for `ls docs/sprints/SPRINT-*.md`, not only `gemini -p` and `--yolo`.
- Use repo-relative paths in the final sprint doc, not machine-specific `file://` links.
- Keep `/df-sprint-execute agy` as a follow-up question, not a Sprint 0001 implementation item.

## Critiques Rejected

- Do not include Codex auth standardization in this sprint by default. It is real drift, but it touches a different agent lane and can change expectations around `--full-auto` vs `--sandbox workspace-write`. Record it as follow-up work unless the human planner explicitly widens scope.
- Do not add a separate `tests/` tree for validator checks. This repo can get enough proof from a focused helper-function refactor plus inline `--selftest`, keeping the surface small.
- Do not ban all `gemini` binary mentions categorically. The guard should target legacy command forms and stale preflight shapes.

## Interview Refinements Applied

No live human interview was collected during this smoke-test planning run. The merge applies conservative defaults from the critique consensus:

- Keep sprint scope narrow to `agy` durability, zsh probe hardening, and proof.
- Defer Codex auth standardization and `agy` selector aliasing.
- Prefer inline validator self-test over new test files.
- Preserve the current `Gemini` lane semantics and artifact names.

## Final Decisions

- Final sprint document path: `docs/sprints/SPRINT-0001.md`, matching the bundled YAML ledger canonical ID.
- Draft artifacts remain under `docs/sprints/drafts/SPRINT-001-*` from the fan-out workflow.
- No chapter link is selected.
- `docs/SEMANTIC-INDEX.md` remains `Status: Not configured`.
