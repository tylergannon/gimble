# Sprint 001 Codex Critique

## Context

The sprint intent is narrow: make the `agy` migration durable across both skill
mirrors while preserving the stable `Gemini` lane identity, `GEMINI` artifact
filenames, and user-facing `gemini` selector. The plan also needs to cover the
zsh first-sprint glob failure surfaced during orientation, run the full skills
repo SOP gates, and prove the underlying `agy` CLI still works.

## Claude Draft Critique

### Architectural soundness

Claude's draft is architecturally strong. It treats the migration as a durable
repo invariant rather than a one-time text replacement, and puts the main
enforcement in `scripts/validate-skills.py`, which is already the repo's
canonical mirror and content gate. The best architectural choice is banning
legacy command forms rather than the word `gemini`, because the intent explicitly
preserves the `Gemini` lane name, `GEMINI` filenames, and `gemini` selector.

The proposed `check_skill_content(label, content) -> list[str]` extraction is
also a good seam: it keeps mirror validation, frontmatter checks, and content
checks in one script while making the new guard testable without filesystem
fixtures. The `--selftest` mode is a practical way to make the validator guard
self-proving without adding a larger test harness to this small repo.

The `find docs/sprints -maxdepth 1 -name 'SPRINT-*.md' 2>/dev/null | sort |
tail -N` replacement is the right shell-portability fix. It avoids zsh
`nomatch`, handles absent directories, and preserves the old lexical ordering
semantics by adding `sort`.

### Completeness

Claude covers the core sprint surfaces well:

- Validator guard for `gemini -p`, `--yolo`, `which gemini`, `gemini --version`,
  and `gemini exec`.
- Guard self-test with both failing legacy and passing migrated fixtures.
- All known zsh-fragile sprint discovery probes, including the one in
  `df-chapter-create`, not just `df-sprint-plan`.
- Compatibility-boundary documentation in the final sprint doc and skill prose.
- Full SOP gates, Claude critique gate, and real `agy` proof.

This draft is more complete than Gemini's on durable proof. It explicitly adds
`validate-skills.py --selftest` to the gates and DoD, and it records both
positive proof (`agy` works) and negative proof (legacy forms fail).

One completeness gap: the draft says to add one-line compatibility notes to both
sprint skills and possibly `README.md`, but it does not specify exact insertion
points beyond "near" the preflight or selector sections. That is acceptable for
implementation, but the merged plan should be precise enough to avoid duplicated
or awkward prose.

### Phasing and ordering

The ordering is sound:

1. Build the validator guard first so future skill edits can be checked.
2. Fix the shell-portable discovery probes.
3. Document the compatibility boundary.
4. Run the full repo gates and real CLI proof.

This sequence makes the most important invariant testable before touching more
mirrored skill text. It also keeps the zsh fix in its own phase, which is useful
because those edits reach `df-chapter-create` as well as `df-sprint-plan`.

### Risk coverage

Risk coverage is strong. Claude identifies the important failure modes:
false-positive `gemini` detection, behavior drift from refactoring the validator,
`find` ordering differences, self-test drift, mirror divergence, and a self-test
flag that exists but is not routinely run. The mitigations are concrete and map
directly to verification steps.

The best risk item is the explicit warning that a blanket `\bgemini\b` ban would
break valid lane identifiers. That risk is central to this sprint.

### Feasibility

The plan is feasible as a focused sprint. The implementation is small, local,
and compatible with the existing validator style. The only extra complexity is
the `--selftest` refactor, but that complexity buys real durability and remains
inside a single top-level script.

One feasibility caveat: if the validator self-test embeds strings containing
legacy forms, the new guard must only scan skill mirror content during normal
validation, not the validator script itself. Claude's draft implicitly has that
right because `scripts/validate-skills.py` is not itself a SKILL.md file; the
final plan should keep that boundary explicit.

### Definition of done

Claude's DoD is the strongest part of the draft. It is concrete, falsifiable,
and aligned to the repo SOP. It includes:

- Positive clean-tree validator pass.
- Negative regression detection for legacy command forms.
- No false positives for stable `Gemini` identifiers.
- Self-test pass.
- zsh proof for empty sprint discovery.
- Mirror parity.
- Full SOP gates and Claude critique.
- Real `agy` version and prompt proof.

This is the definition of done the merged sprint should mostly keep.

### Strongest ideas to keep

- Use Claude's validator design as the backbone: form-specific legacy command
  regexes, not a blanket `gemini` ban.
- Keep the pure `check_skill_content` refactor and `--selftest` mode.
- Keep the broader legacy-form list, especially `which gemini` and
  `gemini --version`, because stale preflight examples are exactly where this
  migration can regress.
- Keep the `df-chapter-create` zsh fix in scope.
- Keep the explicit compatibility-boundary statement and the full SOP/real-CLI
  proof in the DoD.

### Weaknesses or gaps

- The draft may slightly over-document compatibility boundary edits across skill
  files. The merged plan should avoid adding repetitive prose if the current
  skill text already states the `agy` mapping clearly enough.
- The insertion points for documentation changes should be tightened before
  implementation.
- The draft correctly rejects Codex auth standardization as likely out of scope,
  but it should still preserve it as a follow-up note because Gemini identified
  a real drift point.

## Gemini Draft Critique

### Architectural soundness

Gemini's draft has the right high-level architecture for the core migration:
extend `scripts/validate-skills.py`, detect legacy invocation patterns, avoid
banning the bare word `gemini`, replace zsh-fragile `ls` globs with `find`, and
preserve the user-facing `gemini` selector.

However, it is less sound than Claude's draft in the validator design. It
proposes adding checks "inside the SKILL.md content inspection loop" but does
not separate the content check into a testable unit. That leaves the guard
harder to verify except by mutating real skill files or relying on manual
inspection. It also narrows the pattern list mostly to `gemini -p` and `--yolo`,
missing stale preflight forms such as `which gemini` and `gemini --version` that
the intent specifically cares about.

The draft also brings Codex preflight standardization into the main architecture.
That is a plausible cleanup, but it is orthogonal to the `agy` migration. Since
the intent asks for migration hardening, mirror parity, and shell portability,
standardizing `codex exec --full-auto` to `--sandbox workspace-write` should be
treated as a follow-up unless the human planner explicitly widens the sprint.

### Completeness

Gemini covers the main user-visible work but misses several proof and scope
details:

- It does not include a validator self-test or another durable negative-test
  mechanism.
- It does not mention `which gemini`, `gemini --version`, or `gemini exec`
  regressions.
- It fixes `df-sprint-plan` zsh probes but omits the same fragile probe in
  `df-chapter-create`.
- It lists validator and py_compile gates but makes the Claude critique gate
  optional, even though AGENTS.md makes it required for non-trivial skill edits.
- It does not require real `agy --version` and `agy -p "say ok"` proof in the
  DoD.
- It does not mention `README.md` or final sprint documentation as places to
  record the Gemini-lane/agy-CLI compatibility boundary.

The file summaries also use absolute `file://` links. For repo planning docs,
repo-relative paths are more portable and match the local sprint-doc style.

### Phasing and ordering

The phasing is simple and mostly workable: validator hardening, skill hardening,
then gates. The issue is that Phase 2 combines unrelated changes:

- zsh `find` updates in `df-sprint-plan`;
- Codex preflight standardization in `df-sprint-execute`;
- verification that `agy` examples remain migrated.

That mix makes it harder to reason about what the sprint is proving. The merged
plan should split shell portability, compatibility documentation, and any
optional Codex cleanup into separate decisions. If Codex preflight
standardization is accepted, it should have its own risk and DoD item because it
changes execution-agent behavior beyond the Google/Gemini lane.

### Risk coverage

Gemini identifies two real risks: false positives from over-broad validation and
breaking existing `gemini` selectors. Those are important, but the risk coverage
is thin.

Missing risks include:

- Refactoring or modifying the validator could change existing validation
  behavior.
- A guard without self-test coverage can silently rot.
- Raw `find` output order is nondeterministic unless sorted; Gemini's command
  includes `sort`, but the risk is not called out.
- Mirror divergence across `.claude` and `.agents`.
- `--yolo` may appear only as historical prose in non-skill docs; the validator
  scope must stay limited to SKILL.md content.
- Codex preflight standardization could be a behavioral change, not just a docs
  cleanup.

### Feasibility

The core of Gemini's plan is feasible and smaller than Claude's. It would likely
produce a useful hardening patch quickly.

The feasibility problem is scope discipline. Adding Codex preflight
standardization is easy mechanically, but it changes a second agent lane and can
force additional verification. That makes the sprint broader without improving
the `agy` migration guard. The merged sprint should not include it by default.

Gemini also underspecifies how to prove the regression guard. The DoD says
introducing a bad string should fail, but without a self-test or scratch-copy
procedure, that proof depends on manual mutation during implementation.

### Definition of done

Gemini's DoD is directionally good but incomplete. It includes validator pass,
negative legacy-string proof, no shell errors, and mirror parity. It should be
expanded with:

- No false positives for stable `Gemini` identifiers.
- Self-test or equivalent negative-test proof.
- Full AGENTS.md SOP gates, including Claude critique as required, not optional.
- Real `agy` version and prompt proof.
- Coverage of all known zsh-fragile probes, including `df-chapter-create`.
- Explicit final compatibility-boundary documentation.

### Strongest ideas to keep

- Keep the clear statement that the validator must ban command forms without
  banning the word `gemini`.
- Keep the zsh explanation; it is concise and accurately explains why
  `2>/dev/null` on `ls` cannot suppress shell `nomatch`.
- Keep the proposal to preserve `/df-sprint-execute gemini` and not add an
  `agy` alias in this sprint.
- Keep Codex preflight standardization as a named follow-up or interview
  question, not as default sprint scope.

### Weaknesses or gaps

- Validator coverage is too narrow and lacks a self-test.
- The zsh fix misses `df-chapter-create`.
- The plan widens scope into Codex auth behavior without tying it to the
  migration hardening objective.
- Required SOP gates are under-specified; the Claude critique gate is marked
  optional.
- DoD lacks real `agy` proof and compatibility-boundary documentation.
- Absolute `file://` links make the draft less portable as a repo artifact.

## Merge Recommendation

Use Claude's draft as the primary structure for the final sprint. It best fits
the intent because it turns the migration into a validator-enforced invariant,
adds a self-test, covers all known shell-portability sites, and defines a
falsifiable DoD.

From Gemini, keep the crisp zsh explanation and the explicit recommendation not
to add `/df-sprint-execute agy` as an alias during this sprint. Treat Codex
preflight standardization as an open question or follow-up, not part of the
default merged plan unless the human planner chooses to broaden scope.

The final sprint should require:

- Form-specific legacy command detection in `scripts/validate-skills.py`.
- A self-test or equivalent scratch-proof path for negative validation.
- `find ... | sort | tail -N` replacements in both mirrors for
  `df-sprint-plan` and `df-chapter-create`.
- Clear compatibility-boundary prose: stable `Gemini` identity, `agy` as the
  implementation CLI.
- Full AGENTS.md SOP gates, including Claude critique, plus real `agy` proof.
