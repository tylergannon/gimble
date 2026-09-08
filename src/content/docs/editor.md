---
title: Graph editor
description: Open a pipeline file in the browser, edit nodes and edges with inline lint, and keep sharing the same YAML with the agent that is writing it.
eyebrow: Operator guide
order: 6
sourceLabel: Browse the editor source
sourceUrl: https://github.com/tylergannon/tractor/tree/main/web/editor
---

A pipeline is a YAML file, and the editor is a view of that file. `tractor
edit` serves a page that draws the graph, lets you change it, and writes the
result straight back to disk. Nothing is stored anywhere else.

## Start it

```sh
tractor edit examples/loops/bake-off.yaml
```

The command listens on loopback only, prints the URL, opens your browser, and
keeps running until you interrupt it. The default is `127.0.0.1:7331`, the
origin the SvelteKit application is built for. Pass `--addr 127.0.0.1:0` to
pick a port explicitly; any address that is not loopback is refused.
`--no-open` prints the URL without opening a browser, which is what you want
over SSH.

Browse the editor at the URL it prints. A save is a cross-site request from
any other origin, and the server refuses it: opening the same port as
`http://localhost:7331/` shows the graph but every save fails with 403.

The pipeline must end in `.yaml` or `.yml`. The page always writes YAML, so a
`.json` pipeline is rejected rather than silently rewritten.

## What the page does

The canvas shows every node and edge. Selecting a node opens an inspector for
its type: `agent`, `fan_out`, `fan_in`, `command`, `supervisor`, and `loop`
each get the fields the schema gives them, with inherited defaults shown
greyed. Selecting nothing shows the graph settings and file-level `defaults`.

Lint is the same validator `tractor validate` runs. Its diagnostics appear on
the node they belong to, and graph-wide problems appear beside the settings,
so what the page accepts is what `start_run` will accept.

Each change is saved about half a second after you make it. Edits go back into
the parsed document at the changed path, so comments, key order, and the
wrapping of prose you did not touch all survive. The header shows when a save
is in flight.

## Sharing the file with an agent

The server watches the file. When anything else writes it (an agent working
in the repository, another editor, `git checkout`), the page reloads and shows
the new graph. An agent can keep editing a pipeline while a person has it open
and both see one file.

If the file changes on disk while the page has edits that have not been saved
yet, the page keeps the on-disk version, drops its own, and says so in a
banner. The disk always wins; the window is short because saves are quick.

A YAML syntax error on disk puts the page in read-only mode. The canvas shows
the error and its line, every field is disabled, and the page returns to
normal on the next on-disk change that parses.

## Where positions live

Node positions are not part of the pipeline. They are stored beside it in
`<stem>.layout.json`, so `bake-off.yaml` gets `bake-off.layout.json`. A
pipeline with no sidecar is laid out automatically: columns by distance from
`start`, supervisors in a row beneath, terminals on the right. The sidecar is
written the first time you move or resize a node; commit it or ignore it.

## How the page talks to the server

The page is a SvelteKit app served by [skgo](https://github.com/tylergannon/skgo):
Go owns the socket, renders the initial page in-process, and answers every
endpoint the page calls. A Go server load reads the pipeline before rendering,
so the first HTML response already contains the graph and hydrates without an
extra fetch. Later changes use three SvelteKit remote functions written in Go
beside the route, in `web/editor/src/routes/editor.remote.go`:

- `getDoc`, a query: the file, its layout sidecar, the lint diagnostics, and a
  version hash of both files.
- `saveDoc`, a command: writes the file, and the sidecar when a node moved,
  provided the version the page quotes is still the one on disk. A stale save
  writes nothing and returns what is on disk, which is how the page learns the
  agent got there first.
- `watchDoc`, a live query: a stream that announces the version on disk now
  and again after every change, which is what makes the page follow the agent.

`skgo generate` writes the `editor.remote.ts` the page imports, the
TypeScript types in `web/editor/src/lib/skgo/editor/types.ts`, and the Go
registration the server mounts. Every body in the generated `.remote.ts`
throws, so a graph on screen is proof that Go answered.

## Rebuilding the page

The page lives in `web/editor`, is built by the skgo adapter into
`web/editor/build`, and is compiled into the binary from there. The build is
committed so `go install` needs no Node. After changing anything under
`web/editor/src`, or any Go type that crosses to the page, run:

```sh
cd web/editor && mise run build
```

which installs the page's dependencies, runs `go generate` for the editor,
builds the page, and rebuilds `tractor`. The bundle uses a constant version
string instead of a build timestamp, so rebuilding unchanged source produces
byte-identical files and the committed build shows no spurious diff.

To work on the page with hot reload, run `mise run dev:web` in one terminal
and `mise run dev:go` in another. The second starts `tractor edit` on port
7331 with a hidden `--proxy` flag that forwards page requests to the vite dev
server; remote functions are still answered by Go, so the page you are
editing talks to the real server.
