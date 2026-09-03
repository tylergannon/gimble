# Tractor provider and model selection

## Purpose
Where `llm_provider`, `llm_model`, and `reasoning_effort` are accepted, how the
engine resolves them at dispatch, what each backend does with them, and where a
role-to-model table could be applied without touching the engine. Written by a
read-only research branch (R4); transcribed by the planner.

## Pinned
HEAD 96be12f2d7d547bfb8ec154724081bcbf70473ab

## Key concepts
- Six inheritable file-level defaults, including all three model fields:
  graph/graph.go:34.
- `LLMNodeFields` (prompt, max_retries, fidelity, thread_id, timeout, model,
  provider, effort) is embedded by codergen (graph/graph.go:104), fan-in
  (graph/graph.go:121), and parallel (graph/graph.go:222): graph/graph.go:133.
- Per-branch overrides of all of these: `CodergenOverride` graph/graph.go:148,
  applied at graph/parse.go:133.
- Supervisor accepts model/provider/effort but no fidelity or thread_id:
  graph/graph.go:243. Loop accepts the same three, for the infer judge only:
  graph/graph.go:271. Tool nodes accept none: graph/graph.go:169.
- `defaults:` inheritance is per node type: graph/parse.go:196, LLM fields at
  graph/parse.go:218. Supervisor and loop inherit only timeout plus the three
  model fields.
- Lint validates fidelity (lint/rules.go:591) and that nodes sharing a thread
  key resolve to one harness (lint/rules.go:638, resolution at
  lint/analysis.go:358 and lint/analysis.go:382, resolver wired at
  cmd/tractor/root.go:253). No rule validates a model name or effort value;
  effort is constrained to low|medium|high only by the generated schema:
  graph/internal/schemafix/main.go:99.
- Dispatch resolution, node then file then system: engine/codergen.go:125,
  helpers at engine/codergen.go:163 and engine/codergen.go:174. Alias expansion
  and provider-conflict rejection: internal/modelalias/modelalias.go:38.
- Provider inference when nothing is set: `DetectProvider` engine/codergen.go:198
  (claude*→anthropic, gpt-/o1/o3/o4/codex→openai, gemini*→gemini).
- System defaults: model `gpt-5.6-sol`, effort `high` at cmd/tractor/root.go:23,
  wired at cmd/tractor/root.go:184. `DefaultProvider` (engine/runner.go:112,
  engine/codergen.go:21) is never set by the CLI, so provider comes from the
  node, the file defaults, or detection.
- Provider to harness routing: harness/backend.go:16; unroutable provider is a
  terminal error at harness/backend.go:197.
- Backends pass model through unvalidated. Codex: `codex app-server --stdio`
  (harness/codex/rpc.go:320), model in `thread/start` (harness/codex/adapter.go:89)
  and per turn (harness/codex/adapter.go:433). Claude: SDK `WithModel` /
  `WithEffort` (harness/claude/native.go:31, harness/claude/adapter.go:143).
  Agy: binary `agy` (harness/agy/adapter.go:118), `--model` / `--effort`
  (harness/agy/adapter.go:516). Only non-emptiness is checked:
  harness/validate.go:11 and harness/validate.go:29.
- Loop `infer` judge builds a synthetic codergen turn from the loop node's own
  model/provider/effort with fidelity forced to `none`: engine/loop.go:257.
- Supervisor turn resolution mirrors codergen: engine/supervisor.go:373.
- On this machine: codex-cli 0.153.0-alpha.5, claude 2.1.252, agy 1.1.22.
  ~/.codex/config.toml sets `gpt-5.6-sol` with effort `high`;
  ~/.claude/settings.json sets `sonnet`; `agy models` lists gemini-3.8/3.7/3.6
  flash (low|medium|high suffixes), gemini-3.1-pro-high|low, claude-sonnet-4-6,
  claude-opus-4-6-thinking, gpt-oss-120b-medium.

## Gotchas
- Agy model names must carry an explicit `llm_provider: gemini`; detection sends
  `claude-sonnet-4-6` to the claude harness instead.
- Agy suppresses `--effort` when the model name ends in -low/-medium/-high:
  harness/agy/adapter.go:808.
- `fable*` aliases hard-bind provider anthropic; a conflicting `llm_provider` is
  a terminal error: internal/modelalias/modelalias.go:43.
- Embedded workflows set no model fields at all (workflow/plan.yaml,
  workflow/medium.yaml, workflow/large.yaml) and declare no supervisor nodes.
- A loop node's model fields affect only its judge, never its body.
- `claude-sonnet-5` is not a name any backend here is known to accept; the
  reachable claude names are the `sonnet`/`opus` aliases and `claude-fable-5-1`
  / `claude-fable-5` via the alias table.

## Recipe
To assign a model per role in the workflow package, start at
workflow/workflow.go:79 (`Build`); fetch typed nodes with
workflow/workflow.go:126 and workflow/workflow.go:138, then set
`LLMProvider`/`LLMModel`/`ReasoningEffort` as `jsonschema.Optional` values on
each node before returning, exactly as prompts are set at workflow/workflow.go:113.
Always set the provider explicitly; never rely on `DetectProvider`.
