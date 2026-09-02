# Sprint 001 Gemini Critique: Agy Migration Hardening

## Pyramid Index

- L0: Evaluation of the competing Claude and Codex SPRINT-001 drafts, analyzing their architecture, completeness, risks, feasibility, and Definition of Done, followed by a merged recommendation.
- L1:
  - Evaluate [SPRINT-001-CLAUDE-DRAFT.md](file:///Users/jmccarthy/code/skills/docs/sprints/drafts/SPRINT-001-CLAUDE-DRAFT.md): Strong architectural refactoring, inline validator `--selftest` design, granular task lists, but leaves Codex auth drift out of scope.
  - Evaluate [SPRINT-001-CODEX-DRAFT.md](file:///Users/jmccarthy/code/skills/docs/sprints/drafts/SPRINT-001-CODEX-DRAFT.md): Strong identification of sandbox permission risks and baseline audit step, but introduces file pollution with a separate test script and lists risky phasing.
  - Synthesize merge recommendations: Adopt Claude's refactoring, inline `--selftest`, and task order, but integrate Codex's baseline scope audit and sandbox permissions mitigations, and standardize Codex auth checks.
- L2:
  - Claude draft evaluation details → [Claude Draft Critique](#claude-draft-critique).
  - Codex draft evaluation details → [Codex Draft Critique](#codex-draft-critique).
  - Synthesis and recommendations for the merge notes → [Synthesis & Merge Recommendations](#synthesis--merge-recommendations).

---

## Overview & Executive Summary

Both drafts represent high-quality planning efforts that address the core seed requirement: making the legacy `gemini` to new `agy` CLI migration durable and shell-portable. Both plans respect the critical mirror invariant defined in [AGENTS.md](file:///Users/jmccarthy/code/skills/AGENTS.md) and correctly identify the need to preserve the `Gemini` planning lane name, `GEMINI` filenames, and the `gemini` execute selector while hardening the underlying runnable code examples to use the new `agy` CLI binary.

The primary differences lie in **testing architecture** (inline self-test vs. separate test file), **phasing safety** (refactoring validator before or after skill text edits), and **scope decisions** (whether to standardize the Codex auth checks).

---

## Claude Draft Critique

Reference: [SPRINT-001-CLAUDE-DRAFT.md](file:///Users/jmccarthy/code/skills/docs/sprints/drafts/SPRINT-001-CLAUDE-DRAFT.md)

### 1. Architectural Soundness
- **Highly Sound.** Banning specific command invocation shapes (`gemini -p`, `--yolo`, `which gemini`, etc.) via a list of compiled regex patterns is the correct way to allow legitimate "Gemini" lane prose while blocking legacy runnable CLI commands.
- **Excellent Refactoring.** Isolating content check rules into a pure helper function `check_skill_content(label, content)` in [validate-skills.py](file:///Users/jmccarthy/code/skills/scripts/validate-skills.py) makes the script modular and testable without affecting CLI behavior.
- **Brilliant Testing Design.** The `--selftest` CLI flag using embedded inline string constants is elegant. Since [validate-skills.py](file:///Users/jmccarthy/code/skills/scripts/validate-skills.py) is a top-level validation helper and is not mirrored, keeping the tests inside it prevents adding extra files or folders (like `tests/` or `scripts/test_*.py`) that could complicate the mirror invariant checks or pollute the project.

### 2. Completeness
- **Extremely Complete.** Includes precise file lists, estimated line numbers, task descriptions, regex patterns, and detailed command strings. It leaves nothing to guess.

### 3. Phasing/Ordering
- **Excellent.** Claude structures the phases to build the validation guard and `--selftest` logic first (Phase 1), and then modify the mirrored skill text (Phase 2). This ensures that when the developer writes the skill modifications, the validator is already armed and can immediately be run to verify the correctness of the changes.

### 4. Risk Coverage
- **Very Good.** Identifies risks of regex over-broadness, refactoring side-effects, sorting discrepancies in shell commands, and mirror drift during edits. It proposes strong mitigations for each.

### 5. Feasibility
- **High.** The plan is straightforward and contains all the details needed to execute without ambiguity.

### 6. Definition of Done
- **Robust and Testable.** Every item on the DoD checklist is clear, deterministic, and mapped directly to a success criterion.

### Strongest Ideas to Keep:
- **Refactoring content checks** into a testable function `check_skill_content`.
- **Inline `--selftest` execution** inside [validate-skills.py](file:///Users/jmccarthy/code/skills/scripts/validate-skills.py) to avoid filesystem pollution.
- **Phase 1-first validator hardening** to enable validator-assisted edits in subsequent phases.

### Weaknesses & Gaps:
- Recommends leaving Codex auth preflight standardisation out of scope. Although standardizing Codex is technically separate, the drift between `df-sprint-plan` and `df-sprint-execute` was highlighted in the intent, and fixing it represents a low-risk, high-value alignment.

---

## Codex Draft Critique

Reference: [SPRINT-001-CODEX-DRAFT.md](file:///Users/jmccarthy/code/skills/docs/sprints/drafts/SPRINT-001-CODEX-DRAFT.md)

### 1. Architectural Soundness
- **Sound, but adds file pollution.** Codex suggests adding a separate test file `tests/test_validate_skills.py` (or `scripts/test_validate_skills.py`) using `unittest`. While this is standard Python practice, it introduces additional directory maintenance to a project that relies on a flat, simple script layout. It also requires modifying the gate command sequence in [AGENTS.md](file:///Users/jmccarthy/code/skills/AGENTS.md) to run the new test file, adding friction.
- **Portable Globbing.** The `find` command addition `-type f -print` is slightly more precise than Claude's, ensuring we only match files and explicitly print paths.

### 2. Completeness
- **Moderately Complete.** The draft outlines the tasks but lacks the specific regex definitions, line numbers, or exact implementation details found in Claude's draft.

### 3. Phasing/Ordering
- **Sub-optimal Phasing.** Codex schedules skill text edits (Phase 2) *before* the validator guards are implemented (Phase 3). This is a riskier order of operations. If a developer makes an error in Phase 2, there is no tool to check it until Phase 3 or 4.
- **Good Prep Step.** Phase 1 includes a "Baseline and Scope Audit" using `rg` to map the exact locations of legacy patterns before modifying anything. This is a very useful safety check.

### 4. Risk Coverage
- **Excellent and Detailed.** Codex identifies environment-specific risks that Claude missed:
  - *Sandbox permission risks:* Sandboxed compilers might fail to write compiled `__pycache__` files in read-only directories, which could break `py_compile` checks. Codex proposes a workaround using `PYTHONPYCACHEPREFIX`.
  - *Regex scanning false positives:* Warns that mock strings or bad-command fixtures could cause production recursive `rg` scans to fail.

### 5. Feasibility
- **High**, though maintaining a separate test suite adds minor complexity.

### 6. Definition of Done
- **Good**, covers the core criteria, but lacks details on how the validator's test suite should be integrated into the final merge criteria.

### Strongest Ideas to Keep:
- **Baseline scope audit** (Phase 1 tasks using `rg` to find occurrences).
- **`__pycache__` compilation permission warning** and mitigation for sandbox constraints.
- **Standardizing the Codex auth preflight commands** across `df-sprint-plan` and `df-sprint-execute`.

### Weaknesses & Gaps:
- Proposing a separate test file introduces file pollution.
- Implementing validation guards *after* changing the skill files increases risk during the edit phase.
- Missing concrete regex definitions for the validator.

---

## Synthesis & Merge Recommendations

### Consensus Strengths
Both plans correctly implement the zsh-portable `find` command replacement and the regex-based validation of specific legacy CLI invocation forms (to preserve `Gemini` lane naming).

### Key Tensions & Resolution
1. **Testing File Structure (Inline `--selftest` vs. External `test_validate_skills.py`):**
   - *Resolution:* **Keep tests inline.** Claude's `--selftest` design is superior for this codebase because it prevents adding new test runner dependencies or files that must be exempted from the mirror-invariant logic.
2. **Standardizing Codex Preflight Auth:**
   - *Resolution:* **Standardize it.** Codex and Gemini drafts both recommend standardizing the Codex preflight to `codex exec --sandbox workspace-write "echo ok"`. This resolves the drift and provides a safer auth check. We should include this in the final plan.
3. **Phasing Order:**
   - *Resolution:* **Validator First.** Harden the validator (Phase 1) first as proposed by Claude, but prepend Codex's baseline scope audit to the beginning of Phase 1.

### Proposed Merged Sprint Path
- **Phase 1: Scope Audit & Validator Hardening** (Run baseline `rg` scans, then refactor [validate-skills.py](file:///Users/jmccarthy/code/skills/scripts/validate-skills.py) to add the legacy patterns and inline `--selftest`).
- **Phase 2: Skill Text Hardening** (Update the 3 `find` discovery probes, update the Codex auth commands to be uniform, and verify `agy` command usage).
- **Phase 3: Documentation & Compatibility** (Add notes about the compatibility boundary).
- **Phase 4: Gate Execution & Sandbox Proof** (Execute all gate commands, handle any `__pycache__` writing constraints in sandbox environments, and verify real CLI actions).
