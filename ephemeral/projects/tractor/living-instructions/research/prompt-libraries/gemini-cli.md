# google-gemini/gemini-cli

## Purpose
TypeScript agent CLI whose system prompt is composed from code snippets, but which ships the two mechanisms we care about: an env var that writes the materialized system prompt to a file, and snapshot tests of the rendered prompt. Looked at for the show-command and render-test shapes.

## Pinned
- Repo: https://github.com/google-gemini/gemini-cli
- Commit: `55b495d6db1794bf5b7f37a9bc03ebcab5103673` (main)
- License: Apache-2.0

## Key concepts
- Layout: `packages/core/src/prompts/{promptProvider,snippets,snippets.legacy,utils}.ts`; facade `packages/core/src/core/prompts.ts` L23-35. Prompt bodies are TypeScript template literals inside `render*` functions (`snippets.ts` L183 `renderPreamble` … L598 `renderPlanningWorkflow`); `getCoreSystemPrompt` L136 concatenates them. Builtin skills are files: `packages/core/src/skills/builtin/<name>/SKILL.md`.
- Section toggles by name: `isSectionEnabled(key)` reads `GEMINI_PROMPT_<KEY>` — https://raw.githubusercontent.com/google-gemini/gemini-cli/55b495d6db1794bf5b7f37a9bc03ebcab5103673/packages/core/src/prompts/utils.ts L109-113; used via `withSection` in `promptProvider.ts` L311-317.
- Full override from a file: `GEMINI_SYSTEM_MD` (`promptProvider.ts` L52-54, L110-129) reads `.gemini/system.md` or a path, errors if missing, then `applySubstitutions` fills `${AgentSkills}`, `${SubAgents}`, `${AvailableTools}`, `${<tool>_ToolName}` (`utils.ts` L64). Documented at `docs/cli/system-prompt.md`.
- Materialized dump: `GEMINI_WRITE_SYSTEM_MD` writes the exact string the provider is about to return (`maybeWriteSystemMd`, `promptProvider.ts` L286-291 and L321-335), so the file is byte-equal by construction. Docs call this "Export the default prompt (recommended)".
- Render tests: `packages/core/src/core/prompts.test.ts` uses `expect(prompt).toMatchSnapshot()` for each configuration (L172, L192, L230, L266-375) with `GEMINI_SYSTEM_MD` stubbed off (L84); snapshots in `packages/core/src/core/__snapshots__/prompts.test.ts.snap`.
- Skills: `SKILL.md` with YAML frontmatter, fallback simple parser (`skills/skillLoader.ts` L38-64, L115-127, L164-184).

## Bounded comparison
Like ours in wanting a byte-equal export of the materialized prompt, but only as an env-var side effect inside the live call, and unlike ours because prompt bodies are code strings.

## Gotchas
- `GEMINI_WRITE_SYSTEM_MD` runs the real startup (needs config, tool registry); it is not a standalone subcommand.
- Snapshot tests pin the whole rendered text; every wording change touches the `.snap` file, which is the cost of the render-all approach.
- Two parallel snippet sets (`snippets.ts`, `snippets.legacy.ts`) chosen by model — a fork our library must avoid or make explicit.

## Recipe
- To write "dump the materialized prompt from the same code path", start at `promptProvider.ts` L321-335 and L286-291.
- To write a per-configuration render snapshot test, start at `packages/core/src/core/prompts.test.ts` L144-192.
