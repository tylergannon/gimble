# block/goose

## Purpose
Rust agent whose prompts are Jinja (`minijinja`) `.md` files embedded with `include_dir!`, registered in an explicit allowlist, overridable per-user from disk, and unit-tested for registry/file agreement. Closest non-Go analogue to an `embed.FS` + template library with a coverage test.

## Pinned
- Repo: https://github.com/block/goose
- Commit: `794b04a0b1f4c58378ef3738dade297c13690b77` (main)
- License: Apache-2.0

## Key concepts
- Layout: `crates/goose/src/prompts/*.md` (`system`, `subagent_system`, `plan`, `recipe`, `permission_judge`, `session_name`, `apps_create`, `apps_iterate`, `tiny_model_system`); builtin skills `crates/goose/src/skills/builtins/*.md`.
- Embed: `static CORE_PROMPTS_DIR: Dir = include_dir!("$CARGO_MANIFEST_DIR/src/prompts")` — https://raw.githubusercontent.com/block/goose/794b04a0b1f4c58378ef3738dade297c13690b77/crates/goose/src/prompt_template.rs L7; skills the same way in `crates/goose/src/skills/builtin.rs` L3-11.
- Explicit registry `TEMPLATE_REGISTRY: &[(&str, &str)]` of name + description, L9-51; `render_template` refuses unregistered names (L114-120) and `save_template`/`reset_template` do too (L183-212).
- User override by file name: `~/.config/goose/prompts/<name>` wins over the embedded copy (L74-76, L122-131); `template_source` L142-158 returns whichever source is live; `get_template`/`list_templates` L160-243 report `is_customized`.
- Rendering: a fresh `minijinja::Environment` per call with `trim_blocks`, `lstrip_blocks`, one custom filter `code_fence` (L84-112); output is `.trim()`ed. `system.md` uses `{% if %}`/`{% for %}` over a serialized context (L4-35). No `{% include %}`/`{% extends %}` — one template per environment, so shared fragments are not possible in this design.
- Coverage test: `test_list_templates` asserts `templates.len() == TEMPLATE_REGISTRY.len()` and non-empty content for each (L265-283); `test_render_template` renders `system.md` with an empty context (L257-263). Direction is registry→file only; a stray file in `src/prompts/` is not detected.
- Registry drift already visible: `compaction.md` and `compaction_summary.md` are registered but sourced from another crate via fallback (`builtin_content`, L67-72).
- Override sanitization test: `prompt_manager.rs` L284-297 (`test_build_system_prompt_sanitizes_override`).

## Bounded comparison
Like ours in an embedded directory plus a registry that a test checks, but only in the registry→file direction and with no fragment includes and no CLI show.

## Gotchas
- Building a new `Environment` per render means template parse errors surface at render time; the coverage test only renders `system.md`.
- `include_dir!` is a proc-macro crate (MIT); the Go equivalent is `embed.FS`, whose `fs.WalkDir` gives the reverse (file→registry) check for free.

## Recipe
- To write the registry-vs-files test, start at `crates/goose/src/prompt_template.rs` L265-283 and add the inverse walk.
- To offer a per-user file override that shadows an embedded template by name, start at L114-140 and L183-212.
