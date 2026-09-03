# Workflow package inventory (R5)

## Purpose
Name every seam the `workflow/library/` migration and `tractor workflow show`
touch, so chapter 4 sprint docs can be written against real file:line anchors.
Written by a read-only research branch; transcribed by the planner.

## Pinned
HEAD `96be12f2d7d547bfb8ec154724081bcbf70473ab` (worktree `goal-gates`).

## Inventory

### 1. Prompt strings in Go
- `workflow/workflow.go:250-299` `plannerPrompt(params)` — one `fmt.Sprintf`, 8 verbs.
  Interpolates `Project`, `Plan.Seed`, `Workdir`, `Executable`, derived `projectDir`
  (all through `strconv.Quote`), then `questionCommand`, and `Project` twice raw
  (lines 293-294, the `Next:` contract).
- `workflow/workflow.go:252` builds `questionCommand` as `shellQuote(Executable) + " ask <question-file>"`.
- `workflow/workflow.go:301-324` `mediumPrompt(params, projectDir)` — derives `briefPath`,
  `checklistPath`, `interviewDir`; 6 verbs, all `strconv.Quote`d.
- `workflow/workflow.go:305` `questionCommand` = `TRACTOR_INTERVIEW_DIR=<shellQuote dir> <shellQuote exe> ask <question-file>`.
- `workflow/workflow.go:326-349` `largePlanPrompt` — same derived paths, same quoting (line 330).
- `workflow/workflow.go:351-374` `largeImplementPrompt` — same derived paths, same quoting (line 355).
- `workflow/workflow.go:376-383` `validatorCommand` — not a prompt but the same quoting
  discipline; emits `<exe> workflow validate-plan --project <p> --workdir <w>`.
- `workflow/workflow.go:385-388` `shellQuote` — POSIX single-quote escaping, the only
  quoting primitive; `strconv.Quote` (Go double-quote form) is used for prompt display only.
- `cmd/tractor/*.go` holds no agent prompts. Only cobra help text
  (`cmd/tractor/workflow.go:19-33`, `36-52`, `59-79`, `262-285`) and the MCP server
  instructions string `cmd/tractor/mcp.go:139`.

### 2. The `//go:embed` of the three YAMLs, and what `Build` mutates
- `workflow/workflow.go:23-30` — `plan.yaml`, `medium.yaml`, `large.yaml` into three
  `[]byte` package vars (not an `embed.FS`; tests reassign these vars, see §3).
- `workflow/workflow.go:79-96` `Build` dispatches after `validateParameters`.
- `buildPlan` `workflow/workflow.go:150-179`: sets `planner.Prompt.Value` (176) and
  `validate.ToolCommand` (177). Nothing else.
- `buildMedium` `workflow/workflow.go:181-207`: sets `sprints.Checklist.Value` (204,
  absolute `<projectDir>/checklist.md`) and `implement.Prompt.Value` (205).
- `buildLarge` `workflow/workflow.go:98-124`: sets `chapters.Checklist.Value` (120),
  `plan.Prompt.Value` (121), `implement.Prompt.Value` (122). The nested `sprints` loop is
  fetched (111) but deliberately left without a checklist — it resolves from the chapter item.
- YAML placeholder prompts that Build always overwrites: `workflow/plan.yaml:7`,
  `workflow/medium.yaml:15`, `workflow/large.yaml:15` and `:32`. `plan.yaml:17`
  `tool_command: "false"` is likewise always overwritten.

### 3. Tests that assert on prompt text or the embedded YAML
- `workflow/workflow_test.go:19-96` `TestBuiltInPlan` — 18 substring assertions on
  `planner.Prompt.Value` (51-63); shell-quoting assertions on the validator command (72-76);
  graph shape/fidelity/thread/timeout (35-49); `lint.ValidateOrError` (77); and
  reassigns `planDefinition` to garbage (90-95) to prove parse failure surfaces.
- `workflow/workflow_test.go:98-165` `TestBuiltInMedium` — list ordering (100-107), loop
  checklist path (124), 13 prompt substrings (139-148) including `TRACTOR_INTERVIEW_DIR=`,
  lint (149), `mediumDefinition` swap (159-164).
- `workflow/workflow_test.go:167-254` `TestBuiltInLarge` — 4-node shape (177), 14 plan-prompt
  substrings (198-208), nested `sprints.Checklist` must be absent (211-213), 14
  implement-prompt substrings (226-236), lint (238), `largeDefinition` swap (248-253).
- `workflow/workflow_test.go:256-417` `TestLargeRunsNestedPlanningChecklist` and
  `418-516` `TestMediumRunsPlanningChecklist` — run the built graph through `engine`
  end to end; assert checklist bodies survive byte-for-byte (`workflow_test.go:305`).
- `workflow/workflow_test.go:517-563` `TestPlanArtifacts`, `564-712` `TestPlanExecutionShapes`,
  `713-801` `TestRecommendation` — cover `validate-plan` (see §4).
- `cmd/tractor/workflow_test.go:16-64` `TestWorkflowList` — exact stdout for `workflow list`
  (21-26) and help-text substrings for `workflow`, `workflow list`, `workflow run <name>`.
- `cmd/tractor/workflow_test.go:66-128` `TestWorkflowRun` — asserts the materialized planner
  prompt contains the absolute seed, workdir, and `ephemeral/projects/demo` (96-100), and that
  the validator node command contains `workflow validate-plan` (101-104).
- `cmd/tractor/workflow_test.go:300-364` `TestWorkflowHandoff` — exact stdout, and calls
  `workflow validate-plan` as a CLI subcommand (327).
- `cmd/tractor/root_test.go:25-38` — root help must list `workflow` and must not leak `completion`.
- `examples/skill_bundle_test.go:13-37` — byte-equality tripwire between `examples/loops/*.yaml`
  and `skills/tractor/examples/`. `examples/examples_test.go:12-46` lints every example graph.
  Neither reads `workflow/*.yaml` today.

### 4. `validate-plan`
- Library: `workflow/artifacts.go:29-67` `ValidatePlanArtifacts` — `brief.md` non-empty (34),
  `recommendation.md` non-empty + parsed + field-validated (38-48), `checklist.md` non-empty and
  loadable (50-57), no engine-owned `done` on any item (58-62), then size-shape (63).
- Shape rules `workflow/artifacts.go:69-135`: SIMPLE ≤1 flat item, MEDIUM >1 flat item,
  LARGE >1 chapter each with a unique non-empty `doc` and a unique empty sprint ledger.
- Path safety `workflow/artifacts.go:137-179` — relative only, no `..`, must resolve under the
  project root both before and after `EvalSymlinks`.
- Contract parse `workflow/artifacts.go:203-246`; `Next` must equal the size's exact string
  `workflow/artifacts.go:248-264`.
- Hidden CLI subcommand: `cmd/tractor/workflow.go:262-285` (`Hidden: true`, `--project`, `--workdir`).
  Wired at `cmd/tractor/workflow.go:31`; the plan graph calls it via `workflow/workflow.go:376-383`.

### 5. The `tractor workflow` cobra tree
- Root registration: `cmd/tractor/root.go:36`.
- Parent: `cmd/tractor/workflow.go:18-33`; children added at line 31 — this is where `show` hangs.
- `list`: `cmd/tractor/workflow.go:35-52`. `run`: `54-79`, flags `--project/--seed/--workdir/--logs`
  at `74-77`. `validate-plan`: `262-285`.
- `Parameters` is built in exactly one place: `cmd/tractor/workflow.go:124-127`, fed by
  `os.Executable()`+`filepath.Abs` (108-115), `absoluteDirectory(workdir)` (92), and
  `readableSeed` (101, 195-219).
- XDG state root: `cmd/tractor/mcp_run.go:70-83` `tractorStateRoot()`; used for logs at
  `cmd/tractor/workflow.go:180-192`.

### 6. Existing embedding / generation conventions to match
- `graph/jsonschema_gen.go:14-19` — the repo's only `embed.FS` (directory embed, generated).
- `go:generate` steps: `graph/graph.go:13-14`, `harness/codex/schema/generate.go:5-9`,
  `harness/agy/schema/generate.go:5`.
- No `text/template` or `html/template` import exists anywhere in the Go tree today — the
  library introduces the first one.
- Content-drift tripwire pattern to copy: `examples/skill_bundle_test.go:13-37`.

### 7. Doc surfaces to update
- `docs/spec.md:543-661` §3.1.2 "Built-In Workflows" (prompt behavior described in prose at
  580-587, artifacts 589-604, medium 612-626, large 628-651).
- `src/content/docs/planning.md:1-72`; frontmatter `sourceUrl` pins spec §3.1.2 at line 6;
  run block 15-33, handoff 49-72.
- `skills/tractor/SKILL.md:3` (description) and `38-75` ("Plan before choosing a shape").
- `llms.txt:15-55` ("Plan before choosing a workflow").
- `README.md:118-150` ("Start with a plan").

## Gotchas
- The three `*Definition` vars are reassigned by tests (`workflow_test.go:90-95, 159-164,
  248-253`). Converting to a single `embed.FS` breaks all three unless the invalid-definition
  path gets a new seam.
- `strconv.Quote` and `shellQuote` do different jobs. `strconv.Quote` renders the display of a
  path inside prose; `shellQuote` builds an executable word. A template that pipes both
  through one filter silently changes `'/tmp/tractor'\''s binary'`, which
  `workflow_test.go:72` asserts literally.
- Templates default to `text/template`'s no-escaping, but `{{` inside prompt prose (the checklist
  YAML sample at `workflow.go:269-279` is safe, `$goal` is not) must be audited: the engine also
  does its own `$goal` substitution (`engine/supervisor.go:589`).
- `plan.yaml:17` ships `tool_command: "false"` so an unmutated graph fails closed. Keep that.
- `lint.ValidateOrError` runs on every built graph in tests; an empty template render produces an
  empty prompt, and `PromptValue` (`graph/graph.go:165-170`) then silently substitutes the node
  label rather than erroring. An orphan/empty-render test must be explicit.
- `workflow list` stdout is asserted exactly (`cmd/tractor/workflow_test.go:21-26`); adding a
  `show` subcommand must not alter `list`'s output or the root help set (`root_test.go:33`).
- The engine composes the final prompt as frame + `$goal`-expanded prompt
  (`engine/codergen.go:42-45`, written at `:58`); frames inline the item, the last failure,
  and the doc file (`engine/frames.go:108-168`). `show` cannot reproduce those.

## Recipe
- To move the planner prompt: start at `workflow/workflow.go:250`; its only caller is
  `workflow/workflow.go:176`; its assertions are `workflow/workflow_test.go:51-63` and
  `cmd/tractor/workflow_test.go:96-100`.
- Medium implement prompt: `workflow/workflow.go:301`, caller `:205`, assertions
  `workflow_test.go:139-148`.
- Large chapter-planning prompt: `workflow/workflow.go:326`, caller `:121`, assertions
  `workflow_test.go:198-208`.
- Large sprint prompt: `workflow/workflow.go:351`, caller `:122`, assertions
  `workflow_test.go:226-236`.
- Validator command: `workflow/workflow.go:376`, caller `:177`, assertions
  `workflow_test.go:72-76`.
- The YAML embeds: `workflow/workflow.go:23-30`; the mutation points are `:120-122`, `:176-177`,
  `:204-205`.
- To add `show`: register at `cmd/tractor/workflow.go:31`; reuse the `Parameters` construction at
  `cmd/tractor/workflow.go:108-127` and `tractorStateRoot` at `cmd/tractor/mcp_run.go:70`.
