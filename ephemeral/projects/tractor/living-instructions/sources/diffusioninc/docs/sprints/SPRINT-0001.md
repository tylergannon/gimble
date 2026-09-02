# Sprint 0001: Agy Migration Hardening

## Pyramid Index

- L0: Make the Google/Gemini lane migration to `agy` durable through validator guards, shell-portable sprint discovery, and real CLI proof.
- L1:
  - Preserve the compatibility boundary: `Gemini` remains the lane name, `GEMINI` remains the artifact infix, and `gemini` remains the user-facing selector.
  - Add deterministic validation for legacy runnable command forms such as `gemini -p`, `--yolo`, `which gemini`, and `gemini --version`.
  - Replace zsh-fragile `ls docs/sprints/SPRINT-*.md` examples with `find ... | sort | tail -N` in all affected mirrored skills.
  - Verify with the full `AGENTS.md` SOP gates, inline validator self-test, Claude critique, and real `agy` invocation.
- L2:
  - Compatibility boundary: Overview and Architecture.
  - Validator guard: Implementation Phase 1.
  - Shell portability: Implementation Phase 2.
  - Proof and acceptance: Definition of Done.

## Overview

The current working tree has already switched the runnable Google/Gemini lane examples from the legacy `gemini` CLI to Google's `agy` CLI. The stable product-facing contract remains unchanged: the planning lane is named `Gemini`, draft and critique files use the `GEMINI` infix, and `/df-sprint-execute gemini` remains the selector. Only the underlying CLI binary changed.

Sprint 0001 hardens that state. It should make legacy command examples fail validation, fix the first-sprint zsh `no matches found` probes surfaced during this planning run, and record proof that `agy` works in the local environment. This is a narrow skills-repo hardening sprint with no chapter link and no semantic index configured.

## Use Cases

1. A maintainer pastes an old `gemini -p "..." --yolo` example into a mirrored `SKILL.md`; `python3 scripts/validate-skills.py` fails with a file-scoped error.
2. A future edit keeps valid `Gemini` lane prose, `GEMINI-DRAFT` filenames, and `/df-sprint-execute gemini`; validation passes because the guard targets command forms, not the word `gemini`.
3. A planner runs first-sprint discovery under zsh in a repo with no sprint docs; documented snippets return cleanly instead of emitting `zsh: no matches found`.
4. A reviewer can rerun `which agy && agy --version` and `agy -p "say ok" --dangerously-skip-permissions` to prove the real CLI surface.

## Architecture

### Compatibility Boundary

The `Gemini` lane is a stable planning identity. The `GEMINI` filename infix and `gemini` selector are compatibility surfaces and must not be renamed in this sprint. `agy` is the implementation CLI for that lane and should appear only in runnable command examples and explanatory CLI mapping prose.

### Validator Guard

`scripts/validate-skills.py` remains the single deterministic gate for mirrored skill content. Add a small, explicit set of legacy-command checks to the existing per-file content validation path. The checks should reject invocation shapes, not the bare word `gemini`.

Recommended rejected patterns:

- `gemini -p`
- `--yolo`
- `which gemini`
- `gemini --version`
- `gemini exec`
- `ls docs/sprints/SPRINT-*.md`

Refactor the existing per-file content checks into a pure helper such as `check_skill_content(label, content) -> list[str]`. The default CLI behavior should remain unchanged, except that it now emits new errors when legacy forms appear.

Add an inline `--selftest` mode to `scripts/validate-skills.py`. It should exercise the same helper with embedded legacy and migrated fixtures. Keep this inside the existing script rather than adding a new test tree.

### Shell-Portable Discovery

The documented `ls docs/sprints/SPRINT-*.md 2>/dev/null | tail -N` snippets are fragile under zsh because unmatched globs fail before `ls` runs. Replace them with a `find` form:

```bash
find docs/sprints -maxdepth 1 -type f -name 'SPRINT-*.md' -print 2>/dev/null | sort | tail -N
```

Use the matching tail count for each current snippet. Include both `df-sprint-plan` and `df-chapter-create` mirrors.

## Implementation Plan

### Phase 1: Baseline Audit and Validator Guard

Files:

- `scripts/validate-skills.py`

Tasks:

- Run targeted scans to capture the current state:
  - `rg -n "gemini -p|--yolo|which gemini|gemini --version|gemini exec" .agents/skills .claude/skills README.md AGENTS.md scripts`
  - `rg -n "ls docs/sprints/SPRINT-\\*" .agents/skills .claude/skills README.md AGENTS.md scripts`
- Refactor existing string-based per-file checks into `check_skill_content(label, content) -> list[str]`.
- Add command-form checks for the legacy Google/Gemini CLI patterns and zsh-fragile sprint discovery pattern.
- Add `python3 scripts/validate-skills.py --selftest` with embedded fixtures:
  - legacy fixture must fail;
  - migrated fixture with valid `Gemini` prose and `GEMINI` filenames must pass.
- Confirm `python3 scripts/validate-skills.py` still passes on the current tree before moving to skill edits.

### Phase 2: Shell-Portable Skill Text

Files:

- `.agents/skills/df-sprint-plan/SKILL.md`
- `.claude/skills/df-sprint-plan/SKILL.md`
- `.agents/skills/df-chapter-create/SKILL.md`
- `.claude/skills/df-chapter-create/SKILL.md`

Tasks:

- Replace the Phase 1 recent-sprints probe in both `df-sprint-plan` mirrors with the `find ... | sort | tail -3` form.
- Replace the Phase 2 latest-sprint probe in both `df-sprint-plan` mirrors with the `find ... | sort | tail -1` form.
- Replace the recent-sprints probe in both `df-chapter-create` mirrors with the `find ... | sort | tail -5` form.
- Run zsh smoke checks for the new snippets in an empty or absent `docs/sprints` scenario.

### Phase 3: Compatibility Documentation

Files:

- `.agents/skills/df-sprint-plan/SKILL.md`
- `.claude/skills/df-sprint-plan/SKILL.md`
- `.agents/skills/df-sprint-execute/SKILL.md`
- `.claude/skills/df-sprint-execute/SKILL.md`
- `README.md`

Tasks:

- Ensure the existing `agy` CLI mapping is clear without adding repetitive prose.
- Add or tighten one concise sentence only where needed: `Gemini` is the lane and `agy` is the CLI binary.
- Do not add `/df-sprint-execute agy` aliasing in this sprint.
- Do not standardize Codex auth or execution commands in this sprint; record that as follow-up work.

### Phase 4: Verification and Critique Gate

Tasks:

- Run the full local gate sequence from `AGENTS.md`:
  - `git add -N README.md .claude/skills .agents/skills scripts 2>/dev/null || true`
  - `python3 scripts/validate-skills.py`
  - `python3 scripts/validate-skills.py --selftest`
  - `python3 -m py_compile scripts/validate-skills.py $(find .claude/skills .agents/skills -path '*/scripts/*.py' -type f | sort)`
  - `git diff --check --cached`
  - `git diff --check`
- Prove the CLI surface:
  - `which agy && agy --version`
  - `agy -p "say ok" --dangerously-skip-permissions`
- Run production negative scans scoped away from sprint drafts and inline fixtures:
  - `rg -n "gemini -p|--yolo|which gemini|gemini --version|gemini exec|ls docs/sprints/SPRINT-\\*" .agents/skills .claude/skills README.md AGENTS.md`
- Run the Claude critique gate from `AGENTS.md`.
- Fix actionable findings and repeat gates until clean.

## Files Summary

- `scripts/validate-skills.py` - Add command-form and zsh-fragile-snippet guards; add `--selftest`.
- `.agents/skills/df-sprint-plan/SKILL.md` and `.claude/skills/df-sprint-plan/SKILL.md` - Replace two sprint-discovery `ls` snippets with `find` snippets; keep `agy` mapping clear.
- `.agents/skills/df-chapter-create/SKILL.md` and `.claude/skills/df-chapter-create/SKILL.md` - Replace the recent-sprints `ls` snippet with `find`.
- `.agents/skills/df-sprint-execute/SKILL.md` and `.claude/skills/df-sprint-execute/SKILL.md` - Keep the `gemini` selector mapped to `agy`; only tighten prose if needed.
- `README.md` - Keep the Google/Gemini lane `agy` summary accurate.

## Definition of Done

- `python3 scripts/validate-skills.py` passes on the finished tree.
- `python3 scripts/validate-skills.py --selftest` proves legacy command fixtures fail and migrated compatibility fixtures pass.
- No mirrored skill contains legacy runnable examples matching `gemini -p`, `--yolo`, `which gemini`, `gemini --version`, `gemini exec`, or `ls docs/sprints/SPRINT-*.md`.
- Valid `Gemini` lane prose, `GEMINI` artifact filenames, and `/df-sprint-execute gemini` selector prose remain allowed.
- All affected `.claude` and `.agents` mirrors differ only by the permitted mechanical substitutions.
- New `find ... | sort | tail -N` snippets run cleanly under zsh with no matching sprint docs.
- `which agy && agy --version` and `agy -p "say ok" --dangerously-skip-permissions` pass.
- Full `AGENTS.md` local gates pass, including `py_compile`, cached and working-tree `diff --check`, and the Claude critique gate.
- No chapter link is introduced, and `docs/SEMANTIC-INDEX.md` remains accurate.

## Risks & Mitigations

- Risk: The validator rejects legitimate `Gemini` lane text.
  Mitigation: Match command forms, not the bare token; cover allowed prose in `--selftest`.
- Risk: Inline negative fixtures poison production scans.
  Mitigation: Scope production `rg` scans to skill and README surfaces, not `scripts/validate-skills.py` when inline fixtures intentionally contain bad examples.
- Risk: `find` output order differs from `ls`.
  Mitigation: Pipe through `sort` before `tail`.
- Risk: Mirror edits drift between `.claude` and `.agents`.
  Mitigation: Edit mirror pairs together and run `python3 scripts/validate-skills.py` after each pair or before continuing.
- Risk: Codex auth standardization broadens the sprint.
  Mitigation: Defer it as follow-up unless the human planner explicitly expands scope.
- Risk: `py_compile` hits a local `__pycache__` write constraint.
  Mitigation: Record the exact failure and optionally run `PYTHONPYCACHEPREFIX` as supplemental syntax signal, but do not replace the required SOP gate without an explicit decision.

## Dependencies

- Installed and authenticated `agy` CLI.
- Python 3 for `validate-skills.py`, `--selftest`, and `py_compile`.
- Claude CLI for the required critique gate.
- No semantic index, token cache, or chapter context is required.

## Open Questions

- Should `/df-sprint-execute agy` become an accepted alias for `/df-sprint-execute gemini`? Default answer for Sprint 0001: no.
- Should Codex auth examples be standardized across `df-sprint-plan` and `df-sprint-execute`? Default answer for Sprint 0001: defer.
- Should inline `--selftest` become part of the permanent `AGENTS.md` SOP after implementation proves low-noise? Default answer: include it in this sprint's DoD first, then decide.
