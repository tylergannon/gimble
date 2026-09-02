# Sprint 001 Draft: Agy Migration Hardening (Gemini)

## Pyramid Index

- L0: Harden the Diffusion skills repo after migrating the Google/Gemini lane from the legacy `gemini` CLI to the default `agy` CLI, making the migration durable, shell-portable, and regression-proof.
- L1:
  - Add validator coverage in `scripts/validate-skills.py` to ban legacy `gemini` CLI arguments like `-p` and `--yolo` across both mirrors.
  - Fix shell-portability issues with first-sprint checks by replacing `ls` globbing with `find`.
  - Standardize Codex auth preflight examples to use `--sandbox workspace-write` uniformly.
  - Maintain the existing `Gemini` planning lane naming conventions and selector semantics while documenting the `agy` CLI mapping.
- L2:
  - See [Architecture](#architecture) for validation regex design and shell-portability solutions.
  - See [Implementation Plan](#implementation-plan) for Phase 1 (Validator updates) and Phase 2 (Skill modifications).
  - See [Open Questions](#open-questions) for compatibility boundaries and standardisation decisions.

## Overview

This sprint plan addresses the hardening and durability of the Diffusion skills repository following the migration of the Google/Gemini planning and execution lanes from the legacy `gemini` CLI tool to the new default `agy` CLI tool.

The primary goal of this sprint is to lock in this migration by preventing regressions (such as the reintroduction of legacy `gemini -p` or `gemini --yolo` commands) via automated validations, resolving zsh shell compatibility warnings when discovering sprint files during bootstraps, and resolving inconsistencies in other preflight checks (like Codex).

## Use Cases

1. **Preventing CLI Command Regressions**: A developer edits a skill file and accidentally references the legacy `gemini -p` or `--yolo` parameters. The local gate (`python3 scripts/validate-skills.py`) detects this and blocks the commit.
2. **First-Sprint Bootstrap Portability**: In a newly initialized repo where no sprint files exist, a user runs `/df-sprint-plan`. Under `zsh` (the default macOS shell), the directory globbing logic checks for sprints without throwing a shell `no matches found` error, allowing the orientation phase to complete gracefully.
3. **Consistent Verification**: A user runs the preflight checks for planning or execution. The commands used to verify the tools (e.g. `codex` sandbox writes) are uniform, easy to understand, and secure.

## Architecture

### 1. Hardening the Skill Validator

We will update the `scripts/validate-skills.py` script to inspect the content of all validated `SKILL.md` files. It will check for:
- Text patterns matching legacy executable invocation shapes like `gemini -p` or `--yolo`.
- Regex matching:
  - `gemini\s+-p` (matches direct usage of the legacy prompt flag on the old command)
  - `\b--yolo\b` (matches legacy Codex/Gemini yolo execution flag)
- If either pattern is matched, the script outputs a clear error and exits with code `1`.
- We will *not* ban the word "gemini" in general, since it is needed to document the `Gemini` planning lane, select the `gemini` option in `/df-sprint-execute`, and name the `GEMINI` draft/critique filenames.

### 2. Solving Shell Portability (`ls` vs `find`)

Under `zsh`, running a command with a glob (like `docs/sprints/SPRINT-*.md`) when no files match results in:
`zsh: no matches found: docs/sprints/SPRINT-*.md`
before the command can even suppress output via `2>/dev/null`.

To resolve this portable shell issue, we will replace:
- `ls docs/sprints/SPRINT-*.md 2>/dev/null | tail -3`
- `ls docs/sprints/SPRINT-*.md 2>/dev/null | tail -1`

with a `find`-based command that suppresses errors and returns paths safely:
- `find docs/sprints -maxdepth 1 -name "SPRINT-*.md" 2>/dev/null | sort | tail -3`
- `find docs/sprints -maxdepth 1 -name "SPRINT-*.md" 2>/dev/null | sort | tail -1`

This executes correctly without shell expansion errors even if:
- The `docs/sprints` directory does not exist yet.
- The directory exists but is completely empty.

### 3. Codex Preflight Standardisation

Currently, `df-sprint-plan` uses:
`codex exec --sandbox workspace-write "echo ok"`
While `df-sprint-execute` uses:
`codex exec --full-auto "echo ok"`

We will standardize both skills to use:
`codex exec --sandbox workspace-write "echo ok"`
This is the safer, sandboxed check that validates file-writing capability without executing completely unconstrained commands.

---

## Implementation Plan

### Phase 1: Validator Hardening
Harden the validation script to catch legacy command constructs.

- **File**: [scripts/validate-skills.py](file:///Users/jmccarthy/code/skills/scripts/validate-skills.py)
  - **Task 1.1**: Define patterns to detect legacy usage of `gemini` commands, specifically `gemini -p` and any parameters like `--yolo`.
  - **Task 1.2**: Implement checks inside the `SKILL.md` content inspection loop. If legacy patterns are found, append formatted errors.
  - **Task 1.3**: Add error documentation indicating the migrated command: `agy -p ... --dangerously-skip-permissions`.

### Phase 2: Skill Hardening (Plan & Execute)
Apply changes to both mirrors (`.agents/` and `.claude/`) for both `df-sprint-plan` and `df-sprint-execute`.

- **Files**:
  - [.agents/skills/df-sprint-plan/SKILL.md](file:///Users/jmccarthy/code/skills/.agents/skills/df-sprint-plan/SKILL.md)
  - [.claude/skills/df-sprint-plan/SKILL.md](file:///Users/jmccarthy/code/skills/.claude/skills/df-sprint-plan/SKILL.md)
  - [.agents/skills/df-sprint-execute/SKILL.md](file:///Users/jmccarthy/code/skills/.agents/skills/df-sprint-execute/SKILL.md)
  - [.claude/skills/df-sprint-execute/SKILL.md](file:///Users/jmccarthy/code/skills/.claude/skills/df-sprint-execute/SKILL.md)

  - **Task 2.1**: Update Phase 1 (Orient) and Phase 2 (Intent) in both `df-sprint-plan` mirrors to use the portable `find` commands instead of `ls` globbing.
  - **Task 2.2**: Update the Codex auth check command in the `df-sprint-execute` preflight section in both mirrors to match the standard sandbox write command from `df-sprint-plan`.
  - **Task 2.3**: Verify that all examples in both mirrors use `agy -p ... --dangerously-skip-permissions` for Gemini execution.

### Phase 3: Validation and Gate Execution
Run the full gate suite to verify that changes adhere to repository conventions.

  - **Task 3.1**: Run `python3 scripts/validate-skills.py` locally.
  - **Task 3.2**: Compile supporting scripts: `python3 -m py_compile scripts/validate-skills.py`.
  - **Task 3.3**: Ensure `git diff --check` displays no warnings.
  - **Task 3.4**: (Optional/Orchestrator task) Feed changes to the Claude critique gate as described in `AGENTS.md`.

---

## Files Summary

| File Path | Description of Changes |
|-----------|------------------------|
| [scripts/validate-skills.py](file:///Users/jmccarthy/code/skills/scripts/validate-skills.py) | Added checks to reject `gemini -p` and `--yolo` patterns. |
| [.agents/skills/df-sprint-plan/SKILL.md](file:///Users/jmccarthy/code/skills/.agents/skills/df-sprint-plan/SKILL.md) | Replaced `ls` globbing with `find` for shell compatibility. |
| [.claude/skills/df-sprint-plan/SKILL.md](file:///Users/jmccarthy/code/skills/.claude/skills/df-sprint-plan/SKILL.md) | Mirrored `find` changes. |
| [.agents/skills/df-sprint-execute/SKILL.md](file:///Users/jmccarthy/code/skills/.agents/skills/df-sprint-execute/SKILL.md) | Standardized Codex preflight auth command. |
| [.claude/skills/df-sprint-execute/SKILL.md](file:///Users/jmccarthy/code/skills/.claude/skills/df-sprint-execute/SKILL.md) | Mirrored Codex preflight changes. |

---

## Definition of Done

1. **Automated Verification**: `python3 scripts/validate-skills.py` runs and passes successfully.
2. **Regression Blocking**: Introducing a string containing `gemini -p` or `--yolo` into any `SKILL.md` file causes the validator script to return exit code 1 with descriptive errors.
3. **No Shell Errors**: Running the newly documented `find` commands on a clean machine (where `docs/sprints/` is absent or empty) returns nothing on stdout/stderr, and does not trigger zsh syntax or expansion failures.
4. **Mirror Parity**: Claude and Agents mirrors contain identical content except for the expected mechanical swaps (e.g. `AGENTS.md` vs `CLAUDE.md`, and `.agents/skills` vs `.claude/skills`).

---

## Risks & Mitigations

- **Risk: Strict validation flags false positives.**
  - *Mitigation*: We do not do a blanket ban on the term `gemini`. We only ban specific CLI execution patterns like `gemini -p` and `--yolo` to prevent incorrect command suggestions.
- **Risk: Breaking existing setups or aliases.**
  - *Mitigation*: We maintain `gemini` as the agent selector in `df-sprint-execute` (e.g. `/df-sprint-execute gemini`), preserving user muscle memory and scripting layers. We merely change the command run underneath the hood to use the verified `agy` executable.

---

## Dependencies

- **`agy` CLI**: The system running execution must have `agy` installed and configured under `~/.local/bin/agy` or another PATH location.
- **`python3`**: The validation script runs on python 3.8+.

---

## Open Questions

1. **Should `/df-sprint-execute agy` become an accepted alias for `/df-sprint-execute gemini`?**
   - *Proposal*: To preserve maximum backwards-compatibility and minimize documentation noise, we keep the user-facing option name as `gemini`. Adding an `agy` alias is technically trivial but increases option complexity. We recommend leaving it as `gemini` only.
2. **Should `scripts/validate-skills.py` explicitly ban all `gemini` command binaries?**
   - *Proposal*: No, banning the bare word `gemini` is too broad since the planning lane name is `Gemini` and filenames are `GEMINI-DRAFT`. Banning the specific invocation patterns `gemini -p` and `--yolo` is sufficient to protect against the legacy CLI examples.
3. **Should this sprint standardize the differing Codex auth examples?**
   - *Proposal*: Yes, standardizing both to `codex exec --sandbox workspace-write "echo ok"` reduces drift and secures execution checks. We have included this task in Phase 2 of this plan.
4. **Should the zsh no-match behavior be fixed in this sprint?**
   - *Proposal*: Yes, replacing `ls` with `find` provides a robust, portable shell experience. We have included this task in Phase 2 of this plan.
