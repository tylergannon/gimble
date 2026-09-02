# Sprint 001 Intent: Agy Migration Hardening

## Seed

Harden the Diffusion skills repo after migrating the Google/Gemini lane from the legacy `gemini` CLI examples to the new Google default CLI, `agy`. The sprint should make the migration durable, testable, and easy to verify across both skill mirrors.

## Context

- This is the first planned sprint in the skills repo; `docs/sprints/` did not exist before this planning run, and the sprint ledger reported zero sprints.
- The repo owns four mirrored `df-*` filesystem skills under `.claude/skills/` and `.agents/skills/`. `AGENTS.md` makes mirror parity the critical invariant: edit both mirrors together, with only `.claude/skills/...` vs `.agents/skills/...` and `CLAUDE.md` vs `AGENTS.md` substitutions.
- The current work migrated runnable Google/Gemini CLI examples to `agy -p ... --dangerously-skip-permissions` while preserving the `Gemini` planning lane name, `GEMINI` draft/critique filenames, and the user-facing `gemini` selector in `df-sprint-execute`.
- Required gates are `python3 scripts/validate-skills.py`, `python3 -m py_compile scripts/validate-skills.py $(find .claude/skills .agents/skills -path '*/scripts/*.py' -type f | sort)`, `git diff --check --cached`, `git diff --check`, and the Claude critique gate from `AGENTS.md`.
- Real CLI preflight succeeded during this test: `which agy && agy --version` returned `/Users/jmccarthy/.local/bin/agy` and `1.0.10`; `agy -p "say ok" --dangerously-skip-permissions` returned `ok`.
- First-sprint orientation also exposed a shell-portability wrinkle: under zsh, the documented `ls docs/sprints/SPRINT-*.md 2>/dev/null | tail -3` probe emits `zsh: no matches found` when no sprint docs exist. The sprint should decide whether to harden that example.

## Pyramid Index

- L0: Make the `agy` CLI migration durable by adding regression coverage, tightening examples, and preserving mirror parity.
- L1:
  - Add validator coverage for legacy Google/Gemini command examples such as `gemini -p` and `--yolo`.
  - Review sprint planning and execution docs for shell-portable first-sprint bootstrap commands.
  - Keep the `Gemini` lane semantics stable while documenting that `agy` is the underlying CLI.
  - Preserve the skills repo mirror invariant and run the full SOP gates, including Claude critique.
- L2:
  - See `AGENTS.md` for mirror and gate requirements.
  - See `.agents/skills/df-sprint-plan/SKILL.md` and `.claude/skills/df-sprint-plan/SKILL.md` for preflight, draft, and critique examples.
  - See `.agents/skills/df-sprint-execute/SKILL.md` and `.claude/skills/df-sprint-execute/SKILL.md` for the execution-agent selector and launch table.
  - See `scripts/validate-skills.py` for mirror parity and content guardrails.

## Semantic Index

- **Not configured** — No token cache or semantic index is set up for this project. See `docs/SEMANTIC-INDEX.md`.

Planning agents drafting this sprint should proceed from repo-local files and targeted `rg` searches. There is no index entrypoint to read.

## Chapter Context

No chapter link selected. This repo has no `docs/chapters/` ledger or active chapter context, and the requested work is a narrow skills-repo hardening sprint.

## Recent Sprint Context

First sprint. There are no existing `docs/sprints/SPRINT-*.md` files to summarize.

## Relevant Codebase Areas

- `AGENTS.md` and `README.md` define the repo-level mirror and validation workflow.
- `.agents/skills/df-sprint-plan/SKILL.md` and `.claude/skills/df-sprint-plan/SKILL.md` contain the multi-agent planning preflight and the `agy` draft/critique commands.
- `.agents/skills/df-sprint-execute/SKILL.md` and `.claude/skills/df-sprint-execute/SKILL.md` contain the agent selector and launch examples that map `gemini` to `agy`.
- `scripts/validate-skills.py` enforces skill naming, mirror content normalization, stale path checks, and selected wording regressions.
- `evals/df-semantic-index/evals.json` shows the repo has precedent for skill eval prompts, though this sprint can likely rely on deterministic validator checks rather than a full eval-viewer workflow.

## Constraints

- Must follow project conventions in AGENTS.md.
- Must update `.claude` and `.agents` mirrors together with only the allowed mechanical substitutions.
- Must not rename the `Gemini` planning lane, `GEMINI` artifact filenames, or user-facing `gemini` selector unless the plan explicitly justifies that larger compatibility break.
- Must prove `agy` command examples against the installed CLI, not only by text inspection.
- Must preserve existing helper paths and avoid introducing cross-mirror path leaks.

## Success Criteria

- The validator fails if legacy `gemini -p` or `--yolo` examples are reintroduced in any skill mirror.
- Sprint-plan and sprint-execute skill examples consistently use `agy -p ... --dangerously-skip-permissions` for the Google/Gemini lane.
- First-sprint discovery examples do not produce confusing zsh glob failures when no sprint docs exist, or the sprint documents why that is out of scope.
- Full SOP gates pass, including the Claude critique gate.
- The final sprint document states the compatibility boundary between the `Gemini` lane name and the `agy` CLI.

## Open Questions

- Should `/df-sprint-execute agy` become an accepted alias for `/df-sprint-execute gemini`, or should the user-facing selector remain `gemini` only?
- Should `scripts/validate-skills.py` explicitly ban all `gemini` command binaries, or only legacy command forms such as `gemini -p` and `--yolo`?
- Should this sprint standardize the differing Codex auth examples in `df-sprint-plan` and `df-sprint-execute`, or leave that as a separate cleanup?
- Should the zsh no-match behavior be fixed in this sprint by changing documented examples to `find` or another shell-portable probe?
