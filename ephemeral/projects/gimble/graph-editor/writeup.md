# Graph editor: what landed, what did not, who decided

## What was asked

Tyler put a collaborative workflow editor at the front of the
workflow-designer initiative: a page where either the person or the agent
edits the pipeline, including nodes and edges. He supplied the Claude Design
export `Workflow Graph Editor.dc.html`, asked for SvelteKit with no SSR and
every page prerendered, built into the Go binary, and asked that the work be
done in a worktree by subagents.

## What landed

`tractor edit <pipeline.yaml>` starts a loopback-only server, prints the URL,
opens the browser, and serves the editor page from the binary.

- Go: `internal/editor` serves `GET /api/doc` (yaml, layout sidecar, version
  hash, lint diagnostics, parse error), `PUT /api/doc` (atomic write, 409 on a
  stale version), `GET /api/events` (server-sent `change` events from a
  500 ms poll), and the embedded SvelteKit bundle with an SPA fallback and
  immutable cache headers. Lint is the same validator `tractor validate`
  uses. Tests cover the API, the SSE path, the fallback, and the CLI.
- Page: `web/editor`, SvelteKit 2 with adapter-static, `ssr = false`,
  `prerender = true`, output written to `internal/editor/dist` and committed
  so `go install` needs no Node. The page ports the design: header with
  jump-to and add-node selects, zoom, fit, dark mode; dotted canvas with
  pan, zoom, drag, resize, curved edges with condition labels, edge handles,
  start arrow, minimap; inspector with per-type field sections, transitions
  or outgoing edges, supervises, branches, schema issues. Edits go into the
  parsed YAML document at the changed path, so comments and key order
  survive. Node positions live in `<stem>.layout.json` beside the pipeline.
- Two-writer behavior: the page reloads on disk changes, keeps edits when a
  save fails, drops them and shows a banner when disk wins, and goes
  read-only on a YAML syntax error.
- Docs: `src/content/docs/editor.md` and a paragraph in the tractor skill.
- Proof: light and dark screenshots under `proof/`; `go test ./...`,
  `golangci-lint`, `svelte-check`, and the SvelteKit build pass.

## What did not land

- The design MCP import failed for lack of authorization in this session;
  the export was taken from the zip in Downloads instead.
- No automated browser test. The proofs in `brief.md` items 2 to 4 were run
  by hand with curl and sed; item 5 was checked by screenshot, not scripted.
- No agent-side integration beyond the file: the agent edits the YAML with
  ordinary tools and the page follows. No chat panel, no `tractor ask`
  surface in the page.
- The workflow-designer pipeline itself (branch `workflow-designer-v1`) is
  untouched and unmerged.

## Decisions taken by agents, open to veto

- The YAML file is the only shared state; lint is the referee; no locks.
- Positions in a sidecar, never in the YAML, named by stem:
  `bake-off.yaml` gets `bake-off.layout.json`.
- The browser owns the YAML document (the `yaml` npm package); Go only reads,
  writes, lints, and watches.
- `web/editor` is its own pnpm workspace root so the docs site's vite-plus
  and rsvelte overrides do not apply to it. The SvelteKit config lives in
  `vite.config.ts` rather than a separate `svelte.config.js`.
- The built bundle is committed, with a constant Kit version string so
  rebuilding unchanged source yields byte-identical files.
- `.json` pipelines are refused by `tractor edit` because the page always
  writes YAML.
- The design's field names predate the schema rename in #41; the page uses
  the real names (`agent`, `fan_out`, `fan_in`, `command`, `edges.success`,
  `edges.loop`, `branch_edges`) and lets a supervisor watch any node but
  itself, as the spec says, rather than the three types the design allowed.

## Process note

The reading phase of this session cost far more than it should have, and the
first attempt to import the design stalled on authorization. After the
restart the work was delegated: one agent researched SvelteKit-in-Go
embedding (`research.md`), one built the Go side, one built the page. Their
final reports were partly lost when the session process exited; this file is
reconstructed from the commits and the docs they wrote.
