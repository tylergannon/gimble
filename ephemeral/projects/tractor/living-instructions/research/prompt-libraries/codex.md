# openai/codex

## Purpose
Rust agent CLI with a dedicated `prompts` crate of `include_str!` markdown templates and a `codex debug prompt-input` subcommand that renders the model-visible input through the same builder the live turn uses. Looked at for the show-command and the template-crate layout.

## Pinned
- Repo: https://github.com/openai/codex
- Commit: `93053c7f5ddc1c26e649e9e7ffc0d9e853c633cf` (main)
- License: Apache-2.0

## Key concepts
- Layout: `codex-rs/prompts/templates/{compact,permissions/approval_policy,permissions/sandbox_mode,realtime,review}/*.md|xml` with one Rust module per group (`codex-rs/prompts/src/{compact,permissions_instructions,realtime,review_exit,review_request}.rs`, re-exported from `lib.rs` L1-20). Base instructions: `codex-rs/protocol/src/prompts/base_instructions/default.md`, `codex-rs/models-manager/prompt.md`; model-specific bodies `codex-rs/core/gpt_5*_prompt.md` and `codex-rs/core/templates/{agents,collab,model_instructions,personalities,review,search_tool}/*.md`; bundled skills `codex-rs/skills/src/assets/samples/<name>/SKILL.md`.
- Embed: `include_str!` per file, e.g. `pub const BASE_INSTRUCTIONS_DEFAULT: &str = include_str!("prompts/base_instructions/default.md")` — https://raw.githubusercontent.com/openai/codex/93053c7f5ddc1c26e649e9e7ffc0d9e853c633cf/codex-rs/protocol/src/models.rs L1503; `prompts/src/compact.rs` L1-2; `permissions_instructions.rs` L21-36; `collaboration-mode-templates/src/lib.rs` L1-2.
- Substitution is hand-rolled: a single placeholder constant `"{{ network_access }}"` (`permissions_instructions.rs` L29) and `"{{ personality }}"` (`models-manager/src/model_info.rs` L22); unknown placeholders are preserved verbatim, tested in `permissions_instructions_tests.rs` L86-122.
- Per-model base instructions also live as JSON strings in `codex-rs/models-manager/models.json` (10 `base_instructions` entries) — a second copy of prompt text outside markdown.
- Show command: `codex debug prompt-input [PROMPT]` — `codex-rs/cli/src/main.rs` L257-258, handler L2211-2306 prints `serde_json::to_string_pretty(&prompt_input)`. It calls `codex_core::build_prompt_input`, which starts a real (ephemeral) thread and reuses `build_prompt` from the turn path (`codex-rs/core/src/prompt_debug.rs` L24-32, L83-109; `session/turn.rs` L1389). Output is the full input list (developer/user items), not one system string.
- Test of the show path: `codex-rs/core/tests/suite/prompt_debug_tests.rs` L19-79 asserts the AGENTS.md text and the user message appear.
- Override: `model_instructions_file` (config, "STRONGLY DISCOURAGED") and `compact_prompt` — `codex-rs/config/src/config_toml.rs` L242-249.

## Bounded comparison
Like ours in printing the materialized prompt from the same builder the engine uses, but only as a JSON item list after a live session bootstrap, and unlike ours because substitution is string replace, not a template engine.

## Gotchas
- `debug prompt-input` needs auth, config and sandbox paths; it is not a pure function of the template set.
- Prompt text is spread over four crates plus `models.json`; no test walks the template directories for orphans.

## Recipe
- To model a show command that reuses the engine's builder, start at `codex-rs/core/src/prompt_debug.rs` L83-109.
- To see placeholder-preservation tests for a substitution step, start at `codex-rs/prompts/src/permissions_instructions_tests.rs` L86-122.
