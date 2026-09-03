# sst/opencode (TypeScript)

## Purpose
TypeScript agent CLI that keeps every system prompt and tool description as a `.txt` file imported as a string at build time (Bun text imports), with per-model prompt selection and a `debug agent` command. Looked at for layout and the show-command shape.

## Pinned
- Repo: https://github.com/sst/opencode
- Commit: `b578b7261fc9ec4917fe272df5cc4bd8a056cd5d` (dev branch)
- License: MIT

## Key concepts
- Layout: system prompts in `packages/opencode/src/session/prompt/*.txt` (`anthropic`, `beast`, `codex`, `default`, `gemini`, `gpt`, `kimi`, `meta`, `trinity`, `plan`, `plan-mode`, `build-switch`); agent prompts in `packages/opencode/src/agent/prompt/*.txt` (`compaction`, `explore`, `summary`, `title`) plus `agent/generate.txt`; each tool's description beside its code as `packages/opencode/src/tool/<name>.txt`; slash-command templates in `packages/opencode/src/command/template/*.txt`.
- Import-as-string is the embed mechanism: https://raw.githubusercontent.com/sst/opencode/b578b7261fc9ec4917fe272df5cc4bd8a056cd5d/packages/opencode/src/session/system.ts L6-14; tool example `packages/opencode/src/tool/read.ts` L7 and L380 (`description: DESCRIPTION`).
- Per-model selection is code, not template logic: `system.ts` L27-49 (`provider(model)` returns one prompt file; `PROMPT_META.replaceAll("{{MODEL_NAME}}", ...)` L30 is the only substitution).
- The `<env>` block, references, skills list and MCP instructions are assembled as string arrays in code, not in the `.txt` files: `system.ts` L67-83, L105-118, L120-136.
- Shared fragments: none inside files; composition is `[providerPrompt, env, skills, mcp].join` in code. Agent prompt is overridable from config (`item.prompt = value.prompt ?? item.prompt`, `packages/opencode/src/agent/agent.ts` L283).
- Show command: `opencode debug agent <name>` prints the agent record (including its `prompt` field) and resolved tools as JSON — `packages/opencode/src/cli/cmd/debug/agent.ts` L4-27, handler `agent.handler.ts` L60-64. It does not print the assembled system prompt. `opencode debug skill` lists discovered skills as JSON (`debug/skill.ts` L6-15).
- Skills are `SKILL.md` discovered by glob (`skill/index.ts` L23-25), formatted into the system prompt (`system.ts` L108-117).

## Bounded comparison
Like ours in one-file-per-prompt with the tool description beside the tool, but only string imports with no template engine, so everything conditional lives in TypeScript.

## Gotchas
- Bun `import x from "./file.txt"` needs a `.d.ts` module declaration and a Bun/bundler that supports text loaders; not portable to Go, but `embed.FS` is the direct equivalent.
- `debug agent` shows configuration, not the materialized prompt; do not cite it as a byte-equal show.

## Recipe
- To copy the "tool description beside the tool" layout, start at `packages/opencode/src/tool/read.ts` L7 and `packages/opencode/src/tool/read.txt`.
- To see a config-level prompt override merged over a builtin, start at `packages/opencode/src/agent/agent.ts` L270-290.
