# Graph editor: brief

## Goal

`tractor edit <pipeline.yaml>` opens a local web page where a person and an
agent edit the same pipeline. The page renders the YAML as a graph, edits
write straight back to the file, lint runs in-process and shows inline, and
the page reloads when the agent changes the file on disk. Zero engine changes.

Decided by Tyler: the interface exists and is the front of the workflow-designer
initiative; the UI is the Claude Design export at `design/Workflow Graph
Editor.dc.html` in this directory; work happens in this worktree; SvelteKit,
no SSR, everything prerendered, built into the Go binary.

Decided by agent (Claude), open to Tyler's veto: the YAML file is the only
shared state and lint is the referee, no locks; node positions live in a
sidecar `<pipeline>.layout.json` beside the YAML, never in the YAML; the
browser owns the YAML document (the `yaml` npm package's Document API, so
comments and key order survive edits) and Go only reads, writes, lints, and
watches; the web app lives in `web/editor/` as its own pnpm workspace root so
the docs site's vite-plus/rsvelte overrides do not touch it; built output is
committed under `internal/editor/dist` so `go install` works without Node.

## Real schema (the design used older names; use these)

Node `type` values: `agent`, `fan_out`, `fan_in`, `command`, `supervisor`, `loop`.
Terminal targets: `success`, `failure`. Ids match `^[A-Za-z_][A-Za-z0-9_]*$`.

| type | required | fields |
|---|---|---|
| agent | id | label, prompt, edges[{to, condition}], llm_provider, llm_model, reasoning_effort, fidelity, timeout, max_retries, max_visits, thread_id |
| fan_in | id | same as agent |
| fan_out | id, branches | label, prompt, branches[{id, artifacts[], agent{label, prompt, llm_provider, llm_model, reasoning_effort, fidelity, timeout, max_retries, max_visits, thread_id}}], branch_edges[{to, condition}], workspace (isolated, shared), max_parallel, llm_* , fidelity, timeout, max_retries, max_visits, thread_id |
| command | id, command, edges.success | label, command, edges{success, error}, timeout, max_visits |
| supervisor | id, prompt, supervises | label, prompt, supervises[], interval, llm_provider, llm_model, reasoning_effort, timeout |
| loop | id, edges.loop, edges.exit | label, checklist, edges{loop, exit}, max_visits, timeout, llm_provider, llm_model, reasoning_effort (infer judge), evaluator_llm_provider, evaluator_llm_model, evaluator_reasoning_effort |

Top level: `name`, `goal`, `start`, `defaults{llm_provider, llm_model,
reasoning_effort, fidelity, timeout, max_retries}`, `nodes[]`.
Enums: reasoning_effort low, medium, high; fidelity full, compacted, none.
Durations: integer followed by ms, s, m, h, or d. A fan_out has no `edges`;
its outgoing arrows are `branch_edges`. A supervisor may supervise any node
except itself. The authoritative schema: `go run ./cmd/tractor print-schema`.

Design-to-schema mapping: codergen → agent, parallel → fan_out,
parallel.fan_in → fan_in, tool → command (tool_command → command,
on_success/on_error → edges.success/edges.error), loop body/on_done →
edges.loop/edges.exit, branch `codergen` override → `agent`.

## HTTP contract between Go and the page

All JSON. Served on 127.0.0.1 only.

- `GET /api/doc` → `{ "path": string, "yaml": string, "layout": object|null,
  "version": string, "diagnostics": [Diagnostic], "parse_error": string|"" }`.
  `version` identifies the on-disk content (hash of yaml bytes plus layout
  bytes). Diagnostics come from `graph.ParseYAML` then the CLI validator
  (`cliValidator()` in cmd/tractor/root.go); a parse failure yields
  `parse_error` and empty diagnostics.
- `PUT /api/doc` body `{ "yaml": string, "layout": object|null, "version": string }`.
  If `version` is not the current on-disk version, respond 409 with the
  current document (same shape as GET). Otherwise write the YAML atomically
  (temp file + rename), write the sidecar when layout is non-null, and respond
  200 with the new document.
- `GET /api/events` → server-sent events. Emit `event: change` with data
  `{ "version": string }` whenever the YAML or sidecar changes on disk
  (poll every 500 ms; no new dependency). Send a comment line every 15 s to
  keep the connection alive.
- Everything else: serve the embedded site from `internal/editor/dist`; unknown
  paths fall back to `index.html`.

Diagnostic: `{ rule, severity ("error"|"warning"|"info"), message, node_id?, edge? [from, to], fix? }`
exactly as `lint.Diagnostic` marshals today.

CLI: `tractor edit <pipeline.(yaml|yml|json)> [--addr 127.0.0.1:0] [--no-open]`.
Prints the URL on stdout, opens the browser unless `--no-open`, runs until
Ctrl-C. Register in `cmd/tractor/root.go` beside the other commands.

## Page behavior (from the design, adapted)

Header: wordmark, pipeline name, node count, error count badge, jump-to-node
select, add-node select (six real types), zoom controls, Fit, Graph, dark
toggle. Canvas: dotted grid, pan by drag or scroll, zoom by Ctrl/Cmd+scroll,
nodes as cards (type badge, label, id, meta chips showing inherited defaults
in muted color), success/failure pills, curved edges with condition labels,
draggable edge handles on the selected node, start arrow, minimap. Right
aside: graph settings (name, goal, start, defaults) or node inspector (id,
delete, schema issues alert, start switch, field sections, transitions or
outgoing edges, supervises checklist, branches editor). Delete/Backspace
deletes the selected node. Error ring on nodes with diagnostics; hovering the
warning icon shows the messages. Server diagnostics are merged into the same
per-node error display as the client-side checks the design performs.

Nodes without a layout entry get an automatic layered position: columns by
depth from `start` along all routes, rows by order; supervisors go in a row
beneath. Layout changes save with the same debounce as content changes.

Saving: every edit debounces 400 ms and PUTs. On 409 or on an SSE change
while the page has no pending edit, reload the document and re-render,
keeping selection and viewport. On 409 with pending edits, drop the local
edits, reload, and show a one-line banner saying the file changed on disk.

Styling: the design's shadcn-svelte neutral tokens and `.cn-*` classes from
`design/_ds/*/tokens/*.css` and `components.css`, copied into the app.
Geist via Google Fonts with the fallback stack. Dark mode by `.dark` on the
root, remembered in localStorage. Lucide icons.

## Proof

1. `go build ./... && go vet ./... && go test ./...` pass.
2. `tractor edit examples/loops/bake-off.yaml --no-open` prints a URL;
   `curl` of `/api/doc` returns the YAML, no parse error, and lint diagnostics.
3. A `PUT` that changes one node's prompt rewrites the file with every
   comment and key order intact except the changed value, and the response
   carries fresh diagnostics.
4. Editing the file with `sed` while the page is open produces an SSE
   `change` event and the page shows the new value.
5. In a browser: open the page, add a node, wire an edge, see the lint error
   for the unreachable node, fix it, see it clear, reload the page, positions
   persist.
