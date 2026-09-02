# Sprint 001 Critique: Agy Migration Hardening (Claude)

Reviewing `SPRINT-001-CODEX-DRAFT.md` and `SPRINT-001-GEMINI-DRAFT.md` against
`SPRINT-001-INTENT.md`. I grounded the critique by reading
`scripts/validate-skills.py` and the current state of the mirrored skills.

## Shared baseline (both drafts get this right)

- Both correctly preserve the compatibility boundary: keep the `Gemini` lane
  name, `GEMINI-*` artifact filenames, and the `/df-sprint-execute gemini`
  selector; only the runnable command shape moves to `agy`.
- Both reject a blanket ban on the word `gemini` and instead target
  command-shaped legacy forms — the right call given the intent's stable
  compatibility surfaces.
- Both fix the zsh `no matches found` probe by swapping `ls ... glob` for a
  `find ... | sort | tail -N` form, preserving lexicographic ordering for
  zero-padded `SPRINT-NNN.md` names.
- Both keep `/df-sprint-execute agy` aliasing out of scope, matching the
  intent's preferred answer to that open question.

**Verified fact that reframes both drafts:** the production mirrors already
contain *zero* legacy `gemini -p` / `--yolo` / `which gemini` strings (current
`rg` over `.claude/skills .agents/skills README.md AGENTS.md scripts` returns no
matches). The runnable-example migration is effectively done. So Sprint 001 is
not "migrate the commands" — it is "lock the migration with a regression guard,
fix the remaining shell-portability snippets, and verify." Both drafts treat the
text edits as larger than they are; the durable value is concentrated in the
validator guard and the verification gates.

---

## Codex draft

### Architectural soundness — strong
Stays entirely inside the existing architecture: mirrored `SKILL.md` as source,
`scripts/validate-skills.py` as the single deterministic guard, no second
required tool. The instruction to add command-form checks **before** mirror
normalization is correct and matches the validator's actual structure — per-file
content checks already run in the `skill_dir` loop (lines 46–71) ahead of the
normalization block (73–97), so file-scoped error paths will report precisely.
This shows the draft was written against the real code, not from assumption.

### Completeness — strongest of the two
- Catches that `df-chapter-create` (`.claude`/`.agents`, line 62) has the *same*
  `ls docs/sprints/SPRINT-*.md ... | tail -5` zsh probe. Gemini misses this
  file entirely — a genuine completeness gap, since it's the identical bug in a
  third skill pair.
- Adds a validator guard for the `ls docs/sprints/SPRINT-*.md` pattern, not just
  a text fix. Gemini fixes the text but adds **no** guard against its
  reintroduction (see below).
- Covers `which gemini && gemini --version` and bare `--yolo` as distinct
  legacy forms.
- Calls out the verification self-poisoning trap: negative fixtures or `rg`
  scans that match the bad strings will themselves trip the gates. Its mitigation
  (keep bad-command fixtures in unit-test strings, scope production scans to
  specific paths) is the right defensive instinct and something Gemini ignores.

### Phasing / ordering — sound
Audit → harden text → add guards → tests → verify. Putting guards *after* text
hardening is the safe order (guards can't fire on not-yet-fixed text). Given the
verified-clean baseline this ordering is low-risk either way, but it's the
defensible sequence.

### Risk coverage — thorough
Six named risks with mitigations: false positives, scope creep into selector
redesign, ordering semantics of `find`, fixture-poisoned `rg` scans, the
`__pycache__` sandbox-write blocker, and mid-edit mirror drift. The
`PYTHONPYCACHEPREFIX` note is a real environment concern and correctly framed as
a supplemental signal, not a substitute for the SOP gate.

### Feasibility — high, with one scope caution
All steps are runnable today. The one risk is **Phase 4 over-scoping**: adding a
`tests/` tree and refactoring the validator to expose "pure helper functions" is
the largest single piece of new surface in the plan, and the intent itself
hedges ("can likely rely on deterministic validator checks rather than a full
eval workflow"). Codex's own Open Question recommends *deferring* the unit-test
SOP expansion — good — but Phase 4 still front-loads the refactor. For a narrow
hardening sprint this is the part most likely to balloon. Recommend: keep the
guard (Phase 3) mandatory and the unit tests minimal/optional, or fold a couple
of inline assertions rather than a new test tree.

### Definition of Done — rigorous
Best DoD of the two. Proves the guard *positively* (inject a representative
legacy snippet into a fixture → validator exits 1), lists the full gate set
including the mirrored-script `py_compile` form from the intent, `git diff
--check [--cached]`, and treats the Claude critique gate as required. Explicitly
requires the final doc to state the lane/CLI boundary — a named success
criterion in the intent.

### Weaknesses
- Phase 4 scope (above) is the main one.
- Long and somewhat repetitive across Files Summary / DoD / Risks; an
  implementer must dedupe. More planning overhead than the task strictly needs.

---

## Gemini draft

### Architectural soundness — adequate
Same correct core: extend `validate-skills.py`, target command shapes not the
bare word, preserve the selector. Regexes are concrete and reasonable
(`gemini\s+-p`, `\b--yolo\b`). The `\b--yolo\b` form is actually a clean way to
catch bare `--yolo` and `gemini --yolo` in one pattern. It does not mention the
before-normalization ordering point, but its checks would land in the same loop
by default, so this is a documentation omission rather than a design error.

### Completeness — weaker; two real gaps
1. **No guard for the `ls docs/sprints` zsh probe.** Gemini's validator
   (Phase 1) only bans `gemini -p` / `--yolo`. It fixes the `ls` snippets in
   text (Phase 2) but adds nothing to prevent their reintroduction. The intent's
   third success criterion is about not producing zsh glob noise; a guard makes
   that durable, and Gemini leaves it un-protected.
2. **Misses `df-chapter-create` entirely.** Phase 2 touches only
   `df-sprint-plan` and `df-sprint-execute`. The identical `ls ... | tail -5`
   probe in both `df-chapter-create` mirrors (line 62) is left unfixed. This is a
   concrete coverage miss.
3. No `which gemini && gemini --version` handling.

### Distinct strong idea worth keeping — Codex auth standardization
Gemini is right that there's real drift: `df-sprint-execute` uses
`codex exec --full-auto "echo ok"` (line 54) while `df-sprint-plan` uses
`codex exec --sandbox workspace-write "echo ok"` (line 63). I verified both.
Standardizing the *preflight auth check* on the sandboxed form is a sensible,
low-risk cleanup and Gemini folds it into the plan rather than deferring.

**But the nuance Gemini misses:** `--full-auto` also appears in
`df-sprint-execute`'s *execution* command table (line 158), not just the auth
check (line 54). Changing only line 54 leaves the execute skill internally
inconsistent (auth checks under `--sandbox workspace-write`, actual execution
under `--full-auto`). That may be fine — execution legitimately wants broader
permissions than a write-capability probe — but Gemini doesn't acknowledge the
split, so an implementer could either over-reach (also change line 158, altering
Codex execution behavior) or leave a confusing mismatch. Codex's instinct to
*defer* this because it "could change Codex execution behavior" is the more
cautious read. Best synthesis: do the standardization, but scope it explicitly
to the auth/preflight check and state that the execution command's `--full-auto`
is intentionally left alone.

### Phasing / ordering — minor risk
Validator-hardening first (Phase 1), then skill text (Phase 2). In principle
this means running the validator between phases would fail on not-yet-fixed
text; in practice the baseline is already clean of `gemini -p`/`--yolo`, so it's
harmless here. Still the less-defensive order than Codex's text-before-guard.

### Risk coverage — thin
Two risks only (false positives, breaking aliases). Omits the
verification/fixture-poisoning trap, the `py_compile`/`__pycache__` sandbox
concern, mirror drift during multi-file edits, and `find` ordering semantics —
all of which a hardening sprint actually hits.

### Feasibility — high
Smaller, all steps runnable. Lower scope-creep risk than Codex precisely because
it skips the test tree. The leanness is a genuine virtue if the team wants a
tight sprint.

### Definition of Done — lighter, with gaps vs the intent
- Good: explicit positive regression check (inject `gemini -p`/`--yolo` → exit 1).
- Gaps: the gate list omits the mirrored-script `py_compile` form
  (`$(find .claude/skills .agents/skills -path '*/scripts/*.py' ...)`) and
  `git diff --check --cached` from the intent's required gates; it runs only
  `py_compile scripts/validate-skills.py`.
- **Contradicts the intent on the Claude critique gate:** Gemini labels it
  "(Optional/Orchestrator task)" in Phase 3 and omits it from DoD. The intent
  lists the Claude critique gate as *required*. This is a substantive miss.
- DoD #3 claims the `find` commands return "nothing on stdout/stderr" on a fresh
  machine — only true because of `2>/dev/null`; worth stating that the
  suppression is load-bearing (on a repo with no `docs/sprints/`, `find` emits a
  "No such file or directory" to stderr otherwise).

### Other weaknesses
- Uses `file:///Users/jmccarthy/...` absolute links throughout — machine-specific
  and non-portable; should be repo-relative paths.
- Does not require the final doc to state the lane/CLI boundary (a named intent
  success criterion). Codex does.

---

## Recommendation for the merge

Take **Codex as the structural base** (completeness, guard-plus-text coverage,
`df-chapter-create` inclusion, rigorous DoD, before-normalization ordering note)
and graft in from **Gemini**:

1. **The Codex auth standardization** — but scope it explicitly to the
   preflight/auth check in `df-sprint-execute`, and state that execution's
   `--full-auto` (line 158) is intentionally unchanged. This resolves the
   Codex-defer vs Gemini-include disagreement in the safest way.
2. **Gemini's leaner instinct on tests** — make the `ls`-pattern and command-form
   guards in `validate-skills.py` the mandatory deliverable; treat a `tests/`
   tree (Codex Phase 4) as optional/minimal so the sprint doesn't balloon.

Must-fix items the merged plan should not drop:
- Add a validator guard for `ls docs/sprints/SPRINT-*.md` (Codex has it, Gemini
  lacks it).
- Fix the `df-chapter-create` zsh probe in both mirrors (Gemini omits it).
- Keep the Claude critique gate **required**, not optional (Gemini downgrades it).
- Use the full intent gate set in DoD, including the mirrored-script `py_compile`
  form and `git diff --check --cached`.
- Scope production `rg`/negative scans away from draft artifacts and fixtures so
  the verification step doesn't poison itself (Codex's point).
- Require the final sprint doc to state the `Gemini` lane / `agy` CLI boundary.
