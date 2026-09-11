# Gimble

Gimble is a Go library for writing agent workflows as ordinary Go, with a
web page that shows every run live. It restarted from an empty tree on
2026-09-10. `ephemeral/legacy/` is the old code: inspiration, not the API.

## The rule

As simple as possible. Add a name only when a workflow that exists needs
it. Everything else is ordinary Go written in the workflow.

## Read first

- `ephemeral/research/api/API.md`: the API and the reasons behind it.
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

## Information lands locally

Everything an agent needs to finish its task is on the local filesystem
before the agent starts: an issue's text, a spec, a prior result. The
filesystem is the store and the cache for the task, and agents find what
is there. A prompt never points an agent at a remote source to fetch
("read issue 120 with gh"); the workflow fetches it into a file first.

## Build what was asked

Implement only what was asked. A reviewer's objection is not a
requirement. No backwards compatibility, no deprecation paths, no shims:
delete what is replaced. Unit tests for what you are building are fine;
the proof of a workflow is a live run and what it showed.

## Working here

- Commit and push whenever there is work to commit.
- Attestation runs use the cheapest models: Codex `gpt-5.6-luna`, Claude
  Haiku, Gemini flash. Say which model a run used.
- Ports from `ephemeral/legacy/` are rewritten against the new contract by
  hand, never spliced by script.
- No reflection and no `runtime.Caller` to recover a call site. Every node
  is named at its call site with a constant.
- The layout: the API is the root package `gimble`. The page is a
  `tylergannon/skgo` app: `web/` is the SvelteKit app, Go beside its pages
  in `web/src/routes/*.remote.go`, `web/server.go` is the one `NewHandler`
  the binary and the tests share, `generated/` is written by `go generate
  ./...` and never by hand, `cmd/` is the binary. `just build` builds all
  of it. The page's Go imports `gimble`, so `gimble` never imports the
  page; that is why `Serve` is in package `web`.
