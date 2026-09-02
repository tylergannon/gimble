# Sprint 001: Agy Migration Hardening

## Pyramid Index

- L0: Make the `gemini`→`agy` CLI migration durable by adding a validator regression guard (with its own self-test), fixing zsh-fragile first-sprint discovery probes, and documenting the Gemini-lane/agy-CLI compatibility boundary — all across both skill mirrors with full SOP gates green.
- L1:
  - Add a content guardrail to `scripts/validate-skills.py` that fails on legacy Google CLI **command forms** (`gemini -p`, `--yolo`, `which gemini`, `gemini --version`, `gemini exec`) while leaving the legitimate `gemini` lane name, selector, and `GEMINI` artifact names untouched.
  - Make the validator's new guard self-proving: refactor per-file checks into a pure function and add a `--selftest` mode that asserts the guard fails on a legacy fixture and passes on the migrated fixture.
  - Replace the three zsh-fragile `ls docs/sprints/SPRINT-*.md` discovery probes (in `df-sprint-plan` ×2 and `df-chapter-create` ×1) with a shell-portable `find` form, mirrored in `.claude` and `.agents`.
  - State the compatibility boundary explicitly: the `Gemini` lane name, `GEMINI` draft/critique filenames, and user-facing `gemini` selector are stable; `agy` is only the underlying CLI binary.
  - Preserve mirror parity and pass every SOP gate, including `git diff --check`, `py_compile`, the validator (now with `--selftest`), the Claude critique gate, and a real `agy` preflight.
- L2:
  - Validator guard + self-test → **Phase 1** and `## Architecture › Validator guardrail`.
  - Shell portability → **Phase 2** and `## Architecture › Shell-portable discovery`.
  - Compatibility-boundary documentation → **Phase 3** and `## Architecture › Compatibility boundary`.
  - Gate execution + proof → **Phase 4** and `## Definition of Done`.

## Overview

The Gemini planning/execution lane has been migrated off the legacy `gemini` CLI
to Google's new default CLI, `agy`. The text edits are already in the working
tree (`gemini -p … --yolo` → `agy -p … --dangerously-skip-permissions`) across
all four `df-*` skill mirrors, the lane name `Gemini` and the user-facing
`gemini` selector were intentionally preserved, and a real preflight confirmed
`agy 1.0.10` answers `agy -p "say ok" --dangerously-skip-permissions`.

What is **not** yet done is making that migration *durable*. Today nothing stops
a future edit (or a stale copy/paste from old docs) from silently reintroducing
`gemini -p` or `--yolo`. The validator has no opinion about CLI command forms,
and the migration's correctness rests entirely on text inspection.

This sprint converts the migration from "done in text" to "protected by a
test." It adds a narrow, self-proving regression guard to the validator, fixes a
first-sprint shell-portability wrinkle exposed during orientation (zsh
`no matches found` on empty `SPRINT-*.md` globs), and writes down the
compatibility boundary so the deliberate split between the `Gemini` lane name
and the `agy` binary is unambiguous to future maintainers.

This is a hardening sprint: surgical, low-risk, fully gated, and confined to the
skills repo. There is no chapter link and no semantic index configured, so the
plan proceeds from repo-local files and targeted searches only.

## Use Cases

1. **A maintainer reintroduces a legacy command form.** Someone copies an old
   example containing `gemini -p "…" --yolo` into a SKILL.md. `python3
   scripts/validate-skills.py` fails with a clear message naming the file and
   the offending form, before the change can merge.
2. **A maintainer trusts but verifies the guard.** Running `python3
   scripts/validate-skills.py --selftest` proves the guard actually fails on a
   legacy fixture and passes on the migrated fixture — the guard itself is
   tested, not just asserted.
3. **A new repo plans its first sprint under zsh.** During Phase 1 orientation
   the discovery probe runs cleanly and returns nothing, instead of emitting
   `zsh: no matches found`, so the operator is not misled into thinking the
   tooling is broken.
4. **A future contributor asks "why is it still called Gemini if it runs
   agy?"** The sprint doc and skill prose state the compatibility boundary, so
   the answer is documented rather than tribal knowledge.
5. **The legitimate `gemini` selector keeps working.** `/df-sprint-execute
   gemini` and the `Gemini` planning lane continue to function and validate;
   the guard never flags the lane name, selector, or `GEMINI` filenames.

## Architecture

### Validator guardrail (the durable core)

`scripts/validate-skills.py` already performs per-skill content checks via a
series of `if "<substring>" in content` tests inside `main()`. The migration
guard extends this, with three design constraints:

1. **Ban command forms, not the word.** The bare token `gemini` is legitimate
   and pervasive — it is the lane name, the user-facing selector, the
   `GEMINI-DRAFT`/`GEMINI-CRITIQUE` filenames, and ordinary prose ("the `gemini`
   agent runs through Google's `agy` CLI"). The guard therefore matches only
   *invocation shapes* that are unique to the legacy CLI:
   - `gemini -p` (legacy prompt invocation)
   - `--yolo` (legacy permission flag; unique to old gemini, safe to ban
     outright)
   - `which gemini` and `gemini --version` (legacy preflight)
   - `gemini exec` (defensive: a plausible future legacy form)

   These are expressed as a small list of compiled regexes so the intent is
   explicit and extensible. Each pattern carries a human-readable label used in
   the error message.

2. **Testable in isolation.** The per-file checks are refactored out of
   `main()` into a pure function, e.g.
   `check_skill_content(label: str, content: str) -> list[str]`, that takes a
   file label and raw content and returns error strings. `main()` calls it for
   each real SKILL.md; the self-test calls it with synthetic fixtures. No
   behavior change for the existing checks — they move verbatim into the
   function.

3. **Self-proving.** A `--selftest` CLI flag (argparse-free; a simple
   `sys.argv` check keeps the script dependency-light) runs the guard against
   two in-file fixtures:
   - a **legacy fixture** containing `agy ... ` *plus* a planted `gemini -p "x"
     --yolo` line → must produce at least one error;
   - a **migrated fixture** containing only `agy -p "x"
     --dangerously-skip-permissions` and legitimate `gemini` selector prose →
     must produce zero errors.

   `--selftest` prints a pass/fail line and exits non-zero on any mismatch, so
   it can be wired into the SOP gate sequence. Default invocation (no flag)
   preserves today's behavior exactly.

`validate-skills.py` lives at top-level `scripts/` (not inside a skill bundle),
so it is **not** subject to the mirror-parity rule and does not need a `.claude`
/ `.agents` twin. Its self-test fixtures are embedded in the script as string
constants, so no fixture files leak into the skill trees.

### Shell-portable discovery

Three discovery probes use an unquoted glob that zsh's default `nomatch` option
turns into a hard error when no sprint docs exist (confirmed:
`zsh: no matches found: docs/sprints/SPRINT-*.md`). Because zsh aborts the
command *before* `ls` runs, the `2>/dev/null` on `ls` does not suppress the
message. The replacement uses `find`, which globs internally and is unaffected
by shell `nomatch`:

```bash
find docs/sprints -maxdepth 1 -name 'SPRINT-*.md' 2>/dev/null | sort | tail -N
```

This returns nothing (exit 0) when the directory is empty or absent, matches the
original `tail -N` semantics after an explicit `sort` (mirroring `ls`'s implicit
lexical order), and is portable across bash and zsh. The three sites:

| File | Line (approx) | Current | `tail` |
|---|---|---|---|
| `df-sprint-plan/SKILL.md` (Phase 1, find existing sprints) | 108 | `ls docs/sprints/SPRINT-*.md 2>/dev/null \| tail -3` | `-3` |
| `df-sprint-plan/SKILL.md` (Phase 2, next sprint number) | 272 | `ls docs/sprints/SPRINT-*.md 2>/dev/null \| tail -1` | `-1` |
| `df-chapter-create/SKILL.md` (recent sprints) | 62 | `ls docs/sprints/SPRINT-*.md 2>/dev/null \| tail -5` | `-5` |

Each is edited in both `.claude` and `.agents` mirrors (6 edits total). The
adjacent prose ("If no sprint directory exists yet, note this is the first
sprint…") already handles the empty case, so only the command line changes.

### Compatibility boundary

The deliberate split is: **lane identity is stable; the binary changed.** The
sprint document records this explicitly, and a one-line note is added to the
skills' relevant sections (and `README.md`'s `df-sprint-plan`/`df-sprint-execute`
descriptions, which already mention "using `agy` for the Google/Gemini lane").
The canonical statement:

> The planning lane is named **Gemini**, draft/critique artifacts use the
> **`GEMINI`** infix, and the `df-sprint-execute` selector is **`gemini`**.
> These identifiers are stable and are **not** renamed. Only the underlying CLI
> binary changed: the Gemini lane now shells out to Google's **`agy`** CLI
> (`agy -p … --dangerously-skip-permissions`), replacing the legacy `gemini -p …
> --yolo` form.

No identifier rename is in scope; renaming the selector or filenames would be a
larger, separately-justified compatibility break (see Open Questions).

## Implementation Plan

### Phase 1 — Validator regression guard + self-test

Goal: the validator fails on legacy Google CLI command forms and proves it.

Files:
- `scripts/validate-skills.py`

Tasks:
1. Add a module-level list of `(compiled_regex, label)` pairs for legacy
   command forms: `r"gemini\s+-p\b"`, `r"--yolo\b"`, `r"\bwhich\s+gemini\b"`,
   `r"gemini\s+--version\b"`, `r"gemini\s+exec\b"`.
2. Refactor the existing per-file content checks (lines ~53–71 of `main()`)
   into a pure function `check_skill_content(label, content) -> list[str]`,
   moving the current checks verbatim and appending the new legacy-form scan.
   Each legacy match yields `f"{label}: legacy Google CLI command form '<label>'
   — use the agy CLI (agy -p … --dangerously-skip-permissions)"`.
3. Update `main()` to call `check_skill_content(str(skill_md), content)` and
   extend `errors`, preserving the name/dir, frontmatter, and mirror-parity
   logic unchanged.
4. Add a `selftest()` function with two embedded fixtures (legacy → expects ≥1
   error; migrated → expects 0 errors) and wire a `--selftest` argv branch in
   `__main__` that runs it, prints a summary, and exits 0/1 accordingly.
5. Verify: `python3 scripts/validate-skills.py` still prints "Skill naming
   validation passed" on the current tree; `python3 scripts/validate-skills.py
   --selftest` passes.

### Phase 2 — Shell-portable first-sprint discovery

Goal: no `zsh: no matches found` noise on empty `SPRINT-*.md` globs.

Files:
- `.claude/skills/df-sprint-plan/SKILL.md`
- `.agents/skills/df-sprint-plan/SKILL.md`
- `.claude/skills/df-chapter-create/SKILL.md`
- `.agents/skills/df-chapter-create/SKILL.md`

Tasks:
1. In `df-sprint-plan` Phase 1 (≈line 108), replace the `ls … | tail -3` probe
   with `find docs/sprints -maxdepth 1 -name 'SPRINT-*.md' 2>/dev/null | sort |
   tail -3`. Apply in both mirrors.
2. In `df-sprint-plan` Phase 2 (≈line 272), replace the `… | tail -1` probe with
   the `find … | sort | tail -1` form. Apply in both mirrors.
3. In `df-chapter-create` (≈line 62), replace the `… | tail -5` probe with the
   `find … | sort | tail -5` form. Apply in both mirrors.
4. Confirm under zsh that each new probe exits 0 and emits nothing when
   `docs/sprints/` has no matching files.

### Phase 3 — Document the compatibility boundary

Goal: the Gemini-lane/agy-CLI split is written down, not implied.

Files:
- `docs/sprints/SPRINT-001.md` (the final merged sprint doc — carries the
  canonical boundary statement; created at merge time)
- `.claude/skills/df-sprint-plan/SKILL.md` and `.agents/…` (one-line note near
  the Phase 0 `agy` preflight)
- `.claude/skills/df-sprint-execute/SKILL.md` and `.agents/…` (one-line note near
  the agent-selector parsing where `gemini` maps to `agy`)
- `README.md` (confirm the existing "using `agy` for the Google/Gemini lane"
  phrasing states the boundary; tighten if needed)

Tasks:
1. Add a single, mirror-safe sentence to each sprint skill stating that the
   `Gemini` lane name / `GEMINI` filenames / `gemini` selector are stable and
   that `agy` is only the CLI binary. Keep wording identical across mirrors
   except the mechanical `.claude`/`.agents` and `CLAUDE.md`/`AGENTS.md` swaps
   (none are needed in this sentence, so it is byte-identical).
2. Verify `README.md` already conveys the boundary; add a clause only if absent.
3. Ensure the final `SPRINT-001.md` contains the canonical boundary statement
   verbatim (Success Criteria requirement).

### Phase 4 — Full SOP gates + real-CLI proof

Goal: every gate green, migration proven against the installed CLI.

Tasks:
1. `git add -N README.md .claude/skills .agents/skills scripts 2>/dev/null || true`
2. `python3 scripts/validate-skills.py`
3. `python3 scripts/validate-skills.py --selftest`
4. `python3 -m py_compile scripts/validate-skills.py $(find .claude/skills .agents/skills -path '*/scripts/*.py' -type f | sort)`
5. `git diff --check --cached` and `git diff --check`
6. Real CLI preflight: `which agy && agy --version` and
   `agy -p "say ok" --dangerously-skip-permissions`.
7. Negative proof: temporarily plant `gemini -p "x" --yolo` in a scratch copy
   (or rely on `--selftest`) to confirm the guard fires, then discard.
8. Run the Claude critique gate from `AGENTS.md`; fix any actionable findings;
   repeat Phases 1–4 until clean.

## Files Summary

| File | Change | Mirror? |
|---|---|---|
| `scripts/validate-skills.py` | Add legacy-form regex guard; refactor checks into `check_skill_content`; add `--selftest` with embedded fixtures | No (top-level) |
| `.claude/skills/df-sprint-plan/SKILL.md` | `find`-based discovery probes (×2); compatibility-boundary note | Yes |
| `.agents/skills/df-sprint-plan/SKILL.md` | Same as above (mirror) | Yes |
| `.claude/skills/df-chapter-create/SKILL.md` | `find`-based discovery probe (×1) | Yes |
| `.agents/skills/df-chapter-create/SKILL.md` | Same (mirror) | Yes |
| `.claude/skills/df-sprint-execute/SKILL.md` | Compatibility-boundary note near selector parsing | Yes |
| `.agents/skills/df-sprint-execute/SKILL.md` | Same (mirror) | Yes |
| `README.md` | Confirm/tighten boundary phrasing | No |
| `docs/sprints/SPRINT-001.md` | Final sprint doc with canonical boundary statement | No |

## Definition of Done

- [ ] `python3 scripts/validate-skills.py` fails with a clear, file-attributed
      message if any skill mirror contains `gemini -p`, `--yolo`, `which
      gemini`, `gemini --version`, or `gemini exec`.
- [ ] The guard does **not** flag the `Gemini` lane name, the `gemini` selector,
      `GEMINI-DRAFT`/`GEMINI-CRITIQUE` filenames, or boundary prose.
- [ ] `python3 scripts/validate-skills.py --selftest` passes (legacy fixture
      errors; migrated fixture is clean) and exits non-zero on any mismatch.
- [ ] `python3 scripts/validate-skills.py` (no flag) prints "Skill naming
      validation passed" on the current tree.
- [ ] All three first-sprint discovery probes use the `find … | sort | tail -N`
      form in both mirrors and emit no `zsh: no matches found` when
      `docs/sprints/` is empty (verified under zsh).
- [ ] Sprint-plan and sprint-execute examples consistently use
      `agy -p … --dangerously-skip-permissions` for the Gemini lane; no legacy
      form remains anywhere in `.claude`, `.agents`, or `README.md`.
- [ ] `SPRINT-001.md` states the `Gemini` lane / `agy` CLI compatibility
      boundary verbatim.
- [ ] Mirror parity holds: validator's `.claude`/`.agents` body-diff check
      passes; no cross-mirror path leaks.
- [ ] Full SOP gates pass: `py_compile`, `git diff --check` (cached + working),
      validator, validator `--selftest`, and the Claude critique gate reports no
      actionable findings.
- [ ] Real CLI proof recorded: `agy --version` and
      `agy -p "say ok" --dangerously-skip-permissions` succeed.

## Risks & Mitigations

- **Over-broad guard flags legitimate `gemini` usage.** A regex like `\bgemini\b`
  would break the lane name, selector, and filenames. *Mitigation:* match only
  command-form shapes (`gemini -p`, `which gemini`, etc.) and `--yolo`; validate
  against the current clean tree as a regression check (must still pass).
- **Refactor changes existing validator behavior.** Moving checks into a
  function could subtly alter error output. *Mitigation:* move the checks
  verbatim, keep error strings byte-identical, and confirm the current tree
  still prints exactly "Skill naming validation passed".
- **`find` semantics differ from `ls`.** `ls` sorts lexically by default; raw
  `find` order is filesystem-dependent. *Mitigation:* pipe through explicit
  `sort` before `tail` to preserve "most recent N by name" semantics.
- **Self-test fixtures drift from real guard logic.** Fixtures embedded as
  strings could rot. *Mitigation:* the self-test calls the *same*
  `check_skill_content` function the real run uses, so there is one code path,
  not two.
- **Mirror divergence during multi-file edits.** Six SKILL.md edits across two
  mirrors risk asymmetry. *Mitigation:* edit each pair together and rely on the
  validator's body-diff parity check to catch any drift before gates pass.
- **`--selftest` not run in practice.** A flag nobody invokes adds no safety.
  *Mitigation:* add it to the SOP gate sequence (Phase 4) and the DoD so it is
  part of the standard check, not optional.

## Dependencies

- **`agy` CLI** installed and authenticated (`/Users/jmccarthy/.local/bin/agy`,
  v1.0.10) — confirmed during intent preflight. Required for Phase 4 real-CLI
  proof.
- **`claude` CLI** — required for the Claude critique gate in the Change SOP.
- **Python 3** — runs `validate-skills.py` and the `py_compile` gate.
- **zsh** — used to verify the shell-portability fix (the platform default
  shell here).
- **`df-sprint-plan` ↔ `df-sprint-execute` coupling** — unchanged; no ledger or
  helper-path edits in this sprint, so the existing dependency
  (`scripts/ledger.py`) is untouched.
- No semantic index / token cache (Not configured) and no chapter link — no
  retrieval or chapter dependencies.

## Open Questions

1. **`agy` selector alias?** Should `/df-sprint-execute agy` become an accepted
   alias for `/df-sprint-execute gemini`? *Recommendation: out of scope.* Keep
   the selector `gemini` only this sprint; an alias is a user-facing surface
   change that belongs in its own sprint, and the compatibility boundary
   (lane name stable) argues against it.
2. **Ban scope — forms vs. binary.** Should the validator ban every `gemini`
   command binary, or only legacy forms? *Recommendation: legacy command forms
   only* (`gemini -p`, `--yolo`, `which gemini`, `gemini --version`,
   `gemini exec`). Banning the bare word would break the lane name, selector,
   and `GEMINI` filenames the intent explicitly preserves.
3. **Codex auth example drift.** `df-sprint-plan` uses
   `codex exec --sandbox workspace-write` while `df-sprint-execute` uses
   `codex exec --full-auto`. Standardize now, or leave as separate cleanup?
   *Recommendation: leave as a separate cleanup* — it is orthogonal to the agy
   migration and would widen this sprint's blast radius; flag it for a follow-up.
   (Optionally fold in only if the interview prioritizes a single consistent
   Codex form.)
4. **zsh fix scope.** Fix the `no matches found` probes in this sprint?
   *Recommendation: yes* — it was surfaced by first-sprint orientation, the fix
   is a one-line-per-site `find` swap, and leaving it would mislead the very
   next operator. Included as Phase 2.
5. **Self-test delivery — flag vs. separate file.** `--selftest` on the existing
   script keeps the `py_compile` gate and mirror rules untouched (recommended).
   A separate `scripts/test_validate_skills.py` would need adding to the
   `py_compile` glob and a CI hook. *Recommendation: `--selftest` flag.*
6. **Committed `__pycache__` dirs.** `*/scripts/__pycache__/*.pyc` files are
   tracked in both mirrors. Out of scope here, but worth a follow-up `.gitignore`
   cleanup so they don't drift or trip future parity checks.
