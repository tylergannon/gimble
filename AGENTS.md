# Gimble

Gimble is a Go library for writing agent workflows as ordinary Go, with a
web page that shows every run live. It restarted from an empty tree on
2026-09-10. `ephemeral/legacy/` is the old code: inspiration, not the API.

## The rule

As simple as possible. Add a name only when a workflow that exists needs
it. Everything else is ordinary Go written in the workflow.

## Read first

- `go doc -all .`: the current public API and its behavioral contract.
- `ephemeral/research/api/API.md`: the design record and reasons behind the API.
- `ephemeral/research/api/SPRINTS.md`: what is being built, in what order,
  and how each sprint is proven.
- `docs/definition-of-done.md`: how work is gated, validated, and merged.

## No wrappers

Tyler: "WE ARE NOT DOING HIGH LEVEL WRAPPERS OF FUNCTIONALITY THAT OBSCURES
THE MEANING OF THE CODE. There is NO SUCH THING as a `workflows.BakeOff`
function." A workflow reads like a page of pseudocode. A tactic (a
bake-off, a critique round, a worktree, a merge, a retry) is written inline
in the workflow that needs it, with `Group`, `Generate`, and git through
`os/exec`. A new exported name exists only when Tyler asks for it by name.
Propose "write program X that does Y", never "add function Z".

## Build what was asked

Implement only what was asked. A reviewer's objection is not a
requirement. No backwards compatibility, no deprecation paths, no shims:
delete what is replaced. Unit tests for what you are building are fine;
the proof of a workflow is a live run and what it showed.

## Working here

- Commit only when Tyler asks.
- Attestation runs use the cheapest models: Codex `gpt-5.6-luna`, Claude
  Haiku, Gemini flash. Say which model a run used.
- Ports from `ephemeral/legacy/` are rewritten against the new contract by
  hand, never spliced by script.
- No reflection and no `runtime.Caller` to recover a call site. Every node
  is named at its call site with a constant.
- The layout: the API is the root package `gimble`. The runtime owns the
  embedded page and starts it automatically unless `WithNoWeb` is used.
  `web/` is the SvelteKit app; its server assembly and tracked embedded build
  are in `internal/webapp/` and `internal/webembed/`. Go lives beside pages in
  `web/src/routes/*.remote.go`, `internal/generated/` is written by `go
  generate ./...` and never by hand, and `cmd/` contains binaries. `just
  build` builds all of it.
