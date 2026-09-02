# Agent Instructions — Diffusion Skills Repo

This repo contains Claude Code skills (`/df-sprint-plan`, `/df-sprint-execute`,
`/df-semantic-index`, `/df-chapter-create`, `/df-promise`,
`/df-easy-loop-e2e`, `/df-easy-loop-simple`) distributed as filesystem skill
bundles.

## Critical invariant: always update both mirrors together

Every skill lives in two places:
- `.claude/skills/df-<name>/SKILL.md` — read by Claude Code
- `.agents/skills/df-<name>/SKILL.md` — read by Codex and other agent runtimes

After editing either mirror, apply equivalent changes to the other. The substitution
rules are mechanical:

| .claude version | .agents version |
|---|---|
| `.claude/skills/…` | `.agents/skills/…` |
| `CLAUDE.md` | `AGENTS.md` |
| helper search `find .claude ~/.claude …` | helper search `find .agents ~/.agents ~/.codex …` |
| user scope `~/.claude/skills/…` | user scopes `~/.agents/skills/…` or `~/.codex/skills/…` |

Nothing else changes. Do not paraphrase or restructure — just swap those strings.
All supporting resources, including scripts and references, must also exist in
both mirrors and differ only by the expected mirror-root substitutions. Binary
resources must be byte-identical. A bundled resource must never mention the
opposite mirror's skill path, even in prose or comments.

Run the validator after every skill change:

```bash
python3 scripts/validate-skills.py
```

The validator will catch: name/directory mismatches, skill-body cross-mirror
path leaks, missing SKILL.md files, skill-body drift, missing bundled
resources, cross-mirror helper paths, and resource drift beyond the
expected root and instruction-file swaps.

## Change SOP (required for all non-trivial edits)

1. Edit both mirrors together — `.claude/skills/df-*` and `.agents/skills/df-*`
2. Run local gates:
   ```bash
   git add -N AGENTS.md README.md .claude/skills .agents/skills scripts 2>/dev/null || true
   python3 scripts/validate-skills.py
   python3 -m py_compile scripts/validate-skills.py $(find .claude/skills .agents/skills -path '*/scripts/*.py' -type f | sort)
   git diff --check --cached
   git diff --check
   ```
3. Run the Claude critique gate:
   ```bash
   git add -N AGENTS.md README.md .claude/skills .agents/skills scripts 2>/dev/null || true
   {
     git status --short
     git diff --cached --stat
     git diff --stat
     git diff --cached -- AGENTS.md README.md .claude/skills .agents/skills scripts
     git diff -- AGENTS.md README.md .claude/skills .agents/skills scripts
   } | claude -p "Review this skills repo change for skill formatting, namespacing, mirror consistency, helper paths, and actionable SOP regressions. Do not modify files, the index, or the worktree. Return concise findings with severity and file/line references, or say no actionable findings." --dangerously-skip-permissions
   ```
4. Fix any actionable findings. Repeat steps 1–3 until Claude reports no actionable findings.

## Naming conventions

- All skills use the `df-` prefix: `df-sprint-plan`, `df-semantic-index`, etc.
- Directory name and `name:` frontmatter field must match exactly
- Lowercase ASCII kebab-case only — no colons, no camelCase

## Adding a new skill

1. Create `.claude/skills/df-<name>/SKILL.md` with the required frontmatter
2. Create `.agents/skills/df-<name>/SKILL.md` with the mirror substitutions applied
3. If the skill bundles resources, add them under both skill mirrors
4. Run the full Change SOP above

## Skill dependencies

`df-sprint-execute` depends on `df-sprint-plan`: it invokes the ledger CLI at
`.claude/skills/df-sprint-plan/scripts/ledger.py` (or `.agents/skills/…` in the
agents mirror). Always install both skills together.

## What not to do

- Do not edit one mirror and leave the other stale — the validator will catch it,
  but it creates confusing intermediate states
- Do not reference `.claude/skills/…` paths from inside an `.agents/` skill, or
  vice versa — the validator explicitly checks for this
- Do not add skills without running `validate-skills.py` — the skill-set parity
  check will fail on the next run
