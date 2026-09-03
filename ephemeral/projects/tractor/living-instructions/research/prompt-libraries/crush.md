# charmbracelet/crush

## Purpose
Go agent CLI (successor of opencode-ai/opencode) whose system prompts and tool descriptions are `.md`/`.md.tpl` files rendered with `text/template`, and whose builtin skills are an `embed.FS`. Closest Go analogue to our library.

## Pinned
- Repo: https://github.com/charmbracelet/crush
- Commit: `e3c970336d7ca889b75dd9bf8c1c4ffd42d65396` (main, 2026-09-02)
- License: FSL-1.1-MIT (Functional Source License, converts to MIT after two years) — https://raw.githubusercontent.com/charmbracelet/crush/e3c970336d7ca889b75dd9bf8c1c4ffd42d65396/LICENSE.md L1-5

## Key concepts
- Layout: agent prompts in `internal/agent/templates/` (`coder.md.tpl`, `task.md.tpl`, `initialize.md.tpl`, `summary.md`, `title.md`, `agentic_fetch*.md`); tool descriptions sit beside their Go file in `internal/agent/tools/*.md[.tpl]` (e.g. `bash.md.tpl`, `edit.md`); builtin skills in `internal/skills/builtin/<name>/SKILL.md`.
- Embedding is per-file `//go:embed`, not a directory `embed.FS`, for prompts: https://raw.githubusercontent.com/charmbracelet/crush/e3c970336d7ca889b75dd9bf8c1c4ffd42d65396/internal/agent/prompts.go L11-18 (`coderPromptTmpl`, `taskPromptTmpl`, `initializePromptTmpl` as `[]byte`). Tool descriptions same pattern: `internal/agent/tools/bash.go` L58-64 (`//go:embed bash.md.tpl` + `template.Must(template.New("bashDescription").Parse(...))`), `edit.go` L47.
- Skills use `embed.FS`: https://raw.githubusercontent.com/charmbracelet/crush/e3c970336d7ca889b75dd9bf8c1c4ffd42d65396/internal/skills/embed.go L14-15 (`//go:embed builtin/*`), walked with `fs.WalkDir` L35-56; virtual path prefix `crush://skills/` L12 so the View tool can read embedded skills.
- Rendering: `prompt.Build` parses the template on every call (`template.New(p.name).Parse(p.template)`) and executes against a `PromptDat` struct — https://raw.githubusercontent.com/charmbracelet/crush/e3c970336d7ca889b75dd9bf8c1c4ffd42d65396/internal/agent/prompt/prompt.go L82-97, data struct L31-43. No `FuncMap`, no `{{template}}`/`{{define}}` includes: shared content is passed as data (`AvailSkillXML` L173-205, `ContextFiles`), not as template fragments.
- Template surface in `coder.md.tpl` is only `{{if}}`/`{{range}}`/field access, lines 372-434 (`<env>`, `<lsp>`, `<skills_usage>`, `<file path=...>` blocks).
- Determinism for tests: `coderAgent` in `internal/agent/common_test.go` L127-135 pins `WithTimeFunc`, `WithPlatform("linux")`, `WithWorkingDir` so VCR cassettes under `internal/agent/testdata/TestCoderAgent/<model>/*.yaml` match; the prompt otherwise contains today's date and live `git status` (`prompt.go` L215, L240-257).
- No render-all test, no orphan-file test, no `show`/`prompt` command: `internal/cmd/root.go` L67-79 registers `run, dirs, projects, update-providers, logs, logout, schema, login, stats, session` only.

## Bounded comparison
Like ours in keeping bodies as embedded `text/template` files, but only per-file `go:embed` with data-passing instead of fragment includes, and with no show command or coverage test.

## Gotchas
- FSL-1.1-MIT is not OSI-open; copying template text verbatim is a license question, copying the *pattern* is not.
- `template.Parse` at `Build` time means a syntax error surfaces at first use, not at startup or in a test; only the VCR-backed agent tests exercise `coder.md.tpl`.
- Date and git state are inside the rendered prompt; byte-equal replay needs the `Option` hooks (`WithTimeFunc`, `WithPlatform`).

## Recipe
- To see how a tool description template is embedded and rendered once at init, start at `internal/agent/tools/bash.go` L58-64 and L148-155.
- To see embedded skills exposed through a virtual path, start at `internal/skills/embed.go` L12-60 and the `<skills_usage>` block in `internal/agent/templates/coder.md.tpl` L390-410.
