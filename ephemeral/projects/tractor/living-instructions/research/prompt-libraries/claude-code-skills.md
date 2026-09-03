# Claude Code skills (SKILL.md format)

## Purpose
The de-facto file format for agent skills (Agent Skills open standard plus Claude Code extensions): a directory with `SKILL.md`, YAML frontmatter, supporting files loaded on demand. Looked at because crush, gemini-cli, goose, codex and sst/opencode all consume this format for their builtin skill directories.

## Pinned
- Source: https://code.claude.com/docs/en/skills (fetched 2026-09-02; no repository or license — vendor documentation). Standard: https://agentskills.io.

## Key concepts
- Unit is a directory: `my-skill/SKILL.md` (required) plus `reference.md`, `examples.md`, `scripts/` ("Add supporting files"). Body loads only when invoked; supporting files are read by the model when `SKILL.md` points at them.
- Frontmatter is honored only if `---` is the first line of the file ("Frontmatter reference"); fields: `name`, `description`, `disable-model-invocation`, `user-invocable`, `allowed-tools`, `disallowed-tools`, `model`, `context: fork`, `agent`, `background`, `arguments`, `paths`, `shell`, `metadata`, `license`, `compatibility`. The portable subset (Agent Skills spec, Skills API) is six: `name, description, license, compatibility, metadata, allowed-tools`.
- Listing cost is bounded: "the combined `description` and `when_to_use` text is truncated at 1,536 characters in the skill listing"; descriptions are in context, bodies are not, unless `disable-model-invocation: true` (then neither) — table under "Control who invokes a skill".
- Substitutions in the body: `$ARGUMENTS`, `$ARGUMENTS[N]`/`$N`, named `$name`, `${CLAUDE_SKILL_DIR}`, `${CLAUDE_PROJECT_DIR}`, plugin `${CLAUDE_PLUGIN_ROOT}`; `!`-prefixed lines run shell and inject output before the model sees the text ("Inject dynamic context"). Argument values containing `$1` are inserted literally, then `${CLAUDE_*}` is expanded.
- Lifecycle: "the rendered `SKILL.md` content enters the conversation as a single message and stays there across later turns"; the file is not re-read.
- Discovery: `~/.claude/skills/`, `.claude/skills/` in cwd and every parent to repo root, `--add-dir` roots, plugins; changes are hot-reloaded. Command name comes from the directory name (plugins: from `name`).

## Bounded comparison
Like ours in "frontmatter + markdown body + sibling reference files", but only as a discovery/loading convention with ad-hoc `$VAR` substitution, never a template engine or a render test.

## Gotchas
- `$` and `!` have meaning in the body; a doctrine page copied into a skill must not contain a stray `$1` or line-leading `!`.
- `name` in frontmatter does not set the command for personal/project skills; the directory does.

## Recipe
- To reuse the frontmatter parser semantics (first-line `---`, unknown-key errors for the spec subset), start at the "Frontmatter reference" and "Using skill frontmatter outside Claude Code" sections of the page.
- To see the same format parsed in Go with `embed.FS`, start at crush `internal/skills/skills.go` L149-211 (`Parse`, `ParseContent`, `splitFrontmatter`).
