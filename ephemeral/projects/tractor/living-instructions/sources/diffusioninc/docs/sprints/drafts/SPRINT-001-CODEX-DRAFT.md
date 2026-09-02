# Sprint 001: Agy Migration Hardening

## Pyramid Index

- L0: Make the `agy` migration durable by enforcing command-form regressions, hardening first-sprint shell examples, and preserving the existing Gemini lane compatibility contract.
- L1:
  - Treat `Gemini` / `gemini` as the stable planning lane, artifact name, and execution selector while documenting that the executable CLI is now `agy`.
  - Extend deterministic validation so legacy runnable examples such as `gemini -p ... --yolo` cannot return in mirrored skills.
  - Replace zsh-sensitive sprint glob probes with shell-portable `find` examples where first-sprint bootstrap can encounter an empty `docs/sprints/` directory.
  - Prove the actual `agy` command form with the installed CLI and run the repo SOP gates, including mirror validation and Claude critique.
- L2:
  - See Overview and Architecture for the compatibility boundary.
  - See Implementation Plan phases 1-3 for text changes, validator hardening, and regression tests.
  - See Definition of Done for the exact proof commands and acceptance gates.

## Overview

This sprint should finish the migration from legacy Google/Gemini CLI examples to Google's `agy` CLI without breaking the repo's established planning vocabulary. The user-facing lane remains `Gemini`, generated draft and critique files remain `SPRINT-NNN-GEMINI-*`, and `/df-sprint-execute gemini` remains the execution selector. The compatibility boundary is that those names describe the agent lane; runnable shell examples for that lane must use `agy`.

The current repo state already includes mirrored edits in `df-sprint-plan`, `df-sprint-execute`, and `README.md` that swap `gemini -p ... --yolo` examples for `agy -p ... --dangerously-skip-permissions`. The sprint should make that state durable by adding validator coverage, tightening shell-portable bootstrap examples, and recording verification commands that future maintainers can rerun.

There is no chapter link for this sprint. `docs/SEMANTIC-INDEX.md` is explicitly not configured, so this plan is grounded in repo-local files, targeted `rg` searches, the current mirror invariant in `AGENTS.md`, and the first-sprint intent.

## Use Cases

1. A planner runs `/df-sprint-plan` on a fresh repo with no final sprint docs and does not see zsh `no matches found` noise from the documented sprint discovery examples.

2. A planner uses the Gemini lane in the multi-agent planning workflow and every runnable example checks or invokes `agy`, while draft and critique filenames still use `GEMINI` for continuity.

3. An executor runs `/df-sprint-execute gemini 001`; the skill validates `agy`, authenticates with `agy -p "say ok" --dangerously-skip-permissions`, and launches the execution prompt through `agy`.

4. A maintainer accidentally reintroduces `gemini -p`, `gemini ... --yolo`, bare `--yolo`, or the zsh-sensitive `ls docs/sprints/SPRINT-*.md` example in a mirrored skill; `python3 scripts/validate-skills.py` fails with a concrete file-scoped error.

5. A reviewer can distinguish allowed compatibility text such as `Gemini lane`, `GEMINI-DRAFT`, and the user selector `gemini` from disallowed runnable legacy commands.

## Architecture

The hardening should stay inside the repo's existing architecture:

- Mirrored skills remain the source artifacts. Any skill text edits happen in both `.claude/skills/df-*` and `.agents/skills/df-*`, with only the allowed `.claude/skills/...` vs `.agents/skills/...` and `CLAUDE.md` vs `AGENTS.md` substitutions.
- `scripts/validate-skills.py` remains the deterministic local guard. It already validates skill naming, mirror parity, path leaks, missing references, and selected wording regressions. This sprint should add focused text-regression checks there rather than introducing a second required validation tool for mirror correctness.
- `agy` migration validation should be command-form aware. The validator should reject runnable legacy forms, not every occurrence of the word `gemini`, because the lane name, file names, and selector are intentionally stable compatibility surfaces.
- Shell-portability hardening should be documentation-level and narrow. Replace command snippets like `ls docs/sprints/SPRINT-*.md 2>/dev/null | tail -3` with `find docs/sprints -maxdepth 1 -type f -name 'SPRINT-*.md' -print 2>/dev/null | sort | tail -3`. Do not change Python `Path.glob("SPRINT-*.md")` usage in the ledger helper; that is not affected by shell `nomatch`.
- Regression testing can use standard-library Python. If a new `tests/` tree is added, keep it small and focused on validator pattern behavior so the repo does not acquire a framework dependency.

The recommended scope is to accept the zsh probe hardening and defer `/df-sprint-execute agy` as an alias. Adding the alias would create a new user-facing contract and should be planned separately after deciding whether skills should expose provider names, agent lane names, or both. For Sprint 001, `gemini` remains the only selector for the Google/Gemini lane.

## Implementation Plan

### Phase 1: Baseline and Scope Audit

Files:
- `docs/sprints/drafts/SPRINT-001-INTENT.md`
- `AGENTS.md`
- `README.md`
- `.agents/skills/df-sprint-plan/SKILL.md`
- `.claude/skills/df-sprint-plan/SKILL.md`
- `.agents/skills/df-sprint-execute/SKILL.md`
- `.claude/skills/df-sprint-execute/SKILL.md`
- `.agents/skills/df-chapter-create/SKILL.md`
- `.claude/skills/df-chapter-create/SKILL.md`
- `scripts/validate-skills.py`

Tasks:
- Confirm there are no final `docs/sprints/SPRINT-*.md` files and no selected chapter context.
- Confirm `docs/SEMANTIC-INDEX.md` remains `Status: Not configured`; do not invent index prior art.
- Capture the current migration surface with targeted searches:
  - `rg -n "gemini -p|--yolo|\\bagy\\b|\\bgemini\\b" .agents/skills .claude/skills README.md AGENTS.md scripts`
  - `rg -n "ls docs/sprints/SPRINT-\\*|find docs/sprints" .agents/skills .claude/skills README.md AGENTS.md docs`
- Record the expected compatibility boundary in the implementation notes: `Gemini` lane text is allowed; runnable legacy command forms are not.

### Phase 2: Harden Mirrored Skill Text

Files:
- `.agents/skills/df-sprint-plan/SKILL.md`
- `.claude/skills/df-sprint-plan/SKILL.md`
- `.agents/skills/df-sprint-execute/SKILL.md`
- `.claude/skills/df-sprint-execute/SKILL.md`
- `.agents/skills/df-chapter-create/SKILL.md`
- `.claude/skills/df-chapter-create/SKILL.md`
- `README.md`

Tasks:
- Audit `df-sprint-plan` to ensure all Gemini lane preflight, auth, draft, and critique runnable examples use:
  - `which agy && agy --version`
  - `agy -p "say ok" --dangerously-skip-permissions`
  - `agy -p "<prompt>" --dangerously-skip-permissions`
- Audit `df-sprint-execute` to ensure the parser still accepts `gemini`, maps that selector to `agy`, and uses the same `agy` command form for auth and execution.
- Replace zsh-sensitive sprint-discovery snippets in `df-sprint-plan` with `find` examples:
  - recent sprint docs: `find docs/sprints -maxdepth 1 -type f -name 'SPRINT-*.md' -print 2>/dev/null | sort | tail -3`
  - latest sprint doc: `find docs/sprints -maxdepth 1 -type f -name 'SPRINT-*.md' -print 2>/dev/null | sort | tail -1`
- Replace the same zsh-sensitive recent-sprint snippet in `df-chapter-create` with the matching `find ... | tail -5` form because chapter creation also reads recent sprint docs and can encounter an empty `docs/sprints/` directory.
- Update `README.md` only as needed to state the compatibility boundary: `df-sprint-plan` and `df-sprint-execute` keep the Gemini lane but run it through `agy`.
- Apply mirror substitutions mechanically. Do not paraphrase `.claude` and `.agents` copies differently.

### Phase 3: Add Validator Regression Guards

Files:
- `scripts/validate-skills.py`

Tasks:
- Add named regex checks for legacy runnable command forms inside mirrored `SKILL.md` files:
  - reject `gemini -p`
  - reject `gemini ... --yolo`
  - reject bare `--yolo`
  - reject `which gemini && gemini --version` when it appears as the Gemini lane CLI check
- Add a named regex check for zsh-sensitive sprint glob examples:
  - reject `ls docs/sprints/SPRINT-*.md`
- Keep the errors actionable and file-scoped, matching the validator's existing style.
- Do not reject these allowed compatibility strings:
  - `Gemini`
  - `GEMINI-DRAFT`
  - `GEMINI-CRITIQUE`
  - `/df-sprint-execute gemini`
  - `Agent must be one of: claude, codex, or gemini`
- Ensure the new checks run before mirror normalization so both mirrors report precise source paths.

### Phase 4: Add Focused Validator Tests

Files:
- `tests/test_validate_skills.py` or `scripts/test_validate_skills.py`
- `scripts/validate-skills.py`

Tasks:
- Refactor only enough of `scripts/validate-skills.py` to expose pure helper functions for text-regression checks. Avoid changing the validator's CLI behavior.
- Add standard-library `unittest` coverage for:
  - legacy `gemini -p "say ok" --yolo` is rejected
  - bare `--yolo` is rejected
  - `which gemini && gemini --version` is rejected in a skill snippet
  - `ls docs/sprints/SPRINT-*.md 2>/dev/null | tail -3` is rejected
  - allowed `Gemini` lane prose and `SPRINT-NNN-GEMINI-DRAFT.md` are not rejected
  - allowed `/df-sprint-execute gemini` selector prose is not rejected
- Keep tests fixture-local; do not mutate the real mirrored skill directories.
- Add the test command to this sprint's Definition of Done, but do not rewrite the repo-wide AGENTS SOP unless the team decides all future skill changes must run the new unit test.

### Phase 5: Verify Real CLI and Repo Gates

Files:
- no source changes expected unless verification exposes issues

Tasks:
- Prove the installed `agy` CLI and auth path:
  - `which agy && agy --version`
  - `agy -p "say ok" --dangerously-skip-permissions`
- Run deterministic local checks:
  - `python3 scripts/validate-skills.py`
  - `python3 -m unittest tests/test_validate_skills.py` if the test file is added under `tests/`
  - `python3 -m py_compile scripts/validate-skills.py $(find .claude/skills .agents/skills -path '*/scripts/*.py' -type f | sort)`
  - `git diff --check --cached`
  - `git diff --check`
- Run targeted negative scans against production docs, excluding sprint draft artifacts and test fixtures:
  - `rg -n "gemini -p|--yolo|ls docs/sprints/SPRINT-\\*.md" .agents/skills .claude/skills README.md AGENTS.md scripts`
- Run the Claude critique gate from `AGENTS.md` and fix actionable findings.
- If `py_compile` fails only because a sandbox cannot write `__pycache__` under mirrored skill directories, rerun with an allowed pycache prefix for local signal and record the exact blocker separately. The final acceptance should still include the normal SOP command passing in a writable environment.

## Files Summary

- `scripts/validate-skills.py` - Add focused text-regression checks for legacy Google/Gemini command forms and zsh-sensitive sprint glob examples; optionally expose pure helper functions for unit tests.
- `tests/test_validate_skills.py` - Add small standard-library tests for validator allow/deny behavior, if the implementation chooses a separate test file.
- `.agents/skills/df-sprint-plan/SKILL.md` and `.claude/skills/df-sprint-plan/SKILL.md` - Audit `agy` examples and replace first-sprint `ls docs/sprints/SPRINT-*.md` snippets with shell-portable `find` snippets.
- `.agents/skills/df-sprint-execute/SKILL.md` and `.claude/skills/df-sprint-execute/SKILL.md` - Audit selector-to-CLI wording and runnable `agy` examples while preserving the `gemini` selector.
- `.agents/skills/df-chapter-create/SKILL.md` and `.claude/skills/df-chapter-create/SKILL.md` - Replace the recent-sprint `ls docs/sprints/SPRINT-*.md` snippet with a shell-portable `find` snippet.
- `README.md` - Keep or tighten the concise statement that the Gemini lane runs through `agy`.
- `docs/sprints/SPRINT-001.md` and `docs/sprints/ledger.yaml` - Created later by the merge/orchestrator phase, not by this Codex draft.

## Definition of Done

- Both skill mirrors contain equivalent changes with only the allowed `.claude/skills/...` vs `.agents/skills/...` and `CLAUDE.md` vs `AGENTS.md` substitutions.
- `python3 scripts/validate-skills.py` passes on the finished repo and fails when representative legacy snippets are injected into a temporary skill fixture or covered by unit tests.
- No mirrored skill contains runnable legacy examples matching `gemini -p`, `--yolo`, `which gemini && gemini --version`, or `ls docs/sprints/SPRINT-*.md`.
- Allowed compatibility text remains present where appropriate: `Gemini` lane naming, `GEMINI` artifact filenames, and `/df-sprint-execute gemini`.
- `which agy && agy --version` and `agy -p "say ok" --dangerously-skip-permissions` pass with the installed CLI.
- Local gates pass:
  - `python3 scripts/validate-skills.py`
  - `python3 -m unittest tests/test_validate_skills.py` if tests are added
  - `python3 -m py_compile scripts/validate-skills.py $(find .claude/skills .agents/skills -path '*/scripts/*.py' -type f | sort)`
  - `git diff --check --cached`
  - `git diff --check`
- The Claude critique gate from `AGENTS.md` returns no actionable findings, or any accepted non-blocking tradeoff is documented.
- The final sprint document explicitly states that `Gemini` remains the lane/selector while `agy` is the executable CLI.

## Risks & Mitigations

- Risk: Overzealous validation rejects legitimate `Gemini` lane prose or `GEMINI` filenames.
  Mitigation: Validate command-shaped legacy forms, not the word `gemini`; add positive tests for allowed compatibility strings.

- Risk: The sprint expands from hardening into a selector redesign by adding `/df-sprint-execute agy`.
  Mitigation: Keep `agy` aliasing out of scope for Sprint 001. Record it as a future compatibility decision.

- Risk: zsh portability work changes semantics of sprint selection order.
  Mitigation: Use `find ... -print | sort | tail -N`, matching the prior lexicographic behavior closely enough for zero-padded `SPRINT-NNN.md` filenames.

- Risk: Bad-command fixtures make targeted `rg` scans fail.
  Mitigation: Keep negative fixtures in unit-test strings, not production skill files, and scope production scans to `.agents/skills`, `.claude/skills`, `README.md`, `AGENTS.md`, and `scripts/validate-skills.py`.

- Risk: The py_compile gate writes `__pycache__` in locations the local sandbox cannot modify.
  Mitigation: Treat that as an environment blocker only if the normal command fails in the actual repo environment; use `PYTHONPYCACHEPREFIX` only as supplemental syntax signal, not as a replacement for the SOP gate.

- Risk: Mirror drift appears while editing multiple skills.
  Mitigation: Edit one mirror pair at a time, then run `python3 scripts/validate-skills.py` before moving to the next pair.

## Dependencies

- Installed and authenticated `agy` CLI. The intent records a passing local preflight for `/Users/jmccarthy/.local/bin/agy` version `1.0.10`, but execution should verify it again.
- Existing repo mirror invariant and validator behavior from `AGENTS.md`.
- Python standard library only for validator tests, unless the implementer explicitly chooses a different test harness and updates the plan.
- Claude CLI availability for the required critique gate.
- No semantic index or chapter context is required for this sprint.

## Open Questions

- Should the final merged sprint include the `df-chapter-create` zsh-sensitive sprint probe, or keep the shell-portability fix limited to `df-sprint-plan`? My recommendation is to include it because it is the same documented pattern and a low-risk mirror edit.
- Should the new unit test command become part of the permanent `AGENTS.md` SOP? My recommendation is to defer that SOP expansion until the first implementation proves the tests are low-noise.
- Should `/df-sprint-execute agy` be accepted as an alias for `/df-sprint-execute gemini`? My recommendation is no for this sprint; preserve the current user-facing selector and revisit aliasing separately.
- Should differing Codex auth examples in `df-sprint-plan` and `df-sprint-execute` be standardized now? My recommendation is to defer because it is not part of the `agy` migration and could change Codex execution behavior.
